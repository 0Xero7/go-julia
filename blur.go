package main

import (
	"image"
	"image/color"
	"math"
)

func GaussianBlur(img *image.RGBA, radius float64) *image.RGBA {
	// Create kernel
	size := int(2*radius + 1)
	kernel := makeGaussianKernel(radius)

	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	// Create output image
	output := image.NewRGBA(bounds)

	// Apply convolution
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var r, g, b, a float64
			var sum float64

			// Apply kernel to neighboring pixels
			for ky := -size / 2; ky <= size/2; ky++ {
				for kx := -size / 2; kx <= size/2; kx++ {
					// Get neighboring pixel coordinates
					ix := x + kx
					iy := y + ky

					// Skip if outside image bounds
					if ix < 0 || ix >= width || iy < 0 || iy >= height {
						continue
					}

					// Get kernel value
					kernelVal := kernel[ky+size/2][kx+size/2]

					// Get pixel color
					pixel := img.RGBAAt(ix, iy)

					// Accumulate weighted values
					r += float64(pixel.R) * kernelVal
					g += float64(pixel.G) * kernelVal
					b += float64(pixel.B) * kernelVal
					a += float64(pixel.A) * kernelVal
					sum += kernelVal
				}
			}

			// Normalize and set output pixel
			if sum != 0 {
				output.Set(x, y, color.RGBA{
					R: uint8(r/sum + 0.5),
					G: uint8(g/sum + 0.5),
					B: uint8(b/sum + 0.5),
					A: uint8(a/sum + 0.5),
				})
			}
		}
	}

	return output
}

// Helper function to create Gaussian kernel
func makeGaussianKernel(radius float64) [][]float64 {
	size := int(2*radius + 1)
	kernel := make([][]float64, size)
	for i := range kernel {
		kernel[i] = make([]float64, size)
	}

	// Calculate kernel values
	sigma := radius / 3.0
	twoSigmaSquare := 2 * sigma * sigma
	for y := -size / 2; y <= size/2; y++ {
		for x := -size / 2; x <= size/2; x++ {
			distance := float64(x*x + y*y)
			kernel[y+size/2][x+size/2] = math.Exp(-distance/twoSigmaSquare) / (math.Pi * twoSigmaSquare)
		}
	}

	return kernel
}
