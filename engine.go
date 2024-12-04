package main

import "context"

type Engine interface {
	Perform(context context.Context, x, y int32)
	GetExplodesAt(x, y int32) int
	GetMaxExplodesAt() int
}
