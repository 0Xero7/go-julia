package main

type Engine interface {
	Perform(x, y int32)
	GetExplodesAt(x, y int32) int
	GetMaxExplodesAt() int
}
