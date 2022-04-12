package main

import "fmt"

// Parser errors
func unexpectedTokErr(pos Pos) error {
	return fmt.Errorf("unexpected token %+v", pos)
}

func unexpectedEofErr(pos Pos) error {
	return fmt.Errorf("unexpected end of file %+v", pos)}

// Semantic errors
func mismatchTypeErr(expected, actual SemantTy, pos Pos) error {
	return fmt.Errorf("expected %s, but found %s at %s", expected.TypeName(), actual.TypeName(), pos.String())
}

func undefinedVarErr(v string, pos Pos) error {
	return fmt.Errorf("undefined variable %s at %s", v, pos.String())
}

func expectedVarButFoundFunErr(v string, pos Pos) error {
	return fmt.Errorf("expected %s is a variable but found function %s", v, pos.String())
}

func typeNotFoundErr(ty string, pos Pos) error {
	return fmt.Errorf("type %s not found %s", ty, pos.String())
}

func typeMismatchWhenDeclErr(expected, got SemantTy, pos Pos) error {
	return fmt.Errorf("expected type %s, but expression has type %s at %s", expected.TypeName(), got.TypeName(), pos.String())
}

func baseTypeNotFoundErr(base string, pos Pos) error {
	return fmt.Errorf("base type not found %s at %s", base, pos.String())
}
