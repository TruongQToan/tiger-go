package main

import (
	"strings"
)

type reg = string

type Instr interface {
	assemStr() string
	srcRegs() []Temp
	dstRegs() []Temp
	jumpLabels() []Label
}

type OperInstr struct {
	assem string
	dst   []Temp
	src   []Temp
	jumps []Label
}

func (o *OperInstr) assemStr() string {
	return o.assem
}

func (o *OperInstr) srcRegs() []Temp {
	return o.src
}

func (o *OperInstr) dstRegs() []Temp {
	return o.dst
}

func (o *OperInstr) jumpLabels() []Label {
	return o.jumps
}

type LabelInstr struct {
	assem string
	lab   Label
}

func (l *LabelInstr) assemStr() string {
	return l.assem
}

func (l *LabelInstr) srcRegs() []Temp {
	return nil
}

func (l *LabelInstr) dstRegs() []Temp {
	return nil
}

func (l *LabelInstr) jumpLabels() []Label {
	return nil
}

type MoveInstr struct {
	assem string
	dst   Temp
	src   Temp
}

func (m *MoveInstr) assemStr() string {
	return m.assem
}

func (m *MoveInstr) srcRegs() []Temp {
	return []Temp{m.src}
}

func (m *MoveInstr) dstRegs() []Temp {
	return []Temp{m.dst}
}

func (m *MoveInstr) jumpLabels() []Label {
	return nil
}

// in assem, we may have something like "addi `d0, `s0, 3".
// This function replaces `d0, `s0 with actual registers.
func formatAssem(i Instr, frame Frame) {
	sources, dests, jumpLabels := i.srcRegs(), i.dstRegs(), i.jumpLabels()
	assem := i.assemStr()
	sb := strings.Builder{}
	for i := 0; i < len(assem); i++{
		if assem[i] == '`' {
			i++
			switch assem[i] {
			case 's':
				i++
				if assem[i] < '0' || assem[i] > '9' {
					panic("invalid source register number")
				}

				n := assem[i] - '0'
				sb.WriteString(frame.TempMap(sources[n]))

			case 'd':
				i++
				if assem[i] < '0' || assem[i] > '9' {
					panic("invalid destination register number")
				}

				n := assem[i] - '0'
				sb.WriteString(frame.TempMap(dests[n]))

			case 'j':
				i++
				if assem[i] < '0' || assem[i] > '9' {
					panic("invalid destination register number")
				}

				n := assem[i] - '0'
				sb.WriteString(tm.LabelString(jumpLabels[n]))

			case '`':
				sb.WriteByte('`')

			default:
				sb.WriteByte(assem[i])
			}
		}
	}
}
