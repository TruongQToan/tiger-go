package main

type BinOpIr int

const (
	PlusIr BinOpIr = iota
	MinusIr
	MulIr
	DivIr
	AndIr
	OrIr
	LshiftIr
	RshiftIr
	ArshiftIr
	Xor
)

type RelOpIr int

const (
	EqIr RelOpIr = iota
	NeIr
	LtIr
	GtIr
	LeIr
	GeIr
	UltIr
	UleIr
	UgtIr
	UgeIr
)

type ExpIr interface {
	IsExpIr()
}

type ConstExpIr struct {
	c int32
}

func (e *ConstExpIr) IsExpIr() {}

// NameExpIr assembly language label
type NameExpIr struct {
	label Label
}

func (e *NameExpIr) IsExpIr() {}

// TempExpIr corresponds to registers of machines, but in the abstract language, there are an infinite amount of registers.
type TempExpIr struct {
	temp Temp
}

func (e *TempExpIr) IsExpIr() {}

type BinOpExpIr struct {
	binop BinOpIr
	left ExpIr
	right ExpIr
}

func (e *BinOpExpIr) IsExpIr() {}

type MemExpIr struct {
	mem ExpIr
}

func (e *MemExpIr) IsExpIr() {}

type CallExpIr struct {
	exp  ExpIr
	args []ExpIr
}

func (e *CallExpIr) IsExpIr() {}

type EsEqExpIr struct {
	stm StmIr
	exp ExpIr
}

func (e *EsEqExpIr) IsExpIr() {}

type StmIr interface {
	IsStmIr()
}

type MoveStmIr struct {
	dst ExpIr
	src ExpIr
}

func (s *MoveStmIr) IsStmIr() {}

type ExpStmIr struct {
	exp ExpIr
}

func (s *ExpStmIr) IsStmIr() {}

type JumpStmIr struct {
	exp    ExpIr
	labels []Label
}

func (s *JumpStmIr) IsStmIr() {}

type CJumpStmIr struct {
	relop RelOpIr
	left  ExpIr
	right      ExpIr
	trueLabel  Label
	falseLabel Label
}

func (s *CJumpStmIr) IsStmIr() {}

type SeqStmIr struct {
	first StmIr
	second StmIr
}

func (s *SeqStmIr) IsStmIr() {}

type LabelStmIr struct {
	label Label
}

func (s *LabelStmIr) IsStmIr() {}

func isNullStm(s StmIr) bool {
	if v, ok := s.(*ExpStmIr); ok {
		if v, ok := v.exp.(*ConstExpIr); ok {
			return v.c == 0
		}
	}

	return false
}
