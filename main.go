package main

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
)

const width = 1024
const height = 1024

const centerX float64 = -0.6596510985176695 //.75 // -1.04180483110546
const centerY float64 = -0.3362177249890653 //0.1               // 0.346342664848392
const scale = 256                           // 128

const scaleFactorX = float64(3) / (width * scale)
const scaleFactorY = float64(3) / (height * scale)

var iterations = 1
var value [width][height]AComplex
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

	// x := (float64(pair.x-(width/2)) / float64(scaleFactorX)) - offsetX
	// y := (float64(pair.y-(height/2)) / float64(scaleFactorY)) - offsetY
	z := value[pair.x][pair.y]
	c := New(x, y)

	for i := range 100 {
		z = Add(Mul(z, z), *c)
		if Gt2(z) {
			explodesAt[pair.x][pair.y] = 100*(iterations-1) + i
			if explodesAt[pair.x][pair.y] > maxExplodesAt {
				maxExplodesAt = explodesAt[pair.x][pair.y]
			}
			break
		}
	}

	value[pair.x][pair.y] = z

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

	q := make([]Pair, width*height)
	for x := range width {
		for y := range height {
			q = append(q, Pair{x: x, y: y})

			dx := float64(x - (width / 2))
			dy := float64(y - (height / 2))

			x1 := centerX + dx*scaleFactorX
			y1 := centerY + dy*scaleFactorY

			init := New(x1, y1)
			value[x][y] = *init // - complex(-0.10714993602959, -0.91210639328364)
		}
	}

	go func() {
		for {
			exited := false
			select {
			case <-ticker.C:
				// imageLock.Lock()
				// bImg := GaussianBlur(image, 2)
				img := canvas.NewImageFromImage(image)
				w.SetContent(img)
				// imageLock.Unlock()

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
				// imageLock.Lock()
				if explodesAt[m.x][m.y] < 0 {
					image.Set(m.x, m.y, color.Black)
				} else {
					div := float64(300) / float64(maxExplodesAt)
					col := spectral_color(float64(300) + div*float64(explodesAt[m.x][m.y]))
					image.Set(m.x, m.y, col)
				}
				// imageLock.Unlock()

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
			// index := 0
			for x := range width {
				for y := range height {
					q = append(q, Pair{x: x, y: y})
				}
			}

			go updater()

			for range 1000 {
				go DoProcess(q[0], workerCompleted)
				q = q[1:]
			}

			<-passCompleted

			fmt.Println("Iteration", iterations, "completed successfully")
			iterations++
		}
	}()

	w.Resize(fyne.NewSize(1024, 1024))
	w.ShowAndRun()
}
