package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
)

const width = 1024
const height = 1024

// const centerX float64 = -0.6596510987176695 //-1.8466284581716412 //-0.6596510985176695 //.75 // -1.04180483110546
// const centerY float64 = -0.3362177249890653 //0                   // 0.441169538038593e-06 //-0.3362177249890653 //0.1               // 0.346342664848392
// const scale = 20480 * 800000

// 340ms - non-cached hilbert
// 306ms - linear                  46487ms
//                                 51209ms

func main() {
	a := app.New()
	w := a.NewWindow("Mandelbrot")

	// imageLock := sync.Mutex{}

	quitCh := make(chan bool)
	workerCompleted := make(chan Message)
	passCompleted := make(chan bool)

	image := image.NewRGBA(image.Rect(0, 0, width, height))
	ticker := time.NewTicker(time.Millisecond * 128)
	// 0.07318231460617092 + 0.6137973663828865
	engine := NewFastFloatEngine(FastFloatEngineParams{
		Width:         width,
		Height:        height,
		CenterX:       Ptr(0.07318231460617092),
		CenterY:       Ptr(0.6137973663828865),
		Scale:         Ptr(57868),
		SubIterations: Ptr(4000),
	})

	sampler := UnCachedHilbertCurveSampler{}
	// sampler := LinearSampler{}
	// sampler := *NewHilbertCurveSampler(height, width)

	// color_picker := SpectralColor{}
	// color_picker := NewHistogram("gradient.png")
	color_picker := NewHistogram("Behongo.jpg")
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

	index := atomic.Int32{}
	updater := func() {
		for {
			select {
			case <-quitCh:
				return

			case m := <-workerCompleted:
				explodesAt := engine.GetExplodesAt(Pair{x: m.x, y: m.y})
				if explodesAt == 0 {
					image.Set(int(m.x), int(m.y), color.Black)
				} else {
					fac := math.Log(1+float64(explodesAt)) / math.Log(1+float64(engine.GetMaxExplodesAt()))
					col := color_picker.Get(fac)
					image.Set(int(m.x), int(m.y), col)
				}

				if index.Load() == width*height {
					passCompleted <- true
					return
				}

				_index := index.Load()
				go engine.Perform(sampler.Sample(int(_index), height, width), workerCompleted)
				index.Add(1)
			}
		}
	}

	go func() {
		for iterations := range 30 {
			fmt.Println("Iteration", iterations, "started")
			startTime := time.Now()

			index.Store(0)
			engine.IncreaseIteration()

			go updater()

			for range 1000 {
				_index := index.Load()
				go engine.Perform(sampler.Sample(int(_index), height, width), workerCompleted)
				index.Add(1)
			}

			<-passCompleted

			endTime := time.Now()
			duration := endTime.Sub(startTime).Milliseconds()
			totalTime += int(duration)
			metric := math.Round(float64(1000*totalTime) / float64(engine.GetIterations()))
			w.SetTitle("Mandelbrot: [" + fmt.Sprint(width, "x", height) + "] " + fmt.Sprint(engine.GetIterations()) + " iterations (" + fmt.Sprint(metric) + "ms / 1000 iterations)")

			img := canvas.NewImageFromImage(image)
			w.SetContent(img)

			fmt.Println("Iteration", iterations, "completed successfully in ", duration, " ms")
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
		}

	})
	w.Resize(fyne.NewSize(width, height))
	w.ShowAndRun()
}
