package main

type Engine interface {
	Perform(pair Pair, message Message)
	GetExplodesAt(pair Pair)
}
