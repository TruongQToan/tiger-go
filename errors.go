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

func invalidNumberOfRecordFieldErr(pos Pos) error {
	return fmt.Errorf("invalid number of record fields when constructing new record at %s", pos.String())
}

func fieldNotFoundErr(field string, recordTy string, pos Pos) error {
	if len(recordTy) > 0 {
		return fmt.Errorf("field %s not found in record %s at %s", field, recordTy, pos.String())
	}

	return fmt.Errorf("field %s not found at %s", field, pos.String())
}

func undefinedVarErr(v string, pos Pos) error {
	return fmt.Errorf("undefined variable %s at %s", v, pos.String())
}

// TODO: change this to "variable not found"
func expectedVarButFoundFunErr(v string, pos Pos) error {
	return fmt.Errorf("expected %s is a variable but found function at %s", v, pos.String())
}

func functionNotFoundErr(f string, pos Pos) error {
	return fmt.Errorf("function not found %s at %s", f, pos.String())
}

func mismatchNumberOfParameters(f string, pos Pos) error {
	return fmt.Errorf("function %s has a different number of params at %s", f, pos.String())
}

func typeNotFoundErr(ty string, pos Pos) error {
	return fmt.Errorf("type %s not found %s", ty, pos.String())
}

func typeMismatchWhenDeclErr(expected, got SemantTy, pos Pos) error {
	return fmt.Errorf("expected type %s, but expression has type %s at %s", expected.TypeName(), got.TypeName(), pos.String())
}

func duplicateRecordDefinition(pos Pos) error {
	return fmt.Errorf("duplicate record definition %s", pos.String())
}

func baseTypeNotFoundErr(base string, pos Pos) error {
	return fmt.Errorf("base type not found %s at %s", base, pos.String())
}
