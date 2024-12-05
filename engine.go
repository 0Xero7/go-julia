package main

import (
	"context"
	"image"
)

type Engine interface {
	Perform(context context.Context, x, y int32)
	GetExplodesAt(x, y int32) int
	GetMaxExplodesAt() int
	ResetImage()
	IncreaseIteration()
	GetIterations() int
	GetChunkedArea() int
	GetImage() *image.RGBA

	IsStopped() bool
	Stop()
}
