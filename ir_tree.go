package main

import (
	"strconv"
	"strings"
)

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

func (r RelOpIr) repr() string {
	switch r {
	case EqIr:
		return "eq"
	case NeIr:
		return "ne"
	case LtIr:
		return "lt"
	case GtIr:
		return "gt"
	case LeIr:
		return "le"
	case GeIr:
		return "ge"
	}

	return ""
}

type ExpIr interface {
	printExpIr(sb *strings.Builder, level int)
}

type ConstExpIr struct {
	c int32
}

func (e *ConstExpIr) printExpIr(sb *strings.Builder, level int) {
	indent(sb, level)
	sb.WriteString("Const\n")
	indent(sb, level+1)
	sb.WriteString(strconv.Itoa(int(e.c)) + "\n")
}

// NameExpIr assembly language label
type NameExpIr struct {
	label Label
}

func (e *NameExpIr) printExpIr(sb *strings.Builder, level int) {
	indent(sb, level)
	sb.WriteString("NameExp\n")
	indent(sb, level+1)
	sb.WriteString(strs.Get(Symbol(e.label)) + "\n")
}

// TempExpIr corresponds to registers of machines, but in the abstract language, there are an infinite amount of registers.
type TempExpIr struct {
	temp Temp
}

func (e *TempExpIr) printExpIr(sb *strings.Builder, level int) {
	indent(sb, level)
	sb.WriteString("Temp\n")
	indent(sb, level+1)
	sb.WriteString(strs.Get(Symbol(e.temp)) + "\n")
}

type BinOpExpIr struct {
	binop BinOpIr
	left  ExpIr
	right ExpIr
}

func (e *BinOpExpIr) printExpIr(sb *strings.Builder, level int) {
	indent(sb, level)
	switch e.binop {
	case PlusIr:
		sb.WriteString("Plus\n")
	case MinusIr:
		sb.WriteString("Minus\n")
	case MulIr:
		sb.WriteString("Mul\n")
	case DivIr:
		sb.WriteString("Div\n")
	}

	indent(sb, level+1)
	sb.WriteString("Left\n")
	e.left.printExpIr(sb, level+2)
	indent(sb, level+1)
	sb.WriteString("Right\n")
	e.right.printExpIr(sb, level+2)
}

type MemExpIr struct {
	mem ExpIr
}

func (e *MemExpIr) printExpIr(sb *strings.Builder, level int) {
	indent(sb, level)
	sb.WriteString("Mem\n")
	e.mem.printExpIr(sb, level+1)
}

type CallExpIr struct {
	exp  ExpIr
	args []ExpIr
}

func (e *CallExpIr) printExpIr(sb *strings.Builder, level int) {
	indent(sb, level)
	sb.WriteString("Call\n")
	indent(sb, level+1)
	sb.WriteString("FuncName\n")
	e.exp.printExpIr(sb, level+2)
	indent(sb, level+1)
	sb.WriteString("Arguments\n")
	for _, arg := range e.args {
		arg.printExpIr(sb, level+2)
	}
}

type EsEqExpIr struct {
	stm StmIr
	exp ExpIr
}

func (e *EsEqExpIr) printExpIr(sb *strings.Builder, level int) {
	indent(sb, level)
	sb.WriteString("EsEq\n")
	indent(sb, level+1)
	sb.WriteString("Stm\n")
	e.stm.printStm(sb, level+2)
	indent(sb, level+1)
	sb.WriteString("Exp\n")
	e.exp.printExpIr(sb, level+2)
}

type StmIr interface {
	printStm(sb *strings.Builder, level int)
}

type MoveStmIr struct {
	dst ExpIr
	src ExpIr
}

func (s *MoveStmIr) printStm(sb *strings.Builder, level int) {
	indent(sb, level)
	sb.WriteString("Move\n")
	indent(sb, level+1)
	sb.WriteString("Src\n")
	s.src.printExpIr(sb, level+2)
	indent(sb, level+1)
	sb.WriteString("Dest\n")
	s.dst.printExpIr(sb, level+2)
}

type ExpStmIr struct {
	exp ExpIr
}

func (s *ExpStmIr) printStm(sb *strings.Builder, level int) {
	indent(sb, level)
	sb.WriteString("ExpStm\n")
	s.exp.printExpIr(sb, level+1)
}

type JumpStmIr struct {
	exp    ExpIr
	labels []Label
}

func (s *JumpStmIr) printStm(sb *strings.Builder, level int) {
	indent(sb, level)
	sb.WriteString("Jump\n")
	s.exp.printExpIr(sb, level+1)
}

type CJumpStmIr struct {
	relop      RelOpIr
	left       ExpIr
	right      ExpIr
	trueLabel  Label
	falseLabel Label
}

func (s *CJumpStmIr) printStm(sb *strings.Builder, level int) {
	indent(sb, level)
	sb.WriteString("CJump\n")
	indent(sb, level+1)
	sb.WriteString("Op\n")
	indent(sb, level+2)
	sb.WriteString(s.relop.repr() + "\n")
	indent(sb, level+1)
	sb.WriteString("Left\n")
	s.left.printExpIr(sb, level+2)
	indent(sb, level+1)
	sb.WriteString("Right\n")
	s.right.printExpIr(sb, level+2)
	indent(sb, level+1)
	sb.WriteString("True Label\n")
	indent(sb, level+2)
	sb.WriteString(strs.Get(Symbol(s.trueLabel)) + "\n")
	indent(sb, level+1)
	sb.WriteString("False Label\n")
	indent(sb, level+2)
	sb.WriteString(strs.Get(Symbol(s.falseLabel)) + "\n")
}

type SeqStmIr struct {
	first  StmIr
	second StmIr
}

func (s *SeqStmIr) printStm(sb *strings.Builder, level int) {
	indent(sb, level)
	sb.WriteString("SeqStm\n")
	indent(sb, level+1)
	sb.WriteString("First\n")
	s.first.printStm(sb, level+2)
	indent(sb, level+1)
	sb.WriteString("Second\n")
	s.second.printStm(sb, level+2)
}

type LabelStmIr struct {
	label Label
}

func (s *LabelStmIr) printStm(sb *strings.Builder, level int) {
	indent(sb, level)
	sb.WriteString("Label\n")
	indent(sb, level+1)
	sb.WriteString(strs.Get(Symbol(s.label)) + "\n")
}

func isNullStm(s StmIr) bool {
	if v, ok := s.(*ExpStmIr); ok {
		if v, ok := v.exp.(*ConstExpIr); ok {
			return v.c == 0
		}
	}

	return false
}
