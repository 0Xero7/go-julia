package main

import (
	"context"
	"image"
)

type ComplexEngine struct {
	fz            [][][][]complex128
	fz2           [][][][]complex128
	explodesAt    [][][][]int
	image         *image.RGBA
	maxExplodesAt int

	width, height              int
	scale                      int
	scaleFactorX, scaleFactorY float64

	centerX, centerY float64

	subIterations int

	chunkSizeX, chunkSizeY int

	iterations int

	stopped bool
}

type ComplexEngineParams struct {
	Width, Height          int
	CenterX, CenterY       *float64
	Scale                  *int
	SubIterations          *int
	ChunkSizeX, ChunkSizeY *int
}

func NewComplexEngine(params ComplexEngineParams) *ComplexEngine {
	engine := ComplexEngine{
		width:         params.Width,
		height:        params.Height,
		fz:            Create4D[complex128](params.Height / *params.ChunkSizeY, params.Width / *params.ChunkSizeX, Elvis(params.ChunkSizeX, 1), Elvis(params.ChunkSizeY, 1)),
		fz2:           Create4D[complex128](params.Height / *params.ChunkSizeY, params.Width / *params.ChunkSizeX, Elvis(params.ChunkSizeX, 1), Elvis(params.ChunkSizeY, 1)),
		explodesAt:    Create4D[int](params.Height / *params.ChunkSizeY, params.Width / *params.ChunkSizeX, Elvis(params.ChunkSizeX, 1), Elvis(params.ChunkSizeY, 1)),
		maxExplodesAt: 1,
		scale:         Elvis(params.Scale, 1),
		scaleFactorX:  float64(3) / float64(params.Width*Elvis(params.Scale, 1)),
		scaleFactorY:  (float64(3*params.Height) / float64(params.Width)) / float64(params.Width*Elvis(params.Scale, 1)),
		centerX:       Elvis(params.CenterX, 0.75),
		centerY:       Elvis(params.CenterY, 0),
		subIterations: Elvis(params.SubIterations, 100),
		iterations:    1,
		chunkSizeX:    Elvis(params.ChunkSizeX, 1),
		chunkSizeY:    Elvis(params.ChunkSizeY, 1),
		image:         image.NewRGBA(image.Rect(0, 0, params.Width, params.Height)),
	}

	return &engine
}

func (f *ComplexEngine) Perform(context context.Context, x, y int32) {
	X := x * int32(f.chunkSizeX)
	Y := y * int32(f.chunkSizeY)

	for _x := 0; _x < f.chunkSizeX; _x++ {
		for _y := 0; _y < f.chunkSizeY; _y++ {
			select {
			case <-context.Done():
				return

			default:
				XX := X + int32(_x)
				YY := Y + int32(_y)

				if f.explodesAt[x][y][_x][_y] > 0 {
					continue
				}

				dXX := float64(XX - int32(f.width/2))
				dYY := float64(YY - int32(f.height/2))
				_XX := f.centerX + dXX*f.scaleFactorX
				_YY := f.centerY + dYY*f.scaleFactorY

				for i := range f.subIterations {
					z1 := f.fz2[x][y][_x][_y]

					if real(z1)*real(z1)+imag(z1)*imag(z1) > 4 {
						f.explodesAt[x][y][_x][_y] = f.iterations + i
						if f.explodesAt[x][y][_x][_y] > f.maxExplodesAt {
							f.maxExplodesAt = f.explodesAt[x][y][_x][_y]
						}
						break
					}

					z3 := f.fz[x][y][_x][_y]*f.fz[x][y][_x][_y] + complex(_XX, _YY)

					f.fz[x][y][_x][_y] = z3

					f.fz2[x][y][_x][_y] = complex(real(z3)*real(z3), imag(z3)*imag(z3))
				}
			}
		}
	}
}

func (f *ComplexEngine) GetChunkedArea() int {
	return (f.width / f.chunkSizeX) * (f.height / f.chunkSizeY)
}

func (f *ComplexEngine) GetExplodesAt(x, y int32) int {
	xx := x / int32(f.chunkSizeX)
	xy := x % int32(f.chunkSizeX)
	yx := y / int32(f.chunkSizeY)
	yy := y % int32(f.chunkSizeY)

	ans := f.explodesAt[xx][yx][xy][yy]
	return ans
}

func (f ComplexEngine) GetMaxExplodesAt() int {
	return f.maxExplodesAt
}

func (f *ComplexEngine) IncreaseIteration() {
	f.iterations += f.subIterations
}

func (f *ComplexEngine) GetIterations() int {
	return f.iterations
}

func (f *ComplexEngine) ResetImage() {
	f.image = image.NewRGBA(image.Rect(0, 0, f.width, f.height))
}

func (f *ComplexEngine) GetImage() *image.RGBA {
	return f.image
}

func (f *ComplexEngine) IsStopped() bool {
	return f.stopped
}

func (f *ComplexEngine) Stop() {
	f.stopped = true
}
