package main

import (
	"fmt"
)

type SemantTy interface {
	TypeName() string
}

type RecordSemantTy struct {
	symbols []Symbol
	ty      []SemantTy
}
func (t *RecordSemantTy) TypeName() string {
	return "record"
}

func (t *RecordSemantTy) IsEnvEntry() {}

type NilSemantTy struct{}
func (t *NilSemantTy) TypeName() string {
	return "nil"
}

func (t *NilSemantTy) IsEnvEntry() {}

type UnitSemantTy struct{}
func (t *UnitSemantTy) TypeName() string {
	return "unit"
}

func (t *UnitSemantTy) IsEnvEntry() {}

type IntSemantTy struct{}
func (t *IntSemantTy) TypeName() string {
	return "int"
}

func (t *IntSemantTy) IsEnvEntry() {}

type StringSemantTy struct{}
func (t *StringSemantTy) TypeName() string {
	return "string"
}

func (t *StringSemantTy) IsEnvEntry() {}

type ArrSemantTy struct {
	baseTy SemantTy
}
func (t *ArrSemantTy) TypeName() string {
	return fmt.Sprintf("array of %s", t.baseTy.TypeName())
}

func (t *ArrSemantTy) IsEnvEntry() {}

type NameSemantTy struct {
	baseTy SemantTy
	name   Symbol
}
func (t *NameSemantTy) TypeName() string {
	return fmt.Sprintf("type name of %s", t.baseTy.TypeName())
}

func (t *NameSemantTy) IsEnvEntry() {}

func isSameType(ty1, ty2 SemantTy) bool {
	switch v1 := ty1.(type) {
	case *IntSemantTy:
		return isInt(ty2)
	case *StringSemantTy:
		return isString(ty2)
	case *RecordSemantTy:
		return isRecord(ty2)
	case *ArrSemantTy:
		switch v2 := ty2.(type) {
		case *ArrSemantTy:
			return isSameType(v1.baseTy, v2.baseTy)
		default:
			return false
		}
	case *NameSemantTy:
		switch v2 := ty2.(type) {
		case *NameSemantTy:
			return v1.name == v2.name
		default:
			return false
		}
	}

	return false
}

func isInt(ty SemantTy) bool {
	switch ty.(type) {
	case *IntSemantTy:
		return true
	default:
		return false
	}
}

func isString(ty SemantTy) bool {
	switch ty.(type) {
	case *StringSemantTy:
		return true
	default:
		return false
	}
}

func isRecord(ty SemantTy) bool {
	switch ty.(type) {
	case *StringSemantTy:
		return true
	default:
		return false
	}
}
