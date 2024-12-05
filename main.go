package main

import (
	"context"
	"flag"
	"fmt"
	"image/png"
	"log"
	"math"
	"os"
	"slices"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"github.com/alitto/pond/v2"
)

type cliParams struct {
	width, height          int
	chunkSizeX, chunkSizeY int
	centerX, centerY       float64
	scale                  int
	subiterations          int
	iterations             int
	sampler                string
	colorOf                string
	colorGradientPath      string
}

func verify(params cliParams) {
	if params.width <= 0 || params.height <= 0 {
		log.Fatal("Width and height must be positive integers")
	}

	if (params.width&(params.width-1) != 0) || (params.height&(params.height-1) != 0) {
		log.Fatal("Width and height must be powers of two")
	}

	if params.chunkSizeX <= 0 || params.chunkSizeY <= 0 {
		log.Fatal("Chunk size X and Y must be positive integers")
	}

	if (params.chunkSizeX&(params.chunkSizeX-1) != 0) || (params.chunkSizeY&(params.chunkSizeY-1) != 0) {
		log.Fatal("Width and height must be powers of two")
	}

	if params.scale <= 0 {
		log.Fatal("Scale must be a positive integer")
	}

	if params.subiterations <= 0 || params.iterations <= 0 {
		log.Fatal("Sub-iterations and iterations must be positive integers")
	}

	if !slices.Contains([]string{"linear", "hilbert", "cachedhilbert"}, params.sampler) {
		log.Fatalf("Invalid sampler: %s. Supported samplers are %s", params.sampler, strings.Join([]string{"linear", "hilbert", "cachedhilbert"}, ","))
	}

	if params.colorOf != "spectral" && params.colorOf != "gradient" {
		log.Fatalf("Invalid color pickers: %s. Supported color pickers are spectral and gradient", params.sampler)
	}

	if params.colorGradientPath == "" && params.colorOf == "gradient" {
		log.Fatal("Gradient color picker requires a valid gradient image path")
	}
}

func main() {
	params := cliParams{}
	flag.IntVar(&params.width, "width", 1024, "width of the image")
	flag.IntVar(&params.height, "height", 1024, "height of the image")
	flag.IntVar(&params.chunkSizeX, "chunkX", 256, "chunk size in X direction")
	flag.IntVar(&params.chunkSizeY, "chunkY", 256, "chunk size in Y direction")
	flag.Float64Var(&params.centerX, "x", -0.75, "center offset X")
	flag.Float64Var(&params.centerY, "y", 0, "center offset Y")
	flag.IntVar(&params.scale, "scale", 1, "zoom level")
	flag.IntVar(&params.subiterations, "subit", 200, "sub-iterations per chunk")
	flag.IntVar(&params.iterations, "it", 10, "max iterations")
	flag.StringVar(&params.sampler, "sampler", "linear", "which sampler to use (linear/hilbert)")
	flag.StringVar(&params.colorOf, "color", "spectral", "which color picker to use (spectral/gradient)")
	flag.StringVar(&params.colorGradientPath, "path", ".", "if gradient color picker, the path of the image from which to sample the colors")
	flag.Parse()

	verify(params)

	a := app.New()
	w := a.NewWindow("Mandelbrot")

	var width = params.width
	var height = params.height

	quitCh := make(chan bool)
	// ticker := time.NewTicker(time.Millisecond * 2500)

	chunkSizeX, chunkSizeY := params.chunkSizeX, params.chunkSizeY

	engineParams := DerbailEngineParams{
		Width:   width,
		Height:  height,
		CenterX: &params.centerX, //Ptr(0.07318231460617092),
		CenterY: &params.centerY, //Ptr(0.6137973663828865),
		// CenterX: Ptr(0.3290059999116987),
		// CenterY: Ptr(0.5184159787873756),
		Scale: &params.scale, //Ptr(14586006),
		// Scale:         Ptr(49259936850),
		SubIterations: &params.subiterations, //Ptr(500),
		ChunkSizeX:    &params.chunkSizeX,    //Ptr(chunkSizeX),
		ChunkSizeY:    &params.chunkSizeY,    //Ptr(chunkSizeY),
	}

	var sampler Sampler
	if params.sampler == "linear" {
		sampler = LinearSampler{
			n: height / chunkSizeY,
			m: width / chunkSizeX,
		}
	} else if params.sampler == "hilbert" {
		sampler = UnCachedHilbertCurveSampler{
			n: height / chunkSizeX,
			m: width / chunkSizeY,
		}
	} else if params.sampler == "cachedhilbert" {
		sampler = NewHilbertCurveSampler(height, width)
	}

	color_converter := ExponentialMappedModuloColorRangeConverer{
		S:     1.1,
		Steps: 20,
	}

	var color_picker ColorOf
	if params.colorOf == "spectral" {
		color_picker = SpectralColor{}
	} else if params.colorOf == "gradient" {
		color_picker = NewHistogram(params.colorGradientPath)
	}

	totalTime := 0

	// iterationStoppedChannel := make(chan bool)
	iterationContext, iterationContextCancel := context.WithCancel(context.TODO())

	// engineX := NewFastFloatEngine(engineParams)
	engineX := NewDerbailEngine(engineParams)

	// func() {
	// 	for {
	// 		exited := false
	// 		select {
	// 		case <-ticker.C:
	// 			img := canvas.NewImageFromImage(engineX.GetImage())
	// 			w.SetContent(img)

	// 		case <-quitCh:
	// 			exited = true
	// 		}

	// 		if exited {
	// 			break
	// 		}
	// 	}
	// }()

	iterate := func(engineInstance Engine, iteration int) {
		if engineInstance.IsStopped() {
			return
		}

		fmt.Println("Iteration", iteration, "started")
		engineInstance.IncreaseIteration()
		startTime := time.Now()

		workerPool := pond.NewPool(128, pond.WithContext(iterationContext))

		for k := range engineInstance.GetChunkedArea() {
			P := sampler.Sample(k)
			j, i := P.x, P.y
			workerPool.Submit(func() {
				engineInstance.Perform(iterationContext, int32(j), int32(i))

				if engineInstance.IsStopped() {
					return
				}

				X := chunkSizeX * int(j)
				Y := chunkSizeY * int(i)

				for x := 0; x < chunkSizeX; x++ {
					for y := 0; y < chunkSizeY; y++ {
						px := X + x
						py := Y + y

						if engineInstance.IsStopped() {
							return
						}

						updateImage(engineInstance.GetImage(), px, py, color_converter, color_picker, engineInstance)
					}
				}
			})
		}
		workerPool.StopAndWait()

		for j := range width {
			for i := range height {
				px := j
				py := i

				updateImage(engineInstance.GetImage(), px, py, color_converter, color_picker, engineInstance)
			}
		}

		endTime := time.Now()
		duration := endTime.Sub(startTime).Milliseconds()
		totalTime += int(duration)
		metric := math.Round(float64(1000*totalTime) / float64(engineInstance.GetIterations()))
		w.SetTitle("Mandelbrot: [" + fmt.Sprint(width, "x", height) + "] " + fmt.Sprint(engineInstance.GetIterations()) + " iterations (" + fmt.Sprint(metric) + "ms / 1000 iterations)")

		img := canvas.NewImageFromImage(engineInstance.GetImage())
		w.SetContent(img)

		fmt.Println("Iteration", iteration, "completed successfully in ", duration, " ms")
	}

	iterationLoop := func(engine Engine) {
		for iterations := range params.iterations {
			select {
			case <-iterationContext.Done():
				return

			default:
				iterate(engine, iterations)
			}
		}

		fmt.Println("All iterations completed in ", totalTime, " ms")
	}
	go iterationLoop(engineX)

	resetWith := func(newParams DerbailEngineParams) {
		iterationContextCancel()
		engineX.Stop()

		iterationContext, iterationContextCancel = context.WithCancel(context.TODO())
		engineX = NewDerbailEngine(newParams)
		go iterationLoop(engineX)
	}

	w.Canvas().SetOnTypedKey(func(ke *fyne.KeyEvent) {
		switch ke.Name {
		case fyne.KeyQ:
			quitCh <- true
			return

		case fyne.KeyS:
			im := engineX.GetImage()

			f, err := os.Create("mandelbrot.png")
			if err != nil {
				log.Println(err)
				return
			}

			if err = png.Encode(f, im); err != nil {
				log.Println(err)
			}
			f.Close()

		case fyne.KeyReturn:
			iterate(engineX, engineX.GetIterations()+1)

		case fyne.KeyR:
			resetWith(engineParams)

		case fyne.KeyPlus:
			*engineParams.Scale++
			resetWith(engineParams)

		case fyne.KeyMinus:
			*engineParams.Scale--
			resetWith(engineParams)

		case fyne.KeyLeft:
			delta := 0.01 / float64(*engineParams.Scale)
			engineParams.CenterX = Ptr(*engineParams.CenterX - delta)
			resetWith(engineParams)

		case fyne.KeyRight:
			delta := 0.01 / float64(*engineParams.Scale)
			engineParams.CenterX = Ptr(*engineParams.CenterX + delta)
			resetWith(engineParams)

		case fyne.KeyUp:
			delta := 0.01 / float64(*engineParams.Scale)
			engineParams.CenterY = Ptr(*engineParams.CenterY - delta)
			resetWith(engineParams)

		case fyne.KeyDown:
			delta := 0.01 / float64(*engineParams.Scale)
			engineParams.CenterY = Ptr(*engineParams.CenterY + delta)
			resetWith(engineParams)
		}

	})
	w.Resize(fyne.NewSize(float32(width), float32(height)))
	w.ShowAndRun()
}
