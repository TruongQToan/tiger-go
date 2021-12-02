package tiger

import "fmt"

type TreeStm interface {
	IsStm()
}

type SeqStm struct {
	First TreeStm
	Next  TreeStm
}

func (s *SeqStm) IsStm() {}

type LabelStm struct {
	Label Label
}

func (s *LabelStm) IsStm() {}

type JmpStm struct {
	exp    TreeExp
	labels []Label
}

func (s *JmpStm) IsStm() {}

type CJmpStm struct{}

func (s *CJmpStm) IsStm() {}

type MoveStm struct {
	firstExp TreeExp
	rightExp TreeExp
}

func (s *MoveStm) IsStm() {}

type ExpStm struct {
	exp TreeStm
}

func (s *ExpStm) IsStm() {}

type TreeExp interface {
	IsTreeExp()
}

type BinOpTreeExp struct {
	binOp BinOp
	left  TreeExp
	right TreeExp
}

func (e *BinOpTreeExp) IsTreeExp() {}

type MemTreeExp struct {
	exp TreeExp
}

func (e *MemTreeExp) IsTreeExp() {}

type TempTreeExp struct {
	temp Temp
}

func (e *TempTreeExp) IsTreeExp() {}

type EseqTreeExp struct {
	stm TreeStm
	exp TreeExp
}

func (e *EseqTreeExp) IsTreeExp() {}

type NameTreeExp struct {
	label Label
}

func (e *NameTreeExp) IsTreeExp() {}

type ConstTreeExp struct {
	cnst int32
}

func (e *ConstTreeExp) IsTreeExp() {}

type CallTreeExp struct {
	exp     TreeExp
	expList []TreeExp
}

func (e *CallTreeExp) IsTreeExp() {}

type BinOpExp struct {
	binOp    BinOp
	leftExp  TreeExp
	rightExp TreeExp
}

type BinOp int16

const (
	PlusBinOp BinOp = iota + 1
	MinusBinOp
	MulBinOp
	DivBinOp
	AndBinOp
	OrBinOp
	LShiftBinOp
	RShiftBinOp
	ArShiftBinOp
	XorBinOp
)

type ReOp int64

const (
	EqReOp ReOp = iota + 1
	NeReOp
	LtReOp
	GtReOp
	LeReOp
	GeReOp
	UltReOp
	UleReOp
	UgtReOp
	UgeReOp
)

func (op ReOp) NotRel() ReOp {
	switch op {
	case EqReOp:
		return NeReOp
	case NeReOp:
		return EqReOp
	case LtReOp:
		return GeReOp
	case GeReOp:
		return LtReOp
	case GtReOp:
		return LeReOp
	case LeReOp:
		return GtReOp
	case UleReOp:
		return UgtReOp
	case UgtReOp:
		return UleReOp
	case UltReOp:
		return UgeReOp
	case UgeReOp:
		return UltReOp
	default:
		panic(fmt.Sprintf("invalid Re operation %d", op))
	}
}

func (op ReOp) Commute() ReOp {
	switch op {
	case EqReOp:
		return EqReOp
	case NeReOp:
		return NeReOp
	case LtReOp:
		return GtReOp
	case GeReOp:
		return LeReOp
	case LeReOp:
		return GeReOp
	case GtReOp:
		return LtReOp
	case UltReOp:
		return UgtReOp
	case UleReOp:
		return UgeReOp
	case UgtReOp:
		return UltReOp
	case UgeReOp:
		return UleReOp
	default:
		panic(fmt.Sprintf("invalid Re operation %d", op))
	}
}
