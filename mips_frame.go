package main

import (
	"fmt"
	"strings"
)

const (
	wordSize = 4
)

var (
	zero = tm.NewTemp() // $0

	// expression evaluation and results of a function
	v0 = tm.NewTemp() // $2
	v1 = tm.NewTemp() // $3

	// function arguments registers
	a0 = tm.NewTemp() // $4
	a1 = tm.NewTemp() // $5
	a2 = tm.NewTemp() // $6
	a3 = tm.NewTemp() // $7

	// temporaries - not preserved across call
	t0 = tm.NewTemp() // $8
	t1 = tm.NewTemp() // $9
	t2 = tm.NewTemp() // $10
	t3 = tm.NewTemp() // $11
	t4 = tm.NewTemp() // $12
	t5 = tm.NewTemp() // $13
	t6 = tm.NewTemp() // $14
	t7 = tm.NewTemp() // $15

	// save values - preserved across calls
	s0 = tm.NewTemp() // $16
	s1 = tm.NewTemp() // $17
	s2 = tm.NewTemp() // $18
	s3 = tm.NewTemp() // $19
	s4 = tm.NewTemp() // $20
	s5 = tm.NewTemp() // $21
	s6 = tm.NewTemp() // $22
	s7 = tm.NewTemp() // $23

	// temporaries - not preserved across call
	t8 = tm.NewTemp() // $24
	t9 = tm.NewTemp() // $25

	// pointer to the global area
	gp = tm.NewTemp() // $28

	// stack pointer
	sp = tm.NewTemp() // $29

	// frame pointer
	fp = tm.NewTemp() // $30

	// return address
	ra = tm.NewTemp() // $31
	rv = v0

	specialArgs = []Temp{rv, fp, sp, ra}

	// registers to save arguments
	argRegs = []Temp{a0, a1, a2, a3}

	// callee-saved registers
	calleeSaves = []Temp{s0, s1, s2, s3, s4, s5, s6, s7}

	// caller-saved registers
	callerSaves = []Temp{t0, t1, t2, t3, t4, t5, t6, t7, t8, t9}

	tempMap = map[Temp]string{
		a0: "$a0",
		a1: "$a1",
		a2: "$a2",
		a3: "$a3",
		t0: "$t0",
		t1: "$t1",
		t2: "$t2",
		t3: "$t3",
		t4: "$t4",
		t5: "$t5",
		t6: "$t6",
		t7: "$t7",
		t8: "$t8",
		t9: "$t9",
		s1: "$s1",
		s2: "$s2",
		s3: "$s3",
		s4: "$s4",
		s5: "$s5",
		s6: "$s6",
		s7: "$s7",
		fp: "$fp",
		v0: "$rv",
		sp: "$sp",
		ra: "$ra",
	}
)

func tempName(t Temp) string {
	v, ok := tempMap[t]
	if ok {
		return v
	}

	return tm.MakeTempString(t)
}

type InFrameMipsAccess struct {
	offset int32
}

// exp converts InFrameMipsAccess into ExpIr. The argument is the address of the stack frame that the access lives in.
func (a *InFrameMipsAccess) exp(frameAddress ExpIr) ExpIr {
	return &MemExpIr{mem: &BinOpExpIr{
		binop: PlusIr,
		left:  frameAddress,
		right: &ConstExpIr{c: a.offset},
	}}
}

type InRegMipsAccess struct {
	temp Temp
}

func (a *InRegMipsAccess) exp(_ ExpIr) ExpIr {
	return &TempExpIr{temp: a.temp}
}

type MipsFrame struct {
	name       Label
	accesses   []FrameAccess
	shiftInsts *SeqStmIr
	locals     int32
}

func NewMipsFrame(name Label, escapes []bool) Frame {
	frame := MipsFrame{
		name: name,
	}

	frame.createAccesses(0, escapes)
	return &frame
}

func (f *MipsFrame) TempMap() map[Temp]string {
	return tempMap
}

func (f *MipsFrame) TempName(t Temp) string {
	return tempName(t)
}

func (f *MipsFrame) createAccesses(i int32, escapes []bool) {
	if int(i) >= len(escapes) {
		return
	}

	// after the first 4 arguments, always escape
	if int(i) >= len(argRegs) {
		f.accesses = append(f.accesses, &InFrameMipsAccess{offset: i * wordSize})
		f.createAccesses(i+1, escapes)
		return
	}

	var acc FrameAccess
	// the first 4 arguments are passed in 4 registers [$a0, $a1, $a2, $a3]
	if escapes[i] {
		acc = &InFrameMipsAccess{offset: i * wordSize}
	} else {
		acc = &InRegMipsAccess{tm.NewTemp()}
	}

	f.accesses = append(f.accesses, acc)
	f.shiftInsts = &SeqStmIr{
		first: f.shiftInsts,
		second: &MoveStmIr{
			dst: acc.exp(&TempExpIr{fp}),
			src: &TempExpIr{temp: argRegs[i]},
		},
	}

	f.createAccesses(i+1, escapes)
}

func (f *MipsFrame) Name() Label {
	return f.name
}

func (f *MipsFrame) Formals() []FrameAccess {
	return f.accesses
}

func (f *MipsFrame) AllocLocal(escape bool) FrameAccess {
	if escape {
		f.locals++
		return &InFrameMipsAccess{offset: -wordSize * f.locals}
	}

	return &InRegMipsAccess{tm.NewTemp()}
}

func (f *MipsFrame) FP() Temp {
	return fp
}

// ProcEntryExit1 is procedure entry and exit statement
// 4. save "escaping" arguments (including static link) into the frame, move nonescaping arguments into fresh temporary registers.
// 5. store instructions to save any calle-save registers - including the return address register - used within the function.
// 8. load instructions to restore the calle-save registers
func (f *MipsFrame) ProcEntryExit1(body StmIr) StmIr {
	shifts := f.shiftInsts

	calleeSaveRegs := append(calleeSaves, ra)
	accesses := make([]FrameAccess, len(calleeSaveRegs))
	for i := 0; i < len(calleeSaveRegs); i++ {
		accesses[i] = f.AllocLocal(false)
	}

	saves := make([]StmIr, 0, len(calleeSaveRegs))
	for i, reg := range calleeSaveRegs {
		saves = append(saves, &MoveStmIr{
			dst: accesses[i].exp(&TempExpIr{fp}),
			src: &TempExpIr{reg},
		})
	}

	restores := make([]StmIr, 0, len(calleeSaveRegs))
	for i := len(calleeSaveRegs) - 1; i >= 0; i-- {
		restores = append(restores, &MoveStmIr{
			src: accesses[i].exp(&TempExpIr{fp}),
			dst: &TempExpIr{calleeSaveRegs[i]},
		})
	}

	return &SeqStmIr{
		first:  shifts,
		second: seqStm(append(append(saves, body), restores...)...),
	}
}

func (f *MipsFrame) ProcEntryExit3() (string, string) {
	offset := (int(f.locals) + len(argRegs)) * wordSize
	prolog := fmt.Sprintf("%s:\n\tsw\t$fp\t0($sp)\n\tmove\t$fp\t$sp\n\taddiu\t$sp\t$sp\t-%d\n",
		tm.LabelString(f.Name()), offset)

	epilog := fmt.Sprintf("\tmove\t$sp\t$fp\n\tlw\t$fp\t0($sp)\n\tjr\t$ra\n\n")
	return prolog, epilog
}

// ProcEntryExit2 notifies the register allocation that zero, ra, sp, calleSaves are live out at the end of the function.
func ProcEntryExit2(body []Instr) []Instr {
	return append(body, &OperInstr{
		src: append([]Temp{zero, ra, sp}, calleeSaves...),
	})
}

func StringFrag(sb *strings.Builder, frag *StrFrag) string {
	sb.WriteString(".data\n")
	sb.WriteString(tm.LabelString(frag.label))
	sb.WriteString(":\t.word\t")
	sb.WriteString(fmt.Sprintf("%d", len(frag.str)))
	sb.WriteString("\n\t.ascii\t\"")
	for _, c := range frag.str {
		switch c {
		case '\n':
			sb.WriteString("\\n")
		case '\t':
			sb.WriteString("\\t")
		case '0':
			sb.WriteString("\\0")
		case '"':
			sb.WriteString("\\\"")
		case '\'':
			sb.WriteString("\\'")
		case '\\':
			sb.WriteString("\\\\")
		default:
			if c >= ' ' || c < 127 {
				sb.WriteByte(byte(c))
			} else {
				sb.WriteByte('\\')
				sb.WriteByte('0' + (byte(c) >> 6))
				sb.WriteByte('0' + ((byte(c) >> 3) & 8))
				sb.WriteByte('0' + (byte(c) & 8))
			}
		}
	}

	sb.WriteString("\"\n\t.align\t2\n")
	return sb.String()
}
