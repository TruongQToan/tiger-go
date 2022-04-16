package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/stretchr/testify/require"
	"os"
	"strings"
	"testing"
)

func TestParser_Hello2(t *testing.T) {
	t.Parallel()

	testFile(t, "./test_files/hello2.tig")
}

func TestParser_Integers(t *testing.T) {
	t.Parallel()

	testFile(t, "./test_files/integers.tig")
}

func TestParser_Loops(t *testing.T) {
	t.Parallel()

	testFile(t, "./test_files/loops.tig")
}

func TestParser_Cycle(t *testing.T) {
	t.Parallel()

	testFile(t, "./test_files/cycle.tig")
}

func TestParser_Functions(t *testing.T) {
	t.Parallel()

	testFile(t, "./test_files/functions.tig")
}

func testFile(t *testing.T, fileName string) {
	f, err := os.ReadFile(fileName)
	require.NoError(t, err)
	buf := bufio.NewReader(bytes.NewReader(f))
	lexer := NewLexer(fileName, buf)

	strs := NewStrings()
	parser := NewParser(lexer, strs)
	exp, err := parser.Parse()
	require.NoError(t, err)
	strBuilder := strings.Builder{}
	exp.String(strs, &strBuilder, 0)
	fmt.Println(strBuilder.String())
}
