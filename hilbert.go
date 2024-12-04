package main

type HilbertCurveSampler struct {
	n, m        int
	samplePoint []Pair
}

func NewHilbertCurveSampler(n, m int) *HilbertCurveSampler {
	hc := &HilbertCurveSampler{
		samplePoint: make([]Pair, n*m),
		n:           n,
		m:           m,
	}

	// Initialize Hilbert curve points
	for i := range n * m {
		hc.samplePoint[i] = hc.Sample2(i, n, m)
	}

	return hc
}

func (h *HilbertCurveSampler) Sample(i int) Pair {
	return h.samplePoint[i]
}

// HilbertPoint returns the i-th point in the Hilbert curve
// for a 2D space of size n x Samplem
func (h *HilbertCurveSampler) Sample2(i, n, m int) Pair {
	// Find the smallest power of 2 that covers both dimensions
	size := 1
	for size < m || size < n {
		size *= 2
	}

	// Initialize point
	var x, y int

	// Convert i to x,y coordinates using bit manipulation
	for s := 1; s < size; s *= 2 {
		rx := 1 & (i / 2)
		ry := 1 & (i ^ rx)

		// Rotate quadrant if needed
		if ry == 0 {
			if rx == 1 {
				x = s - 1 - x
				y = s - 1 - y
			}
			// Swap x and y
			x, y = y, x
		}

		x += s * rx
		y += s * ry
		i /= 4
	}

	// Ensure the point is within bounds
	if x >= m {
		x = m - 1
	}
	if y >= n {
		y = n - 1
	}

	return Pair{x: int32(x), y: int32(y)}
}
