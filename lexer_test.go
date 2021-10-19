package tiger

import (
	"bufio"
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLexer_HappyCase(t *testing.T) {
	t.Parallel()

	fn := "./test_files/hello.tig"
	f, err := os.ReadFile(fn)
	require.NoError(t, err)
	buf := bufio.NewReader(bytes.NewReader(f))
	lexer := NewLexer(fn, buf)
	token, err := lexer.Token()
	require.NoError(t, err)
	require.Equal(t, "ident", token.tok)
	require.EqualValues(t, "print", token.value)

	token, err = lexer.Token()
	require.NoError(t, err)
	require.Equal(t, "(", token.tok)

	token, err = lexer.Token()
	require.NoError(t, err)
	require.Equal(t, "str", token.tok)
	require.EqualValues(t, "Hello, World!\n", token.value)

	token, err = lexer.Token()
	require.NoError(t, err)
	require.Equal(t, ")", token.tok)
}

func TestLexer_Comment(t *testing.T) {
	t.Parallel()

	t.Run("unclosed_comment", func(t *testing.T) {
		t.Parallel()

		str := `  /* `
		buf := bufio.NewReader(bytes.NewReader([]byte(str)))
		lexer := NewLexer("", buf)
		_, err := lexer.Token()
		require.Error(t, err)
	})

	t.Run("invalid_comment", func(t *testing.T) {
		t.Parallel()

		str := `  \* `
		buf := bufio.NewReader(bytes.NewReader([]byte(str)))
		lexer := NewLexer("", buf)
		_, err := lexer.Token()
		require.Error(t, err)
	})

	t.Run("nested_comment", func(t *testing.T) {
		t.Parallel()

		str := `  /* /*  */ */ `
		buf := bufio.NewReader(bytes.NewReader([]byte(str)))
		lexer := NewLexer("", buf)
		tok, err := lexer.Token()
		require.NoError(t, err)
		require.EqualValues(t, "eof", tok.tok)
	})
}

func TestLexer_String(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		str := "  \" string \\\" \n \t \\123 \" "
		buf := bufio.NewReader(bytes.NewReader([]byte(str)))
		lexer := NewLexer("", buf)
		token, err := lexer.Token()
		require.NoError(t, err)
		require.Equal(t, "str", token.tok)
		require.EqualValues(t, " string \" \n \t 123 ", token.value)
	})

	t.Run("invalid_ascii", func(t *testing.T) {
		t.Parallel()

		str := "  \" string \\\" \n \t \\567 \" "
		buf := bufio.NewReader(bytes.NewReader([]byte(str)))
		lexer := NewLexer("", buf)
		_, err := lexer.Token()
		require.Error(t, err)
	})

	t.Run("unclosed_string", func(t *testing.T) {
		t.Parallel()

		fn := "./test_files/unclosed_string.tig"
		f, err := os.ReadFile(fn)
		require.NoError(t, err)
		buf := bufio.NewReader(bytes.NewReader(f))
		lexer := NewLexer(fn, buf)
		token, err := lexer.Token()
		require.NoError(t, err)
		require.Equal(t, "let", token.tok)

		token, err = lexer.Token()
		require.NoError(t, err)
		require.Equal(t, "var", token.tok)

		token, err = lexer.Token()
		require.NoError(t, err)
		require.Equal(t, "ident", token.tok)
		require.Equal(t, "N", token.value)

		token, err = lexer.Token()
		require.NoError(t, err)
		require.Equal(t, ":=", token.tok)

		token, err = lexer.Token()
		require.NoError(t, err)
		require.Equal(t, "int", token.tok)
		require.EqualValues(t, 8, token.value)

		token, err = lexer.Token()
		require.NoError(t, err)
		require.Equal(t, "in", token.tok)

		token, err = lexer.Token()
		require.NoError(t, err)
		require.Equal(t, "ident", token.tok)
		require.Equal(t, "print", token.value)

		token, err = lexer.Token()
		require.NoError(t, err)
		require.Equal(t, "(", token.tok)

		token, err = lexer.Token()
		require.Error(t, err)
	})
}
