package main

import "math/rand"

type FGraphNode struct {
	id int64

	instr  Instr
	use    TempSet
	def    TempSet
	isMove bool

	succ, pred, adj          []GraphNode
	succMap, predMap, adjMap map[int64]int

	liveOut, liveIn TempSet
}

func Instr2FGraphNode(instr Instr) *FGraphNode {
	var (
		use, def []Temp
		isMove   bool
	)

	switch v := instr.(type) {
	case *OperInstr:
		use, def, isMove = v.src, v.dst, false
	case *LabelInstr:
		use, def, isMove = nil, nil, false
	case *MoveInstr:
		use, def, isMove = []Temp{v.src}, []Temp{v.dst}, true
	}

	return &FGraphNode{
		id:     rand.Int63(),
		instr:  instr,
		use:    NewTempSet(use...),
		def:    NewTempSet(def...),
		isMove: isMove,

		succMap: make(map[int64]int),
		predMap: make(map[int64]int),
		adjMap:  make(map[int64]int),
	}
}

func (node *FGraphNode) NodeName() string {
	return node.instr.assemStr()
}

func (node *FGraphNode) Succ() []GraphNode {
	return node.succ
}

func (node *FGraphNode) Adj() []GraphNode {
	return append(append([]GraphNode{}, node.pred...), node.succ...)
}

func (node *FGraphNode) Pred() []GraphNode {
	return node.pred
}

func (node *FGraphNode) AddSucc(succ *FGraphNode) {
	if _, ok := node.succMap[succ.id]; ok {
		return
	}

	node.succMap[succ.id] = len(node.succ)
	node.succ = append(node.succ, succ)
}

func (node *FGraphNode) AddPred(pred *FGraphNode) {
	if _, ok := node.predMap[pred.id]; ok {
		return
	}

	node.predMap[pred.id] = len(node.pred)
	node.pred = append(node.pred, pred)
}

func (node *FGraphNode) RmPred(pred *FGraphNode) {
	v, ok := node.predMap[pred.id]
	if !ok {
		return
	}

	node.pred = append(node.pred[:v], node.pred[v+1:]...)
	delete(node.predMap, pred.id)
}

func (node *FGraphNode) RmSucc(succ *FGraphNode) {
	v, ok := node.succMap[succ.id]
	if !ok {
		return
	}

	node.succ = append(node.succ[:v], node.succ[v+1:]...)
	delete(node.succMap, succ.id)
}

func (node *FGraphNode) Equal(other GraphNode) bool {
	v, ok := other.(*FGraphNode)
	if !ok {
		return false
	}

	return v.id == node.id
}

func Instrs2FGraph(instrs []Instr) FGraph {
	flowGraph := make(FGraph, 0, len(instrs))
	label2Node := map[Label]*FGraphNode{}
	for _, instr := range instrs {
		node := Instr2FGraphNode(instr)
		flowGraph = append(flowGraph, node)
		if v, ok := instr.(*LabelInstr); ok {
			label2Node[v.lab] = node
		}
	}

	for i, instr := range instrs {
		if len(instr.jumpLabels()) > 0 {
			for _, label := range instr.jumpLabels() {
				node := label2Node[label]
				flowGraph.MkEdge(flowGraph[i], node)
			}

			continue
		}

		if i < len(instrs)-1 {
			flowGraph.MkEdge(flowGraph[i], flowGraph[i+1])
		}
	}

	return flowGraph
}

type FGraph []*FGraphNode

func (g *FGraph) MkEdge(from, to GraphNode) {
	from1 := from.(*FGraphNode)
	to1 := to.(*FGraphNode)
	from1.AddSucc(to1)
	to1.AddPred(from1)
}

func (g *FGraph) RmEdge(from, to GraphNode) {
	from1 := from.(*FGraphNode)
	to1 := to.(*FGraphNode)
	from1.AddSucc(to1)
	to1.AddPred(from1)
}
