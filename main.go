package main

import (
	"fmt"
	"image"
	"image/color"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
)

const width = 1024
const height = 1024

const centerX float64 = -0.6596510985176695 //.75 // -1.04180483110546
const centerY float64 = -0.3362177249890653 //0.1               // 0.346342664848392
const scale = 20480

const scaleFactorX = float64(3) / (width * scale)
const scaleFactorY = (float64(3*height) / float64(width)) / (height * scale)

var iterations = 1

const subIterations = 100 * 2

// var value [width][height]AComplex

// var zr [width][height]decimal.Big
// var zi [width][height]decimal.Big
// var zr2 [width][height]decimal.Big
// var zi2 [width][height]decimal.Big

var fzr [width][height]float64
var fzi [width][height]float64
var fzr2 [width][height]float64
var fzi2 [width][height]float64

var explodesAt [width][height]int
var maxExplodesAt = 1

type Pair struct {
	x int
	y int
}

type Message struct {
	x        int
	y        int
	explodes int
}

func DoProcess(pair Pair, completed chan Message) {
	if explodesAt[pair.x][pair.y] > 0 {
		completed <- Message{
			x:        pair.x,
			y:        pair.y,
			explodes: explodesAt[pair.x][pair.y],
		}
		return
	}

	dx := float64(pair.x - (width / 2))
	dy := float64(pair.y - (height / 2))
	x := centerX + dx*scaleFactorX
	y := centerY + dy*scaleFactorY

	for i := range subIterations {
		z1r := fzr2[pair.x][pair.y]
		z1i := fzi2[pair.x][pair.y]

		if z1i+z1r > 4 { // CompareDecimal(AddDecimal(z1r, z1i), Dec(4)) == 1 {
			explodesAt[pair.x][pair.y] = subIterations*(iterations-1) + i
			if explodesAt[pair.x][pair.y] > maxExplodesAt {
				maxExplodesAt = explodesAt[pair.x][pair.y]
			}
			break
		}

		// Calculate next iteration
		// z3i := AddDecimal(MulDecimal(MulDecimal(Dec(2), zr[pair.x][pair.y]), zi[pair.x][pair.y]), Dec(y)) // 2.0*zr*zi + ci
		// z3r := AddDecimal(SubDecimal(zr2[pair.x][pair.y], zi2[pair.x][pair.y]), Dec(x)) // zr2 - zi2 + cr

		z3i := float64(2)*fzr[pair.x][pair.y]*fzi[pair.x][pair.y] + y
		z3r := fzr2[pair.x][pair.y] - fzi2[pair.x][pair.y] + x

		// zr[pair.x][pair.y] = z3r
		// zi[pair.x][pair.y] = z3i

		fzr[pair.x][pair.y] = z3r
		fzi[pair.x][pair.y] = z3i

		// zr2[pair.x][pair.y] = MulDecimal(z3r, z3r)
		// zi2[pair.x][pair.y] = MulDecimal(z3i, z3i)

		fzr2[pair.x][pair.y] = z3r * z3r
		fzi2[pair.x][pair.y] = z3i * z3i
	}

	completed <- Message{
		x:        pair.x,
		y:        pair.y,
		explodes: explodesAt[pair.x][pair.y],
	}
}

func main() {
	a := app.New()
	w := a.NewWindow("Images")

	// imageLock := sync.Mutex{}

	quitCh := make(chan bool)
	workerCompleted := make(chan Message)
	passCompleted := make(chan bool)

	image := image.NewRGBA(image.Rect(0, 0, width, height))
	ticker := time.NewTicker(time.Millisecond * 1020)

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
				if explodesAt[m.x][m.y] == 0 {
					image.Set(m.x, m.y, color.Black)
				} else {
					div := float64(300) / float64(maxExplodesAt)
					col := spectral_color(float64(300) + div*float64(explodesAt[m.x][m.y]))
					image.Set(m.x, m.y, col)
				}

				if index.Load() == width*height {
					passCompleted <- true
					return
				}

				_index := index.Load()
				_x := _index % width
				_y := _index / width
				go DoProcess(Pair{
					x: int(_x),
					y: int(_y),
				}, workerCompleted)
				index.Add(1)
			}
		}
	}

	go func() {
		for range 30 {
			fmt.Println("Iteration", iterations, "started")
			startTime := time.Now()

			index.Store(0)

			go updater()

			for range 100 {
				_index := index.Load()
				_x := _index % width
				_y := _index / width

				go DoProcess(Pair{
					x: int(_x),
					y: int(_y),
				}, workerCompleted)
				index.Add(1)
			}

			<-passCompleted

			endTime := time.Now()
			duration := endTime.Sub(startTime).Milliseconds()
			fmt.Println("Iteration", iterations, "completed successfully in ", duration, " ms")
			iterations++
		}
	}()

	w.Resize(fyne.NewSize(width, height))
	w.ShowAndRun()
}
