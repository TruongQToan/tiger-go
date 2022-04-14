package main

import (
	"fmt"
	"math/rand"
)

type Semant struct {
	venv    *VarST
	tenv    *TypeST
	strings *Strings
}

func NewSemant(strings *Strings, vent *VarST, tenv *TypeST) *Semant {
	return &Semant{
		strings: strings,
		venv:    vent,
		tenv:    tenv,
	}
}

func (s *Semant) TransProg(exp Exp) error {
	_, ty, err := s.transExp(exp)
	if err != nil {
		return err
	}

	fmt.Printf("Parse type %s\n", ty.TypeName())
	return nil
}

func (s *Semant) actualTy(ty SemantTy, pos Pos) (SemantTy, error) {
	switch v := ty.(type) {
	case *NameSemantTy:
		if v.baseTy == nil {
			return nil, typeNotFoundErr(v.name, pos)
		}

		return s.actualTy(v.baseTy, pos)
	case *ArrSemantTy:
		baseTy, err := s.actualTy(v.baseTy, pos)
		if err != nil {
			return nil, err
		}

		return &ArrSemantTy{baseTy: baseTy, u: rand.Int63()}, nil
	default:
		return ty, nil
	}
}

func (s *Semant) transVar(variable Var) (TransExp, SemantTy, error) {
	switch v := variable.(type) {
	case *SimpleVar:
		entry, err := s.venv.Look(v.symbol)
		if err == errSTNotFound {
			return struct{}{}, nil, undefinedVarErr(s.strings.Get(v.symbol), v.VarPos())
		}

		if e, ok := entry.(*VarEntry); ok {
			return struct{}{}, e.ty, nil
		} else {
			return struct{}{}, nil, expectedVarButFoundFunErr(s.strings.Get(v.symbol), v.VarPos())
		}
	}

	// TODO: update this
	return struct{}{}, nil, nil
}

func (s *Semant) transExp(exp Exp) (TransExp, SemantTy, error) {
	pos := exp.ExpPos()
	switch v := exp.(type) {
	case *OperExp:
		_, leftTy, err := s.transExp(v.left)
		if err != nil {
			return struct{}{}, nil, err
		}

		_, rightTy, err := s.transExp(v.right)
		if err != nil {
			return struct{}{}, nil, err
		}

		if v.op.IsArith() {
			if !isInt(leftTy) {
				return struct{}{}, nil, mismatchTypeErr(&IntSemantTy{}, leftTy, v.left.ExpPos())
			}

			if !isInt(rightTy) {
				return struct{}{}, nil, mismatchTypeErr(&IntSemantTy{}, rightTy, v.right.ExpPos())
			}

			return struct{}{}, &IntSemantTy{}, nil
		}

		if v.op.IsEq() {
			switch v1 := leftTy.(type) {
			case *IntSemantTy:
				if !isInt(rightTy) {
					return struct{}{}, nil, mismatchTypeErr(&IntSemantTy{}, rightTy, pos)
				}

				return struct{}{}, &IntSemantTy{}, nil
			case *StringSemantTy:
				if !isString(rightTy) {
					return struct{}{}, nil, mismatchTypeErr(&StringSemantTy{}, rightTy, pos)
				}

				return struct{}{}, &IntSemantTy{}, nil
			case *RecordSemantTy:
				if !isRecord(rightTy) {
					return struct{}{}, nil, mismatchTypeErr(&RecordSemantTy{}, rightTy, pos)
				}

				return struct{}{}, &IntSemantTy{}, nil
			case *ArrSemantTy:
				switch v2 := rightTy.(type) {
				case *ArrSemantTy:
					if !isSameType(v1.baseTy, v2.baseTy) {
						return struct{}{}, nil, mismatchTypeErr(leftTy, rightTy, pos)
					}

					return struct{}{}, &IntSemantTy{}, nil
				default:
					return struct{}{}, nil, mismatchTypeErr(leftTy, rightTy, pos)
				}
			default:
				return struct{}{}, nil, fmt.Errorf("expect type int, string, record, arr; found %s", leftTy.TypeName())
			}
		}

		if v.op.IsComp() {
			switch leftTy.(type) {
			case *IntSemantTy:
				if !isInt(rightTy) {
					return struct{}{}, nil, mismatchTypeErr(&IntSemantTy{}, rightTy, pos)
				}

				return struct{}{}, &IntSemantTy{}, nil
			case *StringSemantTy:
				if !isString(rightTy) {
					return struct{}{}, nil, mismatchTypeErr(&StringSemantTy{}, rightTy, pos)
				}

				return struct{}{}, &IntSemantTy{}, nil
			default:
				return struct{}{}, nil, fmt.Errorf("expect type int, string; found %s", leftTy.TypeName())
			}
		}
	case *StrExp:
		return struct{}{}, &StringSemantTy{}, nil

	case *IntExp:
		return struct{}{}, &IntSemantTy{}, nil

	case *NilExp:
		return struct{}{}, &NilSemantTy{}, nil

	case *VarExp:
		return s.transVar(v.v)

	case *ArrExp:
		ty, err := s.tenv.Look(v.typ)
		if err != nil {
			if err == errSTNotFound {
				return struct{}{}, nil, typeNotFoundErr(ty.TypeName(), v.ExpPos())
			}

			return struct{}{}, nil, err
		}

		aty, ok := ty.(*ArrSemantTy)
		if !ok {
			return struct{}{}, nil, typeMismatchWhenDeclErr(&ArrSemantTy{}, ty, v.ExpPos())
		}

		_, sTy, err := s.transExp(v.size)
		if err != nil {
			return struct{}{}, nil, err
		}

		_, ok = sTy.(*IntSemantTy)
		if !ok {
			return struct{}{}, nil, typeMismatchWhenDeclErr(&IntSemantTy{}, sTy, v.size.ExpPos())
		}

		_, iTy, err := s.transExp(v.init)
		if err != nil {
			return struct{}{}, nil, err
		}

		if !isSameType(iTy, aty.baseTy) {
			return struct{}{}, nil, typeMismatchWhenDeclErr(aty.baseTy, iTy, v.init.ExpPos())
		}

		return struct{}{}, &ArrSemantTy{
			baseTy: aty.baseTy,
			u:      rand.Int63(),
		}, nil
	case *SequenceExp:
		return s.transExp(v.seq[len(v.seq)-1])

	case *LetExp:
		for _, decl := range v.decls {
			if err := s.transDec(decl); err != nil {
				return struct{}{}, nil, err
			}
		}

		return s.transExp(v.body)
	}

	return struct{}{}, nil, nil
}

func (s *Semant) transDec(decl Declaration) error {
	switch v := decl.(type) {
	case *VarDecl:
		_, ty, err := s.transExp(v.init)
		if err != nil {
			return err
		}

		if v.typ == 0 {
			// var id := expr
			switch ty.(type) {
			case *NilSemantTy:
				return fmt.Errorf("cannot use nil here")
			default:
				s.venv.Enter(v.name, &VarEntry{
					ty: ty,
				})
				return nil
			}
		}

		tentry, err := s.tenv.Look(v.typ)
		if err != nil {
			if err == errSTNotFound {
				return typeNotFoundErr(s.tenv.Name(v.typ), v.pos)
			}

			return err
		}

		tentry, ok := tentry.(SemantTy)
		if !ok {
			panic("entry of type ST must be SemantTy")
		}

		actualTy, err := s.actualTy(tentry, v.pos)
		if err != nil {
			return err
		}

		if !isSameType(actualTy, ty) {
			return typeMismatchWhenDeclErr(actualTy, ty, v.init.ExpPos())
		}

		s.venv.Enter(v.name, &VarEntry{
			ty: ty,
		})
		return nil
	case *TypeDecl:
		ty, err := s.transTy(v.ty)
		if err != nil {
			return err
		}

		s.tenv.Enter(v.tyName, ty)
		return nil
	}

	panic("NOT IMPLEMENTED")
}

// transTy translates type expressions as found in the abstract syntax to the digested type descriptions that we
// will put into env
func (s *Semant) transTy(ty Ty) (SemantTy, error) {
	switch v := ty.(type) {
	case *NameTy:
		baseTy, err := s.tenv.Look(v.ty)
		if err != nil {
			if err == errSTNotFound {
				return nil, baseTypeNotFoundErr(s.strings.Get(v.ty), v.TyPos())
			}

			return nil, err
		}

		return &NameSemantTy{
			baseTy:  baseTy,
			nameSym: v.ty,
			name:    s.strings.Get(v.ty),
		}, nil
	case *ArrayTy:
		baseTy, err := s.tenv.Look(v.ty)
		if err != nil {
			if err == errSTNotFound {
				return nil, baseTypeNotFoundErr(s.strings.Get(v.ty), v.TyPos())
			}

			return nil, err
		}

		return &ArrSemantTy{
			baseTy: baseTy,
			u:      rand.Int63(),
		}, nil
	case *RecordTy:
		if v.HasDuplicateField() {
			return nil, duplicateRecordDefinition(v.TyPos())
		}

		return &RecordSemantTy{
			symbols: nil,
			ty:      nil,
		}, nil
	}

	panic("NOT IMPLEMENTED")
}
