package tiger

type FuncInfo struct {
	name  string
	args  []SemantType
	resTy SemantType
}

var baseFuncs  = []*FuncInfo{
	{
		name:  "print",
		args:  []SemantType{&StringSemantType{}},
		resTy: &UnitSemantType{},
	},
	{
		name:  "printi",
		args:  []SemantType{&IntSemantType{}},
		resTy: &UnitSemantType{},
	},
	{
		name:  "flush",
		args:  []SemantType{},
		resTy: &UnitSemantType{},
	},
	{
		name:  "getchar",
		args:  []SemantType{},
		resTy: &StringSemantType{},
	},
	{
		name:  "ord",
		args:  []SemantType{&StringSemantType{}},
		resTy: &IntSemantType{},
	},
	{
		name:  "chr",
		args:  []SemantType{&IntSemantType{}},
		resTy: &StringSemantType{},
	},
	{
		name:  "size",
		args:  []SemantType{&StringSemantType{}},
		resTy: &IntSemantType{},
	},
	{
		name:  "substring",
		args:  []SemantType{&StringSemantType{}, &IntSemantType{}, &IntSemantType{}},
		resTy: &IntSemantType{},
	},
	{
		name:  "concat",
		args:  []SemantType{&StringSemantType{}, &StringSemantType{}},
		resTy: &StringSemantType{},
	},
	{
		name:  "not",
		args:  []SemantType{&IntSemantType{}},
		resTy: &IntSemantType{},
	},
	{
		name:  "exit",
		args:  []SemantType{&IntSemantType{}},
		resTy: &UnitSemantType{},
	},
}

type EnvEntry interface {
	IsEnvEntry()
}

type VarEntry struct {
	access Access
	ty     SemantType
}

func (v *VarEntry) IsEnvEntry() {}

type FunEntry struct {
	level   Level
	label   Label
	formals []SemantType
	Result  SemantType
}

func InitBaseTypeEnv(strings *Strings) *Symbols {
	symbols := NewSymbols(strings)
	symbols.Enter(symbols.Symbol("int"), IntSemantType{})
	symbols.Enter(symbols.Symbol("string"), StringSemantType{})
	return symbols
}

func InitBaseVenv(strings *Strings) *Symbols {
	symbols := NewSymbols(strings)
	for _, finfo := range baseFuncs {
		symbols.Enter(symbols.Symbol(finfo.name), &FunEntry{
			level:   ChildLevel{
				parent: nil,
				frame:  nil,
			},
			label:   0,
			formals: nil,
			Result:  nil,
		})
	}
}
