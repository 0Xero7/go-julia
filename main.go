package main

import (
	"fmt"
	"image"
	"image/color"
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

func main() {
	a := app.New()
	w := a.NewWindow("Mandelbrot")

	const width = 1024
	const height = 1024

	quitCh := make(chan bool)

	image := image.NewRGBA(image.Rect(0, 0, width, height))
	ticker := time.NewTicker(time.Millisecond * 128)

	chunkSizeX, chunkSizeY := 256, 256

	engine := NewFastFloatEngine(FastFloatEngineParams{
		Width:  width,
		Height: height,
		// CenterX: Ptr(0.07318231460617092),
		// CenterY: Ptr(0.6137973663828865),
		CenterX: Ptr(0.3290059999116987),
		CenterY: Ptr(0.5184159787873756),
		Scale:   Ptr(14586006),
		// Scale:         Ptr(49259936850),
		SubIterations: Ptr(80),
		ChunkSizeX:    Ptr(chunkSizeX),
		ChunkSizeY:    Ptr(chunkSizeY),
	})

	// sampler := UnCachedHilbertCurveSampler{
	// 	n: height / chunkSizeX,
	// 	m: width / chunkSizeY,
	// }
	sampler := LinearSampler{
		n: height / chunkSizeY,
		m: width / chunkSizeX,
	}
	// sampler := *NewHilbertCurveSampler(height, width)

	// color_picker := SpectralColor{}
	color_picker := NewHistogram("gradient.png")
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

						explodesAt := engine.GetExplodesAt(int32(px), int32(py))
						if explodesAt == 0 {
							image.Set(int(px), int(py), color.Black)
						} else {
							fac := math.Log(1+float64(explodesAt)) / math.Log(1+float64(engine.GetMaxExplodesAt()))
							col := color_picker.Get(fac)
							image.Set(int(px), int(py), col)
						}
					}
				}
			})
		}
		workerPool.StopAndWait()

		for j := range width {
			for i := range height {
				px := j
				py := i
				explodesAt := engine.GetExplodesAt(int32(px), int32(py))
				if explodesAt == 0 {
					image.Set(int(px), int(py), color.Black)
				} else {
					fac := math.Log(1+float64(explodesAt)) / math.Log(1+float64(engine.GetMaxExplodesAt()))
					col := color_picker.Get(fac)
					image.Set(int(px), int(py), col)
				}
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
		for iterations := range 30 {
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
	w.Resize(fyne.NewSize(width, height))
	w.ShowAndRun()
}
