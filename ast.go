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
	DeclPos()
}

type Field struct {
	name   Symbol
	escape bool
	typ    Symbol
	pos    Pos
}

type FuncDecl struct {
	decls []struct {
		name   Symbol
		params []*Field
		result *SymbolWithPos
		body   Exp
		pos    Pos
	}
	pos Pos
}

func (f *FuncDecl) DeclPos() Pos {
	return f.pos
}

type VarDecl struct {
	name   SymbolWithPos
	escape bool
	typ    SymbolWithPos
	init   Exp
}

func (f *VarDecl) DeclPos() Pos {
	return f.name.pos
}

type TypeDecl struct {
	typ  Ty
	name SymbolWithPos
}

func (f *TypeDecl) DeclPos() Pos {
	return f.name.pos
}

type Ty interface {
	TyPos()
}

type NameTy struct {
	ty SymbolWithPos
}

func (t *NameTy) TyPos() Pos {
	return t.ty.pos
}

type RecordTy struct {
	ty  []*Field
	pos Pos
}

func (t *RecordTy) TyPos() Pos {
	return t.pos
}

type ArrayTy struct {
	ty SymbolWithPos
}

func (t *ArrayTy) TyPos() Pos {
	return t.ty.pos
}

type Exp interface {
	ExpPos() Pos
}

type ArrExp struct {
	init Exp
	size Exp
	typ  SymbolWithPos
}

func (e *ArrExp) ExpPos() Pos {
	return e.typ.pos
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
	function SymbolWithPos
	args     []Exp
}

func (e *CallExp) ExpPos() Pos {
	return e.function.pos
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
	op    OperatorWithPos
	right Exp
}

func (e *OperExp) ExpPos() Pos {
	return e.left.ExpPos()
}

type RecordExp struct {
	fields []RecordField
	typ    SymbolWithPos
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

type Str struct {
	str string
	pos Pos
}

func (e *Str) ExpPos() Pos {
	return e.pos
}

type Var struct {
	sym SymbolWithPos
}

func (e *Var) ExpPos() Pos {
	return e.sym.pos
}

type While struct {
	pred Exp
	body Exp
	pos  Pos
}

func (e *While) ExpPos() Pos {
	return e.pos
}
