package main

import (
	"fmt"
)

type SemantTy interface {
	TypeName() string
}

type RecordSemantTy struct {
	symbols []Symbol
	types   []Symbol
	u       int64
}

func (t *RecordSemantTy) TypeName() string {
	return "record"
}

func (t *RecordSemantTy) HasField(field Symbol) int {
	for i, sym := range t.symbols {
		if sym == field {
			return i
		}
	}

	return -1
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
	if t.baseTy != nil {
		return fmt.Sprintf("array of %s", t.baseTy.TypeName())
	}

	return "array"
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
	case *UnitSemantTy:
		return isUnit(ty2)

	case *NilSemantTy:
		// TODO: is this correct?
		return isRecord(ty2)

	case *IntSemantTy:
		return isInt(ty2)

	case *StringSemantTy:
		return isString(ty2)

	case *RecordSemantTy:
		switch v2 := ty2.(type) {
		case *NilSemantTy:
			return true
		case *RecordSemantTy:
			return v1.u == v2.u
		default:
			return false
		}

	case *ArrSemantTy:
		switch v2 := ty2.(type) {
		case *ArrSemantTy:
			return v2.u == v1.u && isSameType(v1.baseTy, v2.baseTy)
		default:
			return false
		}

	case *NameSemantTy:
		switch v2 := ty2.(type) {
		case *NameSemantTy:
			return isSameType(v1.baseTy, v2.baseTy)
		default:
			return isSameType(v1.baseTy, v2)
		}
	}

	return false
}

func isInt(ty SemantTy) bool {
	switch v := ty.(type) {
	case *IntSemantTy:
		return true
	case *NameSemantTy:
		return isInt(v.baseTy)
	default:
		return false
	}
}

func isString(ty SemantTy) bool {
	switch v := ty.(type) {
	case *StringSemantTy:
		return true
	case *NameSemantTy:
		return isString(v.baseTy)
	default:
		return false
	}
}

func isRecord(ty SemantTy) bool {
	switch v := ty.(type) {
	case *RecordSemantTy:
		return true
	case *NilSemantTy:
		return true
	case *NameSemantTy:
		return isRecord(v.baseTy)
	default:
		return false
	}
}

func isUnit(ty SemantTy) bool {
	switch ty.(type) {
	case *UnitSemantTy:
		return true
	default:
		return false
	}
}
