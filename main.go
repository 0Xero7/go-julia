package main

import (
	"fmt"
	"image"
	"image/color"
	"math/cmplx"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
)

var iterations = 1
var value [1024][1024]complex128
var explodesAt [1024][1024]int
var maxExplodesAt = 1

var scale = 128
var offsetX float64 = .75 // -1.04180483110546
var offsetY float64 = 0.1 // 0.346342664848392

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

	x := (float64(pair.x-512) / float64(200.0*scale)) - offsetX
	y := (float64(pair.y-512) / float64(200.0*scale)) - offsetY
	z := value[pair.x][pair.y]
	c := complex(x, y)

	for range 100 {
		z = cmplx.Pow(z, 2) + c
	}

	value[pair.x][pair.y] = z
	if cmplx.Abs(z) > 2 {
		explodesAt[pair.x][pair.y] = iterations
		if explodesAt[pair.x][pair.y] > maxExplodesAt {
			maxExplodesAt = explodesAt[pair.x][pair.y]
		}
	}

	completed <- Message{
		x:        pair.x,
		y:        pair.y,
		explodes: explodesAt[pair.x][pair.y],
	}
}

func spectral_color(l float64) color.RGBA { // RGB <0,1> <- lambda l <400,700> [nm] {
	var t float64 = 0
	var r float64 = 0
	var g float64 = 0
	var b float64 = 0

	if (l >= 400.0) && (l < 410.0) {
		t = (l - 400.0) / (410.0 - 400.0)
		r = +(0.33 * t) - (0.20 * t * t)
	} else if (l >= 410.0) && (l < 475.0) {
		t = (l - 410.0) / (475.0 - 410.0)
		r = 0.14 - (0.13 * t * t)
	} else if (l >= 545.0) && (l < 595.0) {
		t = (l - 545.0) / (595.0 - 545.0)
		r = +(1.98 * t) - (t * t)
	} else if (l >= 595.0) && (l < 650.0) {
		t = (l - 595.0) / (650.0 - 595.0)
		r = 0.98 + (0.06 * t) - (0.40 * t * t)
	} else if (l >= 650.0) && (l < 700.0) {
		t = (l - 650.0) / (700.0 - 650.0)
		r = 0.65 - (0.84 * t) + (0.20 * t * t)
	}

	if (l >= 415.0) && (l < 475.0) {
		t = (l - 415.0) / (475.0 - 415.0)
		g = +(0.80 * t * t)
	} else if (l >= 475.0) && (l < 590.0) {
		t = (l - 475.0) / (590.0 - 475.0)
		g = 0.8 + (0.76 * t) - (0.80 * t * t)
	} else if (l >= 585.0) && (l < 639.0) {
		t = (l - 585.0) / (639.0 - 585.0)
		g = 0.84 - (0.84 * t)
	}

	if (l >= 400.0) && (l < 475.0) {
		t = (l - 400.0) / (475.0 - 400.0)
		b = +(2.20 * t) - (1.50 * t * t)
	} else if (l >= 475.0) && (l < 560.0) {
		t = (l - 475.0) / (560.0 - 475.0)
		b = 0.7 - (t) + (0.30 * t * t)
	}

	return color.RGBA{uint8(255 * r), uint8(255 * g), uint8(255 * b), 255}
}

func main() {
	a := app.New()
	w := a.NewWindow("Images")

	quitCh := make(chan bool)
	workerCompleted := make(chan Message)
	passCompleted := make(chan bool)

	width := 1024
	height := 1024
	image := image.NewRGBA(image.Rect(0, 0, width, height))
	ticker := time.NewTicker(time.Millisecond * 120)

	q := make([]Pair, width*height)
	for x := range width {
		for y := range height {
			q = append(q, Pair{x: x, y: y})

			x1 := (float64(x-512) / float64(200.0*scale)) - offsetX
			y1 := (float64(y-512) / float64(200.0*scale)) - offsetY
			init := complex(x1, y1)
			value[x][y] = init // - complex(-0.10714993602959, -0.91210639328364)
		}
	}

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

	updater := func() {
		for {
			select {
			case <-quitCh:
				return

			case m := <-workerCompleted:
				if explodesAt[m.x][m.y] < 0 {
					image.Set(m.x, m.y, color.Black)
				} else {
					div := float64(300) / float64(maxExplodesAt)
					col := spectral_color(float64(300) + div*float64(explodesAt[m.x][m.y]))
					image.Set(m.x, m.y, col)
				}

				if len(q) == 0 {
					passCompleted <- true
					return
				}

				go DoProcess(q[0], workerCompleted)
				q = q[1:]
			}
		}
	}

	go func() {
		for range 30 {
			fmt.Println("Iteration", iterations, "started")
			q = q[:0]
			q = make([]Pair, width*height)
			index := 0
			for x := range width {
				for y := range height {
					q = append(q, Pair{x: x, y: y})
				}
			}

			go updater()

			for range 100 {
				go DoProcess(q[index], workerCompleted)
				index += 1
			}

			<-passCompleted

			fmt.Println("Iteration", iterations, "completed successfully")
			iterations++
		}
	}()

	w.Resize(fyne.NewSize(1024, 1024))
	w.ShowAndRun()
}
