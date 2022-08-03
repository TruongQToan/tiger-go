package main

type GraphNode interface {
	NodeName() string
	Succ() []GraphNode
	Pred() []GraphNode
	Adj() []GraphNode
	Equal(other GraphNode) bool
}

type Graph interface {
	MkEdge(from, to GraphNode)
	RmEdge(from, to GraphNode)
}
