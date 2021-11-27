package tiger

import (
	"fmt"
)

type ErrUnexpectedTok struct {
	pos *Pos
}

func (err ErrUnexpectedTok) Error() string {
	return fmt.Sprintf("unexpected token %+v", err.pos)
}

type Parser struct {
	lexer     *Lexer
	lookahead *Token
	symbols   *Symbols
}

func NewParser(lexer *Lexer, symbols *Symbols) *Parser {
	return &Parser{
		lexer:     lexer,
		lookahead: nil,
		symbols:   symbols,
	}
}

func (p *Parser) peekNext() (*Token, error) {
	if err := p.nextToken(); err != nil {
		return nil, err
	}

	return p.peekToken(), nil
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

func (p *Parser) breakExp() (Exp, error) {
	pos := p.peekToken().pos
	if err := p.nextToken(); err != nil {
		return nil, err
	}

	return &BreakExp{
		pos: *pos,
	}, nil
}

func (p *Parser) forExp() (Exp, error) {
	pos := p.peekToken().pos
	// Pass "for"
	if err := p.nextToken(); err != nil {
		return nil, err
	}

	tok := p.peekToken()
	if tok.tok != "ident" {
		return nil, &ErrUnexpectedTok{pos: tok.pos}
	}

	varName, ok := tok.value.(string)
	if !ok {
		panic("type of identity must be a string")
	}

	sym := p.symbols.Symbol(varName)
	itVar := VarExp{
		sym: sym,
		pos: *tok.pos,
	}

	tok, err := p.peekNext()
	if err != nil {
		return nil, err
	}

	if tok.tok != ":=" {
		return nil, fmt.Errorf("unexpected token %+v", tok.pos)
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	start, err := p.Exp()
	if err != nil {
		return nil, err
	}

	tok = p.peekToken()
	if tok.tok != "to" {
		return nil, fmt.Errorf("unexpected token %+v", tok.pos)
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	end, err := p.Exp()
	if err != nil {
		return nil, err
	}

	tok = p.peekToken()
	if tok.tok != "do" {
		return nil, fmt.Errorf("unexpected token %+v", tok.pos)
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	body, err := p.Exp()
	if err != nil {
		return nil, err
	}

	startSymbol := p.symbols.Symbol(varName + "_start")
	endSymbol := p.symbols.Symbol(varName + "_end")

	declarations := []Declaration{
		&VarDecl{
			name: startSymbol,
			init: start,
		},
		&VarDecl{
			name: endSymbol,
			init: end,
		},
	}

	whileBody := IfExp{
		predicate: &OperExp{
			left: &itVar,
			op: &OperatorWithPos{
				op: Le,
			},
			right: &VarExp{
				sym: sym,
			},
		},
		then: &WhileExp{
			pred: &IntExp{
				val: 1,
			},
			body: &SequenceExp{
				seq: []Exp{
					body,
					&IfExp{
						predicate: &OperExp{
							left: &itVar,
							op: &OperatorWithPos{
								op: Lt,
							},
							right: &VarExp{sym: endSymbol},
						},
						then: &AssignExp{
							exp: &OperExp{
								left:  &itVar,
								op:    &OperatorWithPos{op: Plus},
								right: &IntExp{val: 1},
							},
							variable: &itVar,
						},
						els: &BreakExp{},
					},
				},
			},
		},
		els: nil,
	}

	return &LetExp{
		body:  &whileBody,
		decls: declarations,
		pos:   *pos,
	}, nil
}

func (p *Parser) ifExp() (Exp, error) {
	pos := p.peekToken().pos
	if err := p.nextToken(); err != nil {
		return nil, err
	}

	test, err := p.Exp()
	if err != nil {
		return nil, err
	}

	tok := p.peekToken()
	if tok.tok != "then" {
		return nil, fmt.Errorf("unexpected token %+v", tok.pos)
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	then, err := p.Exp()
	if err != nil {
		return nil, err
	}

	var elseExp Exp
	tok = p.peekToken()
	if tok.tok == "else" {
		if err := p.nextToken(); err != nil {
			return nil, err
		}

		elseExp, err = p.Exp()
		if err != nil {
			return nil, err
		}
	}

	return &IfExp{
		predicate: test,
		then:      then,
		els:       elseExp,
		pos:       *pos,
	}, nil
}

func (p *Parser) funcArgs() ([]Exp, error) {
	args := make([]Exp, 0)
	for true {
		tok := p.peekToken()
		if tok.tok == ")" {
			break
		}

		arg, err := p.Exp()
		if err != nil {
			return nil, err
		}

		args = append(args, arg)
		tok = p.peekToken()
		if tok.tok != "," && tok.tok != ")" {
			return nil, fmt.Errorf("unexpected token %+v", tok.pos)
		}

		if tok.tok == "," {
			if err := p.nextToken(); err != nil {
				return nil, err
			}
		}
	}

	return args, nil
}

func (p *Parser) oneField() (*RecordField, error) {
	tok := p.peekToken()
	name := tok.value.(string)
	sym := p.symbols.Symbol(name)
	pos := tok.pos

	tok, err := p.peekNext()
	if err != nil {
		return nil, err
	}

	if tok.tok != "=" {
		return nil, fmt.Errorf("unexpected token %+v", tok.pos)
	}

	exp, err := p.Exp()
	if err != nil {
		return nil, err
	}

	return &RecordField{
		expr:  exp,
		ident: sym,
		pos:   *pos,
	}, nil
}

func (p *Parser) createRecord(ty Symbol, pos *Pos) (Exp, error) {
	if err := p.nextToken(); err != nil {
		return nil, err
	}

	field, err := p.oneField()
	if err != nil {
		return nil, err
	}

	fields := []*RecordField{field}
	for true {
		tok := p.peekToken()
		if tok.tok != "," {
			break
		}

		if err := p.nextToken(); err != nil {
			return nil, err
		}

		field, err := p.oneField()
		if err != nil {
			return nil, err
		}

		fields = append(fields, field)
	}

	return &RecordExp{
		fields: fields,
		typ:    ty,
		pos:    *pos,
	}, err
}

func (p *Parser) funcOrIdent() (Exp, error) {
	tok := p.peekToken()
	name := tok.value.(string)
	sym := p.symbols.Symbol(name)

	tok, err := p.peekNext()
	if err != nil {
		return nil, err
	}

	switch tok.tok {
	case "(":
		if err := p.nextToken(); err != nil {
			return nil, err
		}

		args, err := p.funcArgs()
		if err != nil {
			return nil, err
		}

		if p.peekToken().tok == ")" {
			if err := p.nextToken(); err != nil {
				return nil, err
			}
		}

		return &CallExp{
			function: sym,
			args:     args,
			pos:      *tok.pos,
		}, nil
	case "{":
		return p.createRecord(sym, tok.pos)
	default:
		varExp := &VarExp{
			sym: sym,
			pos: *tok.pos,
		}

		return p.lvalueOrAssign(varExp)
	}
}

func (p *Parser) subscript(firstExp Exp) (Exp, error) {
	if err := p.nextToken(); err != nil {
		return nil, err
	}

	exp, err := p.Exp()
	if err != nil {
		return nil, err
	}

	tok := p.peekToken()
	if tok.tok != "]" {
		return nil, ErrUnexpectedTok{pos: tok.pos}
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	subscriptExp := SubscriptExp{
		subscript: exp,
		firstExp:  firstExp,
		pos:       exp.ExpPos(),
	}

	return p.lvalue(&subscriptExp)
}

func (p *Parser) fieldExp(exp Exp) (Exp, error) {
	if err := p.nextToken(); err != nil {
		return nil, err
	}

	tok := p.peekToken()
	if tok.tok != "ident" {
		return nil, ErrUnexpectedTok{tok.pos}
	}

	fieldName := p.symbols.Symbol(tok.value.(string))
	return &FieldExp{
		firstExp:  exp,
		fieldName: fieldName,
		pos:       *tok.pos,
	}, nil
}

func (p *Parser) lvalue(exp Exp) (Exp, error) {
	switch p.peekToken().tok {
	case "[":
		return p.subscript(exp)
	case ".":
		return p.fieldExp(exp)
	default:
		return exp, nil
	}
}

func (p *Parser) array(exp *VarExp, size Exp) (Exp, error) {
	if err := p.nextToken(); err != nil {
		return nil, err
	}

	init, err := p.Exp()
	if err != nil {
		return nil, err
	}

	return &ArrExp{
		init: init,
		size: size,
		typ:  exp.sym,
		pos:  exp.pos,
	}, nil
}

func (p *Parser) lvalueOrAssign(exp Exp) (Exp, error) {
	varExp, err := p.lvalue(exp)
	if err != nil {
		return nil, err
	}

	tok := p.peekToken()
	switch tok.tok {
	case "of":
		if exp, ok := varExp.(*SubscriptExp); ok {
			if firstExp, ok := exp.firstExp.(*VarExp); ok {
				return p.array(firstExp, exp.subscript)
			}

			pos := exp.ExpPos()
			return nil, ErrUnexpectedTok{&pos}
		}

		pos := varExp.ExpPos()
		return nil, ErrUnexpectedTok{&pos}
	case ":=":
		if err := p.nextToken(); err != nil {
			return nil, err
		}

		exp, err := p.Exp()
		if err != nil {
			return nil, err
		}

		return &AssignExp{
			exp:      exp,
			variable: varExp,
		}, nil
	default:
		return varExp, nil
	}
}

func (p *Parser) intConst() (Exp, error) {
	tok := p.peekToken()
	if err := p.nextToken(); err != nil {
		return nil, err
	}

	return &IntExp{
		val: tok.value.(int64),
		pos: *tok.pos,
	}, nil
}

func (p *Parser) optionalType() (Symbol, error) {
	tok := p.peekToken()
	if tok.tok != ":" {
		return 0, nil
	}

	tok, err := p.peekNext()
	if err != nil {
		return 0, err
	}

	if tok.tok != "ident" {
		return 0, fmt.Errorf("unexpected token %+v", tok.pos)
	}

	return p.symbols.Symbol(tok.value.(string)), p.nextToken()
}

func (p *Parser) fieldDecl() (*Field, error) {
	tok := p.peekToken()
	if tok.tok != "ident" {
		return nil, fmt.Errorf("unexpected token %+v", tok.pos)
	}

	pos := tok.pos

	varSym := p.symbols.Symbol(tok.value.(string))
	tok, err := p.peekNext()
	if err != nil {
		return nil, err
	}

	if tok.tok != ":" {
		return nil, fmt.Errorf("unexpected token %+v", tok.pos)
	}

	tok, err = p.peekNext()
	if err != nil {
		return nil, err
	}

	if tok.tok != "ident" {
		return nil, fmt.Errorf("unexpected token %+v", tok.pos)
	}

	typSym := p.symbols.Symbol(tok.value.(string))
	return &Field{
		name:   varSym,
		escape: false,
		typ:    typSym,
		pos:    *pos,
	}, nil
}

func (p *Parser) fields(endTok string) ([]*Field, error) {
	fields := make([]*Field, 0)
	for true {
		tok := p.peekToken()
		if tok.tok == endTok {
			return fields, nil
		}

		if tok.tok == "," {
			if err := p.nextToken(); err != nil {
				return nil, err
			}

			continue
		}

		field, err := p.fieldDecl()
		if err != nil {
			return nil, err
		}

		fields = append(fields, field)
	}

	return fields, nil
}

func (p *Parser) funcDecl() (*FuncDecl, error) {
	tok := p.peekToken()
	funcPos := tok.pos

	ident, err := p.peekNext()
	if err != nil {
		return nil, err
	}

	if ident.tok != "ident" {
		return nil, fmt.Errorf("unexpected token %+v", ident.pos)
	}

	functionNameSym := p.symbols.Symbol(ident.value.(string))

	openCur, err := p.peekNext()
	if err != nil {
		return nil, err
	}

	if openCur.tok != "(" {
		return nil, fmt.Errorf("unexpected token %+v", ident.pos)
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	params, err := p.fields(")")
	if err != nil {
		return nil, err
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	ty, err := p.optionalType()
	if err != nil {
		return nil, err
	}

	tok = p.peekToken()
	if tok.tok != "=" {
		return nil, fmt.Errorf("unexpected token %+v", tok.pos)
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	body, err := p.Exp()
	if err != nil {
		return nil, err
	}

	return &FuncDecl{
		name:     functionNameSym,
		params:   params,
		resultTy: ty,
		body:     body,
		pos:      *funcPos,
	}, nil
}

//func (p *Parser) funcDecls() (Declaration, error) {
//	functions := make(FuncDecls, 0)
//	for true {
//		if tok := p.peekToken(); tok.tok != "funcDecl" {
//			break
//		}
//
//		fun, err := p.funcDecl()
//		if err != nil {
//			return nil, err
//		}
//
//		functions = append(functions, fun)
//	}
//
//	return functions, nil
//}
//
func (p *Parser) arrayTy() (Ty, error) {
	pos := p.peekToken().pos

	tok, err := p.peekNext()
	if err != nil {
		return nil, err
	}

	if tok.tok != "of" {
		return nil, fmt.Errorf("unexpected token %+v", tok.pos)
	}

	tok, err = p.peekNext()
	if err != nil {
		return nil, err
	}

	if tok.tok != "ident" {
		return nil, fmt.Errorf("unexpected token %+v", tok.pos)
	}

	sym := p.symbols.Symbol(tok.value.(string))
	return &ArrayTy{
		ty:  sym,
		pos: *pos,
	}, nil
}

func (p *Parser) recTy() (Ty, error) {
	pos := p.peekToken().pos
	fields, err := p.fields("}")
	if err != nil {
		return nil, err
	}

	tok, err := p.peekNext()
	if err != nil {
		return nil, err
	}

	if tok.tok != "}" {
		return nil, fmt.Errorf("unexpected token %+v", tok.pos)
	}

	return &RecordTy{
		ty:  fields,
		pos: *pos,
	}, nil
}

func (p *Parser) nameTy() (Ty, error) {
	tyName := p.peekToken().value.(string)
	pos := p.peekToken().pos
	tySym := p.symbols.Symbol(tyName)
	return &NameTy{
		ty:  tySym,
		pos: *pos,
	}, nil
}

func (p *Parser) tyDecl() (Declaration, error) {
	pos := p.peekToken().pos
	if err := p.nextToken(); err != nil {
		return nil, err
	}

	tok := p.peekToken()
	if tok.tok != "ident" {
		return nil, fmt.Errorf("unexpected token %+v", tok.pos)
	}

	tyName := p.symbols.Symbol(tok.value.(string))

	tok, err := p.peekNext()
	if err != nil {
		return nil, err
	}

	if tok.tok != "=" {
		return nil, fmt.Errorf("unexpected token %+v", tok.pos)
	}

	tok, err = p.peekNext()
	if err != nil {
		return nil, err
	}

	var (
		ty Ty
	)

	switch tok.tok {
	case "array":
		ty, err = p.arrayTy()
	case "{":
		ty, err = p.recTy()
	case "ident":
		ty, err = p.nameTy()
	default:
		return nil, fmt.Errorf("unexpected token %+v", tok.pos)
	}

	return &TypeDecl{
		tyName: tyName,
		typ:    ty,
		pos:    *pos,
	}, nil
}

func (p *Parser) varDecl() (Declaration, error) {
	pos := p.peekToken().pos
	tok, err := p.peekNext()
	if err != nil {
		return nil, err
	}

	if tok.tok != "ident" {
		return nil, &ErrUnexpectedTok{tok.pos}
	}

	varName := p.symbols.Symbol(tok.value.(string))
	ty, err := p.optionalType()
	if err != nil {
		return nil, err
	}

	tok, err = p.peekNext()
	if err != nil {
		return nil, err
	}

	if tok.tok != ":=" {
		return nil, ErrUnexpectedTok{tok.pos}
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	init, err := p.Exp()
	if err != nil {
		return nil, err
	}

	return &VarDecl{
		name:   varName,
		escape: false,
		typ:    ty,
		init:   init,
		pos:    *pos,
	}, nil
}

func (p *Parser) declarations() (Declaration, error) {
	tok := p.peekToken()
	switch tok.tok {
	case "function":
		return p.funcDecl()
	case "type":
		return p.tyDecl()
	case "var":
		return p.varDecl()
	default:
		return nil, &ErrUnexpectedTok{tok.pos}
	}
}

func (p *Parser) letExp() (Exp, error) {
	pos := p.peekToken().pos
	if err := p.nextToken(); err != nil {
		return nil, err
	}

	decl, err := p.declarations()
	if err != nil {
		return nil, err
	}

	decls := []Declaration{decl}
LOOP:
	for {
		switch p.peekToken().tok {
		case "function":
			fallthrough
		case "type":
			fallthrough
		case "var":
			decl, err := p.declarations()
			if err != nil {
				return nil, err
			}

			decls = append(decls, decl)
		default:
			break LOOP
		}
	}

	if p.peekToken().tok != "in" {
		return nil, &ErrUnexpectedTok{pos: p.peekToken().pos}
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	exp, err := p.Exp()
	if err != nil {
		return nil, err
	}

	firstExpPos := exp.ExpPos()

	exps := []Exp{exp}
	for true {
		if p.peekToken().tok != ";" {
			break
		}

		if err := p.nextToken(); err != nil {
			return nil, err
		}

		exp, err := p.Exp()
		if err != nil {
			return nil, err
		}

		exps = append(exps, exp)
	}

	var seqExp *SequenceExp
	if len(exps) == 1 {
		if e, ok := exps[0].(*SequenceExp); ok {
			seqExp = e
		} else {
			seqExp = &SequenceExp{
				seq: exps,
				pos: firstExpPos,
			}
		}
	} else {
		seqExp = &SequenceExp{
			seq: exps,
			pos: firstExpPos,
		}
	}

	if p.peekToken().tok != "end" {
		return nil, ErrUnexpectedTok{pos: p.peekToken().pos}
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	return &LetExp{
		body:  seqExp,
		decls: decls,
		pos:   *pos,
	}, nil
}

func (p *Parser) nilExp() (Exp, error) {
	pos := p.peekToken().pos
	if err := p.nextToken(); err != nil {
		return nil, err
	}

	return &NilExp{*pos}, nil
}

func (p *Parser) seqExp() (Exp, error) {
	pos := p.peekToken().pos
	if err := p.nextToken(); err != nil {
		return nil, err
	}

	exp, err := p.Exp()
	if err != nil {
		return nil, err
	}

	seqExp := []Exp{exp}
	for true {
		if p.peekToken().tok != ";" {
			break
		}

		if err := p.nextToken(); err != nil {
			return nil, err
		}

		exp, err := p.Exp()
		if err != nil {
			return nil, err
		}

		seqExp = append(seqExp, exp)
	}

	if p.peekToken().tok != ")" {
		return nil, ErrUnexpectedTok{p.peekToken().pos}
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	return &SequenceExp{
		seq: seqExp,
		pos: *pos,
	}, nil
}

func (p *Parser) strExp() (Exp, error) {
	tok := p.peekToken()
	if err := p.nextToken(); err != nil {
		return nil, err
	}

	return &StrExp{
		str: tok.value.(string),
		pos: *tok.pos,
	}, nil
}

func (p *Parser) whileExp() (Exp, error) {
	pos := p.peekToken().pos
	if err := p.nextToken(); err != nil {
		return nil, err
	}

	test, err := p.Exp()
	if err != nil {
		return nil, err
	}

	if p.peekToken().tok != "do" {
		return nil, ErrUnexpectedTok{p.peekToken().pos}
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	body, err := p.Exp()
	if err != nil {
		return nil, err
	}

	return &WhileExp{
		pred: test,
		body: body,
		pos:  *pos,
	}, nil
}

func (p *Parser) primaryExp() (Exp, error) {
	tok := p.peekToken()
	switch tok.tok {
	case "break":
		return p.breakExp()
	case "for":
		return p.forExp()
	case "if":
		return p.ifExp()
	case "ident":
		return p.funcOrIdent()
	case "int":
		return p.intConst()
	case "let":
		return p.letExp()
	case "nil":
		return p.nilExp()
	case "(":
		return p.seqExp()
	case "str":
		return p.strExp()
	case "while":
		return p.whileExp()
	default:
		return nil, ErrUnexpectedTok{tok.pos}
	}
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
			left: &IntExp{
				val: 0,
				pos: *pos,
			},
			op: &OperatorWithPos{
				op:  Minus,
				pos: *pos,
			},
			right: exp,
		}, nil
	}

	return p.primaryExp()
}

func (p *Parser) mulExp() (Exp, error) {
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
				op: &OperatorWithPos{
					op:  Mul,
					pos: *op.pos,
				},
				right: nextExp,
			}
		} else {
			exp = &OperExp{
				left: exp,
				op: &OperatorWithPos{
					op:  Div,
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
				op: &OperatorWithPos{
					op:  Plus,
					pos: *op.pos,
				},
				right: nextExp,
			}
		} else {
			exp = &OperExp{
				left: exp,
				op: &OperatorWithPos{
					op:  Minus,
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
		op: &OperatorWithPos{
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
		op: &OperatorWithPos{
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
		op: &OperatorWithPos{
			op:  Or,
			pos: *op.pos,
		},
		right: right,
	}, nil
}

func (p *Parser) Parse() (Exp, error) {
	if err := p.nextToken(); err != nil {
		return nil, err
	}

	return p.Exp()
}

func (p *Parser) Exp() (Exp, error) {
	return p.orExp()
}
