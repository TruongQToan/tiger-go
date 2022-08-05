package main

import (
	"bufio"
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var fileName = flag.String("source", "./test_files/hello3.tig", "source file to compile")

var (
	strs = NewStrings()
	tm   = NewTempManagement()
)

func addTab(instrs []Instr) {
	for _, instr := range instrs {
		switch v := instr.(type) {
		case *OperInstr:
			v.assem = "\t" + v.assem
		case *MoveInstr:
			v.assem = "\t" + v.assem
		}
	}
}

func emitProc(sb *strings.Builder, procs []*ProcFrag) {
	for _, proc := range procs {
		canon := &Canon{}
		stms, _ := canon.Linearize(proc.body)
		blocks, doneLabel := canon.BasicBlocks(stms)
		stms = canon.TraceSchedule(blocks, doneLabel)

		instrs := make([]Instr, 0)
		for _, stm := range stms {
			codeGen := NewCodeGenerator()
			instrs = append(instrs, codeGen.GenCode(stm)...)
		}

		instrs = ProcEntryExit2(instrs)

		var (
			colored map[Temp]string
		)

		instrs, colored = Alloc(proc.frame, instrs)
		addTab(instrs)
		prolog, epilog := proc.frame.ProcEntryExit3()
		sb.WriteString(prolog)
		for _, instr := range instrs {
			sb.WriteString(formatAssem(instr, func(temp Temp) string {
				return colored[temp]
			}) + "\n")
		}
		sb.WriteString(epilog)
	}
}

func emitString(sb *strings.Builder, strs []*StrFrag) {
	for _, str := range strs {
		StringFrag(sb, str)
	}
}

func emit(frags []Frag) string {
	var (
		procs []*ProcFrag
		strs  []*StrFrag
	)

	for _, frag := range frags {
		if f, ok := frag.(*ProcFrag); ok {
			procs = append(procs, f)
		} else {
			strs = append(strs, frag.(*StrFrag))
		}
	}

	sb := strings.Builder{}
	sb.WriteString("\t.globl main\n")
	sb.WriteString("\t.data\n")
	emitString(&sb, strs)
	sb.WriteString("\n\t.text\n")
	emitProc(&sb, procs)
	return sb.String()
}

func compile(f []byte) {
	buf := bufio.NewReader(bytes.NewReader(f))
	lexer := NewLexer(*fileName, buf)
	parser := NewParser(lexer, strs)
	exp, err := parser.Parse()
	if err != nil {
		log.Fatalf("parsing error %v", err)
	}

	findEscape := NewFindEscape()
	findEscape.FindEscape(exp)
	translate := Translate{frameFactory: NewMipsFrame}
	venv, tenv := InitBaseVarEnv(), InitBaseTypeEnv()
	semant := NewSemant(&translate, venv, tenv)
	frags, err := semant.TransProg(exp)
	if err != nil {
		log.Fatalf("semantic error %v", err)
	}

	fo, err := os.Create(*fileName + ".s")
	if err != nil {
		log.Fatalf("cannot create file %v", err)
	}

	rb, err := ioutil.ReadFile("./runtime/runtime.s")
	if err != nil {
		log.Fatalf("cannot open file %v", err)
	}

	fo.WriteString(string(rb) + "\n" + emit(frags))
}

func main() {
	f, err := os.ReadFile(*fileName)
	if err != nil {
		log.Fatalf("error when reading input file %v", err)
	}

	compile(f)
}
