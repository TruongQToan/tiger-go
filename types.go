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

type NilSemantTy struct{}

func (t *NilSemantTy) TypeName() string {
	return "nil"
}

type UnitSemantTy struct{}

func (t *UnitSemantTy) TypeName() string {
	return "unit"
}

type IntSemantTy struct{}

func (t *IntSemantTy) TypeName() string {
	return "int"
}

type StringSemantTy struct{}

func (t *StringSemantTy) TypeName() string {
	return "string"
}

type ArrSemantTy struct {
	baseTy SemantTy
	u      int64
}

func (t *ArrSemantTy) TypeName() string {
	return fmt.Sprintf("array of %s", t.baseTy.TypeName())
}

type NameSemantTy struct {
	baseTy  SemantTy
	nameSym Symbol
	name    string
}

func (t *NameSemantTy) TypeName() string {
	return t.name
}

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
			return v2.u == v1.u
		default:
			return false
		}
	case *NameSemantTy:
		switch v2 := ty2.(type) {
		case *NameSemantTy:
			return v1.nameSym == v2.nameSym
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
	case *RecordSemantTy:
		return true
	default:
		return false
	}
}
