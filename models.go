package main

type Pair struct {
	x int32
	y int32
}

type Message struct {
	chunkCompleted bool
	x              int32
	y              int32
	explodesAt     [][]int
}
