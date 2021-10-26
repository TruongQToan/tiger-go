package tiger

type Parser struct {
	lexer     Lexer
	lookahead *Token
	symbols   Symbols
}

func NewParser(lexer Lexer, symbols Symbols) *Parser {
	return &Parser{
		lexer:     lexer,
		lookahead: nil,
		symbols:   symbols,
	}
}

func (p *Parser) peekToken() *Token {
	return p.lookahead
}

func (p *Parser) nextToken() error {
	var err error
	p.lookahead, err = p.lexer.Token()
	if err != nil {
		return err
	}

	return nil
}

func (p *Parser) primaryExp() (Exp, error) {
	// TODO: finish this function
	return nil, nil
}

func (p *Parser) unaryExp() (Exp, error) {
	tok := p.peekToken()
	if tok.tok == "-" {
		pos := tok.pos
		if err := p.nextToken(); err != nil {
			return nil, err
		}

		exp, err := p.unaryExp()
		if err != nil {
			return nil, err
		}

		return &OperExp{
			left:  &IntExp{
				val: 0,
				pos: *pos,
			},
			op: OperatorWithPos{
				op: Minus,
				pos: *pos,
			},
			right: exp,
		}, nil
	}

	return p.primaryExp()
}

func (p *Parser) mulExp() (Exp, error) {
	// Can be in the format 2+3-4
	exp, err := p.unaryExp()
	if err != nil {
		return nil, err
	}

	for {
		op := p.peekToken()
		if op.tok != "*" && op.tok != "/" {
			break
		}

		if err := p.nextToken(); err != nil {
			return nil, err
		}

		nextExp, err := p.unaryExp()
		if err != nil {
			return nil, err
		}

		if op.tok == "*" {
			exp = &OperExp{
				left: exp,
				op: OperatorWithPos{
					op: Mul,
					pos: *op.pos,
				},
				right: nextExp,
			}
		} else {
			exp = &OperExp{
				left: exp,
				op: OperatorWithPos{
					op: Div,
					pos: *op.pos,
				},
				right: nextExp,
			}
		}
	}

	return exp, nil
}

func (p *Parser) addExp() (Exp, error) {
	// Can be in the format 2+3-4
	exp, err := p.mulExp()
	if err != nil {
		return nil, err
	}

	for {
		op := p.peekToken()
		if op.tok != "+" && op.tok != "-" {
			break
		}

		if err := p.nextToken(); err != nil {
			return nil, err
		}

		nextExp, err := p.mulExp()
		if err != nil {
			return nil, err
		}

		if op.tok == "+" {
			exp = &OperExp{
				left: exp,
				op: OperatorWithPos{
					op: Plus,
					pos: *op.pos,
				},
				right: nextExp,
			}
		} else {
			exp = &OperExp{
				left: exp,
				op: OperatorWithPos{
					op: Minus,
					pos: *op.pos,
				},
				right: nextExp,
			}
		}
	}

	return exp, nil
}

func (p *Parser) relationalExp() (Exp, error) {
	left, err := p.addExp()
	if err != nil {
		return nil, err
	}

	nextToken := p.peekToken()
	if nextToken.tok != "=" && nextToken.tok != ">" && nextToken.tok != "<" && nextToken.tok != ">=" && nextToken.tok != "<=" && nextToken.tok != "!=" {
		return left, nil
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	right, err := p.addExp()
	if err != nil {
		return nil, err
	}

	var op Operator
	switch nextToken.tok {
	case "=":
		op = Eq
	case ">":
		op = Gt
	case ">=":
		op = Ge
	case "<":
		op = Lt
	case "<=":
		op = Le
	case "!=":
		op = Neq
	}
	return &OperExp{
		left: left,
		op: OperatorWithPos{
			op:  op,
			pos: *nextToken.pos,
		},
		right: right,
	}, nil
}

func (p *Parser) andExp() (Exp, error) {
	left, err := p.relationalExp()
	if err != nil {
		return nil, err
	}

	op := p.peekToken()
	if op.tok != "&" {
		return left, nil
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	right, err := p.relationalExp()
	if err != nil {
		return nil, err
	}

	return &OperExp{
		left: left,
		op: OperatorWithPos{
			op:  And,
			pos: *op.pos,
		},
		right: right,
	}, nil
}

func (p *Parser) orExp() (Exp, error) {
	left, err := p.andExp()
	if err != nil {
		return nil, err
	}

	op := p.peekToken()
	if op.tok != "|" {
		return left, nil
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	right, err := p.andExp()
	if err != nil {
		return nil, err
	}

	return &OperExp{
		left: left,
		op: OperatorWithPos{
			op:  Or,
			pos: *op.pos,
		},
		right: right,
	}, nil
}

func (p *Parser) Exp() (Exp, error) {
	return p.orExp()
}
