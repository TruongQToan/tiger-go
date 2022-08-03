package main

func rewrite() {

}

func Alloc(frame Frame, instrs []Instr) ([]Instr, map[Temp]Temp) {
	fGraph := Instrs2FGraph(instrs)
	iGraph, moves := InitIGraph(fGraph)
	coloring := NewColoring(
		iGraph,
		fGraph,
		moves,
		frame.TempMap(),
	)

	colored, spilledNodes := coloring.Color()
}
