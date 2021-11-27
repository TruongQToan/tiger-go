package tiger

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

	fn := "./test_files/hello2.tig"
	f, err := os.ReadFile(fn)
	require.NoError(t, err)
	buf := bufio.NewReader(bytes.NewReader(f))
	lexer := NewLexer(fn, buf)

	symbols := NewSymbols(NewStrings())
	parser := NewParser(lexer, symbols)
	exp, err := parser.Parse()
	require.NoError(t, err)
	strBuilder := strings.Builder{}
	exp.String(symbols, &strBuilder, 0)
	fmt.Println(strBuilder.String())
}

func TestParser_Integers(t *testing.T) {
	t.Parallel()

	fn := "./test_files/integers.tig"
	f, err := os.ReadFile(fn)
	require.NoError(t, err)
	buf := bufio.NewReader(bytes.NewReader(f))
	lexer := NewLexer(fn, buf)

	symbols := NewSymbols(NewStrings())
	parser := NewParser(lexer, symbols)
	exp, err := parser.Parse()
	require.NoError(t, err)
	strBuilder := strings.Builder{}
	exp.String(symbols, &strBuilder, 0)
	fmt.Println(strBuilder.String())
}

func TestParser_Loops(t *testing.T) {
	t.Parallel()

	fn := "./test_files/loops.tig"
	f, err := os.ReadFile(fn)
	require.NoError(t, err)
	buf := bufio.NewReader(bytes.NewReader(f))
	lexer := NewLexer(fn, buf)

	symbols := NewSymbols(NewStrings())
	parser := NewParser(lexer, symbols)
	exp, err := parser.Parse()
	require.NoError(t, err)
	strBuilder := strings.Builder{}
	exp.String(symbols, &strBuilder, 0)
	fmt.Println(strBuilder.String())
}
