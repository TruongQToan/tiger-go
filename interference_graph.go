package main

type IGraph map[Temp]*IGraphNode
type tempEdge struct {
	u, v Temp
}

func computeLiveInOut(fGraph FGraph) {
	same := false
	for !same {
		same = true
		for _, node := range fGraph {
			oldLiveIn, oldLiveOut := node.liveIn.Clone(), node.liveOut.Clone()
			node.liveIn = node.use.Union(node.liveOut.Diff(node.def))
			node.liveOut = NewTempSet()
			for _, succ := range node.succ {
				iNode := succ.(*FGraphNode)
				node.liveOut = node.liveOut.Union(iNode.liveIn)
			}

			if !oldLiveIn.Equal(node.liveIn) || !oldLiveOut.Equal(node.liveOut) {
				same = false
			}
		}
	}
}

func addEdges(node *FGraphNode) []*tempEdge {
	edges := make([]*tempEdge, 0)
	for def := range node.def {
		for liveOut := range node.liveOut {
			if node.isMove && !node.use.Has(liveOut) {
				edges = append(edges, &tempEdge{
					u: def,
					v: liveOut,
				})
			} else if !node.isMove {
				edges = append(edges, &tempEdge{
					u: def,
					v: liveOut,
				})
			}
		}
	}

	return edges
}

func allEdges(fGraph FGraph) []*tempEdge {
	edges := make([]*tempEdge, 0)
	for _, node := range fGraph {
		edges = append(edges, addEdges(node)...)
	}

	return edges
}

func convertTempToIGraph(fGraph FGraph) IGraph {
	allTemps := make(TempSet)
	for _, node := range fGraph {
		allTemps = allTemps.Union(node.use.Union(node.def))
	}

	iNodes := make(map[Temp]*IGraphNode, len(allTemps))
	for temp := range allTemps {
		iNode := &IGraphNode{
			temp: temp,
			adj:  InitIGraphNodeSet(),
		}

		iNodes[temp] = iNode
	}

	return iNodes
}

func allMoves(fGraph FGraph, iNodes map[Temp]*IGraphNode) *MoveSet {
	pairs := InitMoveSet()
	for _, node := range fGraph {
		if !node.isMove {
			continue
		}

		var src, dst Temp
		src, node.use = node.use.Split()
		dst, node.def = node.def.Split()
		pairs.Add(&Move{
			src: iNodes[src],
			dst: iNodes[dst],
		})
	}

	return pairs
}

func InitIGraph(fGraph FGraph) (IGraph, *MoveSet) {
	computeLiveInOut(fGraph)
	edges := allEdges(fGraph)
	iGraph := convertTempToIGraph(fGraph)
	allMoves := allMoves(fGraph, iGraph)
	for _, edge := range edges {
		uNode, vNode := iGraph[edge.u], iGraph[edge.v]
		if !uNode.AdjSet().Has(vNode) {
			uNode.degree++
			uNode.adj.Add(vNode)
		}

		if !vNode.AdjSet().Has(uNode) {
			vNode.degree++
			vNode.adj.Add(uNode)
		}
	}

	return iGraph, allMoves
}
