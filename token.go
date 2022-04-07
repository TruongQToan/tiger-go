package main

import "fmt"

type Token struct {
	pos   *Pos
	tok   string
	value interface{}
}

func (t *Token) IsEor() bool {
	return t.tok == "eof"
}

func (t *Token) String() string {
	if t.value != nil {
		switch t.value.(type) {
		case string:
			return t.value.(string)
		case int:
			return fmt.Sprintf("%d", t.value.(int))
		}
	}

	return t.tok
}

func NewIdent(s string, pos *Pos) *Token {
	return &Token{
		tok:   "ident",
		value: s,
		pos:   pos,
	}
}

func NewStr(s string, pos *Pos) *Token {
	return &Token{
		tok:   "str",
		value: s,
		pos: pos,
	}
}

func NewInt(i int64, pos *Pos) *Token {
	return &Token{
		tok:   "int",
		value: i,
		pos:   pos,
	}
}

func NewAnd(pos *Pos) *Token {
	return &Token{
		tok: "&",
		pos: pos,
	}
}
func NewAssign(pos *Pos) *Token {
	return &Token{
		tok: ":=",
		pos: pos,
	}
}

func NewArray(pos *Pos) *Token {
	return &Token{
		tok: "array",
		pos: pos,
	}
}

func NewBreak(pos *Pos) *Token {
	return &Token{
		tok: "break",
		pos: pos,
	}
}

func NewClass(pos *Pos) *Token {
	return &Token{
		tok: "class",
		pos: pos,
	}
}

func NewCloseCurly(pos *Pos) *Token {
	return &Token{
		tok: "}",
		pos: pos,
	}
}

func NewCloseParen(pos *Pos) *Token {
	return &Token{
		tok: ")",
		pos: pos,
	}
}

func NewCloseBrac(pos *Pos) *Token {
	return &Token{
		tok: "]",
		pos: pos,
	}
}

func NewColon(pos *Pos) *Token {
	return &Token{
		tok: ":",
		pos: pos,
	}
}

func NewComma(pos *Pos) *Token {
	return &Token{
		tok: ",",
		pos: pos,
	}
}
func NewDo(pos *Pos) *Token {
	return &Token{
		tok: "do",
		pos: pos,
	}
}

func NewDot(pos *Pos) *Token {
	return &Token{
		tok: ".",
		pos: pos,
	}
}

func NewElse(pos *Pos) *Token {
	return &Token{
		tok: "else",
		pos: pos,
	}
}

func NewEnd(pos *Pos) *Token {
	return &Token{
		tok: "end",
		pos: pos,
	}
}

func NewEndOfFile(pos *Pos) *Token {
	return &Token{
		tok: "eof",
		pos: pos,
	}
}

func NewEqual(pos *Pos) *Token {
	return &Token{
		tok: "=",
		pos: pos,
	}
}

func NewExtends(pos *Pos) *Token {
	return &Token{
		tok: "extends",
		pos: pos,
	}
}

func NewFor(pos *Pos) *Token {
	return &Token{
		tok: "for",
		pos: pos,
	}
}

func NewFunction(pos *Pos) *Token {
	return &Token{
		tok: "function",
		pos: pos,
	}
}

func NewGreater(pos *Pos) *Token {
	return &Token{
		tok: ">",
		pos: pos,
	}
}

func NewGreaterOrEqual(pos *Pos) *Token {
	return &Token{
		tok: ">=",
		pos: pos,
	}
}

func NewIf(pos *Pos) *Token {
	return &Token{
		tok: "if",
		pos: pos,
	}
}
func NewIn(pos *Pos) *Token {
	return &Token{
		tok: "in",
		pos: pos,
	}
}

func NewLesser(pos *Pos) *Token {
	return &Token{
		tok: "<",
		pos: pos,
	}
}

func NewLesserOrEqual(pos *Pos) *Token {
	return &Token{
		tok: "<=",
		pos: pos,
	}
}
func NewLet(pos *Pos) *Token {
	return &Token{
		tok: "let",
		pos: pos,
	}
}

func NewMethod(pos *Pos) *Token {
	return &Token{
		tok: "method",
		pos: pos,
	}
}

func NewMinus(pos *Pos) *Token {
	return &Token{
		tok: "-",
		pos: pos,
	}
}

func NewNil(pos *Pos) *Token {
	return &Token{
		tok: "nil",
		pos: pos,
	}
}

func NewNotEqual(pos *Pos) *Token {
	return &Token{
		tok: "!=",
		pos: pos,
	}
}

func NewOf(pos *Pos) *Token {
	return &Token{
		tok: "of",
		pos: pos,
	}
}

func NewOpenCurly(pos *Pos) *Token {
	return &Token{
		tok: "{",
		pos: pos,
	}
}

func NewOpenParen(pos *Pos) *Token {
	return &Token{
		tok: "(",
		pos: pos,
	}
}

func NewOpenBrac(pos *Pos) *Token {
	return &Token{
		tok: "[",
		pos: pos,
	}
}

func NewOr(pos *Pos) *Token {
	return &Token{
		tok: "|",
		pos: pos,
	}
}

func NewPlus(pos *Pos) *Token {
	return &Token{
		tok: "+",
		pos: pos,
	}
}

func NewSemicolon(pos *Pos) *Token {
	return &Token{
		tok: ";",
		pos: pos,
	}
}

func NewDiv(pos *Pos) *Token {
	return &Token{
		tok: "/",
		pos: pos,
	}
}

func NewTimes(pos *Pos) *Token {
	return &Token{
		tok: "*",
		pos: pos,
	}
}

func NewThen(pos *Pos) *Token {
	return &Token{
		tok: "then",
		pos: pos,
	}
}

func NewTo(pos *Pos) *Token {
	return &Token{
		tok: "to",
		pos: pos,
	}
}

func NewType(pos *Pos) *Token {
	return &Token{
		tok: "type",
		pos: pos,
	}
}

func NewVar(pos *Pos) *Token {
	return &Token{
		tok: "var",
		pos: pos,
	}
}

func NewWhile(pos *Pos) *Token {
	return &Token{
		tok: "while",
		pos: pos,
	}
}
