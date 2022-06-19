package main

type Canon struct{}

// TraceSchedule is basically a dfs-like algorithm
func (c *Canon) TraceSchedule(blocks [][]StmIr, done Label) []StmIr {
	// Add all the blocks to a table
	table := NewStmBlockST()
	for _, block := range blocks {
		if len(block) > 0 {
			if v, ok := block[0].(*LabelStmIr); ok {
				table.Enter(Symbol(v.label), block)
			}
		}
	}

	return append(c.getNextTrace(table, blocks), &LabelStmIr{done})
}

func (c *Canon) getNextTrace(table *StmBlockST, blocks [][]StmIr) []StmIr {
	if len(blocks) == 0 {
		return nil
	}

	if len(blocks[0]) > 0 {
		if v, ok := blocks[0][0].(*LabelStmIr); ok {
			if _, err := table.Look(Symbol(v.label)); err == nil {
				return c.trace(table, blocks[0], blocks[1:])
			}
		}
	}

	return c.getNextTrace(table, blocks[1:])
}

func (c *Canon) trace(table *StmBlockST, block []StmIr, rest [][]StmIr) []StmIr {
	// Mark label visited
	if len(block) == 0 {
		panic("cannot have empty block")
	}

	if v, ok := block[0].(*LabelStmIr); !ok {
		panic("head of the block must be a label")
	} else {
		table.Replace(Symbol(v.label), nil)
		switch v1 := block[len(block)-1].(type) {
		case *JumpStmIr:
			if _, ok := v1.exp.(*NameExpIr); ok {
				if v, err := table.Look(Symbol(v1.labels[0])); err == nil {
					return append(block[:len(block)-1], c.trace(table, v, rest)...)
				}

				return append(block, c.getNextTrace(table, rest)...)
			}

			return append(block, c.getNextTrace(table, rest)...)

		case *CJumpStmIr:
			if v, err := table.Look(Symbol(v1.falseLabel)); err == nil {
				return append(block, c.trace(table, v, rest)...)
			} else if v, err := table.Look(Symbol(v1.trueLabel)); err == nil {
				return append(block[:len(block)-1], append([]StmIr{&CJumpStmIr{
					relop:      v1.relop.not(),
					left:       v1.left,
					right:      v1.right,
					trueLabel:  v1.falseLabel,
					falseLabel: v1.trueLabel,
				}}, c.trace(table, v, rest)...)...)
			} else {
				f := tm.NewLabel()
				return append(block[:len(block)-1], append([]StmIr{
					&CJumpStmIr{
						relop:      v1.relop,
						left:       v1.left,
						right:      v1.right,
						trueLabel:  v1.trueLabel,
						falseLabel: f,
					},
					&LabelStmIr{f},
					&JumpStmIr{
						exp:    &NameExpIr{f},
						labels: []Label{f},
					},
				}, c.getNextTrace(table, rest)...)...)
			}
		}
	}

	panic("cannot call trace on not jump statement")
}

func (c *Canon) enterBlock(table *StmBlockST, block []StmIr) {
	if v, ok := block[0].(*LabelStmIr); ok {
		table.Enter(Symbol(v.label), block)
	}
}

func (c *Canon) BasicBlocks(stms []StmIr) ([][]StmIr, Label) {
	var blocks [][]StmIr
	done := tm.NewLabel()
	c.blocks(stms, &blocks, done)
	return blocks, done
}

func (c *Canon) blocks(stms []StmIr, blocks *[][]StmIr, done Label) {
	if len(stms) == 0 {
		return
	}

	if _, ok := stms[0].(*LabelStmIr); ok {
		c.nextBlock(stms[1:], []StmIr{stms[0]}, blocks, done)
		return
	}

	c.blocks(append([]StmIr{&LabelStmIr{tm.NewLabel()}}, stms...), blocks, done)
}

func (c *Canon) nextBlock(stms []StmIr, thisBlock []StmIr, blocks *[][]StmIr, done Label) {
	if len(stms) == 0 {
		c.nextBlock([]StmIr{
			&JumpStmIr{
				exp:    &NameExpIr{done},
				labels: []Label{done},
			},
		}, thisBlock, blocks, done)
		return
	}

	switch v := stms[0].(type) {
	case *JumpStmIr:
		c.endBlock(stms[1:], append(thisBlock, v), blocks, done)

	case *CJumpStmIr:
		c.endBlock(stms[1:], append(thisBlock, v), blocks, done)

	case *LabelStmIr:
		c.nextBlock(append([]StmIr{
			&JumpStmIr{
				exp:    &NameExpIr{v.label},
				labels: []Label{v.label},
			},
		}, stms...), thisBlock, blocks, done)
	default:
		c.nextBlock(stms[1:], append(thisBlock, v), blocks, done)
	}
}

func (c *Canon) endBlock(stms []StmIr, thisBlock []StmIr, blocks *[][]StmIr, done Label) {
	*blocks = append(*blocks, thisBlock)
	c.blocks(stms, blocks, done)
}

func (c *Canon) Linearize(stm StmIr) ([]StmIr, StmIr) {
	s := c.doStm(stm)
	return c.linear(s, nil), s
}

func (c *Canon) linear(s1 StmIr, s2 []StmIr) []StmIr {
	switch v := s1.(type) {
	case *SeqStmIr:
		return c.linear(v.first, c.linear(v.second, s2))

	default:
		return append([]StmIr{s1}, s2...)
	}
}

func (c *Canon) reorder(exps []ExpIr) (StmIr, []ExpIr) {
	if len(exps) == 0 {
		return &ExpStmIr{exp: &ConstExpIr{0}}, nil
	}

	if v, ok := exps[0].(*CallExpIr); ok {
		t := tm.NewTemp()
		return c.reorder(append([]ExpIr{&EsEqExpIr{
			stm: &MoveStmIr{
				dst: &TempExpIr{t},
				src: v,
			},
			exp: &TempExpIr{t},
		}}, exps[1:]...))
	}

	s1, e1 := c.doExp(exps[0])
	if len(exps) == 1 {
		return s1, []ExpIr{e1}
	}

	s2, e2 := c.reorder(exps[1:])
	if c.commute(s2, e1) {
		return c.concat(s1, s2), append([]ExpIr{e1}, e2...)
	}

	t := tm.NewTemp()
	return c.concat(c.concat(s1, &MoveStmIr{
		dst: &TempExpIr{t},
		src: e1,
	}), s2), append([]ExpIr{&TempExpIr{t}}, e2...)
}

func (c *Canon) reorderExp(exps []ExpIr, fn func(e1 []ExpIr) ExpIr) (StmIr, ExpIr) {
	s, el := c.reorder(exps)
	return s, fn(el)
}

func (c *Canon) reorderStm(exps []ExpIr, fn func(e1 []ExpIr) StmIr) StmIr {
	s, el := c.reorder(exps)
	return c.concat(s, fn(el))
}

func (c *Canon) doExp(e ExpIr) (StmIr, ExpIr) {
	switch v := e.(type) {
	case *BinOpExpIr:
		return c.reorderExp([]ExpIr{v.left, v.right}, func(e1 []ExpIr) ExpIr {
			return &BinOpExpIr{
				binop: v.binop,
				left:  e1[0],
				right: e1[1],
			}
		})

	case *MemExpIr:
		return c.reorderExp([]ExpIr{v.mem}, func(e1 []ExpIr) ExpIr {
			return &MemExpIr{e1[0]}
		})

	case *EsEqExpIr:
		stm := c.doStm(v.stm)
		stm1, e := c.doExp(v.exp)
		return c.concat(stm, stm1), e

	case *CallExpIr:
		return c.reorderExp(v.args, func(e1 []ExpIr) ExpIr {
			return &CallExpIr{
				exp:  v.exp,
				args: e1,
			}
		})

	default:
		return c.reorderExp([]ExpIr{}, func(e1 []ExpIr) ExpIr {
			return v
		})
	}
}

func (c *Canon) doStm(stm StmIr) StmIr {
	switch v := stm.(type) {
	case *SeqStmIr:
		s1, s2 := c.doStm(v.first), c.doStm(v.second)
		return c.concat(s1, s2)

	case *JumpStmIr:
		return c.reorderStm([]ExpIr{v.exp}, func(e1 []ExpIr) StmIr {
			return &JumpStmIr{
				exp:    e1[0],
				labels: v.labels,
			}
		})

	case *CJumpStmIr:
		return c.reorderStm([]ExpIr{v.left, v.right}, func(e1 []ExpIr) StmIr {
			return &CJumpStmIr{
				relop:      v.relop,
				left:       e1[0],
				right:      e1[1],
				trueLabel:  v.trueLabel,
				falseLabel: v.falseLabel,
			}
		})

	case *MoveStmIr:
		switch v1 := v.dst.(type) {
		case *TempExpIr:
			if v2, ok := v.src.(*CallExpIr); ok {
				return c.reorderStm(append([]ExpIr{v2.exp}, v2.args...), func(e1 []ExpIr) StmIr {
					return &MoveStmIr{
						dst: v1,
						src: &CallExpIr{
							exp:  e1[0],
							args: e1[1:],
						},
					}
				})
			}

			return c.reorderStm([]ExpIr{v.src}, func(e1 []ExpIr) StmIr {
				return &MoveStmIr{
					dst: v.dst,
					src: e1[0],
				}
			})

		case *MemExpIr:
			return c.reorderStm([]ExpIr{v1.mem, v.src}, func(e1 []ExpIr) StmIr {
				return &MoveStmIr{
					dst: &MemExpIr{e1[0]},
					src: e1[1],
				}
			})

		case *EsEqExpIr:
			return c.doStm(&SeqStmIr{
				first: v1.stm,
				second: &MoveStmIr{
					dst: v1.exp,
					src: v.src,
				},
			})
		}

	case *ExpStmIr:
		switch v1 := v.exp.(type) {
		case *CallExpIr:
			return c.reorderStm(v1.args, func(e1 []ExpIr) StmIr {
				return &ExpStmIr{
					&CallExpIr{
						exp:  v1.exp,
						args: e1,
					},
				}
			})

		default:
			return c.reorderStm([]ExpIr{v.exp}, func(e1 []ExpIr) StmIr {
				return &ExpStmIr{e1[0]}
			})
		}

	default:
		return c.reorderStm([]ExpIr{}, func(e1 []ExpIr) StmIr {
			return v
		})
	}

	panic("invalid statement ir type")
}

func (c *Canon) commute(s StmIr, e ExpIr) bool {
	if v, ok := s.(*ExpStmIr); ok {
		if _, ok := v.exp.(*ConstExpIr); ok {
			return true
		}
	}

	if _, ok := e.(*NameExpIr); ok {
		return true
	}

	if _, ok := e.(*ConstExpIr); ok {
		return true
	}

	return false
}

func (c *Canon) concat(s1, s2 StmIr) StmIr {
	if v, ok := s1.(*ExpStmIr); ok {
		if _, ok := v.exp.(*ConstExpIr); ok {
			return s2
		}
	}

	if v, ok := s2.(*ExpStmIr); ok {
		if _, ok := v.exp.(*ConstExpIr); ok {
			return s1
		}
	}

	return &SeqStmIr{
		first:  s1,
		second: s2,
	}
}
