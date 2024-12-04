package main

type Sampler interface {
	Sample(i int) Pair
}
