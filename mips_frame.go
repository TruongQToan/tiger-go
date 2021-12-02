package tiger

import "errors"

type Register string
type Access interface {
	IsAccess()
}

type InFrameAccess struct {
	offset int32
}

func (f *InFrameAccess) IsAccess() {}

type InRegAccess struct {
	temp Temp
}

func (f *InRegAccess) IsAccess() {}

type Frag interface {
	IsFrag()
}

type ProcFrag struct {
	body  TreeStm
	frame MipsFrame
}

func (proc *ProcFrag) IsFrag() {}

type StringFrag struct {
	label Label
	str   string
}

func (proc *StringFrag) IsFrag() {}

const (
	wordSize = 4
)

var (
	// Expression evaluations and results of a function
	v0 = newTemp()
	v1 = newTemp()

	// Arguments
	a0 = newTemp()
	a1 = newTemp()
	a2 = newTemp()
	a3 = newTemp()

	// Temporary - not preserved across calls
	t0 = newTemp()
	t1 = newTemp()
	t2 = newTemp()
	t3 = newTemp()
	t4 = newTemp()
	t5 = newTemp()
	t6 = newTemp()
	t7 = newTemp()

	t8 = newTemp()
	t9 = newTemp()

	// Saved temporary - preserved across call
	s0 = newTemp()
	s1 = newTemp()
	s2 = newTemp()
	s3 = newTemp()
	s4 = newTemp()
	s5 = newTemp()
	s6 = newTemp()
	s7 = newTemp()

	zero = newTemp() // constant 0
	gp   = newTemp() // pointer for global area
	fp   = newTemp() // frame pointer
	sp   = newTemp() // stack pointer
	ra   = newTemp() // return address
	rv   = newTemp()

	specialArgs = []Temp{rv, fp, sp, ra}
	argRegs     = []Temp{a0, a1, a2, a3}
	calleeSaves = []Temp{s0, s1, s2, s3, s4, s5, s6, s7}
	callerSaves = []Temp{t0, t1, t2, t3, t4, t5, t6, t7}

	regList = []struct {
		regName string
		reg     Temp
	}{
		{
			regName: "$a0",
			reg:     a0,
		},
		{
			regName: "$a1",
			reg:     a1,
		},
		{
			regName: "$a2",
			reg:     a2,
		},
		{
			regName: "$a3",
			reg:     a3,
		},
		{
			regName: "$t0",
			reg:     t0,
		},
		{
			regName: "$t1",
			reg:     t1,
		},
		{
			regName: "$t2",
			reg:     t2,
		},
		{
			regName: "$t3",
			reg:     t3,
		},
		{
			regName: "$t4",
			reg:     t4,
		},
		{
			regName: "$t5",
			reg:     t5,
		},
		{
			regName: "$t6",
			reg:     t6,
		},
		{
			regName: "$t7",
			reg:     t7,
		},
		{
			regName: "$s0",
			reg:     s0,
		},
		{
			regName: "$s1",
			reg:     s1,
		},
		{
			regName: "$s2",
			reg:     s2,
		},
		{
			regName: "$s3",
			reg:     s3,
		},
		{
			regName: "$s4",
			reg:     s4,
		},
		{
			regName: "$s5",
			reg:     s5,
		},
		{
			regName: "$s6",
			reg:     s6,
		},
		{
			regName: "$s7",
			reg:     s7,
		},
		{
			regName: "$fp",
			reg:     fp,
		},
		{
			regName: "$rv",
			reg:     rv,
		},
		{
			regName: "$sp",
			reg:     sp,
		},
		{
			regName: "$ra",
			reg:     ra,
		},
	}
)

type MipsFrame struct {
	name    Label
	formals []Access
	locals  int32
	instrs  []TreeStm
}

func NewFrame(name Label, formals []bool) (*MipsFrame, error) {
	if len(formals) > len(argRegs) {
		return nil, errors.New("function has too many arguments")
	}

	frame := &MipsFrame{
		name: name,
	}

	accs := make([]Access, 0, len(formals))
	numEscapes := 0
	for _, formal := range formals {
		var acc Access
		if formal {
			// Escape into memory
			acc = &InFrameAccess{offset: int32(wordSize * (1 + numEscapes))}
			numEscapes++
		} else {
			acc = &InRegAccess{temp: newTemp()}
		}

		accs = append(accs, acc)
	}

	frame.formals = accs
	instructions := make([]TreeStm, 0, len(accs))
	for i, acc := range accs {
		var exp TreeExp
		if v, ok := acc.(*InFrameAccess); ok {
			exp = &MemTreeExp{exp: &BinOpTreeExp{
				binOp: PlusBinOp,
				left:  &TempTreeExp{temp: fp},
				right: &ConstTreeExp{cnst: v.offset},
			}}
		} else {
			v := acc.(*InRegAccess)
			exp = &TempTreeExp{temp: v.temp}
		}

		instructions = append(instructions, &MoveStm{
			firstExp: exp,
			rightExp: &TempTreeExp{temp: argRegs[i]},
		})
	}

	frame.instrs = instructions
	frame.locals = 0
	return frame, nil
}

func (f *MipsFrame) Name() Label {
	return f.name
}

func (f *MipsFrame) Formals() []Access {
	return f.formals
}
