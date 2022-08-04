package main

func isRedundantMove(colored map[Temp]Temp, i Instr) bool {
	i1, ok := i.(*MoveInstr)
	if !ok {
		return false
	}

	dst, src := i1.dst, i1.src
	return colored[dst] == colored[src]
}

func genInst(isDefine bool, accessExp ExpIr, temp Temp) []Instr {
	codeGen := NewCodeGenerator()
	if isDefine {
		return codeGen.GenCode(&MoveStmIr{
			dst: accessExp,
			src: &TempExpIr{temp},
		})
	}

	return codeGen.GenCode(&MoveStmIr{
		dst: &TempExpIr{},
		src: accessExp,
	})
}

func replaceWithNewTemp(temp Temp, tempList []Temp) (Temp, []Temp) {
	nt := tm.NewTemp()
	newTempList := make([]Temp, 0, len(tempList))
	for _, t := range tempList {
		if t == temp {
			newTempList = append(newTempList, nt)
		} else {
			newTempList = append(newTempList, t)
		}
	}

	return nt, newTempList
}

func allocStoreFetch(isDef bool, acc ExpIr, temp Temp, tempList []Temp) ([]Instr, []Temp) {
	hasTemp := false
	for _, t := range tempList {
		if t == temp {
			hasTemp = true
			break
		}
	}

	if !hasTemp {
		return []Instr{}, tempList
	}

	nt, newTempList := replaceWithNewTemp(temp, tempList)
	return genInst(isDef, acc, nt), newTempList
}

func rewriteOne(frame Frame, temp Temp, instrs []Instr) []Instr {
	accessExp := frame.AllocLocal(true).exp(&TempExpIr{temp: frame.FP()})
	newInstrs := make([]Instr, 0)
	for _, instr := range instrs {
		switch t := instr.(type) {
		case *OperInstr:
			stores, dst := allocStoreFetch(true, accessExp, temp, t.dst)
			fetches, src := allocStoreFetch(false, accessExp, temp, t.src)
			newInstrs = append(newInstrs, fetches...)
			newInstrs = append(newInstrs, &OperInstr{
				assem: t.assem,
				dst:   dst,
				src:   src,
				jumps: t.jumps,
			})
			newInstrs = append(newInstrs, stores...)

		case *MoveInstr:
			stores, dst := allocStoreFetch(true, accessExp, temp, []Temp{t.dst})
			fetches, src := allocStoreFetch(false, accessExp, temp, []Temp{t.src})
			newInstrs = append(newInstrs, fetches...)
			newInstrs = append(newInstrs, &MoveInstr{
				assem: t.assem,
				dst:   dst[0],
				src:   src[0],
			})
			newInstrs = append(newInstrs, stores...)

		default:
			newInstrs = append(newInstrs, t)
		}
	}

	return newInstrs
}

func rewrite(frame Frame, spilledNodes *IGraphNodeSet, instrs []Instr) []Instr {
	for _, node := range spilledNodes.All() {
		instrs = rewriteOne(frame, node.temp, instrs)
	}

	return instrs
}

func Alloc(frame Frame, instrs []Instr) ([]Instr, map[Temp]string) {
	fGraph := Instrs2FGraph(instrs)
	iGraph, moves := InitIGraph(fGraph)
	coloring := NewColoring(
		iGraph,
		fGraph,
		moves,
		frame.TempMap(),
	)

	colored, spilledNodes := coloring.Color()
	if spilledNodes.Empty() {
		filteredInstrs := make([]Instr, 0, len(instrs))
		for _, inst := range instrs {
			if !isRedundantMove(colored, inst) {
				filteredInstrs = append(filteredInstrs, inst)
			}
		}

		res := make(map[Temp]string)
		for k, v := range colored {
			res[k] = frame.TempName(v)
		}

		return filteredInstrs, res
	}

	rewrittenInstrs := rewrite(frame, spilledNodes, instrs)
	return Alloc(frame, rewrittenInstrs)
}
