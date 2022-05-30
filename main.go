package main

import (
	"bufio"
	"bytes"
	"flag"
	"log"
	"os"
	"strings"
)

var fileName = flag.String("source", "./test_files/record.tig", "source file to compile")

var (
	strs = NewStrings()
	tm = NewTempManagement()
)

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
}
