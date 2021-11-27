package tiger

import (
	"fmt"
	"strconv"
	"strings"
)

type Operator int

type String interface {
	String(symbols *Symbols, strBuilder *strings.Builder, level int)
}

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

func (op Operator) String() string {
	switch op {
	case And:
		return "&"
	case Div:
		return "/"
	case Eq:
		return "="
	case Ge:
		return ">="
	case Gt:
		return ">"
	case Le:
		return "<="
	case Lt:
		return "<"
	case Neq:
		return "!="
	case Minus:
		return "-"
	case Or:
		return "|"
	case Plus:
		return "+"
	case Mul:
		return "*"
	default:
		// Must not occur
		panic(fmt.Sprintf("operator %d", op))
	}
}

type OperatorWithPos struct {
	op  Operator
	pos Pos
}

func (op *OperatorWithPos) String(symbols *Symbols, strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString(op.op.String() + "\n")
}

type RecordField struct {
	expr  Exp
	ident Symbol
	pos   Pos
}

func (rec *RecordField) String(symbols *Symbols, strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("RecordField\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Ident\n")
	indent(strBuilder, level+2)
	strBuilder.WriteString(symbols.Name(rec.ident) + "\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Exp\n")
	rec.expr.String(symbols, strBuilder, level+2)
}

type Declaration interface {
	String
	DeclPos() Pos
}

type Field struct {
	name   Symbol
	escape bool
	typ    Symbol
	pos    Pos
}

func (f *Field) String(symbols *Symbols, strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("Field\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Name\n")
	indent(strBuilder, level+2)
	strBuilder.WriteString(symbols.Name(f.name) + "\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Type\n")
	indent(strBuilder, level+2)
	strBuilder.WriteString(symbols.Name(f.typ) + "\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Escape\n")
	indent(strBuilder, level+2)
	strBuilder.WriteString(strconv.FormatBool(f.escape))
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

func (f *FuncDecl) String(symbols *Symbols, strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("FuncDecl\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Name\n")
	indent(strBuilder, level+2)
	strBuilder.WriteString(symbols.Name(f.name) + "\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("ResultTy\n")
	indent(strBuilder, level+2)
	strBuilder.WriteString(symbols.Name(f.resultTy) + "\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Params\n")
	for _, field := range f.params {
		field.String(symbols, strBuilder, level+2)
	}

	indent(strBuilder, level+1)
	strBuilder.WriteString("Body\n")
	f.body.String(symbols, strBuilder, level+2)
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

func (f *VarDecl) String(symbols *Symbols, strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("VarDecl\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Name\n")
	indent(strBuilder, level+2)
	strBuilder.WriteString(symbols.Name(f.name) + "\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Type\n")
	indent(strBuilder, level+2)
	strBuilder.WriteString(symbols.Name(f.typ) + "\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Escape\n")
	indent(strBuilder, level+2)
	strBuilder.WriteString(strconv.FormatBool(f.escape) + "\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Init\n")
	f.init.String(symbols, strBuilder, level+2)
}

type TypeDecl struct {
	tyName Symbol
	typ    Ty
	pos    Pos
}

func (f *TypeDecl) DeclPos() Pos {
	return f.pos
}

func (f *TypeDecl) String(symbols *Symbols, strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("TypeDecl\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("TyName\n")
	indent(strBuilder, level+2)
	strBuilder.WriteString(symbols.Name(f.tyName) + "\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Type\n")
	f.typ.String(symbols, strBuilder, level+2)
}

type Ty interface {
	String
	TyPos() Pos
}

type NameTy struct {
	ty  Symbol
	pos Pos
}

func (t *NameTy) TyPos() Pos {
	return t.pos
}

func (t *NameTy) String(symbols *Symbols, strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("NameTy\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Ty\n")
	indent(strBuilder, level+2)
	strBuilder.WriteString(symbols.Name(t.ty) + "\n")
}

type RecordTy struct {
	ty  []*Field
	pos Pos
}

func (t *RecordTy) TyPos() Pos {
	return t.pos
}

func (t *RecordTy) String(symbols *Symbols, strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("RecordType\n")
	for _, field := range t.ty {
		field.String(symbols, strBuilder, level+1)
	}
}

type ArrayTy struct {
	ty  Symbol
	pos Pos
}

func (t *ArrayTy) TyPos() Pos {
	return t.pos
}

func (t *ArrayTy) String(symbols *Symbols, strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("ArrayType\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString(symbols.Name(t.ty) + "\n")
}

type Exp interface {
	String
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

func (e *ArrExp) String(symbols *Symbols, strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("ArrExp\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Type\n")
	indent(strBuilder, level+2)
	strBuilder.WriteString(symbols.Name(e.typ) + "\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Size\n")
	e.size.String(symbols, strBuilder, level+2)
	indent(strBuilder, level+1)
	strBuilder.WriteString("Init\n")
	e.init.String(symbols, strBuilder, level+2)
}

type AssignExp struct {
	exp      Exp
	variable Exp
}

func (e *AssignExp) String(symbols *Symbols, strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("AssignExp\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Variable\n")
	e.variable.String(symbols, strBuilder, level+2)
	indent(strBuilder, level+1)
	strBuilder.WriteString("Exp\n")
	e.exp.String(symbols, strBuilder, level+2)
}

func (e *AssignExp) ExpPos() Pos {
	return e.variable.ExpPos()
}

type BreakExp struct {
	pos Pos
}

func (e *BreakExp) String(_ *Symbols, strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("Break\n")
}

func (e *BreakExp) ExpPos() Pos {
	return e.pos
}

type CallExp struct {
	function Symbol
	args     []Exp
	pos      Pos
}

func (e *CallExp) String(symbols *Symbols, strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("CallExp\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Function\n")
	indent(strBuilder, level+2)
	strBuilder.WriteString(symbols.Name(e.function) + "\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Args\n")
	for _, exp := range e.args {
		exp.String(symbols, strBuilder, level+2)
	}
}

func (e *CallExp) ExpPos() Pos {
	return e.pos
}

type FieldExp struct {
	firstExp  Exp
	fieldName Symbol
	pos       Pos
}

func (e *FieldExp) String(symbols *Symbols, strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("FieldExp\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("FirstExp\n")
	e.firstExp.String(symbols, strBuilder, level+2)
	strBuilder.WriteString("FieldName\n")
	indent(strBuilder, level+2)
	strBuilder.WriteString(symbols.Name(e.fieldName) + "\n")
}

func (e *FieldExp) ExpPos() Pos {
	return e.pos
}

type IfExp struct {
	predicate Exp
	then      Exp
	els       Exp
	pos       Pos
}

func (e *IfExp) String(symbols *Symbols, strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("IfExp\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Predicate\n")
	e.predicate.String(symbols, strBuilder, level+2)
	indent(strBuilder, level+1)
	strBuilder.WriteString("Then\n")
	e.then.String(symbols, strBuilder, level+2)
	if e.els != nil {
		indent(strBuilder, level+1)
		strBuilder.WriteString("Else\n")
		e.els.String(symbols, strBuilder, level+2)
	}
}

func (e *IfExp) ExpPos() Pos {
	return e.pos
}

type IntExp struct {
	val int64
	pos Pos
}

func (e *IntExp) String(_ *Symbols, strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("Int\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString(strconv.FormatInt(e.val, 10) + "\n")
}

func (e *IntExp) ExpPos() Pos {
	return e.pos
}

type LetExp struct {
	body  Exp
	decls []Declaration
	pos   Pos
}

func (e *LetExp) String(symbols *Symbols, strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("Let\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Decls\n")
	for _, decl := range e.decls {
		decl.String(symbols, strBuilder, level+2)
	}
	indent(strBuilder, level+1)
	strBuilder.WriteString("Body\n")
	e.body.String(symbols, strBuilder, level+2)
}

func (e *LetExp) ExpPos() Pos {
	return e.pos
}

type NilExp struct {
	pos Pos
}

func (e *NilExp) String(_ *Symbols, strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("Nil\n")
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

func (e *OperExp) String(symbols *Symbols, strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("OperExp\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Op\n")
	e.op.String(symbols, strBuilder, level+2)
	indent(strBuilder, level+1)
	strBuilder.WriteString("Left\n")
	e.left.String(symbols, strBuilder, level+2)
	indent(strBuilder, level+1)
	strBuilder.WriteString("Right\n")
	e.right.String(symbols, strBuilder, level+2)
}

type RecordExp struct {
	fields []*RecordField
	typ    Symbol
	pos    Pos
}

func (e *RecordExp) String(symbols *Symbols, strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("RecordExp\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Type\n")
	indent(strBuilder, level+2)
	strBuilder.WriteString(symbols.Name(e.typ) + "\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Fields\n")
	for _, field := range e.fields {
		field.String(symbols, strBuilder, level+2)
	}
}

func (e *RecordExp) ExpPos() Pos {
	return e.pos
}

type SequenceExp struct {
	seq []Exp
	pos Pos
}

func (e *SequenceExp) String(symbols *Symbols, strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("SequenceExp\n")
	for _, field := range e.seq {
		field.String(symbols, strBuilder, level+1)
	}
}

func (e *SequenceExp) ExpPos() Pos {
	return e.pos
}

type StrExp struct {
	str string
	pos Pos
}

func (e *StrExp) String(_ *Symbols, strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("String\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString(fmt.Sprintf("\"%s\"\n", e.str))
}

func (e *StrExp) ExpPos() Pos {
	return e.pos
}

type SubscriptExp struct {
	subscript Exp
	firstExp  Exp
	pos       Pos
}

func (e *SubscriptExp) String(symbols *Symbols, strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("SubscriptExp\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("FirstExp\n")
	e.firstExp.String(symbols, strBuilder, level+2)
	indent(strBuilder, level+1)
	strBuilder.WriteString("Subscript\n")
	e.subscript.String(symbols, strBuilder, level+2)
}

func (e *SubscriptExp) ExpPos() Pos {
	return e.pos
}

type VarExp struct {
	sym Symbol
	pos Pos
}

func (e *VarExp) String(symbols *Symbols, strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("VarExp\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString(symbols.Name(e.sym) + "\n")
}

func (e *VarExp) ExpPos() Pos {
	return e.pos
}

type WhileExp struct {
	pred Exp
	body Exp
	pos  Pos
}

func (e *WhileExp) String(symbols *Symbols, strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("WhileExp\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Predicate\n")
	e.pred.String(symbols, strBuilder, level+2)
	strBuilder.WriteString("Body\n")
	e.body.String(symbols, strBuilder, level+2)
}

func (e *WhileExp) ExpPos() Pos {
	return e.pos
}

func indent(builder *strings.Builder, level int) {
	for i := 0; i < level; i++ {
		builder.WriteString("  ")
	}
}
