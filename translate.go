package main

import "math/rand"

var (
	frags []Frag
)

type cxFunc func(tl, fl Label) StmIr

type TransExp interface {
	unEx() ExpIr
	unNx() StmIr
	unCx() cxFunc
}

func isNop(e TransExp) bool {
	if _, ok := e.(*Nx); ok {
		return isNullStm(e.unNx())
	}

	return false
}

type Ex struct {
	exp ExpIr
}

func (e *Ex) unEx() ExpIr {
	return e.exp
}

func (e *Ex) unNx() StmIr {
	return &ExpStmIr{exp: e.exp}
}

func (e *Ex) unCx() cxFunc {
	switch v := e.exp.(type) {
	case *ConstExpIr:
		if v.c == 0 {
			return func(_, fl Label) StmIr {
				return &JumpStmIr{
					exp:    &NameExpIr{label: fl},
					labels: []Label{fl},
				}
			}
		} else if v.c == 1 {
			return func(tl, _ Label) StmIr {
				return &JumpStmIr{
					exp:    &NameExpIr{label: tl},
					labels: []Label{tl},
				}
			}
		}

		return func(tl, fl Label) StmIr {
			return &CJumpStmIr{
				relop:      EqIr,
				left:       v,
				right:      &ConstExpIr{0},
				trueLabel:  fl,
				falseLabel: tl,
			}
		}
	default:
		return func(tl, fl Label) StmIr {
			return &CJumpStmIr{
				relop:      EqIr,
				left:       v,
				right:      &ConstExpIr{0},
				trueLabel:  fl,
				falseLabel: tl,
			}
		}
	}
}

type Nx struct {
	stm StmIr
}

func (s *Nx) unEx() ExpIr {
	return &EsEqExpIr{
		stm: s.stm,
		exp: &ConstExpIr{0},
	}
}

func (s *Nx) unNx() StmIr {
	return s.stm
}

func (s *Nx) unCx() cxFunc {
	panic("not supported")
}

type Cx struct {
	cx cxFunc
}

func (c *Cx) unCx() cxFunc {
	return c.cx
}

func (c *Cx) unEx() ExpIr {
	r, t, f := tm.NewTemp(), tm.NewLabel(), tm.NewLabel()
	return &EsEqExpIr{
		stm: &SeqStmIr{
			first: &MoveStmIr{
				dst: &TempExpIr{r},
				src: &ConstExpIr{1},
			},
			second: &SeqStmIr{
				first: c.cx(t, f),
				second: &SeqStmIr{
					first: &LabelStmIr{f},
					second: &SeqStmIr{
						first: &MoveStmIr{
							dst: &TempExpIr{r},
							src: &ConstExpIr{0},
						},
						second: &LabelStmIr{t},
					},
				},
			},
		},
		exp: &TempExpIr{r},
	}
}

func (c *Cx) unNx() StmIr {
	l := tm.NewLabel()
	return &SeqStmIr{
		first:  c.cx(l, l),
		second: &LabelStmIr{l},
	}
}

type TranslateAccess struct {
	level  *Level
	access FrameAccess
}

type Level struct {
	// top level have nil as parent
	parent *Level
	frame  Frame
	u      int64
}

var OutermostLevel = &Level{
	parent: nil,
	u:      rand.Int63(),
}

func (l *Level) depth() int32 {
	if l.parent == nil {
		return 0
	}

	return 1 + l.parent.depth()
}

func (l *Level) staticLink(from *Level, base ExpIr) ExpIr {
	if from.u == l.u {
		return base
	}

	return l.staticLink(from.parent, from.frame.Formals()[0].exp(base))
}

type Translate struct {
	frameFactory FrameFactoryFunc
}

func (t *Translate) NewLevel(parent *Level, name Label, formals []bool) *Level {
	return &Level{
		parent: parent,
		frame:  t.frameFactory(name, formals),
		u:      rand.Int63(),
	}
}

func (t *Translate) Formals(level *Level) []*TranslateAccess {
	if level.parent == nil {
		// outtermost level
		return nil
	}

	// first formal is static chain offset
	frameAccesses := level.frame.Formals()[1:]
	translateAccesses := make([]*TranslateAccess, 0, len(frameAccesses))
	for _, acc := range frameAccesses {
		translateAccesses = append(translateAccesses, &TranslateAccess{
			level:  level,
			access: acc,
		})
	}

	return translateAccesses
}

func (t *Translate) AllocLocal(level *Level, escape bool) *TranslateAccess {
	if level.parent == nil {
		return nil
	}

	return &TranslateAccess{
		level:  level,
		access: level.frame.AllocLocal(escape),
	}
}

func (t *Translate) simpleVar(level *Level, access *TranslateAccess) TransExp {
	curLevel, defLevel := level, access.level
	var acc ExpIr
	acc = &TempExpIr{fp}

	for curLevel.u != defLevel.u {
		staticLink := curLevel.frame.Formals()[0]
		curLevel = curLevel.parent
		acc = staticLink.exp(acc)
	}

	return &Ex{exp: access.access.exp(acc)}
}

func (t *Translate) memPlus(e1, e2 ExpIr) ExpIr {
	return &MemExpIr{&BinOpExpIr{
		binop: PlusIr,
		left:  e1,
		right: e2,
	}}
}

func (t *Translate) fieldVar(base TransExp, id int32) TransExp {
	return &Ex{exp: t.memPlus(base.unEx(), &BinOpExpIr{
		binop: MulIr,
		left:  &ConstExpIr{id},
		right: &ConstExpIr{wordSize},
	})}
}

func (t *Translate) SubscriptVar(base TransExp, id TransExp) TransExp {
	return &Ex{exp: t.memPlus(base.unEx(), &BinOpExpIr{
		binop: MulIr,
		left:  id.unEx(),
		right: &ConstExpIr{wordSize},
	})}
}

func (t *Translate) BinOp(op Operator, left TransExp, right TransExp) TransExp {
	leftEx, rightEx := left.unEx(), right.unEx()
	var opIr BinOpIr
	switch op {
	case Plus:
		opIr = PlusIr
	case Minus:
		opIr = MinusIr
	case Mul:
		opIr = MulIr
	case Div:
		opIr = DivIr
	}

	return &Ex{&BinOpExpIr{
		binop: opIr,
		left:  leftEx,
		right: rightEx,
	}}
}

func (t *Translate) RelOp(op Operator, left TransExp, right TransExp) TransExp {
	leftEx, rightEx := left.unEx(), right.unEx()
	var opIr RelOpIr
	switch op {
	case Le:
		opIr = LeIr
	case Lt:
		opIr = LtIr
	case Ge:
		opIr = GeIr
	case Gt:
		opIr = GtIr
	case Eq:
		opIr = EqIr
	case Neq:
		opIr = NeIr
	}

	return &Cx{
		func(tl, fl Label) StmIr {
			return &CJumpStmIr{
				relop:      opIr,
				left:       leftEx,
				right:      rightEx,
				trueLabel:  tl,
				falseLabel: fl,
			}
		},
	}
}

func (t *Translate) strExp(s string) TransExp {
	label := tm.NewLabel()
	frags = append(frags, &StringFrag{
		label: label,
		str:   s,
	})

	return &Ex{
		&NameExpIr{
			label: label,
		},
	}
}

func (t *Translate) intExp(i int32) TransExp {
	return &Ex{
		&ConstExpIr{i},
	}
}

func (t *Translate) nilExp() TransExp {
	return &Ex{
		&ConstExpIr{0},
	}
}

func (t *Translate) arrayExp(size, init TransExp) TransExp {
	return &Ex{
		t.externalCall("initArray", size.unEx(), init.unEx()),
	}
}

func (t *Translate) externalCall(name string, args ...ExpIr) *CallExpIr {
	return &CallExpIr{
		exp: &NameExpIr{
			tm.NamedLabel(name),
		},
		args: args,
	}
}

func (t *Translate) letExp(desc []TransExp, body TransExp) TransExp {
	if len(desc) == 0 {
		return body
	}

	stms := make([]StmIr, 0, len(desc))
	for _, d := range desc {
		stms = append(stms, d.unNx())
	}

	return &Ex{
		&EsEqExpIr{
			stm: t.seqStm(stms...),
			exp: body.unEx(),
		},
	}
}

func (t *Translate) call(useLevel, defLevel *Level, label Label, exps []TransExp, isProcedure bool) TransExp {
	args := make([]ExpIr, 0, 1+len(exps))
	args = append(args, defLevel.staticLink(useLevel, &TempExpIr{fp}))
	for _, e := range exps {
		args = append(args, e.unEx())
	}

	if isProcedure {
		return &Ex{
			&CallExpIr{
				exp:  &NameExpIr{label},
				args: args,
			},
		}
	}

	return &Nx{
		&ExpStmIr{
			&CallExpIr{
				exp:  &NameExpIr{label},
				args: args,
			},
		},
	}
}

func (t *Translate) seq(head, tail TransExp) TransExp {
	if isNop(head) {
		return tail
	}

	if isNop(tail) {
		return head
	}

	if v, ok := tail.(*Nx); ok {
		return &Nx{
			&SeqStmIr{
				first:  head.unNx(),
				second: v.stm,
			},
		}
	}

	return &Ex{&EsEqExpIr{
		stm: head.unNx(),
		exp: tail.unEx(),
	}}
}

func (t *Translate) assign(left, right TransExp) TransExp {
	return &Nx{&MoveStmIr{
		dst: right.unEx(),
		src: left.unEx(),
	}}
}

func (t *Translate) ifElse(ifEx, thenEx, elseEx TransExp) TransExp {
	r := tm.NewTemp()
	tl, fl, finish := tm.NewLabel(), tm.NewLabel(), tm.NewLabel()
	testFunc := ifEx.unCx()
	switch v := thenEx.(type) {
	case *Ex:
		if elseEx != nil {
			return &Ex{
				exp: &EsEqExpIr{
					stm: t.seqStm(
						testFunc(tl, fl),
						&LabelStmIr{
							tl,
						},
						&MoveStmIr{
							dst: &TempExpIr{r},
							src: v.exp,
						},
						&JumpStmIr{
							exp:    &NameExpIr{finish},
							labels: []Label{finish},
						},
						&LabelStmIr{fl},
						&MoveStmIr{
							dst: &TempExpIr{r},
							src: elseEx.unEx(),
						},
						&LabelStmIr{finish},
					),
					exp: &TempExpIr{r},
				},
			}
		}

	case *Nx:
		if elseEx != nil {
			return &Nx{
				stm: t.seqStm(
					testFunc(tl, fl),
					&LabelStmIr{
						tl,
					},
					v.stm,
					&JumpStmIr{
						exp:    &NameExpIr{finish},
						labels: []Label{finish},
					},
					&LabelStmIr{fl},
					elseEx.unNx(),
					&LabelStmIr{finish},
				),
			}
		}

		return &Nx{
			t.seqStm(
				testFunc(tl, fl),
				&LabelStmIr{tl},
				v.stm,
				&LabelStmIr{fl},
			),
		}

	case *Cx:
		if elseEx != nil {
			return &Cx{
				func(tl1, fl1 Label) StmIr {
					return t.seqStm(
						testFunc(tl, fl),
						&LabelStmIr{tl},
						v.cx(tl1, fl1),
						&LabelStmIr{fl},
						elseEx.unCx()(tl1, fl1),
					)
				},
			}
		}
	}

	panic("invalid case of if statement")
}

func (t *Translate) whileLoop(tex, bex TransExp, doneLabel Label) TransExp {
	tl, bl := tm.NewLabel(), tm.NewLabel()
	return &Nx{
		t.seqStm(
			&LabelStmIr{tl},
			&CJumpStmIr{
				relop:      EqIr,
				left:       tex.unEx(),
				right:      &ConstExpIr{0},
				trueLabel:  doneLabel,
				falseLabel: bl,
			},
			&LabelStmIr{bl},
			bex.unNx(),
			&JumpStmIr{
				exp:    &NameExpIr{tl},
				labels: []Label{tl},
			},
			&LabelStmIr{doneLabel},
		),
	}
}

func (t *Translate) breakStm(label Label) TransExp {
	return &Nx{
		&JumpStmIr{
			exp:    &NameExpIr{label},
			labels: []Label{label},
		},
	}
}

func (t *Translate) record(fields []TransExp) TransExp {
	r := tm.NewTemp()
	stms := make([]StmIr, len(fields)+1)
	stms[0] = &MoveStmIr{
		dst: &TempExpIr{r},
		src: t.externalCall("allocRecord", &ConstExpIr{
			int32(len(fields) * wordSize),
		}),
	}

	for i, field := range fields {
		stms[i+1] = &MoveStmIr{
			dst: t.memPlus(&TempExpIr{r}, &ConstExpIr{int32(i) * wordSize}),
			src: field.unEx(),
		}
	}

	return &Ex{
		&EsEqExpIr{
			stm: t.seqStm(stms...),
			exp: &TempExpIr{r},
		},
	}
}

func (t *Translate) seqStm(stms ...StmIr) StmIr {
	if len(stms) == 0 {
		return nil
	}

	if len(stms) == 1 {
		return stms[0]
	}

	return &SeqStmIr{
		first:  stms[0],
		second: t.seqStm(stms[1:]...),
	}
}

func ProcEntryExit(level *Level, body TransExp) {
	body1 := ProcEntryExit1(level.frame.(*MipsFrame), &MoveStmIr{
		dst: &TempExpIr{rv},
		src: body.unEx(),
	})

	frags = append(frags, &ProcFrag{
		body:  body1,
		frame: level.frame,
	})
}
