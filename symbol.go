package tiger

import "strings"

type Symbol = uint64

type Strings struct {
	nextSymbol Symbol
	strings    map[Symbol]string
}

func NewStrings() *Strings {
	return &Strings{
		nextSymbol: 0,
		strings:    make(map[Symbol]string),
	}
}

func (s *Strings) Get(sym Symbol) string {
	return s.strings[sym]
}

type Symbols struct {
	stack   [][]Symbol
	strings Strings
	table   map[Symbol][]interface{}
}

func NewSymbols(strings Strings) *Symbols {
	symbols := Symbols{
		strings: strings,
		table:   make(map[Symbol][]interface{}),
	}

	symbols.BeginScope()
	return &symbols
}

func (s *Symbols) BeginScope() {
	s.stack = append(s.stack, make([]Symbol, 0))
}

func (s *Symbols) EndScope() {
	if len(s.stack) == 0 {
		panic("call BeginScope() before EndScope()")
	}

	for _, sym := range s.stack[len(s.stack)-1] {
		v, ok := s.table[sym]
		if !ok {
			panic("table does not contain symbol")
		}

		v = v[:len(v)-1]
		if len(v) == 0 {
			delete(s.table, sym)
		}
	}

	s.stack = s.stack[:len(s.stack)-1]
}

func (s *Symbols) Enter(sym Symbol, data interface{}) {
	s.table[sym] = append(s.table[sym], data)
	if len(s.stack) == 0 {
		panic("call BeginScope() before Enter()")
	}

	s.stack[len(s.stack)-1] = append(s.stack[len(s.stack)-1], sym)
}

func (s *Symbols) Look(sym Symbol) interface{} {
	return s.table[sym]
}

func (s *Symbols) Name(sym Symbol) string {
	return s.strings.strings[sym]
}

func (s *Symbols) Replace(sym Symbol, data interface{}) {
	if _, ok := s.table[sym]; ok {
		s.table[sym] = append(s.table[sym][:len(s.table)-1], data)
	}
}

func (s *Symbols) Symbol(str string) Symbol {
	for v, s := range s.strings.strings {
		if strings.EqualFold(s, str) {
			return v
		}
	}

	s.strings.nextSymbol++
	s.strings.strings[s.strings.nextSymbol] = str
	return s.strings.nextSymbol
}

type SymbolWithPos struct {
	sym Symbol
	pos Pos
}
