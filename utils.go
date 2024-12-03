package main

func Create2D[T any](n, m int) [][]T {
	result := make([][]T, n)
	for i := range result {
		result[i] = make([]T, m)
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
