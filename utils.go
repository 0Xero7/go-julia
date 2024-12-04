package main

import (
	"image"
	"image/color"
)

func Create2D[T any](n, m int) [][]T {
	result := make([][]T, n)
	for i := range result {
		result[i] = make([]T, m)
	}
	return result
}

func Create4D[T any](n, m, a, b int) [][][][]T {
	result := make([][][][]T, n)
	for i := range result {
		result[i] = make([][][]T, m)
	}

	for i := range n {
		for j := range m {
			val := Create2D[T](a, b)
			result[i][j] = val
		}
	}
	return result
}

func Elvis[T any](left *T, right T) T {
	if left == nil {
		return right
	}
	return *left
}

func Ptr[T any](val T) *T {
	return &val
}

func updateImage(img *image.RGBA, px, py int, colorRange ColorRangeConverer, colorPicker ColorOf, engine Engine) {
	explodesAt := engine.GetExplodesAt(int32(px), int32(py))
	if explodesAt == 0 {
		img.Set(int(px), int(py), color.Black)
	} else {
		// fac := math.Log(1+float64(explodesAt)) / math.Log(1+float64(engine.GetMaxExplodesAt()))
		fac := colorRange.Get(float64(explodesAt), float64(engine.GetMaxExplodesAt()))
		col := colorPicker.Get(fac)
		img.Set(int(px), int(py), col)
	}
}
