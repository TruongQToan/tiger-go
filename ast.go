package main

import (
	"fmt"
	"strconv"
	"strings"
)

type String interface {
	String(strBuilder *strings.Builder, level int)
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

type Operator int

func (op Operator) IsArith() bool {
	return op == Plus || op == Minus || op == Mul || op == Div
}

func (op Operator) IsComp() bool {
	return op == Lt || op == Le || op == Gt || op == Ge
}

func (op Operator) IsEq() bool {
	return op == Eq || op == Neq
}

func (op Operator) Repr() string {
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

func (op Operator) String(_ *Strings, strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString(op.Repr() + "\n")
}

type RecordField struct {
	expr  Exp
	ident Symbol
	pos   Pos
}

func (rec *RecordField) String(strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("RecordField\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Ident\n")
	indent(strBuilder, level+2)
	strBuilder.WriteString(strs.Get(rec.ident) + "\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Exp\n")
	rec.expr.String(strBuilder, level+2)
}

type Declaration interface {
	String
	DeclPos() Pos
}

type Field struct {
	name   Symbol
	escape *bool
	typ    Symbol
	pos    Pos
}

func (f *Field) String(strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("Field\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Name\n")
	indent(strBuilder, level+2)
	strBuilder.WriteString(strs.Get(f.name) + "\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Type\n")
	indent(strBuilder, level+2)
	strBuilder.WriteString(strs.Get(f.typ) + "\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Escape\n")
	indent(strBuilder, level+2)
	strBuilder.WriteString(strconv.FormatBool(*f.escape) + "\n")
}

type FuncDecl struct {
	name        Symbol
	params      []*Field
	resultTy    Symbol
	resultTyPos Pos
	body        Exp
	pos         Pos
}

func (f *FuncDecl) DeclPos() Pos {
	return f.pos
}

func (f *FuncDecl) String(strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("FuncDecl\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Name\n")
	indent(strBuilder, level+2)
	strBuilder.WriteString(strs.Get(f.name) + "\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("ResultTy\n")
	indent(strBuilder, level+2)
	strBuilder.WriteString(strs.Get(f.resultTy) + "\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Params\n")
	for _, field := range f.params {
		field.String(strBuilder, level+2)
	}

	indent(strBuilder, level+1)
	strBuilder.WriteString("Body\n")
	f.body.String(strBuilder, level+2)
}

type VarDecl struct {
	name   Symbol
	escape *bool
	typ    Symbol
	init   Exp
	pos    Pos
}

func (f *VarDecl) DeclPos() Pos {
	return f.pos
}

func (f *VarDecl) String(strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("VarDecl\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Name\n")
	indent(strBuilder, level+2)
	strBuilder.WriteString(strs.Get(f.name) + "\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Type\n")
	indent(strBuilder, level+2)
	strBuilder.WriteString(strs.Get(f.typ) + "\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Escape\n")
	indent(strBuilder, level+2)
	strBuilder.WriteString(strconv.FormatBool(*f.escape) + "\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Init\n")
	f.init.String(strBuilder, level+2)
}

type TypeDecl struct {
	tyName Symbol
	ty     Ty
	pos    Pos
}

func (f *TypeDecl) DeclPos() Pos {
	return f.pos
}

func (f *TypeDecl) String(strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("TypeDecl\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("T.Get\n")
	indent(strBuilder, level+2)
	strBuilder.WriteString(strs.Get(f.tyName) + "\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Type\n")
	f.ty.String(strBuilder, level+2)
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

func (t *NameTy) String(strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("NameTy\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Ty\n")
	indent(strBuilder, level+2)
	strBuilder.WriteString(strs.Get(t.ty) + "\n")
}

type RecordTy struct {
	fields []*Field
	pos    Pos
}

func (t *RecordTy) HasDuplicateField() bool {
	for i := range t.fields {
		for j := range t.fields {
			if i != j && t.fields[i].name == t.fields[j].name {
				return true
			}
		}
	}

	return false
}

func (t *RecordTy) TyPos() Pos {
	return t.pos
}

func (t *RecordTy) String(strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("RecordType\n")
	for _, field := range t.fields {
		field.String(strBuilder, level+1)
	}
}

type ArrayTy struct {
	ty  Symbol
	pos Pos
}

func (t *ArrayTy) TyPos() Pos {
	return t.pos
}

func (t *ArrayTy) String(strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("ArrayType\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString(strs.Get(t.ty) + "\n")
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

func (e *ArrExp) String(strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("ArrExp\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Type\n")
	indent(strBuilder, level+2)
	strBuilder.WriteString(strs.Get(e.typ) + "\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Size\n")
	e.size.String(strBuilder, level+2)
	indent(strBuilder, level+1)
	strBuilder.WriteString("Init\n")
	e.init.String(strBuilder, level+2)
}

type AssignExp struct {
	exp      Exp
	variable Var
}

func (e *AssignExp) String(strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("AssignExp\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Variable\n")
	e.variable.String(strBuilder, level+2)
	indent(strBuilder, level+1)
	strBuilder.WriteString("Exp\n")
	e.exp.String(strBuilder, level+2)
}

func (e *AssignExp) ExpPos() Pos {
	return e.variable.VarPos()
}

type BreakExp struct {
	pos Pos
}

func (e *BreakExp) String(strBuilder *strings.Builder, level int) {
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

func (e *CallExp) String(strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("CallExp\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Function\n")
	indent(strBuilder, level+2)
	strBuilder.WriteString(strs.Get(e.function) + "\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Args\n")
	for _, exp := range e.args {
		exp.String(strBuilder, level+2)
	}
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

func (e *IfExp) String(strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("IfExp\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Predicate\n")
	e.predicate.String(strBuilder, level+2)
	indent(strBuilder, level+1)
	strBuilder.WriteString("Then\n")
	e.then.String(strBuilder, level+2)
	if e.els != nil {
		indent(strBuilder, level+1)
		strBuilder.WriteString("Else\n")
		e.els.String(strBuilder, level+2)
	}
}

func (e *IfExp) ExpPos() Pos {
	return e.pos
}

type IntExp struct {
	val int32
	pos Pos
}

func (e *IntExp) String(strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("Int\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString(strconv.FormatInt(int64(e.val), 10) + "\n")
}

func (e *IntExp) ExpPos() Pos {
	return e.pos
}

type LetExp struct {
	body  Exp
	decls []Declaration
	pos   Pos
}

func (e *LetExp) String(strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("Let\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Decls\n")
	for _, decl := range e.decls {
		decl.String(strBuilder, level+2)
	}
	indent(strBuilder, level+1)
	strBuilder.WriteString("Body\n")
	e.body.String(strBuilder, level+2)
}

func (e *LetExp) ExpPos() Pos {
	return e.pos
}

type UnitExp struct {
	pos Pos
}

func (e *UnitExp) String(strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("Unit\n")
}

func (e *UnitExp) ExpPos() Pos {
	return e.pos
}

type NilExp struct {
	pos Pos
}

func (e *NilExp) String(strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("Nil\n")
}

func (e *NilExp) ExpPos() Pos {
	return e.pos
}

type OperExp struct {
	left  Exp
	op    Operator
	right Exp
}

func (e *OperExp) ExpPos() Pos {
	return e.left.ExpPos()
}

func (e *OperExp) String(strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("OperExp\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Op\n")
	e.op.String(strs, strBuilder, level+2)
	indent(strBuilder, level+1)
	strBuilder.WriteString("Left\n")
	e.left.String(strBuilder, level+2)
	indent(strBuilder, level+1)
	strBuilder.WriteString("Right\n")
	e.right.String(strBuilder, level+2)
}

type RecordExp struct {
	fields []*RecordField
	ty     Symbol
	pos    Pos
}

func (e *RecordExp) String(strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("RecordExp\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Type\n")
	indent(strBuilder, level+2)
	strBuilder.WriteString(strs.Get(e.ty) + "\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Fields\n")
	for _, field := range e.fields {
		field.String(strBuilder, level+2)
	}
}

func (e *RecordExp) ExpPos() Pos {
	return e.pos
}

type SequenceExp struct {
	exps []Exp
	pos  Pos
}

func (e *SequenceExp) String(strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("SequenceExp\n")
	for _, field := range e.exps {
		field.String(strBuilder, level+1)
	}
}

func (e *SequenceExp) ExpPos() Pos {
	return e.pos
}

type StrExp struct {
	str string
	pos Pos
}

func (e *StrExp) String(strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("String\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString(fmt.Sprintf("\"%s\"\n", e.str))
}

func (e *StrExp) ExpPos() Pos {
	return e.pos
}

type VarExp struct {
	v Var
}

func (e *VarExp) String(strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("VarExp\n")
	e.v.String(strBuilder, level+1)
}

func (e *VarExp) ExpPos() Pos {
	return e.v.VarPos()
}

type ForExp struct {
	from Exp
	to   Exp
	body Exp
	pos  Pos
	sym  Symbol
}

func (e *ForExp) String(strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("ForExp\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("ItVar\n")
	indent(strBuilder, level+2)
	strBuilder.WriteString(strs.Get(e.sym) + "\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("From\n")
	e.from.String(strBuilder, level+2)
	indent(strBuilder, level+1)
	strBuilder.WriteString("To\n")
	e.to.String(strBuilder, level+2)
	indent(strBuilder, level+1)
	strBuilder.WriteString("Body\n")
	e.body.String(strBuilder, level+2)
}

func (e *ForExp) ExpPos() Pos {
	return e.pos
}

type WhileExp struct {
	pred Exp
	body Exp
	pos  Pos
}

func (e *WhileExp) String(strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("WhileExp\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Predicate\n")
	e.pred.String(strBuilder, level+2)
	indent(strBuilder, level+1)
	strBuilder.WriteString("Body\n")
	e.body.String(strBuilder, level+2)
}

func (e *WhileExp) ExpPos() Pos {
	return e.pos
}

func indent(builder *strings.Builder, level int) {
	for i := 0; i < level; i++ {
		builder.WriteString("  ")
	}
}

type Var interface {
	String
	VarPos() Pos
}

type SimpleVar struct {
	symbol Symbol
	pos    Pos
}

func (v *SimpleVar) String(strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("SimpleVar\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString(fmt.Sprintf("%s\n", strs.Get(v.symbol)))
}

func (v *SimpleVar) VarPos() Pos {
	return v.pos
}

type FieldVar struct {
	variable Var
	field    Symbol
	pos      Pos
}

func (v *FieldVar) String(strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("FieldVar\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Var\n")
	v.variable.String(strBuilder, level+2)
	indent(strBuilder, level+1)
	strBuilder.WriteString("Field\n")
	indent(strBuilder, level+2)
	strBuilder.WriteString(fmt.Sprintf("%s\n", strs.Get(v.field)))
}

func (v *FieldVar) VarPos() Pos {
	return v.pos
}

type SubscriptionVar struct {
	variable Var
	exp      Exp
	pos      Pos
}

func (v *SubscriptionVar) String(strBuilder *strings.Builder, level int) {
	indent(strBuilder, level)
	strBuilder.WriteString("SubscriptionVar\n")
	indent(strBuilder, level+1)
	strBuilder.WriteString("Var\n")
	v.variable.String(strBuilder, level+2)
	indent(strBuilder, level+1)
	strBuilder.WriteString("Subscript\n")
	v.exp.String(strBuilder, level+2)
}

func (v *SubscriptionVar) VarPos() Pos {
	return v.pos
}
