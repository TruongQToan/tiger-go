package main

import "strconv"

type CodeGenerator struct {
	instructions []Instr
	callDefs     []Temp
}

func NewCodeGenerator() *CodeGenerator {
	return &CodeGenerator{
		callDefs: append([]Temp{rv, ra}, argRegs...),
	}
}

func (c *CodeGenerator) GenCode(stm StmIr) []Instr {
	c.munchStm(stm)
	return c.instructions
}

func (c *CodeGenerator) munchStm(s StmIr) {
	switch v := s.(type) {
	case *SeqStmIr:
		c.munchStm(v.first)
		c.munchStm(v.second)

	case *LabelStmIr:
		c.instructions = append(c.instructions, &LabelInstr{
			assem: tm.LabelString(v.label) + ":",
			lab:   v.label,
		})

	case *MoveStmIr:
		switch v1 := v.dst.(type) {

		// Move to memory
		case *MemExpIr:
			switch v2 := v1.mem.(type) {
			case *BinOpExpIr:
				switch v2.binop {
				case PlusIr:
					var instr *OperInstr
					if v3, ok := v2.right.(*ConstExpIr); ok {
						instr = &OperInstr{
							assem: "sw `s0, " + strconv.FormatInt(int64(v3.c), 10) + "(`s1)",
							src:   []Temp{c.munchExp(v.src), c.munchExp(v2.left)},
						}
					} else if v3, ok := v2.left.(*ConstExpIr); ok {
						instr = &OperInstr{
							assem: "sw `s0, " + strconv.FormatInt(int64(v3.c), 10) + "(`s1)",
							src:   []Temp{c.munchExp(v.src), c.munchExp(v2.right)},
						}
					} else {
						// Memory mode must be 3($t5) or -3($t5) for example.
						panic("invalid memory mode")
					}

					c.instructions = append(c.instructions, instr)

				case MinusIr:
					var instr *OperInstr
					if v3, ok := v2.right.(*ConstExpIr); ok {
						instr = &OperInstr{
							assem: "sw `s0, " + strconv.FormatInt(-int64(v3.c), 10) + "(`s1)",
							src:   []Temp{c.munchExp(v.src), c.munchExp(v2.left)},
						}
					} else if v3, ok := v2.left.(*ConstExpIr); ok {
						instr = &OperInstr{
							assem: "sw `s0, " + strconv.FormatInt(-int64(v3.c), 10) + "(`s1)",
							src:   []Temp{c.munchExp(v.src), c.munchExp(v2.right)},
						}
					} else {
						// Memory mode must be 3($t5) or -3($t5) for example.
						panic("invalid memory mode")
					}

					c.instructions = append(c.instructions, instr)

				default:
					panic("invalid memory mode")
				}

			default:
				instr := &OperInstr{
					assem: "sw `s0, 0(`s1)",
					src:   []Temp{c.munchExp(v.src), c.munchExp(v1.mem)},
				}
				c.instructions = append(c.instructions, instr)
			}

		// Load to register
		case *TempExpIr:
			switch v2 := v.src.(type) {
			case *ConstExpIr:
				instr := &OperInstr{
					assem: "li `d0, " + strconv.FormatInt(int64(v2.c), 10),
					dst:   []Temp{c.munchExp(v.dst)},
				}
				c.instructions = append(c.instructions, instr)

			case *BinOpExpIr:
				switch v2.binop {
				case PlusIr:
					if v3, ok := v2.left.(*ConstExpIr); ok {
						instr := &OperInstr{
							assem: "lw `d0, " + strconv.FormatInt(int64(v3.c), 10) + "(`s0)",
							dst:   []Temp{c.munchExp(v.dst)},
							src:   []Temp{c.munchExp(v2.right)},
						}

						c.instructions = append(c.instructions, instr)
					} else if v3, ok := v2.right.(*ConstExpIr); ok {
						instr := &OperInstr{
							assem: "lw `d0, " + strconv.FormatInt(int64(v3.c), 10) + "(`s0)",
							dst:   []Temp{c.munchExp(v.dst)},
							src:   []Temp{c.munchExp(v2.left)},
						}

						c.instructions = append(c.instructions, instr)
					} else {
						panic("invalid memory mode")
					}

				case MinusIr:
					if v3, ok := v2.left.(*ConstExpIr); ok {
						instr := &OperInstr{
							assem: "lw `d0, " + strconv.FormatInt(-int64(v3.c), 10) + "(`s0)",
							dst:   []Temp{c.munchExp(v.dst)},
							src:   []Temp{c.munchExp(v2.right)},
						}

						c.instructions = append(c.instructions, instr)
					} else if v3, ok := v2.right.(*ConstExpIr); ok {
						instr := &OperInstr{
							assem: "lw `d0, " + strconv.FormatInt(-int64(v3.c), 10) + "(`s0)",
							dst:   []Temp{c.munchExp(v.dst)},
							src:   []Temp{c.munchExp(v2.left)},
						}

						c.instructions = append(c.instructions, instr)
					} else {
						panic("invalid memory mode")
					}
				}

			// move register to register
			default:
				instr := &OperInstr{
					assem: "move `d0, `s0",
					dst:   []Temp{v1.temp},
					src:   []Temp{c.munchExp(v.src)},
				}

				c.instructions = append(c.instructions, instr)
			}

		default:
			panic("invalid instruction arguments")
		}

	case *JumpStmIr:
		switch v1 := v.exp.(type) {
		case *NameExpIr:
			instr := &OperInstr{
				assem: "b `j0",
				jumps: []Label{v1.label},
			}

			c.instructions = append(c.instructions, instr)

		default:
			instr := &OperInstr{
				assem: "jr `s0",
				src:   []Temp{c.munchExp(v.exp)},
				jumps: v.labels,
			}

			c.instructions = append(c.instructions, instr)
		}

	case *CJumpStmIr:
		switch v1 := v.right.(type) {
		case *ConstExpIr:
			if v1.c == 0 {
				switch v.relop {
				case EqIr:
					instr := &OperInstr{
						assem: "beqz `s0, `j0\nb `j1",
						src:   []Temp{c.munchExp(v.left)},
						jumps: []Label{v.trueLabel, v.falseLabel},
					}

					c.instructions = append(c.instructions, instr)

				case NeIr:
					instr := &OperInstr{
						assem: "bnez `s0, `j0\nb `j1",
						src:   []Temp{c.munchExp(v.left)},
						jumps: []Label{v.trueLabel, v.falseLabel},
					}

					c.instructions = append(c.instructions, instr)

				case GeIr:
					instr := &OperInstr{
						assem: "bgez `s0, `j0\nb `j1",
						src:   []Temp{c.munchExp(v.left)},
						jumps: []Label{v.trueLabel, v.falseLabel},
					}

					c.instructions = append(c.instructions, instr)

				case GtIr:
					instr := &OperInstr{
						assem: "bgtz `s0, `j0\nb `j1",
						src:   []Temp{c.munchExp(v.left)},
						jumps: []Label{v.trueLabel, v.falseLabel},
					}

					c.instructions = append(c.instructions, instr)

				case LtIr:
					instr := &OperInstr{
						assem: "bltz `s0, `j0\nb `j1",
						src:   []Temp{c.munchExp(v.left)},
						jumps: []Label{v.trueLabel, v.falseLabel},
					}

					c.instructions = append(c.instructions, instr)

				case LeIr:
					instr := &OperInstr{
						assem: "blez `s0, `j0\nb `j1",
						src:   []Temp{c.munchExp(v.left)},
						jumps: []Label{v.trueLabel, v.falseLabel},
					}

					c.instructions = append(c.instructions, instr)
				}
			}

		default:
			switch v.relop {
			case LeIr:
				instr := &OperInstr{
					assem: "ble `s0, `s1, `j0\nb `j1",
					src:   []Temp{c.munchExp(v.left), c.munchExp(v.right)},
					jumps: []Label{v.trueLabel, v.falseLabel},
				}

				c.instructions = append(c.instructions, instr)

			case LtIr:
				instr := &OperInstr{
					assem: "blt `s0, `s1, `j0\nb `j1",
					src:   []Temp{c.munchExp(v.left), c.munchExp(v.right)},
					jumps: []Label{v.trueLabel, v.falseLabel},
				}

				c.instructions = append(c.instructions, instr)

			case GeIr:
				instr := &OperInstr{
					assem: "bge `s0, `s1, `j0\nb `j1",
					src:   []Temp{c.munchExp(v.left), c.munchExp(v.right)},
					jumps: []Label{v.trueLabel, v.falseLabel},
				}

				c.instructions = append(c.instructions, instr)

			case GtIr:
				instr := &OperInstr{
					assem: "bgt `s0, `s1, `j0\nb `j1",
					src:   []Temp{c.munchExp(v.left), c.munchExp(v.right)},
					jumps: []Label{v.trueLabel, v.falseLabel},
				}

				c.instructions = append(c.instructions, instr)

			case EqIr:
				instr := &OperInstr{
					assem: "beq `s0, `s1, `j0\nb `j1",
					src:   []Temp{c.munchExp(v.left), c.munchExp(v.right)},
					jumps: []Label{v.trueLabel, v.falseLabel},
				}

				c.instructions = append(c.instructions, instr)

			case NeIr:
				instr := &OperInstr{
					assem: "bne `s0, `s1, `j0\nb `j1",
					src:   []Temp{c.munchExp(v.left), c.munchExp(v.right)},
					jumps: []Label{v.trueLabel, v.falseLabel},
				}

				c.instructions = append(c.instructions, instr)

			default:
				panic("invalid binary operator")
			}

		}

	case *ExpStmIr:
		c.munchExp(v.exp)
	}
}

func (c *CodeGenerator) munchExp(exp ExpIr) Temp {
	switch t := exp.(type) {
	case *CallExpIr:
		tempCallerSaves := make([]Temp, 0, len(callerSaves))
		for _ = range callerSaves {
			tempCallerSaves = append(tempCallerSaves, tm.NewTemp())
		}

		// Move the caller saves to temporary
		for i, exp := range callerSaves {
			c.munchStm(&MoveStmIr{
				dst: &TempExpIr{tempCallerSaves[i]},
				src: &TempExpIr{exp},
			})
		}

		c.instructions = append(c.instructions, &OperInstr{
			assem: "jalr `s0",
			dst:   c.callDefs,
			src:   append([]Temp{c.munchExp(t.exp)}, c.buildArgs(argRegs, t.args)...),
		})

		return rv

	case *MemExpIr:
		switch t1 := t.mem.(type) {
		case *ConstExpIr:
			return c.gen(func(t Temp) {
				c.instructions = append(c.instructions, &OperInstr{
					assem: "lw `d0, " + strconv.FormatInt(int64(t1.c), 10) + "($zero)",
					dst:   []Temp{t},
				})
			})

		case *BinOpExpIr:
			switch t1.binop {
			case PlusIr:
				if t2, ok := t1.left.(*ConstExpIr); ok {
					return c.gen(func(t Temp) {
						c.instructions = append(c.instructions, &OperInstr{
							assem: "lw `d0, " + strconv.FormatInt(int64(t2.c), 10) + "(`s0)",
							dst:   []Temp{t},
							src:   []Temp{c.munchExp(t1.right)},
						})
					})
				}

				if t2, ok := t1.right.(*ConstExpIr); ok {
					return c.gen(func(t Temp) {
						c.instructions = append(c.instructions, &OperInstr{
							assem: "lw `d0, " + strconv.FormatInt(int64(t2.c), 10) + "(`s0)",
							dst:   []Temp{t},
							src:   []Temp{c.munchExp(t1.left)},
						})
					})
				}

				panic("invalid memory mode")

			case MinusIr:
				if t2, ok := t1.left.(*ConstExpIr); ok {
					c.gen(func(t Temp) {
						c.instructions = append(c.instructions, &OperInstr{
							assem: "lw `d0, " + strconv.FormatInt(int64(-t2.c), 10) + "(`s0)",
							dst:   []Temp{t},
							src:   []Temp{c.munchExp(t1.right)},
						})
					})
				}

				if t2, ok := t1.right.(*ConstExpIr); ok {
					return c.gen(func(t Temp) {
						c.instructions = append(c.instructions, &OperInstr{
							assem: "lw `d0, " + strconv.FormatInt(int64(-t2.c), 10) + "(`s0)",
							dst:   []Temp{t},
							src:   []Temp{c.munchExp(t1.left)},
						})
					})
				}

				panic("invalid memory mode")

			default:
				panic("invalid binary operator when loading memory")
			}

		default:
			c.gen(func(temp Temp) {
				c.instructions = append(c.instructions, &OperInstr{
					assem: "lw `d0, 0(`s0)",
					dst:   []Temp{temp},
					src:   []Temp{c.munchExp(t.mem)},
				})
			})
		}

	case *BinOpExpIr:
		switch t.binop {
		case PlusIr:
			if t1, ok := t.left.(*ConstExpIr); ok {
				return c.gen(func(temp Temp) {
					c.instructions = append(c.instructions, &OperInstr{
						assem: "addi `d0, `s0, " + strconv.FormatInt(int64(t1.c), 10),
						dst:   []Temp{temp},
						src:   []Temp{c.munchExp(t.right)},
					})
				})
			}

			if t1, ok := t.right.(*ConstExpIr); ok {
				return c.gen(func(temp Temp) {
					c.instructions = append(c.instructions, &OperInstr{
						assem: "addi `d0, `s0, " + strconv.FormatInt(int64(t1.c), 10),
						dst:   []Temp{temp},
						src:   []Temp{c.munchExp(t.left)},
					})
				})
			}

			return c.gen(func(temp Temp) {
				c.instructions = append(c.instructions, &OperInstr{
					assem: "add `d0, `s0, `s1",
					dst:   []Temp{temp},
					src:   []Temp{c.munchExp(t.left), c.munchExp(t.right)},
				})
			})

		case MinusIr:
			if t1, ok := t.right.(*ConstExpIr); ok {
				return c.gen(func(temp Temp) {
					c.instructions = append(c.instructions, &OperInstr{
						assem: "addiu `d0, `s0, " + strconv.FormatInt(int64(-t1.c), 10),
						dst:   []Temp{temp},
						src:   []Temp{c.munchExp(t.left)},
					})
				})
			}

			return c.gen(func(temp Temp) {
				c.instructions = append(c.instructions, &OperInstr{
					assem: "sub `d0, `s0, s1",
					dst:   []Temp{temp},
					src:   []Temp{c.munchExp(t.left), c.munchExp(t.right)},
				})
			})

		case MulIr:
			return c.gen(func(temp Temp) {
				c.instructions = append(c.instructions, &OperInstr{
					assem: "mul `d0, `s0, `s1",
					dst:   []Temp{temp},
					src:   []Temp{c.munchExp(t.left), c.munchExp(t.right)},
				})
			})

		case DivIr:
			return c.gen(func(temp Temp) {
				c.instructions = append(c.instructions, &OperInstr{
					assem: "div `d0, `s0, `s1",
					dst:   []Temp{temp},
					src:   []Temp{c.munchExp(t.left), c.munchExp(t.right)},
				})
			})
		}

	case *TempExpIr:
		return t.temp

	case *NameExpIr:
		return c.gen(func(temp Temp) {
			c.instructions = append(c.instructions, &OperInstr{
				assem: "la `d0, " + tm.LabelString(t.label),
				dst:   []Temp{temp},
			})
		})
	}

	panic("invalid IR exp")
}

func (c *CodeGenerator) buildArgs(argsRegisters []Temp, args []ExpIr) []Temp {
	if len(args) == 0 {
		return nil
	}

	n := len(argsRegisters)
	temps := make([]Temp, 0, n)
	for i, exp := range args {
		if i < n {
			c.munchStm(&MoveStmIr{
				dst: &TempExpIr{argsRegisters[i]},
				src: exp,
			})

			temps = append(temps, argsRegisters[i])
		} else {
			c.munchStm(&MoveStmIr{
				dst: &MemExpIr{
					&BinOpExpIr{
						binop: PlusIr,
						left:  &ConstExpIr{int32(i * wordSize)},
						right: &TempExpIr{fp},
					},
				},
				src: exp,
			})
		}
	}

	return temps
}

func (c *CodeGenerator) gen(f func(t Temp)) Temp {
	t := tm.NewTemp()
	f(t)
	return t
}
