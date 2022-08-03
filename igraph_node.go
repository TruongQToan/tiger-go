package main

type IGraphNodeSet struct {
	indices map[Temp]int
	nodes   []*IGraphNode
}

func InitIGraphNodeSet() *IGraphNodeSet {
	return &IGraphNodeSet{
		indices: make(map[Temp]int),
	}
}

func (s *IGraphNodeSet) Has(node *IGraphNode) bool {
	_, ok := s.indices[node.temp]
	return ok
}

func (s *IGraphNodeSet) Diff(s1 *IGraphNodeSet) *IGraphNodeSet {
	diff := InitIGraphNodeSet()
	for _, node := range s.All() {
		diff.Add(node)
	}

	for _, node := range s1.All() {
		diff.Remove(node)
	}

	return diff
}

func (s *IGraphNodeSet) Intersect(s1 *IGraphNodeSet) *IGraphNodeSet {
	u := s.Diff(s1)
	v := s1.Diff(s)
	return s.Union(s1).Diff(u.Union(v))
}

func (s *IGraphNodeSet) Empty() bool {
	return len(s.nodes) == 0
}

func (s *IGraphNodeSet) Split() (*IGraphNode, *IGraphNodeSet) {
	s1 := InitIGraphNodeSet()
	first := s.nodes[0]
	s1.nodes = s.nodes[1:]
	delete(s.indices, first.temp)
	for k, v := range s1.nodes {
		s1.indices[v.temp] = k
	}

	return first, s1
}

func (s *IGraphNodeSet) Reset() {
	s.nodes = nil
	s.indices = make(map[Temp]int)
}

func (s *IGraphNodeSet) Union(s1 *IGraphNodeSet) *IGraphNodeSet {
	union := InitIGraphNodeSet()
	for _, node := range s.All() {
		union.Add(node)
	}

	for _, node := range s1.All() {
		union.Add(node)
	}

	return union
}

func (s *IGraphNodeSet) All() []*IGraphNode {
	return s.nodes
}

func (s *IGraphNodeSet) Add(node *IGraphNode) {
	s.indices[node.temp] = len(s.nodes)
	s.nodes = append(s.nodes, node)
}

func (s *IGraphNodeSet) Remove(node *IGraphNode) {
	idx, ok := s.indices[node.temp]
	if !ok {
		return
	}

	delete(s.indices, node.temp)
	s.nodes = append(s.nodes[:idx], s.nodes[idx+1:]...)
}

type IGraphNode struct {
	temp   Temp
	adj    []GraphNode
	degree int
}

func (node *IGraphNode) NodeName() string {
	return tm.MakeTempString(node.temp)
}

func (node *IGraphNode) Succ() []GraphNode {
	panic("not implemented")
}

func (node *IGraphNode) Pred() []GraphNode {
	panic("not implemented")
}

func (node *IGraphNode) Adj() []GraphNode {
	return node.adj
}

func (node *IGraphNode) AdjSet() *IGraphNodeSet {
	adj := InitIGraphNodeSet()
	for _, n := range node.adj {
		adj.Add(n.(*IGraphNode))
	}

	return adj
}

func (node *IGraphNode) Equal(other GraphNode) bool {
	return node.NodeName() == other.NodeName()
}
