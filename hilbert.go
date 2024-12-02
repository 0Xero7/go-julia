package main

// HilbertPoint returns the i-th point in the Hilbert curve
// for a 2D space of size n x m
func HilbertPoint(i, m, n int) Pair {
	// Find the smallest power of 2 that covers both dimensions
	size := 1
	for size < n || size < m {
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
	if x >= n {
		x = n - 1
	}
	if y >= m {
		y = m - 1
	}

	return Pair{x: x, y: y}
}