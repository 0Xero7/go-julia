package main

type Sampler interface {
	Sample(i, n, m int) Pair
}
