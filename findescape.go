package main


type FindEscape struct {
	escapeEnv *EscapeST
}

func NewFindEscape(strings *Strings) *FindEscape {
	return &FindEscape{escapeEnv: NewEscapeST(strings)}
}

func (t *FindEscape) transVar(v Var, depth int) {
	switch vt := v.(type) {
	case *SimpleVar:
		entry, err := t.escapeEnv.Look(vt.symbol)
		if err == errSTNotFound {
			return
		}

		*entry.escape = true
		if entry.depth < depth {
			t.escapeEnv.Replace(vt.symbol, entry)
		}
	case *FieldVar:
		t.transVar(vt.variable, depth)
	case *SubscriptionVar:
		t.transVar(vt.variable, depth)
	}
}

func (t *FindEscape) transDecs(decl Declaration, depth int) {
	switch dt := decl.(type) {
	case *FuncDecl:
		for _, param := range dt.params {
			entry := EscapeEntry{
				depth:  depth+1,
				escape: param.escape,
			}
			t.escapeEnv.Enter(param.name, &entry)
		}

	case *VarDecl:
		*dt.escape = false
		t.transExp(dt.init, depth)
		entry := EscapeEntry{
			depth:  depth+1,
			escape: dt.escape,
		}
		t.escapeEnv.Enter(dt.name, &entry)
	case *TypeDecl:
		return
	}
}

func (t *FindEscape) transExp(exp Exp, depth int) {
	switch et := exp.(type) {
	case *VarExp:
		t.transVar(et.v, depth)
	case *NilExp:
		return
	case *IntExp:
		return
	case *StrExp:
		return
	case *CallExp:
		for _, exp := range et.args {
			t.transExp(exp, depth)
		}
	case *OperExp:
		t.transExp(et.left, depth)
		t.transExp(et.right, depth)
	case *RecordExp:
		for _, field := range et.fields {
			t.transExp(field.expr, depth)
		}
	case *SequenceExp:
		for _, exp := range et.exps {
			t.transExp(exp, depth)
		}
	case *AssignExp:
		t.transVar(et.variable, depth)
		t.transExp(et.exp, depth)
	case *IfExp:
		t.transExp(et.predicate, depth)
		t.transExp(et.then, depth)
		if et.els != nil {
			t.transExp(et.els, depth)
		}
	case *WhileExp:
		t.transExp(et.pred, depth)
		t.transExp(et.body, depth)
	case *BreakExp:
		return
	case *LetExp:
		for _, decl := range et.decls {
			t.transDecs(decl, depth)
		}

		t.transExp(et.body, depth)
	case *ArrExp:
		t.transExp(et.size, depth)
		t.transExp(et.init, depth)
	}
}

func (t *FindEscape) FindEscape(prog Exp) {
	t.transExp(prog, 0)
}
