package main

type FuncInfo struct {
	name  string
	args  []SemantTy
	resTy SemantTy
}

var baseFuncs = []*FuncInfo{
	{
		name:  "print",
		args:  []SemantTy{&StringSemantTy{}},
		resTy: &UnitSemantTy{},
	},
	{
		name:  "printi",
		args:  []SemantTy{&IntSemantTy{}},
		resTy: &UnitSemantTy{},
	},
	{
		name:  "flush",
		args:  []SemantTy{},
		resTy: &UnitSemantTy{},
	},
	{
		name:  "getchar",
		args:  []SemantTy{},
		resTy: &StringSemantTy{},
	},
	{
		name:  "ord",
		args:  []SemantTy{&StringSemantTy{}},
		resTy: &IntSemantTy{},
	},
	{
		name:  "chr",
		args:  []SemantTy{&IntSemantTy{}},
		resTy: &StringSemantTy{},
	},
	{
		name:  "size",
		args:  []SemantTy{&StringSemantTy{}},
		resTy: &IntSemantTy{},
	},
	{
		name:  "substring",
		args:  []SemantTy{&StringSemantTy{}, &IntSemantTy{}, &IntSemantTy{}},
		resTy: &IntSemantTy{},
	},
	{
		name:  "concat",
		args:  []SemantTy{&StringSemantTy{}, &StringSemantTy{}},
		resTy: &StringSemantTy{},
	},
	{
		name:  "not",
		args:  []SemantTy{&IntSemantTy{}},
		resTy: &IntSemantTy{},
	},
	{
		name:  "exit",
		args:  []SemantTy{&IntSemantTy{}},
		resTy: &UnitSemantTy{},
	},
}

type EnvEntry interface {
	IsEnvEntry()
}

type VarEntry struct {
	ty SemantTy
}

func (v *VarEntry) IsEnvEntry() {}

type FunEntry struct {
	formals []SemantTy
	Result  SemantTy
}

func (v *FunEntry) IsEnvEntry() {}

func InitBaseTypeEnv(strings *Strings) *ST {
	symbols := NewST(strings)
	symbols.Enter(symbols.Symbol("int"), &IntSemantTy{})
	symbols.Enter(symbols.Symbol("string"), &StringSemantTy{})
	return symbols
}

func InitBaseVenv(strings *Strings) *ST {
	symbols := NewST(strings)
	for _, finfo := range baseFuncs {
		symbols.Enter(symbols.Symbol(finfo.name), &FunEntry{
			formals: finfo.args,
			Result:  finfo.resTy,
		})
	}

	return symbols
}
