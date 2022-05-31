package main

import (
	"fmt"
	"strings"
)

const (
	wordSize = 4
)

var (
	argRegs = []Temp{a0, a1, a2, a3}
)

var (
	// function arguments registers
	a0 = tm.NewTemp()
	a1 = tm.NewTemp()
	a2 = tm.NewTemp()
	a3 = tm.NewTemp()

	fp = tm.NewTemp()

	// return value register
	rv = tm.NewTemp()
)

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

func (f *MipsFrame) StringFrag(label Label, str string) string {
	sb := strings.Builder{}
	sb.WriteString(".data\n")
	sb.WriteString(tm.LabelString(label))
	sb.WriteString(":\t.word\t")
	sb.WriteString(fmt.Sprintf("%d", len(str)))
	sb.WriteString("\n\t.ascii\t\"")
	for _, c := range str {
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

// ProcEntryExit1 is procedure entry and exit statement
// 4. save "escaping" arguments (including static link) into the frame, move nonescaping arguments into fresh temporary registers.
// 5. store instructions to save any calle-save registers - including the return address register - used within the function.
// 8. load instructions to restore the calle-save registers
func ProcEntryExit1(frame *MipsFrame, body StmIr) StmIr {
	return body
}
