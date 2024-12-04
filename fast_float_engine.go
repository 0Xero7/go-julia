package main

import "sync"

type FastFloatEngine struct {
	fzr           [][][][]float64
	fzi           [][][][]float64
	fzr2          [][][][]float64
	fzi2          [][][][]float64
	explodesAt    [][][][]int
	maxExplodesAt int

	width, height              int
	scale                      int
	scaleFactorX, scaleFactorY float64

	centerX, centerY float64

	subIterations int

	chunkSizeX, chunkSizeY int

	iterations int

	lock sync.RWMutex
}

type FastFloatEngineParams struct {
	Width, Height          int
	CenterX, CenterY       *float64
	Scale                  *int
	SubIterations          *int
	ChunkSizeX, ChunkSizeY *int
}

func NewFastFloatEngine(params FastFloatEngineParams) *FastFloatEngine {
	engine := FastFloatEngine{
		fzr:           Create4D[float64](params.Height / *params.ChunkSizeY, params.Width / *params.ChunkSizeX, Elvis(params.ChunkSizeX, 1), Elvis(params.ChunkSizeY, 1)),
		fzi:           Create4D[float64](params.Height / *params.ChunkSizeY, params.Width / *params.ChunkSizeX, Elvis(params.ChunkSizeX, 1), Elvis(params.ChunkSizeY, 1)),
		fzr2:          Create4D[float64](params.Height / *params.ChunkSizeY, params.Width / *params.ChunkSizeX, Elvis(params.ChunkSizeX, 1), Elvis(params.ChunkSizeY, 1)),
		fzi2:          Create4D[float64](params.Height / *params.ChunkSizeY, params.Width / *params.ChunkSizeX, Elvis(params.ChunkSizeX, 1), Elvis(params.ChunkSizeY, 1)),
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
		lock:          sync.RWMutex{},
	}

	return &engine
}

func (f *FastFloatEngine) Perform(x, y int32) {
	X := x * int32(f.chunkSizeX)
	Y := y * int32(f.chunkSizeY)

	for _x := 0; _x < f.chunkSizeX; _x++ {
		for _y := 0; _y < f.chunkSizeY; _y++ {
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
				z1r := f.fzr2[x][y][_x][_y]
				z1i := f.fzi2[x][y][_x][_y]

				if z1r+z1i > 4 {
					f.lock.Lock()
					f.explodesAt[x][y][_x][_y] = f.iterations + i
					if f.explodesAt[x][y][_x][_y] > f.maxExplodesAt {
						f.maxExplodesAt = f.explodesAt[x][y][_x][_y]
					}
					f.lock.Unlock()
					break
				}

				z3i := float64(2)*f.fzr[x][y][_x][_y]*f.fzi[x][y][_x][_y] + _YY
				z3r := f.fzr2[x][y][_x][_y] - f.fzi2[x][y][_x][_y] + _XX

				f.fzr[x][y][_x][_y] = z3r
				f.fzi[x][y][_x][_y] = z3i

				f.fzr2[x][y][_x][_y] = z3r * z3r
				f.fzi2[x][y][_x][_y] = z3i * z3i
			}
		}
	}
}

func (f *FastFloatEngine) GetExplodesAt(x, y int32) int {
	f.lock.RLock()
	xx := x / int32(f.chunkSizeX)
	xy := x % int32(f.chunkSizeX)
	yx := y / int32(f.chunkSizeY)
	yy := y % int32(f.chunkSizeY)

	ans := f.explodesAt[xx][yx][xy][yy]
	f.lock.RUnlock()
	return ans
}

func (f *FastFloatEngine) GetMaxExplodesAt() int {
	return f.maxExplodesAt
}

func (f *FastFloatEngine) IncreaseIteration() {
	f.iterations += f.subIterations
}

func (f *FastFloatEngine) GetIterations() int {
	return f.iterations
}
