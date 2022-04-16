package main

import (
	"fmt"
	"math/rand"
)

const (
	IgnorePass = iota
	FirstPass
	SecondPass
	ThirdPass
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
	_, ty, err := s.transExp(exp, false)
	if err != nil {
		return err
	}

	fmt.Printf("Parse type %s\n", ty.TypeName())
	return nil
}

// TODO: refine this one
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
		if err != nil {
			if err == errSTNotFound {
				return struct{}{}, nil, undefinedVarErr(s.strings.Get(v.symbol), v.pos)
			}

			return struct{}{}, nil, err
		}

		e, ok := entry.(*VarEntry)
		if !ok {
			return struct{}{}, nil, expectedVarButFoundFunErr(s.strings.Get(v.symbol), v.pos)
		}

		sTy, err := s.actualTy(e.ty, v.pos)
		if err != nil {
			return struct{}{}, nil, err
		}

		return struct{}{}, sTy, nil

	case *FieldVar:
		v1, ok := v.variable.(*SimpleVar)
		if !ok {
			return struct{}{}, nil, unexpectedTokErr(v.pos)
		}

		entry, err := s.venv.Look(v1.symbol)
		if err != nil {
			return struct{}{}, nil, err
		}

		entry1, ok := entry.(*VarEntry)
		if !ok {
			return struct{}{}, nil, expectedVarButFoundFunErr(s.strings.Get(v1.symbol), v1.VarPos())
		}

		recordTy, ok := entry1.ty.(*RecordSemantTy)
		if !ok {
			return struct{}{}, nil, mismatchTypeErr(&RecordSemantTy{}, recordTy, v.pos)
		}

		for i, field := range recordTy.symbols {
			if field == v.field {
				semantTy, err := s.tenv.Look(recordTy.types[i])
				if err != nil {
					if err == errSTNotFound {
						return struct{}{}, nil, typeNotFoundErr(s.strings.Get(recordTy.types[i]), Pos{})
					}

					return struct{}{}, nil, err
				}

				aTy, err := s.actualTy(semantTy, Pos{})
				if err != nil {
					return struct{}{}, nil, err
				}

				return struct{}{}, aTy, nil
			}
		}

		return struct{}{}, nil, fieldNotFoundErr(s.strings.Get(v.field), "", v.pos)

	case *SubscriptionVar:
		v1, ok := v.variable.(*SimpleVar)
		if !ok {
			return struct{}{}, nil, unexpectedTokErr(v.pos)
		}

		// 1. Look if the array has been declared yet?
		entry, err := s.venv.Look(v1.symbol)
		if err != nil {
			return struct{}{}, nil, err
		}

		// 2. Check if the entry is type array or not
		entry1, ok := entry.(*VarEntry)
		if !ok {
			return struct{}{}, nil, expectedVarButFoundFunErr(s.strings.Get(v1.symbol), v1.VarPos())
		}

		arrTy, ok := entry1.ty.(*ArrSemantTy)
		if !ok {
			return struct{}{}, nil, mismatchTypeErr(&ArrSemantTy{}, arrTy, v.pos)
		}

		// 2. Check the expr is int or not
		_, eTy, err := s.transExp(v.exp, false)
		if err != nil {
			return struct{}{}, nil, err
		}

		_, ok = eTy.(*IntSemantTy)
		if !ok {
			return struct{}{}, nil, mismatchTypeErr(&IntSemantTy{}, eTy, v.exp.ExpPos())
		}

		return struct{}{}, arrTy.baseTy, nil
	}

	return struct{}{}, nil, nil
}

// transExp the output SemantTy must be a real type, not an alias type
func (s *Semant) transExp(exp Exp, canHasBreak bool) (TransExp, SemantTy, error) {
	pos := exp.ExpPos()
	switch v := exp.(type) {
	case *OperExp:
		_, leftTy, err := s.transExp(v.left, false)
		if err != nil {
			return struct{}{}, nil, err
		}

		_, rightTy, err := s.transExp(v.right, false)
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
					if _, ok := rightTy.(*NilSemantTy); !ok {
						return struct{}{}, nil, mismatchTypeErr(&RecordSemantTy{}, rightTy, pos)
					}
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

		_, sTy, err := s.transExp(v.size, false)
		if err != nil {
			return struct{}{}, nil, err
		}

		_, ok = sTy.(*IntSemantTy)
		if !ok {
			return struct{}{}, nil, typeMismatchWhenDeclErr(&IntSemantTy{}, sTy, v.size.ExpPos())
		}

		_, iTy, err := s.transExp(v.init, false)
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
		return s.transExp(v.seq[len(v.seq)-1], false)

	case *LetExp:
		// We need to deal with recursive declarations like in this example type intlist = {first: int, rest: intlist}

		// 1. Perform two passes to parse type declarations
		for _, decl := range v.decls {
			if _, ok := decl.(*TypeDecl); ok {
				if err := s.transDec(decl, FirstPass); err != nil {
					return struct{}{}, nil, err
				}
			}
		}

		for _, decl := range v.decls {
			if _, ok := decl.(*TypeDecl); ok {
				if err := s.transDec(decl, SecondPass); err != nil {
					return struct{}{}, nil, err
				}
			}
		}

		// 2. Perform two passes to parse function declarations
		for _, decl := range v.decls {
			if _, ok := decl.(*FuncDecl); ok {
				if err := s.transDec(decl, FirstPass); err != nil {
					return struct{}{}, nil, err
				}
			}
		}

		for _, decl := range v.decls {
			if _, ok := decl.(*FuncDecl); ok {
				if err := s.transDec(decl, SecondPass); err != nil {
					return struct{}{}, nil, err
				}
			}
		}

		// 3. Parse variable declarations
		for _, decl := range v.decls {
			switch decl.(type) {
			case *VarDecl:
				if err := s.transDec(decl, IgnorePass); err != nil {
					return struct{}{}, nil, err
				}
			default:
				continue
			}
		}

		return s.transExp(v.body, false)

	case *AssignExp:
		_, vTy, err := s.transVar(v.variable)
		if err != nil {
			return struct{}{}, nil, err
		}

		actualTy, err := s.actualTy(vTy, v.variable.VarPos())
		if err != nil {
			return struct{}{}, nil, err
		}

		_, eTy, err := s.transExp(v.exp, false)
		if err != nil {
			return struct{}{}, nil, err
		}

		if !isSameType(actualTy, eTy) {
			return struct{}{}, nil, mismatchTypeErr(vTy, eTy, v.exp.ExpPos())
		}

		return struct{}{}, &UnitSemantTy{}, nil

	case *IfExp:
		_, pTy, err := s.transExp(v.predicate, false)
		if err != nil {
			return struct{}{}, nil, err
		}

		_, ok := pTy.(*IntSemantTy)
		if !ok {
			return struct{}{}, nil, mismatchTypeErr(&IntSemantTy{}, pTy, v.predicate.ExpPos())
		}

		_, tTy, err := s.transExp(v.then, false)
		if err != nil {
			return struct{}{}, nil, err
		}

		if v.els != nil {
			_, eTy, err := s.transExp(v.els, false)
			if err != nil {
				return struct{}{}, nil, err
			}

			if !isSameType(tTy, eTy) {
				return struct{}{}, nil, mismatchTypeErr(tTy, eTy, v.els.ExpPos())
			}
		}

		return struct{}{}, tTy, nil

	case *WhileExp:
		_, tTy, err := s.transExp(v.pred, false)
		if err != nil {
			return struct{}{}, nil, err
		}

		_, ok := tTy.(*IntSemantTy)
		if !ok {
			return struct{}{}, nil, mismatchTypeErr(&IntSemantTy{}, tTy, v.pred.ExpPos())
		}

		_, bTy, err := s.transExp(v.body, true)
		if err != nil {
			return struct{}{}, nil, err
		}

		_, ok = bTy.(*UnitSemantTy)
		if !ok {
			return struct{}{}, nil, mismatchTypeErr(&UnitSemantTy{}, bTy, v.body.ExpPos())
		}

		return struct{}{}, &UnitSemantTy{}, nil

	case *CallExp:
		entry, err := s.venv.Look(v.function)
		if err != nil {
			if err == errSTNotFound {
				return struct{}{}, nil, functionNotFoundErr(s.strings.Get(v.function), v.pos)
			}

			return struct{}{}, nil, err
		}

		fEntry, ok := entry.(*FunEntry)
		if !ok {
			return struct{}{}, nil, functionNotFoundErr(s.strings.Get(v.function), v.pos)
		}

		if len(v.args) != len(fEntry.formals) {
			return struct{}{}, nil, mismatchNumberOfParameters(s.strings.Get(v.function), v.pos)
		}

		for i, arg := range v.args {
			_, ty, err := s.transExp(arg, false)
			if err != nil {
				return struct{}{}, nil, err
			}

			actualTy, err := s.actualTy(fEntry.formals[i], Pos{})
			if err != nil {
				return struct{}{}, nil, err
			}

			if !isSameType(actualTy, ty) {
				return struct{}{}, nil, mismatchTypeErr(actualTy, ty, arg.ExpPos())
			}
		}

		return struct{}{}, fEntry.result, nil

	case *BreakExp:
		// TODO: check if break statement is inside while/for
		if !canHasBreak {
			return struct{}{}, nil, breakOutOfScopeErr(v.pos)
		}

		return struct{}{}, &UnitSemantTy{}, nil

	case *RecordExp:
		tTy, err := s.tenv.Look(v.ty)
		if err != nil {
			if err == errSTNotFound {
				return struct{}{}, nil, typeNotFoundErr(s.strings.Get(v.ty), v.pos)
			}

			return struct{}{}, nil, err
		}

		ty, ok := tTy.(*RecordSemantTy)
		if !ok {
			return struct{}{}, nil, mismatchTypeErr(&RecordSemantTy{}, tTy, v.pos)
		}

		if len(v.fields) != len(ty.types) {
			return struct{}{}, nil, invalidNumberOfRecordFieldErr(v.pos)
		}

		for i := range v.fields {
			for j := range v.fields {
				if i != j && v.fields[i].ident == v.fields[j].ident {
					return struct{}{}, nil, duplicateRecordDefinition(v.fields[i].pos)
				}
			}
		}

		for _, field := range v.fields {
			idx := ty.HasField(field.ident)
			if idx == -1 {
				return struct{}{}, nil, fieldNotFoundErr(s.strings.Get(field.ident), s.strings.Get(v.ty), field.pos)
			}

			_, eTy, err := s.transExp(field.expr, false)
			if err != nil {
				return struct{}{}, nil, err
			}

			semantTy, err := s.tenv.Look(ty.types[idx])
			if err != nil {
				// Should not happen
				return struct{}{}, nil, typeNotFoundErr(s.strings.Get(ty.types[idx]), Pos{})
			}

			fTy, err := s.actualTy(semantTy, Pos{})
			if err != nil {
				return struct{}{}, nil, err
			}

			if !isSameType(fTy, eTy) {
				return struct{}{}, nil, mismatchTypeErr(fTy, eTy, field.expr.ExpPos())
			}
		}

		return struct{}{}, ty, nil
	}

	return struct{}{}, nil, nil
}

func (s *Semant) transDec(decl Declaration, pass int) error {
	switch v := decl.(type) {
	case *FuncDecl:
		resultTy, err := s.tenv.Look(v.resultTy)
		if err != nil {
			if err == errSTNotFound {
				return typeNotFoundErr(s.strings.Get(v.resultTy), v.resultTyPos)
			}

			return err
		}

		resultTy, err = s.actualTy(resultTy, v.pos)
		if err != nil {
			return err
		}

		if pass == SecondPass {
			// Only enter VarEntry at second pass. At first pass, we only gather information (type and params)
			s.venv.BeginScope()
			defer s.venv.EndScope()
		}

		paramsTy := make([]SemantTy, 0, len(v.params))
		for _, param := range v.params {
			ty, err := s.tenv.Look(param.typ)
			if err != nil {
				if err == errSTNotFound {
					return typeNotFoundErr(s.strings.Get(param.typ), param.pos)
				}

				return err
			}

			ty, err = s.actualTy(ty, param.pos)
			if err != nil {
				return err
			}

			paramsTy = append(paramsTy, ty)
			if pass == SecondPass {
				// Only enter VarEntry at second pass. At first pass, we only gather information (type and params)
				s.venv.Enter(param.name, &VarEntry{
					ty: ty,
				})
			}
		}

		if pass == FirstPass {
			s.venv.Enter(v.name, &FunEntry{
				formals: paramsTy,
				result:  resultTy,
			})

			return nil
		}

		_, bTy, err := s.transExp(v.body, false)
		if err != nil {
			return err
		}

		if !isSameType(bTy, resultTy) {
			return mismatchTypeErr(resultTy, bTy, v.body.ExpPos())
		}

		s.venv.Replace(v.name, &FunEntry{
			formals: paramsTy,
			result:  resultTy,
		})

		return nil

	case *VarDecl:
		_, initTy, err := s.transExp(v.init, false)
		if err != nil {
			return err
		}

		initTy, err = s.actualTy(initTy, v.init.ExpPos())
		if err != nil {
			return err
		}

		if v.typ == 0 {
			// var id := expr
			switch initTy.(type) {
			case *NilSemantTy:
				return fmt.Errorf("cannot use nil here")
			default:
				s.venv.Enter(v.name, &VarEntry{
					ty: initTy,
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

		if !isSameType(actualTy, initTy) {
			return typeMismatchWhenDeclErr(actualTy, initTy, v.init.ExpPos())
		}

		s.venv.Enter(v.name, &VarEntry{
			ty: actualTy,
		})
		return nil

	case *TypeDecl:
		ty, err := s.transTypeDecl(v.ty, pass)
		if err != nil {
			return err
		}

		if pass == FirstPass {
			s.tenv.Enter(v.tyName, ty)
		} else {
			s.tenv.Replace(v.tyName, ty)
		}

		return nil
	}

	return nil
}

// transTypeDecl translates type expressions as found in the abstract syntax to the digested type descriptions that we
// will put into env
func (s *Semant) transTypeDecl(ty Ty, pass int) (SemantTy, error) {
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

		if pass == FirstPass {
			return &NameSemantTy{
				baseTy: nil,
			}, nil
		}

		types := make([]Symbol, 0, len(v.fields))
		symbols := make([]Symbol, 0, len(v.fields))
		for _, field := range v.fields {
			_, err := s.tenv.Look(field.typ)
			if err != nil {
				if err == errSTNotFound {
					return nil, typeNotFoundErr(s.strings.Get(field.typ), field.pos)
				}
			}

			symbols = append(symbols, field.name)
			types = append(types, field.typ)
		}

		return &RecordSemantTy{
			symbols: symbols,
			types:   types,
			u:       rand.Int63(),
		}, nil
	}

	return nil, nil
}
