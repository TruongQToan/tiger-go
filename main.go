package main

import (
	"bufio"
	"bytes"
	"flag"
	"log"
	"os"
	"strings"
)

var fileName = flag.String("source", "./test_files/test1.tig", "source file to compile")

var (
	strs = NewStrings()
	tm = NewTempManagement()
)

func emitProc(codeGen *CodeGenerator, canon *Canon, fo *os.File, proc *ProcFrag) {
	stms, _ := canon.Linearize(proc.body)
	blocks, doneLabel := canon.BasicBlocks(stms)
	stms = canon.TraceSchedule(blocks, doneLabel)

	instrs := make([]Instr, 0)
	for _, stm := range stms {
		instrs = append(instrs, codeGen.GenCode(stm)...)
	}

	for _, i := range instrs {
		fo.WriteString(i.assemStr() + "\n")
	}
}

func main() {
	f, err := os.ReadFile(*fileName)
	if err != nil {
		log.Fatalf("error when reading input file %v", err)
	}

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

	transExp, err := semant.TransProg(exp)
	if err != nil {
		log.Fatalf("semantic error %v", err)
	}

	strBuilder := strings.Builder{}
	transExp.print(&strBuilder, 0)

	fo, err := os.Create("tiger_ir.txt")
	if err != nil {
		log.Fatalf("cannot create file %v", err)
	}

	fo.WriteString(strBuilder.String())

	fo, err = os.Create("frags.txt")
	if err != nil {
		log.Fatalf("cannot create file %v", err)
	}

	var (
		procFrags []*ProcFrag
		strFrags  []*StringFrag
	)

	strBuilder.Reset()
	for _, frag := range frags {
		if f, ok := frag.(*ProcFrag); ok {
			strBuilder.WriteString("\nProcedure Frag\n")
			strBuilder.WriteString(strs.Get(Symbol(f.frame.Name())) + "\n")
			f.body.printStm(&strBuilder, 0)
			procFrags = append(procFrags, f)
		}

		if f, ok := frag.(*StringFrag); ok {
			strBuilder.WriteString("\nString Frag\n")
			strBuilder.WriteString(strs.Get(Symbol(f.label)) + "\n")
			strBuilder.WriteString(f.str + "\n")
			strFrags = append(strFrags, f)
		}
	}

	fo.WriteString(strBuilder.String())
	fo, err = os.Create("frags_sche.txt")
	if err != nil {
		log.Fatalf("cannot create file %v", err)
	}

	canon := &Canon{}

	strBuilder.Reset()
	for _, frag := range frags {
		if f, ok := frag.(*ProcFrag); ok {
			_, linearized := canon.Linearize(f.body)
			linearized.printStm(&strBuilder, 0)
			//blocks, doneLabel := canon.BasicBlocks(stms)
			//stms = canon.TraceSchedule(blocks, doneLabel)
			//for _, stm := range stms {
			//	stm.printStm(&strBuilder, 0)
			//}
		}
	}

	fo.WriteString(strBuilder.String())

	//fo, err = os.Create(*fileName + ".s")
	//if err != nil {
	//	log.Fatalf("cannot create file " + *fileName + ".s")
	//}
	//
	//fo.WriteString("\t.global main\n")
	//fo.WriteString("\t.data\n")
	//for _, str := range strFrags {
	//	fo.WriteString(tm.LabelString(str.label) + ": .asciiz \"" + str.str + "\"\n")
	//}
	//
	//codeGen := NewCodeGenerator()
	//fo.WriteString("\n\t.text\n")
	//for _, proc := range procFrags {
	//	emitProc(codeGen, canon, fo, proc)
	//}
}
