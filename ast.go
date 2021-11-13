package tiger

type Operator int

const (
	And Operator = iota + 1
	Div
	Eq
	Ge
	Gt
	Le
	Lt
	Neq
	Minus
	Or
	Plus
	Mul
)

type OperatorWithPos struct {
	op  Operator
	pos Pos
}

type RecordField struct {
	expr  Exp
	ident Symbol
	pos   Pos
}

type Declaration interface {
	DeclPos() Pos
}

type Field struct {
	name   Symbol
	escape bool
	typ    Symbol
	pos    Pos
}

type FuncDecl struct {
	name     Symbol
	params   []*Field
	resultTy Symbol
	body     Exp
	pos      Pos
}

func (f *FuncDecl) DeclPos() Pos {
	return f.pos
}

type VarDecl struct {
	name   Symbol
	escape bool
	typ    Symbol
	init   Exp
	pos    Pos
}

func (f *VarDecl) DeclPos() Pos {
	return f.pos
}

type TypeDecl struct {
	tyName Symbol
	typ  Ty
	pos  Pos
}

func (f *TypeDecl) DeclPos() Pos {
	return f.pos
}

type Ty interface {
	TyPos() Pos
}

type NameTy struct {
	ty  Symbol
	pos Pos
}

func (t *NameTy) TyPos() Pos {
	return t.pos
}

type RecordTy struct {
	ty  []*Field
	pos Pos
}

func (t *RecordTy) TyPos() Pos {
	return t.pos
}

type ArrayTy struct {
	ty  Symbol
	pos Pos
}

func (t *ArrayTy) TyPos() Pos {
	return t.pos
}

type Exp interface {
	ExpPos() Pos
}

type ArrExp struct {
	init Exp
	size Exp
	typ  Symbol
	pos  Pos
}

func (e *ArrExp) ExpPos() Pos {
	return e.pos
}

type AssignExp struct {
	exp      Exp
	variable Exp
}

func (e *AssignExp) ExpPos() Pos {
	return e.variable.ExpPos()
}

type BreakExp struct {
	pos Pos
}

func (e *BreakExp) ExpPos() Pos {
	return e.pos
}

type CallExp struct {
	function Symbol
	args     []Exp
	pos      Pos
}

func (e *CallExp) ExpPos() Pos {
	return e.pos
}

type IfExp struct {
	predicate Exp
	then      Exp
	els       Exp
	pos       Pos
}

func (e *IfExp) ExpPos() Pos {
	return e.pos
}

type IntExp struct {
	val int64
	pos Pos
}

func (e *IntExp) ExpPos() Pos {
	return e.pos
}

type LetExp struct {
	body  Exp
	decls []Declaration
	pos   Pos
}

func (e *LetExp) ExpPos() Pos {
	return e.pos
}

type NilExp struct {
	pos Pos
}

func (e *NilExp) ExpPos() Pos {
	return e.pos
}

type OperExp struct {
	left  Exp
	op    *OperatorWithPos
	right Exp
}

func (e *OperExp) ExpPos() Pos {
	return e.left.ExpPos()
}

type RecordExp struct {
	fields []*RecordField
	typ    Symbol
	pos    Pos
}

func (e *RecordExp) ExpPos() Pos {
	return e.pos
}

type SequenceExp struct {
	seq []Exp
	pos Pos
}

func (e *SequenceExp) ExpPos() Pos {
	return e.pos
}

type StrExp struct {
	str string
	pos Pos
}

func (e *StrExp) ExpPos() Pos {
	return e.pos
}

type VarExp struct {
	sym Symbol
	pos Pos
}

func (e *VarExp) ExpPos() Pos {
	return e.pos
}

type WhileExp struct {
	pred Exp
	body Exp
	pos  Pos
}

func (e *WhileExp) ExpPos() Pos {
	return e.pos
}
