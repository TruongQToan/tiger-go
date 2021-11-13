package tiger

import (
	"bufio"
	"bytes"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestParser_HappyCase(t *testing.T) {
	t.Parallel()

	fn := "./test_files/hello2.tig"
	f, err := os.ReadFile(fn)
	require.NoError(t, err)
	buf := bufio.NewReader(bytes.NewReader(f))
	lexer := NewLexer(fn, buf)

	strings := NewStrings()
	symbols := NewSymbols(strings)
	parser := NewParser(lexer, symbols)
	exp, err := parser.Parse()
	require.NoError(t, err)
	exp = exp.(*LetExp)
}