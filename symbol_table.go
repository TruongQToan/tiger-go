package main

import (
	"errors"
	"strings"
)

var (
	errSTNotFound = errors.New("symbol not found in ST")
)

type Symbol = uint64

type Strings struct {
	nextSymbol Symbol
	strings    map[Symbol]string
}

func NewStrings() *Strings {
	return &Strings{
		strings: make(map[Symbol]string),
	}
}

func (s *Strings) Get(sym Symbol) string {
	return s.strings[sym]
}

func (s *Strings) Symbol(str string) Symbol {
	for v, s := range s.strings {
		if strings.EqualFold(s, str) {
			return v
		}
	}

	s.nextSymbol++
	s.strings[s.nextSymbol] = str
	return s.nextSymbol
}

type VarST struct {
	st *BaseST
}

func NewVarST(strings *Strings) *VarST {
	return &VarST{st: NewST(strings)}
}

func (vst *VarST) BeginScope() {
	vst.st.BeginScope()
}

func (vst *VarST) EndScope() {
	vst.st.EndScope()
}

func (vst *VarST) Enter(sym Symbol, data EnvEntry) {
	vst.st.Enter(sym, data)
}

func (vst *VarST) Look(sym Symbol) (EnvEntry, error) {
	v, err := vst.st.Look(sym)
	if err != nil {
		return nil, err
	}

	v1, ok := v.(EnvEntry)
	if !ok {
		panic("expect env entry in variable ST")
	}

	return v1, nil
}

func (vst *VarST) Replace(sym Symbol, data EnvEntry) {
	vst.st.Replace(sym, data)
}

func (vst *VarST) Name(sym Symbol) string {
	return vst.st.Name(sym)
}

type TypeST struct {
	st *BaseST
}

func NewTypeST(strings *Strings) *TypeST {
	return &TypeST{st: NewST(strings)}
}

func (vst *TypeST) BeginScope() {
	vst.st.BeginScope()
}

func (vst *TypeST) EndScope() {
	vst.st.EndScope()
}

func (vst *TypeST) Enter(sym Symbol, data SemantTy) {
	vst.st.Enter(sym, data)
}

func (vst *TypeST) Look(sym Symbol) (SemantTy, error) {
	v, err := vst.st.Look(sym)
	if err != nil {
		return nil, err
	}

	v1, ok := v.(SemantTy)
	if !ok {
		panic("expect semant type in variable ST")
	}

	return v1, nil
}

func (vst *TypeST) Replace(sym Symbol, data SemantTy) {
	vst.st.Replace(sym, data)
}

func (vst *TypeST) Name(sym Symbol) string {
	return vst.st.Name(sym)
}

type BaseST struct {
	stack   [][]Symbol
	strings *Strings
	table   map[Symbol][]interface{}
}

func NewST(strings *Strings) *BaseST {
	st := BaseST{
		strings: strings,
		table:   make(map[Symbol][]interface{}),
	}

	st.BeginScope()
	return &st
}

func (s *BaseST) BeginScope() {
	s.stack = append(s.stack, make([]Symbol, 0, 100))
}

func (s *BaseST) EndScope() {
	if len(s.stack) == 0 {
		panic("call BeginScope() before EndScope()")
	}

	for _, sym := range s.stack[len(s.stack)-1] {
		v, ok := s.table[sym]
		if !ok {
			panic("table does not contain symbol")
		}

		if len(v) == 0 {
			panic("table does not contain values")
		}

		v = v[:len(v)-1]
		if len(v) == 0 {
			delete(s.table, sym)
		}
	}

	s.stack = s.stack[:len(s.stack)-1]
}

func (s *BaseST) Enter(sym Symbol, data interface{}) {
	s.table[sym] = append(s.table[sym], data)
	if len(s.stack) == 0 {
		panic("call BeginScope() before Enter()")
	}

	s.stack[len(s.stack)-1] = append(s.stack[len(s.stack)-1], sym)
}

func (s *BaseST) Look(sym Symbol) (interface{}, error) {
	if len(s.table[sym]) == 0 {
		return nil, errSTNotFound
	}

	return s.table[sym][len(s.table[sym])-1], nil
}

func (s *BaseST) Name(sym Symbol) string {
	return s.strings.strings[sym]
}

func (s *BaseST) Replace(sym Symbol, data interface{}) {
	if _, ok := s.table[sym]; ok {
		s.table[sym] = append(s.table[sym][:len(s.table)-1], data)
	}
}
