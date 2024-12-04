package main

type Engine interface {
	Perform(x, y int32, message chan Message)
	GetExplodesAt(x, y int32)
}
