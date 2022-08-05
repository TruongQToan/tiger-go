package main

import "fmt"

type Coloring struct {
	iGraph IGraph
	fGraph FGraph

	moves *MoveSet

	registers map[Temp]string

	moveList map[Temp]*MoveSet

	colored map[Temp]Temp

	precolored *IGraphNodeSet

	// temporary registers, not precolored, and not yet processed
	initials *IGraphNodeSet

	// high-degree nodes
	spillWorklist *IGraphNodeSet

	// low-degree move-related nodes
	freezeWorklist *IGraphNodeSet

	// list of low-degree non-move-related nodes
	simplifyWorklist *IGraphNodeSet

	// moves enabled for possible coalescing.
	worklistMoves *MoveSet

	// moves not yet ready for coalescing.
	activeMoves *MoveSet

	// moves that have been coalesced
	coalescedMoves *MoveSet

	// moves that cannot be coalesced
	constrainedMoves *MoveSet

	// frozenMoves moves have been frozen
	frozenMoves *MoveSet
	// K: number of colors
	K int

	selectStack []*IGraphNode

	coalesceNodes *IGraphNodeSet
	alias         map[Temp]*IGraphNode

	spilledNodes *IGraphNodeSet
	coloredNodes *IGraphNodeSet
}

func NewColoring(iGraph IGraph,
	fGraph FGraph,
	moves *MoveSet,
	registers map[Temp]string,
) *Coloring {
	coloring := Coloring{
		iGraph:           iGraph,
		fGraph:           fGraph,
		moves:            moves,
		registers:        registers,
		moveList:         make(map[Temp]*MoveSet),
		colored:          make(map[Temp]Temp),
		precolored:       InitIGraphNodeSet(),
		initials:         InitIGraphNodeSet(),
		spillWorklist:    InitIGraphNodeSet(),
		freezeWorklist:   InitIGraphNodeSet(),
		simplifyWorklist: InitIGraphNodeSet(),
		worklistMoves:    InitMoveSet(),
		activeMoves:      InitMoveSet(),
		coalescedMoves:   InitMoveSet(),
		constrainedMoves: InitMoveSet(),
		frozenMoves:      InitMoveSet(),
		K:                len(registers),
		coalesceNodes:    InitIGraphNodeSet(),
		alias:            make(map[Temp]*IGraphNode),
		spilledNodes:     InitIGraphNodeSet(),
		coloredNodes:     InitIGraphNodeSet(),
	}

	return &coloring
}

func (c *Coloring) addMoves(node *IGraphNode, move *Move) {
	if _, ok := c.moveList[node.temp]; !ok {
		c.moveList[node.temp] = InitMoveSet()
	}

	c.moveList[node.temp].Add(move)
}

func (c *Coloring) initMoveList() {
	for k := range c.iGraph {
		c.moveList[k] = InitMoveSet()
	}

	for _, move := range c.moves.All() {
		if !c.precolored.Has(move.src) {
			c.addMoves(move.src, move)
		}

		if !c.precolored.Has(move.dst) {
			c.addMoves(move.dst, move)
		}

		c.worklistMoves.Add(move)
	}
}

func (c *Coloring) initColoredAndPrecolored() {
	for tmp, node := range c.iGraph {
		if _, ok := c.registers[tmp]; ok {
			c.precolored.Add(node)
			c.colored[tmp] = tmp
			c.coloredNodes.Add(node)
		} else {
			c.initials.Add(node)
		}
	}
}

func (c *Coloring) nodeMoves(n *IGraphNode) *MoveSet {
	return c.moveList[n.temp].Intersect(c.activeMoves.Union(c.worklistMoves))
}

func (c *Coloring) moveRelated(n *IGraphNode) bool {
	return !c.nodeMoves(n).Empty()
}

func (c *Coloring) makeWorklist() {
	for _, node := range c.initials.All() {
		if node.degree >= c.K {
			c.spillWorklist.Add(node)
		} else if c.moveRelated(node) {
			c.freezeWorklist.Add(node)
		} else {
			c.simplifyWorklist.Add(node)
		}
	}

	c.initials.Reset()
}

func (c *Coloring) build() {
	c.initMoveList()
	c.initColoredAndPrecolored()
}

func (c *Coloring) enableMoves(nodes *IGraphNodeSet) {
	for _, n := range nodes.All() {
		for _, move := range c.nodeMoves(n).Moves() {
			if c.activeMoves.Has(move) {
				c.activeMoves.Remove(move)
				c.worklistMoves.Add(move)
			}
		}
	}
}

func (c *Coloring) decrementDegree(node *IGraphNode) {
	d := node.degree
	node.degree--
	if d == c.K {
		adj := c.adj(node)
		adj.Add(node)
		c.enableMoves(adj)
		c.spillWorklist.Remove(node)
		if c.moveRelated(node) {
			c.freezeWorklist.Add(node)
		} else {
			c.simplifyWorklist.Add(node)
		}
	}
}

func (c *Coloring) simplify() {
	var node *IGraphNode
	node, c.simplifyWorklist = c.simplifyWorklist.Split()
	c.selectStack = append(c.selectStack, node)
	for _, adj := range c.adj(node).All() {
		c.decrementDegree(adj)
	}
}

func (c *Coloring) coalesce() {
	var mv *Move
	mv, c.worklistMoves = c.worklistMoves.Split()

	// x = mv.dst, y = mv.src
	x, y := c.findAlias(mv.dst), c.findAlias(mv.src)

	var (
		u, v *IGraphNode
	)

	if c.precolored.Has(y) {
		u, v = y, x
	} else {
		u, v = x, y
	}

	if u.temp == v.temp {
		c.coalescedMoves.Add(mv)
		c.addWorklist(u)
	} else if c.precolored.Has(v) || c.adj(u).Has(v) {
		// if v is precolored, so in this case, both the dst and src of the move is precolored.
		c.constrainedMoves.Add(mv)
		c.addWorklist(u)
		c.addWorklist(v)
	} else if (c.precolored.Has(u) && c.georgeTest(v, u)) || (!c.precolored.Has(u) && c.briggsTest(c.adj(u), c.adj(v))) {
		c.coalescedMoves.Add(mv)
		c.combine(u, v)
		c.addWorklist(u)
	} else {
		c.activeMoves.Add(mv)
	}
}

func (c *Coloring) spillCost(iNode *IGraphNode) float64 {
	useDefines := 0
	for _, fNode := range c.fGraph {
		if fNode.def.Has(iNode.temp) {
			useDefines++
		}

		if fNode.use.Has(iNode.temp) {
			useDefines++
		}
	}

	return float64(useDefines) / float64(iNode.degree)
}

func (c *Coloring) adj(n *IGraphNode) *IGraphNodeSet {
	nodes := c.coloredNodes.Clone()
	for _, node := range c.selectStack {
		nodes.Add(node)
	}

	return n.AdjSet().Diff(nodes)
}

func (c *Coloring) combine(u, v *IGraphNode) {
	fmt.Println("combine u and v", tempMap[u.temp], tm.TempString(v.temp))
	// v isn't in simplifyWorklist when this function is called because that worklist is empty, as in the code of the Main function
	if c.freezeWorklist.Has(v) {
		c.freezeWorklist.Remove(v)
	} else {
		c.spillWorklist.Remove(v)
	}

	c.coalesceNodes.Add(v)
	c.alias[v.temp] = u
	c.moveList[u.temp] = c.moveList[u.temp].Union(c.moveList[v.temp])
	for _, node := range c.adj(v).All() {
		c.addEdge(node, u)
		c.decrementDegree(node)
	}

	if u.degree >= c.K && c.freezeWorklist.Has(u) {
		fmt.Println("spill coalesced", tm.TempString(u.temp))
		c.freezeWorklist.Remove(u)
		c.spillWorklist.Add(u)
	}
}

func (c *Coloring) addEdge(u, v *IGraphNode) {
	if u.temp == v.temp {
		return
	}

	if u.AdjSet().Has(v) {
		return
	}

	if !c.precolored.Has(u) {
		u.AdjSet().Add(v)
		u.degree++
	}

	if !c.precolored.Has(v) {
		v.AdjSet().Add(u)
		v.degree++
	}
}

func (c *Coloring) georgeTest(a, b *IGraphNode) bool {
	// a and b can be coalesced if for every adjacent node t of a. Either t is an insignificant node (degree(t) < K)
	// or t is adjacent with b
	for _, node := range c.adj(a).All() {
		if !c.ok(node, b) {
			return false
		}
	}

	return true
}

func (c *Coloring) briggsTest(adj1, adj2 *IGraphNodeSet) bool {
	k := 0
	for _, node := range adj1.Intersect(adj2).All() {
		if node.degree-1 >= c.K {
			k++
			if k >= c.K {
				return false
			}
		}
	}

	for _, node := range adj1.Diff(adj2).Union(adj2.Diff(adj1)).All() {
		if node.degree >= c.K {
			k++
			if k >= c.K {
				return false
			}
		}
	}

	return true
}

func (c *Coloring) ok(t, b *IGraphNode) bool {
	return t.degree < c.K || c.precolored.Has(t) || t.AdjSet().Has(b)
}

func (c *Coloring) addWorklist(node *IGraphNode) {
	if c.precolored.Has(node) {
		return
	}

	if c.moveRelated(node) {
		return
	}

	if node.degree >= c.K {
		return
	}

	c.simplifyWorklist.Add(node)
	c.freezeWorklist.Remove(node)
}

func (c *Coloring) findAlias(node *IGraphNode) *IGraphNode {
	if c.coalesceNodes.Has(node) {
		return c.alias[node.temp]
	}

	return node
}

func (c *Coloring) freeze() {
	var node *IGraphNode
	node, c.freezeWorklist = c.freezeWorklist.Split()

	// we can do this without checking c.moveRelated(node) is Empty because in the freezeMoves method,
	// we already remove all the move related to node in activeModes.
	// aldo, to this point, workingMoves is empty

	c.simplifyWorklist.Add(node)
	c.freezeWorklist.Remove(node)

	c.freezeMoves(node)
}

func (c *Coloring) freezeMoves(u *IGraphNode) {
	var v *IGraphNode
	for _, mv := range c.moveList[u.temp].Moves() {
		if u.temp == mv.src.temp {
			v = mv.dst
		} else {
			v = mv.src
		}

		c.frozenMoves.Add(mv)
		c.activeMoves.Remove(mv)
		if !c.moveRelated(v) && v.degree < c.K {
			c.freezeWorklist.Remove(v)
			c.simplifyWorklist.Add(v)
		}
	}
}

func (c *Coloring) selectSpill() {
	minNode := c.spillWorklist.All()[0]
	for _, node := range c.spillWorklist.All()[1:] {
		if c.spillCost(node) < c.spillCost(minNode) {
			minNode = node
		}
	}

	c.spillWorklist.Remove(minNode)
	c.simplifyWorklist.Add(minNode)
	c.freezeMoves(minNode)
}

func (c *Coloring) assignColor() {
	for len(c.selectStack) > 0 {
		node := c.selectStack[len(c.selectStack)-1]
		c.selectStack = c.selectStack[:len(c.selectStack)-1]
		okColors := NewTempSet()
		for color := range c.registers {
			okColors.Add(color)
		}

		fmt.Println("len adj", node.AdjSet().Len(), len(c.registers))
		for _, adj := range node.AdjSet().All() {
			fmt.Printf("%s ", tm.TempString(adj.temp))
		}

		fmt.Println()

		for _, adj := range node.AdjSet().All() {
			v := c.findAlias(adj)
			if c.coloredNodes.Has(v) || c.precolored.Has(v) {
				okColors.Remove(c.colored[v.temp])
			}
		}

		if len(okColors) == 0 {
			c.spilledNodes.Add(node)
			continue
		}

		c.coloredNodes.Add(node)
		color, _ := okColors.Split()
		c.colored[node.temp] = color
	}

	for _, node := range c.coalesceNodes.All() {
		c.colored[node.temp] = c.colored[c.findAlias(node).temp]
		fmt.Println("coalesced nodes", tm.TempString(node.temp), tempMap[c.findAlias(node).temp], tempMap[c.colored[c.findAlias(node).temp]])
	}
}

func (c *Coloring) Color() (map[Temp]Temp, *IGraphNodeSet) {
	c.build()
	c.makeWorklist()
	for {
		if !c.simplifyWorklist.Empty() {
			c.simplify()
		} else if c.worklistMoves.Len() != 0 {
			c.coalesce()
		} else if !c.freezeWorklist.Empty() {
			c.freeze()
		} else if !c.spillWorklist.Empty() {
			c.selectSpill()
		} else {
			break
		}
	}

	c.assignColor()
	return c.colored, c.spilledNodes
}
