package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

var fileName = flag.String("source", "./test_files/record.tig", "source file to compile")

var tm *TempManagement

func main() {
	f, err := os.ReadFile(*fileName)
	if err != nil {
		log.Fatalf("error when reading input file %v", err)
	}

	buf := bufio.NewReader(bytes.NewReader(f))
	lexer := NewLexer(*fileName, buf)

	strs := NewStrings()
	parser := NewParser(lexer, strs)
	exp, err := parser.Parse()
	if err != nil {
		log.Fatalf("parsing error %v", err)
	}

	strBuilder := strings.Builder{}
	exp.String(strs, &strBuilder, 0)
	fmt.Println(strBuilder.String())

	findEscape := NewFindEscape(strs)
	findEscape.FindEscape(exp)

	tm = NewTempManagement(strs)

	translate := NewT
	venv, tenv := InitBaseVarEnv(strs), InitBaseTypeEnv(strs)
	semant := NewSemant(strs, venv, tenv)
	if err := semant.TransProg(exp); err != nil {
		log.Fatalf("semantic error %v", err)
	}
}
