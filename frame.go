package main

type FrameAccess interface {
	exp(exp ExpIr) ExpIr
}

type Frame interface {
	Name() Label
	Formals() []FrameAccess
	AllocLocal(escape bool) FrameAccess
	StringFrag(label Label, str string) string
	TempName(t Temp) string
	TempMap() map[Temp]string
	ProcEntryExit1(body StmIr) StmIr
	ProcEntryExit2(body []Instr) []Instr
}

type FrameFactoryFunc func(name Label, formals []bool) Frame

func NewFrameFactory(arch string) FrameFactoryFunc {
	switch arch {
	case "mips":
		return NewMipsFrame
	default:
		panic("not supported yet")
	}
}

type Frag interface {
	IsFragment()
}

type ProcFrag struct {
	body  StmIr
	frame Frame
}

func (frag *ProcFrag) IsFragment() {}

type StringFrag struct {
	label Label
	str   string
}

func (frag *StringFrag) IsFragment() {}
