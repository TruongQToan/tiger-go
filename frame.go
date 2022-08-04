package main

type FrameAccess interface {
	exp(exp ExpIr) ExpIr
}

type Frame interface {
	Name() Label
	Formals() []FrameAccess
	AllocLocal(escape bool) FrameAccess
	TempName(t Temp) string
	TempMap() map[Temp]string
	ProcEntryExit1(body StmIr) StmIr
	ProcEntryExit3() (string, string)
	FP() Temp
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

type StrFrag struct {
	label Label
	str   string
}

func (frag *StrFrag) IsFragment() {}
