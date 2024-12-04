package main

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
