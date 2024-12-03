package main

type LinearSampler struct{}

func (h *LinearSampler) Sample(i, n, m int) Pair {
	x := i % m
	y := i / m
	return Pair{x: int32(x), y: int32(y)}
}
