package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"log"
	"math"
	"os"
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

	if params.sampler != "linear" && params.sampler != "hilbert" {
		log.Fatalf("Invalid sampler: %s. Supported samplers are linear and hilbert", params.sampler)
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

	// Idk why
	params.centerX -= 1.5 / float64(params.scale)
	params.centerY -= 1.5 / float64(params.scale)

	a := app.New()
	w := a.NewWindow("Mandelbrot")

	var width = params.width
	var height = params.height

	quitCh := make(chan bool)

	image := image.NewRGBA(image.Rect(0, 0, width, height))
	ticker := time.NewTicker(time.Millisecond * 128)

	chunkSizeX, chunkSizeY := params.chunkSizeX, params.chunkSizeY
	// S := 1.1
	// COLOR_STEPS := 20

	engine := NewFastFloatEngine(FastFloatEngineParams{
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
	})

	// sampler := UnCachedHilbertCurveSampler{
	// 	n: height / chunkSizeX,
	// 	m: width / chunkSizeY,
	// }

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
	}
	// sampler := NewHilbertCurveSampler(height, width)

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
	// color_picker := SpectralColor{}
	// color_picker := NewHistogram("gradient.png")
	// color_picker := NewHistogram("Behongo.jpg")
	// color_picker := NewHistogram("Grade Grey.jpg")
	// color_picker := NewHistogram("Evening Night.jpg")
	totalTime := 0

	go func() {
		for {
			exited := false
			select {
			case <-ticker.C:
				img := canvas.NewImageFromImage(image)
				w.SetContent(img)

			case <-quitCh:
				exited = true
			}

			if exited {
				break
			}
		}
	}()

	iterate := func(iteration int) {
		fmt.Println("Iteration", iteration, "started")
		engine.IncreaseIteration()
		startTime := time.Now()

		workerPool := pond.NewPool(128)

		for k := range (height / engine.chunkSizeY) * (width / engine.chunkSizeX) {
			P := sampler.Sample(k)
			j, i := P.x, P.y
			workerPool.Submit(func() {
				engine.Perform(int32(j), int32(i))

				X := chunkSizeX * int(j)
				Y := chunkSizeY * int(i)

				for x := 0; x < chunkSizeX; x++ {
					for y := 0; y < chunkSizeY; y++ {
						px := X + x
						py := Y + y

						updateImage(image, px, py, color_converter, color_picker, engine)
					}
				}
			})
		}
		workerPool.StopAndWait()

		for j := range width {
			for i := range height {
				px := j
				py := i

				updateImage(image, px, py, color_converter, color_picker, engine)
			}
		}

		endTime := time.Now()
		duration := endTime.Sub(startTime).Milliseconds()
		totalTime += int(duration)
		metric := math.Round(float64(1000*totalTime) / float64(engine.GetIterations()))
		w.SetTitle("Mandelbrot: [" + fmt.Sprint(width, "x", height) + "] " + fmt.Sprint(engine.GetIterations()) + " iterations (" + fmt.Sprint(metric) + "ms / 1000 iterations)")

		img := canvas.NewImageFromImage(image)
		w.SetContent(img)

		fmt.Println("Iteration", iteration, "completed successfully in ", duration, " ms")
	}

	go func() {
		for iterations := range params.iterations {
			iterate(iterations)
		}

		fmt.Println("All iterations completed in ", totalTime, " ms")
	}()

	w.Canvas().SetOnTypedRune(func(r rune) {
		switch r {
		case 'q':
			quitCh <- true
			return

		case 's':
			im := image

			f, err := os.Create("mandelbrot.png")
			if err != nil {
				log.Println(err)
				return
			}

			if err = png.Encode(f, im); err != nil {
				log.Println(err)
			}
			f.Close()

		case '+':
			iterate(engine.GetIterations() + 1)
		}

	})
	w.Resize(fyne.NewSize(float32(width), float32(height)))
	w.ShowAndRun()
}
