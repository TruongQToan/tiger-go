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

func (s *IGraphNodeSet) Len() int {
	return len(s.indices)
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
	first := s.nodes[0]
	s.nodes = s.nodes[1:]
	delete(s.indices, first.temp)
	for i, v := range s.nodes {
		s.indices[v.temp] = i
	}

	return first, s
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

func (s *IGraphNodeSet) Clone() *IGraphNodeSet {
	clone := InitIGraphNodeSet()
	for _, node := range s.All() {
		clone.Add(node)
	}

	return clone
}

func (s *IGraphNodeSet) Add(node *IGraphNode) {
	if s.Has(node) {
		return
	}

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
	for i := idx; i < len(s.nodes); i++ {
		s.indices[s.nodes[i].temp] = i
	}
}

type IGraphNode struct {
	temp   Temp
	adj    *IGraphNodeSet
	degree int
}

func (node *IGraphNode) Id() int64 {
	panic("not implemented")
}

func (node *IGraphNode) NodeName() string {
	return tm.TempString(node.temp)
}

func (node *IGraphNode) Succ() []GraphNode {
	panic("not implemented")
}

func (node *IGraphNode) Pred() []GraphNode {
	panic("not implemented")
}

func (node *IGraphNode) Adj() []GraphNode {
	panic("don't support")
}

func (node *IGraphNode) AdjSet() *IGraphNodeSet {
	return node.adj
}

func (node *IGraphNode) Equal(other GraphNode) bool {
	return node.NodeName() == other.NodeName()
}
