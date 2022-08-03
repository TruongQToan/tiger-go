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

type TempSet map[Temp]struct{}

func InitTempSet(temps ...Temp) TempSet {
	ts := make(TempSet)
	for _, t := range temps {
		ts.Add(t)
	}

	return ts
}

func (s TempSet) Split() (Temp, TempSet) {
	for k := range s {
		delete(s, k)
		return k, s
	}

	panic("cannot go to there")
}

func (s TempSet) Remove(temp Temp) {
	if _, ok := s[temp]; !ok {
		return
	}

	delete(s, temp)
}

func (s TempSet) GetOneTemp() Temp {
	var t Temp
	for k := range s {
		t = k
		break
	}

	return t
}

func (s TempSet) Add(temp Temp) {
	s[temp] = struct{}{}
}

func (s TempSet) Diff(s1 TempSet) TempSet {
	diff := make(TempSet)
	for k := range s {
		if _, ok := s1[k]; !ok {
			diff[k] = struct{}{}
		}
	}

	return diff
}

func (s TempSet) Intersect(s1 TempSet) TempSet {
	intersect := make(TempSet)
	for k := range s {
		if _, ok := s1[k]; ok {
			intersect[k] = struct{}{}
		}
	}

	return intersect
}

func (s TempSet) Union(s1 TempSet) TempSet {
	union := make(TempSet)
	for k := range s {
		union[k] = struct{}{}
	}

	for k := range s1 {
		union[k] = struct{}{}
	}

	return union
}

func (s TempSet) Empty() bool {
	return len(s) == 0
}

func (s TempSet) Equal(s1 TempSet) bool {
	return s.Diff(s1).Empty() && s1.Diff(s).Empty()
}

func (s TempSet) Clone() TempSet {
	clone := make(TempSet)
	for k := range s {
		clone[k] = struct{}{}
	}

	return clone
}

func (s TempSet) Has(tmp Temp) bool {
	_, ok := s[tmp]
	return ok
}
