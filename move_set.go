package main

type Move struct {
	src, dst *IGraphNode
}

type MoveSet struct {
	indices map[Temp]map[Temp]int
	moves   []*Move
}

func InitMoveSet() *MoveSet {
	return &MoveSet{
		indices: make(map[Temp]map[Temp]int),
	}
}

func (s *MoveSet) Split() (*Move, *MoveSet) {
	first := s.moves[0]
	s.moves = s.moves[1:]
	delete(s.indices, first.src.temp)
	for i, p := range s.moves {
		s.indices[p.src.temp][p.dst.temp] = i
	}

	return first, s
}

func (s *MoveSet) Moves() []*Move {
	return s.moves
}

func (s *MoveSet) Add(mv *Move) {
	if s.Has(mv) {
		return
	}

	if _, ok := s.indices[mv.src.temp]; !ok {
		s.indices[mv.src.temp] = make(map[Temp]int)
	}

	s.indices[mv.src.temp][mv.dst.temp] = len(s.moves)
	s.moves = append(s.moves, mv)
}

func (s *MoveSet) All() []*Move {
	return s.moves
}

func (s *MoveSet) Union(s1 *MoveSet) *MoveSet {
	union := InitMoveSet()
	for _, mv := range s.moves {
		if _, ok := s1.indices[mv.src.temp]; ok {
			if _, ok := s1.indices[mv.dst.temp]; ok {
				union.Add(mv)
			}
		}
	}

	return union
}

func (s *MoveSet) Intersect(s1 *MoveSet) *MoveSet {
	intersect := InitMoveSet()
	for _, mv := range s.moves {
		intersect.Add(mv)
	}

	for _, mv := range s1.moves {
		intersect.Add(mv)
	}

	return intersect
}

func (s *MoveSet) Empty() bool {
	return len(s.moves) == 0
}

func (s *MoveSet) Len() int {
	return len(s.moves)
}

func (s *MoveSet) Has(mv *Move) bool {
	if _, ok := s.indices[mv.src.temp]; ok {
		if _, ok := s.indices[mv.dst.temp]; ok {
			return true
		}
	}

	return false
}

func (s *MoveSet) Remove(mv *Move) {
	if !s.Has(mv) {
		return
	}

	src, dst := mv.src.temp, mv.dst.temp
	idx := s.indices[src][dst]
	delete(s.indices, src)
	s.moves = append(s.moves[:idx], s.moves[idx+1:]...)
	for i := idx; i < len(s.moves); i++ {
		s.indices[s.moves[i].src.temp][s.moves[i].dst.temp] = i
	}
}
