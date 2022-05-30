package main

import "fmt"

type (
	// Label for memory address (function body's code)
	Label Symbol
	// Temp for registers (local variables and parameters)
	Temp Symbol
)

type TempManagement struct {
	tempCnt  int
	labelCnt int
}

func NewTempManagement() *TempManagement {
	return &TempManagement{
		tempCnt: 0,
	}
}

func (t *TempManagement) NewTemp() Temp {
	t.tempCnt++
	return Temp(strs.Symbol(fmt.Sprintf("t%d", t.tempCnt)))
}

func (t *TempManagement) MakeTempString(v Temp) string {
	return strs.Get(Symbol(v))
}

func (t *TempManagement) NewLabel() Label {
	t.tempCnt++
	return Label(strs.Symbol(fmt.Sprintf("L%d", t.tempCnt)))
}

func (t *TempManagement) LabelString(label Label) string {
	return strs.Get(Symbol(label))
}

func (t *TempManagement) NamedLabel(s string) Label {
	return Label(strs.Symbol(s))
}
