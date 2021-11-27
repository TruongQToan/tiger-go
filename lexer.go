package tiger

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Lexer struct {
	buf *bufio.Reader
	pos *Pos
}

func NewLexer(fn string, buf *bufio.Reader) *Lexer {
	return &Lexer{
		buf: buf,
		pos: &Pos{
			fileName: fn,
			line:     1,
			col:      1,
			byte:     0,
			length:   0,
		},
	}
}

func (lex *Lexer) currentChar() (byte, error) {
	chars, err := lex.buf.Peek(1)
	if err != nil {
		return 0, err
	}

	if len(chars) != 1 {
		panic("read 1 byte must return array of length 1")
	}

	return chars[0], nil
}

func (lex *Lexer) advance() error {
	c, err := lex.buf.ReadByte()
	if err != nil {
		return err
	}

	lex.pos.byte++
	if c == '\n' {
		lex.pos.col = 1
		lex.pos.line++
		return nil
	}

	lex.pos.col++
	return nil
}

func (lex *Lexer) takeWhile(cond func(c byte) bool) (string, error) {
	buf := strings.Builder{}
	if c, err := lex.currentChar(); err != nil {
		return "", err
	} else if err := buf.WriteByte(c); err != nil {
		return "", err
	}

	if err := lex.advance(); err != nil {
		return "", err
	}

	c, err := lex.currentChar()
	if err != nil {
		return "", err
	}

	for cond(c) {
		if err := buf.WriteByte(c); err != nil {
			return "", err
		}

		if err := lex.advance(); err != nil {
			return "", err
		}

		c, err = lex.currentChar()
		if err != nil {
			return "", err
		}
	}

	return buf.String(), nil
}

func (lex *Lexer) integer() (*Token, error) {
	pos := *lex.pos
	numRepr, err := lex.takeWhile(isNumeric)
	if err != nil && err != io.EOF {
		return nil, err
	}

	num, err := strconv.ParseInt(numRepr, 10, 64)
	if err != nil {
		return nil, err
	}

	pos.length = len(numRepr)
	return NewInt(num, &pos), nil
}

func (lex *Lexer) identifier() (*Token, error) {
	pos := *lex.pos
	id, err := lex.takeWhile(func(c byte) bool {
		return isAlphaNumeric(c) || isUnderscore(c)
	})
	if err != nil && err != io.EOF {
		return nil, err
	}

	pos.length = len(id)
	switch id {
	case "array":
		return NewArray(&pos), nil
	case "break":
		return NewBreak(&pos), nil
	case "class":
		return NewClass(&pos), nil
	case "do":
		return NewDo(&pos), nil
	case "else":
		return NewElse(&pos), nil
	case "end":
		return NewEnd(&pos), nil
	case "extends":
		return NewExtends(&pos), nil
	case "for":
		return NewFor(&pos), nil
	case "function":
		return NewFunction(&pos), nil
	case "if":
		return NewIf(&pos), nil
	case "in":
		return NewIn(&pos), nil
	case "let":
		return NewLet(&pos), nil
	case "method":
		return NewMethod(&pos), nil
	case "nil":
		return NewNil(&pos), nil
	case "of":
		return NewOf(&pos), nil
	case "then":
		return NewThen(&pos), nil
	case "to":
		return NewTo(&pos), nil
	case "type":
		return NewType(&pos), nil
	case "var":
		return NewVar(&pos), nil
	case "while":
		return NewWhile(&pos), nil
	}

	return NewIdent(id, &pos), nil
}

func (lex *Lexer) simpleToken(f func(pos *Pos) *Token) (*Token, error) {
	pos := *lex.pos
	if err := lex.advance(); err != nil {
		return nil, err
	}

	pos.length = 1
	return f(&pos), nil
}

func (lex *Lexer) twoCharsToken(nextChar byte, nextToken, defaultToken func(pos *Pos) *Token) (*Token, error) {
	pos := *lex.pos
	if err := lex.advance(); err != nil {
		return nil, err
	}

	curChar, err := lex.currentChar()
	if err != nil && err != io.EOF {
		return nil, err
	}

	if curChar == nextChar {
		if err := lex.advance(); err != nil {
			return nil, err
		}

		pos.length = 2
		return nextToken(&pos), nil
	}

	pos.length = 1
	return defaultToken(&pos), nil
}

func (lex *Lexer) colonOrEqual() (*Token, error) {
	return lex.twoCharsToken('=', NewAssign, NewColon)
}

func (lex *Lexer) greaterOrGreaterEq() (*Token, error) {
	return lex.twoCharsToken('=', NewGreaterOrEqual, NewGreater)
}

func (lex *Lexer) lesserOrLesserEq() (*Token, error) {
	return lex.twoCharsToken('=', NewLesserOrEqual, NewLesser)
}

func (lex *Lexer) comment() error {
	depth := 1
	for depth > 0 {
		if err := lex.advance(); err != nil {
			if err == io.EOF {
				// TODO: move error to a new file
				return fmt.Errorf("unclosed comment %+v", lex.pos)
			}

			return err
		}

		curChar, err := lex.currentChar()
		if err != nil {
			return err
		}

		if curChar == '/' {
			if err := lex.advance(); err != nil {
				if err == io.EOF {
					return fmt.Errorf("unclosed comment %+v", lex.pos)
				}

				return err
			}

			curChar, err := lex.currentChar()
			if err != nil {
				return err
			}

			if curChar == '*' {
				depth++
			}
		} else if curChar == '*' {
			if err := lex.advance(); err != nil {
				if err == io.EOF {
					return fmt.Errorf("unclosed comment %+v", lex.pos)
				}

				return err
			}

			curChar, err := lex.currentChar()
			if err != nil {
				if err == io.EOF {
					return fmt.Errorf("unclosed comment %+v", lex.pos)
				}

				return err
			}

			if curChar == '/' {
				depth--
			}
		}
	}

	return nil
}

func (lex *Lexer) divOrCmt() (*Token, error) {
	pos := *lex.pos
	if err := lex.advance(); err != nil {
		return nil, err
	}

	curChar, err := lex.currentChar()
	if err != nil {
		return nil, err
	}

	if curChar == '*' {
		if err := lex.comment(); err != nil {
			return nil, err
		}

		if err := lex.advance(); err != nil && err != io.EOF {
			return nil, err
		}

		return nil, nil
	}

	pos.length = 1
	return NewDiv(&pos), nil
}

func (lex *Lexer) string() (*Token, error) {
	pos := *lex.pos
	buf := strings.Builder{}
	startByte := pos.byte
	if err := lex.advance(); err != nil {
		return nil, err
	}

	curChar, err := lex.currentChar()
	if err != nil {
		return nil, err
	}

	for curChar != '"' {
		if curChar == '\\' {
			pos := *lex.pos
			if err := lex.advance(); err != nil {
				if err == io.EOF {
					return nil, fmt.Errorf("unclosed string %+v", pos)
				}

				return nil, err
			}

			// TODO: bundle advance and currentChar into one funcDecl?
			curChar, err := lex.currentChar()
			if err != nil {
				return nil, err
			}

			if curChar == 'n' || curChar == 't' || curChar == '\\' || curChar == '"' {
				switch curChar {
				case 'n':
					buf.WriteByte('\n')
				case 't':
					buf.WriteByte('\t')
				case '\\':
					buf.WriteByte('\\')
				case '"':
					buf.WriteByte('"')
				}

				if err := lex.advance(); err != nil {
					if err == io.EOF {
						return nil, fmt.Errorf("unclosed string %+v", pos)
					}

					return nil, err
				}
			} else if isNumeric(curChar) {
				num, err := lex.takeWhile(isNumeric)
				if err != nil {
					if err == io.EOF {
						return nil, fmt.Errorf("unclosed string %+v", pos)
					}

					return nil, err
				}

				if len(num) != 3 {
					pos.length = len(num) + 1
					return nil, fmt.Errorf("invalid escape %+v", pos)
				} else {
					num, err := strconv.Atoi(num)
					if err != nil {
						panic("string of digits must be able to converted")
					}

					if num > 255 {
						pos.length = 4
						return nil, fmt.Errorf("invalid ASCII code %+v", pos)
					}

					buf.Write([]byte(strconv.Itoa(num)))
				}
			} else {
				pos.length = 2
				return nil, fmt.Errorf("invalid escape %+v", pos)
			}
		} else {
			if err := buf.WriteByte(curChar); err != nil {
				return nil, err
			}

			if err := lex.advance(); err != nil {
				if err == io.EOF {
					return nil, fmt.Errorf("unclosed string %+v", pos)
				}

				return nil, err
			}
		}

		curChar, err = lex.currentChar()
		if err != nil {
			if err == io.EOF {
				return nil, fmt.Errorf("unclosed string %+v", pos)
			}

			return nil, err
		}
	}

	if err := lex.advance(); err != nil && err != io.EOF {
		return nil, err
	}

	pos.length = lex.pos.length - startByte
	return NewStr(buf.String(), &pos), nil
}

func (lex *Lexer) Token() (*Token, error) {
	pos := *lex.pos
	curChar, err := lex.currentChar()
	if err != nil {
		if err == io.EOF {
			return NewEndOfFile(&pos), nil
		}

		return nil, err
	}

	if isSpace(curChar) {
		if err := lex.advance(); err != nil {
			return nil, err
		}

		return lex.Token()
	}

	// Numbers
	if isNumeric(curChar) {
		return lex.integer()
	}

	// Identifiers
	if isLower(curChar) || isUpper(curChar) || isUnderscore(curChar) {
		return lex.identifier()
	}

	switch curChar {
	case '=':
		return lex.simpleToken(NewEqual)
	case '&':
		return lex.simpleToken(NewAnd)
	case '|':
		return lex.simpleToken(NewOr)
	case '.':
		return lex.simpleToken(NewDot)
	case ',':
		return lex.simpleToken(NewComma)
	case ';':
		return lex.simpleToken(NewSemicolon)
	case '*':
		return lex.simpleToken(NewTimes)
	case '+':
		return lex.simpleToken(NewPlus)
	case '-':
		return lex.simpleToken(NewMinus)
	case '{':
		return lex.simpleToken(NewOpenCurly)
	case '}':
		return lex.simpleToken(NewCloseCurly)
	case '(':
		return lex.simpleToken(NewOpenParen)
	case ')':
		return lex.simpleToken(NewCloseParen)
	case '[':
		return lex.simpleToken(NewOpenBrac)
	case ']':
		return lex.simpleToken(NewCloseBrac)
	case ':':
		return lex.colonOrEqual()
	case '>':
		return lex.greaterOrGreaterEq()
	case '<':
		return lex.lesserOrLesserEq()
	case '!':
		pos := *lex.pos
		if err := lex.advance(); err != nil {
			return nil, err
		}

		curChar, err := lex.currentChar()
		if err != nil {
			return nil, err
		}

		if curChar != '=' {
			return nil, fmt.Errorf("invalid character %+v", pos)
		}

		if err := lex.advance(); err != nil {
			return nil, err
		}

		pos.length = 2
		return NewNotEqual(&pos), nil
	case '/':
		tok, err := lex.divOrCmt()
		if err != nil {
			return nil, err
		}

		if tok != nil {
			return tok, nil
		}

		return lex.Token()
	case '"':
		return lex.string()
	}

	pos.length = 1
	return nil, fmt.Errorf("invalid character %+v", pos)
}
