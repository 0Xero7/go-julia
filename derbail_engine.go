package main

import (
	"context"
	"image"
	"math/cmplx"
)

type DerbailEngine struct {
	zn            [][][][]complex128
	zdashn        [][][][]complex128
	zdashn_sum    [][][][]complex128
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

	bailoutValue float64

	stopped bool
}

type DerbailEngineParams struct {
	Width, Height          int
	CenterX, CenterY       *float64
	Scale                  *int
	SubIterations          *int
	ChunkSizeX, ChunkSizeY *int
	Bailout                *float64
}

func NewDerbailEngine(params DerbailEngineParams) *DerbailEngine {
	engine := DerbailEngine{
		width:         params.Width,
		height:        params.Height,
		zn:            Create4D[complex128](params.Height / *params.ChunkSizeY, params.Width / *params.ChunkSizeX, Elvis(params.ChunkSizeX, 1), Elvis(params.ChunkSizeY, 1)),
		zdashn:        Create4DWithValue[complex128](params.Height / *params.ChunkSizeY, params.Width / *params.ChunkSizeX, Elvis(params.ChunkSizeX, 1), Elvis(params.ChunkSizeY, 1), complex(1, 0)),
		zdashn_sum:    Create4D[complex128](params.Height / *params.ChunkSizeY, params.Width / *params.ChunkSizeX, Elvis(params.ChunkSizeX, 1), Elvis(params.ChunkSizeY, 1)),
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
		bailoutValue:  Elvis(params.Bailout, 1e4),
	}

	return &engine
}

func (f *DerbailEngine) Perform(context context.Context, x, y int32) {
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

				// if f.explodesAt[x][y][_x][_y] > 0 {
				// 	continue
				// }

				dXX := float64(XX - int32(f.width/2))
				dYY := float64(YY - int32(f.height/2))
				_XX := f.centerX + dXX*f.scaleFactorX
				_YY := f.centerY + dYY*f.scaleFactorY

				for i := range f.subIterations {
					new_zdash := (complex(2, 0) * f.zdashn[x][y][_x][_y] * f.zn[x][y][_x][_y]) + complex(1, 0)
					new_zn := cmplx.Pow(f.zn[x][y][_x][_y], 2) + complex(_XX, _YY)
					new_zdashsum := f.zdashn_sum[x][y][_x][_y] + new_zdash

					if real(new_zdashsum)*real(new_zdashsum)+imag(new_zdashsum)*imag(new_zdashsum) > f.bailoutValue {
						f.explodesAt[x][y][_x][_y] = f.iterations + i
						if f.explodesAt[x][y][_x][_y] > f.maxExplodesAt {
							f.maxExplodesAt = f.explodesAt[x][y][_x][_y]
						}
						break
					} else {
						f.explodesAt[x][y][_x][_y] = 0
					}

					f.zdashn[x][y][_x][_y] = new_zdash
					f.zn[x][y][_x][_y] = new_zn
					f.zdashn_sum[x][y][_x][_y] = new_zdashsum
				}
			}
		}
	}
}

func (f *DerbailEngine) GetChunkedArea() int {
	return (f.width / f.chunkSizeX) * (f.height / f.chunkSizeY)
}

func (f *DerbailEngine) GetExplodesAt(x, y int32) int {
	xx := x / int32(f.chunkSizeX)
	xy := x % int32(f.chunkSizeX)
	yx := y / int32(f.chunkSizeY)
	yy := y % int32(f.chunkSizeY)

	ans := f.explodesAt[xx][yx][xy][yy]
	return ans
}

func (f DerbailEngine) GetMaxExplodesAt() int {
	return f.maxExplodesAt
}

func (f *DerbailEngine) IncreaseIteration() {
	f.iterations += f.subIterations
}

func (f *DerbailEngine) GetIterations() int {
	return f.iterations
}

func (f *DerbailEngine) ResetImage() {
	f.image = image.NewRGBA(image.Rect(0, 0, f.width, f.height))
}

func (f *DerbailEngine) GetImage() *image.RGBA {
	return f.image
}

func (f *DerbailEngine) IsStopped() bool {
	return f.stopped
}

func (f *DerbailEngine) Stop() {
	f.stopped = true
}
