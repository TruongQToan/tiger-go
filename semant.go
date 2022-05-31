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
	venv      *VarST
	tenv      *TypeST
	translate *Translate
}

func NewSemant(trans *Translate, vent *VarST, tenv *TypeST) *Semant {
	return &Semant{
		venv:      vent,
		tenv:      tenv,
		translate: trans,
	}
}

func (s *Semant) TransProg(exp Exp) (TransExp, error) {
	mainLevel := Level{
		parent: OutermostLevel,
		frame:  NewMipsFrame(tm.NamedLabel("main"), []bool{true}),
		u: rand.Int63(),
	}

	progExp, ty, err := s.transExp(&mainLevel, exp, tm.NewLabel())
	if err != nil {
		return nil, err
	}

	fmt.Printf("Parse type %s\n", ty.TypeName())
	return progExp, nil
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

func (s *Semant) transVar(level *Level, variable Var, breakLabel Label) (TransExp, SemantTy, error) {
	switch v := variable.(type) {
	case *SimpleVar:
		entry, err := s.venv.Look(v.symbol)
		if err != nil {
			if err == errSTNotFound {
				return nil, nil, undefinedVarErr(strs.Get(v.symbol), v.pos)
			}

			return nil, nil, err
		}

		e, ok := entry.(*VarEntry)
		if !ok {
			return nil, nil, expectedVarButFoundFunErr(strs.Get(v.symbol), v.pos)
		}

		sTy, err := s.actualTy(e.ty, v.pos)
		if err != nil {
			return nil, nil, err
		}

		return s.translate.simpleVar(level, e.access), sTy, nil

	case *FieldVar:
		e1, ty1, err := s.transVar(level, v.variable, breakLabel)
		if err != nil {
			return nil, nil, err
		}

		recordTy, ok := ty1.(*RecordSemantTy)
		if !ok {
			return nil, nil, mismatchTypeErr(&RecordSemantTy{}, recordTy, v.pos)
		}

		for i, field := range recordTy.symbols {
			if field == v.field {
				semantTy, err := s.tenv.Look(recordTy.types[i])
				if err != nil {
					if err == errSTNotFound {
						return nil, nil, typeNotFoundErr(strs.Get(recordTy.types[i]), Pos{})
					}

					return nil, nil, err
				}

				aTy, err := s.actualTy(semantTy, Pos{})
				if err != nil {
					return nil, nil, err
				}

				return s.translate.fieldVar(e1, int32(i)), aTy, nil
			}
		}

		return nil, nil, fieldNotFoundErr(strs.Get(v.field), "", v.pos)

	case *SubscriptionVar:
		ve, ty, err := s.transVar(level, v.variable, breakLabel)
		if err != nil {
			return nil, nil, err
		}

		arrTy, ok := ty.(*ArrSemantTy)
		if !ok {
			return nil, nil, mismatchTypeErr(&ArrSemantTy{}, arrTy, v.pos)
		}

		// 2. Check the expr is int or not
		se, eTy, err := s.transExp(level, v.exp, breakLabel)
		if err != nil {
			return nil, nil, err
		}

		_, ok = eTy.(*IntSemantTy)
		if !ok {
			return nil, nil, mismatchTypeErr(&IntSemantTy{}, eTy, v.exp.ExpPos())
		}

		return s.translate.SubscriptVar(ve, se), arrTy.baseTy, nil
	}

	panic("invalid type")
}

// transExp the output SemantTy must be a real type, not an alias type
func (s *Semant) transExp(level *Level, exp Exp, breakLabel Label) (TransExp, SemantTy, error) {
	pos := exp.ExpPos()
	switch v := exp.(type) {
	case *OperExp:
		le, leftTy, err := s.transExp(level, v.left, breakLabel)
		if err != nil {
			return nil, nil, err
		}

		re, rightTy, err := s.transExp(level, v.right, breakLabel)
		if err != nil {
			return nil, nil, err
		}

		if v.op.IsArith() {
			if !isInt(leftTy) {
				return nil, nil, mismatchTypeErr(&IntSemantTy{}, leftTy, v.left.ExpPos())
			}

			if !isInt(rightTy) {
				return nil, nil, mismatchTypeErr(&IntSemantTy{}, rightTy, v.right.ExpPos())
			}

			return s.translate.BinOp(v.op, le, re), &IntSemantTy{}, nil
		}

		if v.op.IsEq() {
			switch v1 := leftTy.(type) {
			case *IntSemantTy:
				if !isInt(rightTy) {
					return nil, nil, mismatchTypeErr(&IntSemantTy{}, rightTy, pos)
				}

				return s.translate.RelOp(v.op, le, re), &IntSemantTy{}, nil
			case *StringSemantTy:
				if !isString(rightTy) {
					return nil, nil, mismatchTypeErr(&StringSemantTy{}, rightTy, pos)
				}

				return s.translate.RelOp(v.op, le, re), &IntSemantTy{}, nil
			case *RecordSemantTy:
				if !isRecord(rightTy) {
					if _, ok := rightTy.(*NilSemantTy); !ok {
						return nil, nil, mismatchTypeErr(&RecordSemantTy{}, rightTy, pos)
					}
				}

				return s.translate.RelOp(v.op, le, re), &IntSemantTy{}, nil
			case *ArrSemantTy:
				switch v2 := rightTy.(type) {
				case *ArrSemantTy:
					if !isSameType(v1.baseTy, v2.baseTy) {
						return nil, nil, mismatchTypeErr(leftTy, rightTy, pos)
					}

					return s.translate.RelOp(v.op, le, re), &IntSemantTy{}, nil
				default:
					return s.translate.RelOp(v.op, le, re), nil, mismatchTypeErr(leftTy, rightTy, pos)
				}
			default:
				return s.translate.RelOp(v.op, le, re), nil, fmt.Errorf("expect type int, string, record, arr; found %s", leftTy.TypeName())
			}
		}

		if v.op.IsComp() {
			switch leftTy.(type) {
			case *IntSemantTy:
				if !isInt(rightTy) {
					return nil, nil, mismatchTypeErr(&IntSemantTy{}, rightTy, pos)
				}

				return s.translate.RelOp(v.op, le, re), &IntSemantTy{}, nil
			case *StringSemantTy:
				if !isString(rightTy) {
					return nil, nil, mismatchTypeErr(&StringSemantTy{}, rightTy, pos)
				}

				return s.translate.RelOp(v.op, le, re), &IntSemantTy{}, nil
			default:
				return nil, nil, fmt.Errorf("expect type int, string; found %s", leftTy.TypeName())
			}
		}
	case *StrExp:
		return s.translate.strExp(v.str), &StringSemantTy{}, nil

	case *IntExp:
		return s.translate.intExp(v.val), &IntSemantTy{}, nil

	case *NilExp:
		return s.translate.nilExp(), &NilSemantTy{}, nil

	case *VarExp:
		return s.transVar(level, v.v, breakLabel)

	case *ArrExp:
		ty, err := s.tenv.Look(v.typ)
		if err != nil {
			if err == errSTNotFound {
				return nil, nil, typeNotFoundErr(ty.TypeName(), v.ExpPos())
			}

			return nil, nil, err
		}

		aty, ok := ty.(*ArrSemantTy)
		if !ok {
			return nil, nil, typeMismatchWhenDeclErr(&ArrSemantTy{}, ty, v.ExpPos())
		}

		sEx, sTy, err := s.transExp(level, v.size, breakLabel)
		if err != nil {
			return nil, nil, err
		}

		_, ok = sTy.(*IntSemantTy)
		if !ok {
			return nil, nil, typeMismatchWhenDeclErr(&IntSemantTy{}, sTy, v.size.ExpPos())
		}

		iEx, iTy, err := s.transExp(level, v.init, breakLabel)
		if err != nil {
			return nil, nil, err
		}

		if !isSameType(iTy, aty.baseTy) {
			return nil, nil, typeMismatchWhenDeclErr(aty.baseTy, iTy, v.init.ExpPos())
		}

		return s.translate.arrayExp(sEx, iEx), &ArrSemantTy{
			baseTy: aty.baseTy,
			u:      rand.Int63(),
		}, nil

	case *SequenceExp:
		return s.transSeq(level, v.exps, breakLabel)

	case *LetExp:
		// We need to deal with recursive declarations like in this example type intlist = {first: int, rest: intlist}

		// 1. Perform two passes to parse type declarations
		for _, decl := range v.decls {
			if _, ok := decl.(*TypeDecl); ok {
				if _, err := s.transDec(level, decl, FirstPass, breakLabel); err != nil {
					return nil, nil, err
				}
			}
		}

		for _, decl := range v.decls {
			if _, ok := decl.(*TypeDecl); ok {
				if _, err := s.transDec(level, decl, SecondPass, breakLabel); err != nil {
					return nil, nil, err
				}
			}
		}

		// 2. Perform two passes to parse function declarations
		for _, decl := range v.decls {
			if _, ok := decl.(*FuncDecl); ok {
				if _, err := s.transDec(level, decl, FirstPass, breakLabel); err != nil {
					return nil, nil, err
				}
			}
		}

		for _, decl := range v.decls {
			if _, ok := decl.(*FuncDecl); ok {
				if _, err := s.transDec(level, decl, SecondPass, breakLabel); err != nil {
					return nil, nil, err
				}
			}
		}

		vExps := make([]TransExp, 0)
		// 3. Parse variable declarations
		for _, decl := range v.decls {
			switch decl.(type) {
			case *VarDecl:
				exp, err := s.transDec(level, decl, IgnorePass, breakLabel)
				if err != nil {
					return nil, nil, err
				}

				vExps = append(vExps, exp)
			default:
				continue
			}
		}

		bodyExp, ty, err := s.transExp(level, v.body, breakLabel)
		if err != nil {
			return nil, nil, err
		}

		return s.translate.letExp(vExps, bodyExp), ty, nil

	case *AssignExp:
		vex, vTy, err := s.transVar(level, v.variable, breakLabel)
		if err != nil {
			return nil, nil, err
		}

		actualTy, err := s.actualTy(vTy, v.variable.VarPos())
		if err != nil {
			return nil, nil, err
		}

		eex, eTy, err := s.transExp(level, v.exp, breakLabel)
		if err != nil {
			return nil, nil, err
		}

		if !isSameType(actualTy, eTy) {
			return nil, nil, mismatchTypeErr(vTy, eTy, v.exp.ExpPos())
		}

		return s.translate.assign(vex, eex), &UnitSemantTy{}, nil

	case *IfExp:
		ifEx, pTy, err := s.transExp(level, v.predicate, breakLabel)
		if err != nil {
			return nil, nil, err
		}

		_, ok := pTy.(*IntSemantTy)
		if !ok {
			return nil, nil, mismatchTypeErr(&IntSemantTy{}, pTy, v.predicate.ExpPos())
		}

		thenEx, tTy, err := s.transExp(level, v.then, breakLabel)
		if err != nil {
			return nil, nil, err
		}

		var elseEx TransExp
		if v.els != nil {
			var eTy SemantTy
			elseEx, eTy, err = s.transExp(level, v.els, breakLabel)
			if err != nil {
				return nil, nil, err
			}

			if !isSameType(tTy, eTy) {
				return nil, nil, mismatchTypeErr(tTy, eTy, v.els.ExpPos())
			}
		} else {
			if _, ok := tTy.(*UnitSemantTy); !ok {
				return nil, nil, mismatchTypeErr(tTy, &UnitSemantTy{}, v.then.ExpPos())
			}
		}

		return s.translate.ifElse(ifEx, thenEx, elseEx), tTy, nil

	case *WhileExp:
		pex, tTy, err := s.transExp(level, v.pred, breakLabel)
		if err != nil {
			return nil, nil, err
		}

		_, ok := tTy.(*IntSemantTy)
		if !ok {
			return nil, nil, mismatchTypeErr(&IntSemantTy{}, tTy, v.pred.ExpPos())
		}

		doneLabel := tm.NewLabel()
		bex, bTy, err := s.transExp(level, v.body, doneLabel)
		if err != nil {
			return nil, nil, err
		}

		_, ok = bTy.(*UnitSemantTy)
		if !ok {
			return nil, nil, mismatchTypeErr(&UnitSemantTy{}, bTy, v.body.ExpPos())
		}

		return s.translate.whileLoop(pex, bex, doneLabel), &UnitSemantTy{}, nil

	case *CallExp:
		entry, err := s.venv.Look(v.function)
		if err != nil {
			if err == errSTNotFound {
				return nil, nil, functionNotFoundErr(strs.Get(v.function), v.pos)
			}

			return nil, nil, err
		}

		fEntry, ok := entry.(*FunEntry)
		if !ok {
			return nil, nil, functionNotFoundErr(strs.Get(v.function), v.pos)
		}

		if len(v.args) != len(fEntry.formals) {
			return nil, nil, mismatchNumberOfParameters(strs.Get(v.function), v.pos)
		}

		args := make([]TransExp, 0, len(v.args))
		for i, arg := range v.args {
			exp, ty, err := s.transExp(level, arg, breakLabel)
			if err != nil {
				return nil, nil, err
			}

			actualTy, err := s.actualTy(fEntry.formals[i], Pos{})
			if err != nil {
				return nil, nil, err
			}

			if !isSameType(actualTy, ty) {
				return nil, nil, mismatchTypeErr(actualTy, ty, arg.ExpPos())
			}

			args = append(args, exp)
		}

		if _, ok := fEntry.result.(*NilSemantTy); ok {
			return s.translate.call(level, fEntry.level.parent, fEntry.label, args, true), fEntry.result, nil
		}

		return s.translate.call(level, fEntry.level.parent, fEntry.label, args, false), fEntry.result, nil

	case *BreakExp:
		return s.translate.breakStm(breakLabel), &UnitSemantTy{}, nil

	case *RecordExp:
		tTy, err := s.tenv.Look(v.ty)
		if err != nil {
			if err == errSTNotFound {
				return nil, nil, typeNotFoundErr(strs.Get(v.ty), v.pos)
			}

			return nil, nil, err
		}

		ty, ok := tTy.(*RecordSemantTy)
		if !ok {
			return nil, nil, mismatchTypeErr(&RecordSemantTy{}, tTy, v.pos)
		}

		if len(v.fields) != len(ty.types) {
			return nil, nil, invalidNumberOfRecordFieldErr(v.pos)
		}

		for i := range v.fields {
			for j := range v.fields {
				if i != j && v.fields[i].ident == v.fields[j].ident {
					return nil, nil, duplicateRecordDefinition(v.fields[i].pos)
				}
			}
		}

		exps := make([]TransExp, 0, len(v.fields))
		for _, field := range v.fields {
			idx := ty.HasField(field.ident)
			if idx == -1 {
				return nil, nil, fieldNotFoundErr(strs.Get(field.ident), strs.Get(v.ty), field.pos)
			}

			fe, eTy, err := s.transExp(level, field.expr, breakLabel)
			if err != nil {
				return nil, nil, err
			}

			semantTy, err := s.tenv.Look(ty.types[idx])
			if err != nil {
				// Should not happen
				return nil, nil, typeNotFoundErr(strs.Get(ty.types[idx]), Pos{})
			}

			fTy, err := s.actualTy(semantTy, Pos{})
			if err != nil {
				return nil, nil, err
			}

			if !isSameType(fTy, eTy) {
				return nil, nil, mismatchTypeErr(fTy, eTy, field.expr.ExpPos())
			}

			exps = append(exps, fe)
		}

		return s.translate.record(exps), ty, nil
	}

	panic("unexpected expression type")
}

func (s *Semant) transDec(level *Level, decl Declaration, pass int, breakLabel Label) (TransExp, error) {
	switch v := decl.(type) {
	case *FuncDecl:
		resultTy, err := s.tenv.Look(v.resultTy)
		if err != nil {
			if err == errSTNotFound {
				return nil, typeNotFoundErr(strs.Get(v.resultTy), v.resultTyPos)
			}

			return nil, err
		}

		resultTy, err = s.actualTy(resultTy, v.pos)
		if err != nil {
			return nil, err
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
					return nil, typeNotFoundErr(strs.Get(param.typ), param.pos)
				}

				return nil, err
			}

			ty, err = s.actualTy(ty, param.pos)
			if err != nil {
				return nil, err
			}

			paramsTy = append(paramsTy, ty)
			if pass == SecondPass {
				// Only enter VarEntry at second pass. At first pass, we only gather information (type and params)
				s.venv.Enter(param.name, &VarEntry{
					ty: ty,
				})
			}
		}

		es := make([]bool, 0, 1 + len(v.params))
		es = append(es, true)
		for _, p := range v.params {
			es = append(es, *p.escape)
		}

		newLevel := s.translate.NewLevel(level, Label(v.name), es)
		if pass == FirstPass {
			s.venv.Enter(v.name, &FunEntry{
				formals: paramsTy,
				result:  resultTy,
				label:   Label(v.name),
				level:   newLevel,
			})

			return nil, nil
		}

		accesses := s.translate.Formals(newLevel)
		for i, param := range v.params {
			tmp, _ := s.venv.Look(param.name)
			oldEntry := tmp.(*VarEntry)
			oldEntry.access = accesses[i]
			s.venv.Replace(param.name, oldEntry)
		}

		bodyExp, bTy, err := s.transExp(newLevel, v.body, breakLabel)
		if err != nil {
			return nil, err
		}

		if !isSameType(bTy, resultTy) {
			return nil, mismatchTypeErr(resultTy, bTy, v.body.ExpPos())
		}

		//s.venv.Replace(v.name, &FunEntry{
		//	formals: paramsTy,
		//	result:  resultTy,
		//	label:   Label(v.name),
		//	level:   newLevel,
		//})
		//
		ProcEntryExit(newLevel, bodyExp)
		return nil, nil

	case *VarDecl:
		initExp, initTy, err := s.transExp(level, v.init, breakLabel)
		if err != nil {
			return nil, err
		}

		initTy, err = s.actualTy(initTy, v.init.ExpPos())
		if err != nil {
			return nil, err
		}

		if v.typ == 0 {
			// var id := expr
			switch initTy.(type) {
			case *NilSemantTy:
				return nil, fmt.Errorf("cannot use nil here")
			default:
				acc := s.translate.AllocLocal(level, *v.escape)
				varExp := s.translate.simpleVar(level, acc)
				s.venv.Enter(v.name, &VarEntry{
					ty:     initTy,
					access: acc,
				})
				return s.translate.assign(varExp, initExp), nil
			}
		}

		tentry, err := s.tenv.Look(v.typ)
		if err != nil {
			if err == errSTNotFound {
				return nil, typeNotFoundErr(s.tenv.Name(v.typ), v.pos)
			}

			return nil, err
		}

		tentry, ok := tentry.(SemantTy)
		if !ok {
			panic("entry of type ST must be SemantTy")
		}

		actualTy, err := s.actualTy(tentry, v.pos)
		if err != nil {
			return nil, err
		}

		if !isSameType(actualTy, initTy) {
			return nil, typeMismatchWhenDeclErr(actualTy, initTy, v.init.ExpPos())
		}

		acc := s.translate.AllocLocal(level, *v.escape)
		varExp := s.translate.simpleVar(level, acc)
		s.venv.Enter(v.name, &VarEntry{
			ty:     actualTy,
			access: acc,
		})
		return s.translate.assign(varExp, initExp), nil

	case *TypeDecl:
		ty, err := s.transTypeDecl(v.ty, pass)
		if err != nil {
			return nil, err
		}

		if pass == FirstPass {
			s.tenv.Enter(v.tyName, ty)
		} else {
			s.tenv.Replace(v.tyName, ty)
		}

		return nil, nil
	}

	panic("unexpected declaration type")
}

// transTypeDecl translates type expressions as found in the abstract syntax to the digested type descriptions that we
// will put into env
func (s *Semant) transTypeDecl(ty Ty, pass int) (SemantTy, error) {
	switch v := ty.(type) {
	case *NameTy:
		baseTy, err := s.tenv.Look(v.ty)
		if err != nil {
			if err == errSTNotFound {
				return nil, baseTypeNotFoundErr(strs.Get(v.ty), v.TyPos())
			}

			return nil, err
		}

		return &NameSemantTy{
			baseTy:  baseTy,
			nameSym: v.ty,
			name:    strs.Get(v.ty),
		}, nil

	case *ArrayTy:
		baseTy, err := s.tenv.Look(v.ty)
		if err != nil {
			if err == errSTNotFound {
				return nil, baseTypeNotFoundErr(strs.Get(v.ty), v.TyPos())
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
					return nil, typeNotFoundErr(strs.Get(field.typ), field.pos)
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

func (s *Semant) transSeq(level *Level, exps []Exp, breakLabel Label) (TransExp, SemantTy, error) {
	if len(exps) == 0 {
		return &Ex{&ConstExpIr{0}}, &IntSemantTy{}, nil
	}

	if len(exps) == 1 {
		return s.transExp(level, exps[0], breakLabel)
	}

	hex, _, err := s.transExp(level, exps[0], breakLabel)
	if err != nil {
		return nil, nil, err
	}

	tex, tly, err := s.transSeq(level, exps[1:], breakLabel)
	if err != nil {
		return nil, nil, err
	}

	return s.translate.seq(hex, tex), tly, nil
}
