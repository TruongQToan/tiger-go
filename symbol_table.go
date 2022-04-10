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

type ST struct {
	stack   [][]Symbol
	strings *Strings
	table   map[Symbol][]EnvEntry
}

func NewST(strings *Strings) *ST {
	st := ST{
		strings: strings,
		table:   make(map[Symbol][]EnvEntry),
	}

	st.BeginScope()
	return &st
}

func (s *ST) BeginScope() {
	s.stack = append(s.stack, make([]Symbol, 0, 100))
}

func (s *ST) EndScope() {
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

func (s *ST) Enter(sym Symbol, data EnvEntry) {
	s.table[sym] = append(s.table[sym], data)
	if len(s.stack) == 0 {
		panic("call BeginScope() before Enter()")
	}

	s.stack[len(s.stack)-1] = append(s.stack[len(s.stack)-1], sym)
}

func (s *ST) Look(sym Symbol) (EnvEntry, error) {
	if len(s.table[sym]) == 0 {
		return nil, errSTNotFound
	}

	return s.table[sym][len(s.table[sym])-1], nil
}

func (s *ST) Name(sym Symbol) string {
	return s.strings.strings[sym]
}

func (s *ST) Replace(sym Symbol, data EnvEntry) {
	if _, ok := s.table[sym]; ok {
		s.table[sym] = append(s.table[sym][:len(s.table)-1], data)
	}
}

func (s *ST) Symbol(str string) Symbol {
	for v, s := range s.strings.strings {
		if strings.EqualFold(s, str) {
			return v
		}
	}

	s.strings.nextSymbol++
	s.strings.strings[s.strings.nextSymbol] = str
	return s.strings.nextSymbol
}
