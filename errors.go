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
	return fmt.Errorf("undefined error %s at %s", v, pos.String())
}

func expectedVarButFoundFunErr(v string, pos Pos) error {
	return fmt.Errorf("expected %s is a variable but found function %s", v, pos.String())
}