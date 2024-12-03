package main

type Engine interface {
	Perform(pair Pair, message chan Message)
	GetExplodesAt(pair Pair)
}
