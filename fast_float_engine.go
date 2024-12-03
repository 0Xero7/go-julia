package main

type FastFloatEngine struct {
	fzr           [][]float64
	fzi           [][]float64
	fzr2          [][]float64
	fzi2          [][]float64
	explodesAt    [][]int
	maxExplodesAt int

	scale                      int
	scaleFactorX, scaleFactorY float64

	centerX, centerY float64

	subIterations int

	iterations int
}

type FastFloatEngineParams struct {
	Width, Height    int
	CenterX, CenterY *float64
	Scale            *int
	SubIterations    *int
}

func NewFastFloatEngine(params FastFloatEngineParams) *FastFloatEngine {
	engine := FastFloatEngine{
		fzr:           Create2D[float64](height, width),
		fzi:           Create2D[float64](height, width),
		fzr2:          Create2D[float64](height, width),
		fzi2:          Create2D[float64](height, width),
		explodesAt:    Create2D[int](height, width),
		maxExplodesAt: 1,
		scale:         Elvis(params.Scale, 1),
		scaleFactorX:  float64(3) / float64(width*Elvis(params.Scale, 1)),
		scaleFactorY:  (float64(3*height) / float64(width)) / float64(width*Elvis(params.Scale, 1)),
		centerX:       Elvis(params.CenterX, 0.75),
		centerY:       Elvis(params.CenterY, 0),
		subIterations: Elvis(params.SubIterations, 100),
		iterations:    1,
	}

	return &engine
}

func (f *FastFloatEngine) Perform(pair Pair, completed chan Message) {
	if f.explodesAt[pair.x][pair.y] > 0 {
		completed <- Message{
			x:        pair.x,
			y:        pair.y,
			explodes: f.explodesAt[pair.x][pair.y],
		}
		return
	}

	dx := float64(pair.x - (width / 2))
	dy := float64(pair.y - (height / 2))
	x := f.centerX + dx*f.scaleFactorX
	y := f.centerY + dy*f.scaleFactorY

	for i := range f.subIterations {
		z1r := f.fzr2[pair.x][pair.y]
		z1i := f.fzi2[pair.x][pair.y]

		if z1r+z1i > 4 {
			f.explodesAt[pair.x][pair.y] = f.iterations + i
			if f.explodesAt[pair.x][pair.y] > f.maxExplodesAt {
				f.maxExplodesAt = f.explodesAt[pair.x][pair.y]
			}
			break
		}

		z3i := float64(2)*f.fzr[pair.x][pair.y]*f.fzi[pair.x][pair.y] + y
		z3r := f.fzr2[pair.x][pair.y] - f.fzi2[pair.x][pair.y] + x

		f.fzr[pair.x][pair.y] = z3r
		f.fzi[pair.x][pair.y] = z3i

		f.fzr2[pair.x][pair.y] = z3r * z3r
		f.fzi2[pair.x][pair.y] = z3i * z3i

	}

	completed <- Message{
		x:        pair.x,
		y:        pair.y,
		explodes: f.explodesAt[pair.x][pair.y],
	}
}

func (f *FastFloatEngine) GetExplodesAt(pair Pair) int {
	return f.explodesAt[pair.x][pair.y]
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
