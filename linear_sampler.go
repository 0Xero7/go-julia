package main

type LinearSampler struct {
	n, m int
}

func (h *LinearSampler) Sample(i int) Pair {
	x := i % h.m
	y := i / h.m
	return Pair{x: int32(x), y: int32(y)}
}
