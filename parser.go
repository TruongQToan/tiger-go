package main

type Parser struct {
	lexer     *Lexer
	lookahead *Token
	strings   *Strings
}

func NewParser(lexer *Lexer, strs *Strings) *Parser {
	return &Parser{
		lexer:     lexer,
		lookahead: nil,
		strings:   strs,
	}
}

func (p *Parser) peekNext() (*Token, error) {
	if err := p.nextToken(); err != nil {
		return nil, err
	}

	if p.lookahead.IsEof() {
		return nil, unexpectedEofErr(p.lookahead.pos)
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
		pos: pos,
	}, nil
}

func (p *Parser) forExp() (Exp, error) {
	pos := p.peekToken().pos
	// Pass "for"
	if err := p.nextToken(); err != nil {
		return nil, err
	}

	if p.lookahead.IsEof() {
		return nil, unexpectedEofErr(p.lookahead.pos)
	}

	tok := p.peekToken()
	if tok.tok != "ident" {
		return nil, unexpectedTokErr(tok.pos)
	}

	varName, ok := tok.value.(string)
	if !ok {
		panic("type of identity must be a string")
	}

	sym := p.strings.Symbol(varName)
	itVar := &SimpleVar{
		symbol: sym,
		pos:    tok.pos,
	}

	tok, err := p.peekNext()
	if err != nil {
		return nil, err
	}

	if tok.tok != ":=" {
		return nil, unexpectedTokErr(tok.pos)
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	if p.lookahead.IsEof() {
		return nil, unexpectedEofErr(p.lookahead.pos)
	}

	start, err := p.Exp()
	if err != nil {
		return nil, err
	}

	tok = p.peekToken()
	if tok.tok != "to" {
		return nil, unexpectedTokErr(tok.pos)
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	if p.lookahead.IsEof() {
		return nil, unexpectedEofErr(p.lookahead.pos)
	}

	end, err := p.Exp()
	if err != nil {
		return nil, err
	}

	tok = p.peekToken()
	if tok.tok != "do" {
		return nil, unexpectedTokErr(tok.pos)
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	if p.lookahead.IsEof() {
		return nil, unexpectedEofErr(p.lookahead.pos)
	}

	body, err := p.Exp()
	if err != nil {
		return nil, err
	}

	endSymbol := p.strings.Symbol(varName + "_end")
	declarations := []Declaration{
		&VarDecl{
			name: sym,
			init: start,
		},
		&VarDecl{
			name: endSymbol,
			init: end,
		},
	}

	whileBody := IfExp{
		predicate: &OperExp{
			left: &VarExp{itVar},
			op:   Le,
			right: &VarExp{
				&SimpleVar{symbol: endSymbol},
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
							left:  &VarExp{itVar},
							op:    Lt,
							right: &VarExp{&SimpleVar{symbol: endSymbol}},
						},
						then: &AssignExp{
							exp: &OperExp{
								left:  &VarExp{itVar},
								op:    Plus,
								right: &IntExp{val: 1},
							},
							variable: itVar,
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
		pos:   pos,
	}, nil
}

func (p *Parser) ifExp() (Exp, error) {
	pos := p.peekToken().pos
	if err := p.nextToken(); err != nil {
		return nil, err
	}

	if p.lookahead.IsEof() {
		return nil, unexpectedEofErr(p.lookahead.pos)
	}

	test, err := p.Exp()
	if err != nil {
		return nil, err
	}

	tok := p.peekToken()
	if tok.tok != "then" {
		return nil, unexpectedTokErr(tok.pos)
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	if p.lookahead.IsEof() {
		return nil, unexpectedEofErr(p.lookahead.pos)
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

		if p.lookahead.IsEof() {
			return nil, unexpectedEofErr(p.lookahead.pos)
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
		pos:       pos,
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
			return nil, unexpectedTokErr(tok.pos)
		}

		if tok.tok == "," {
			if err := p.nextToken(); err != nil {
				return nil, err
			}

			if p.lookahead.IsEof() {
				return nil, unexpectedEofErr(p.lookahead.pos)
			}
		}
	}

	return args, nil
}

func (p *Parser) oneField() (*RecordField, error) {
	tok := p.peekToken()
	name := tok.value.(string)
	sym := p.strings.Symbol(name)
	pos := tok.pos

	tok, err := p.peekNext()
	if err != nil {
		return nil, err
	}

	if tok.tok != "=" {
		return nil, unexpectedTokErr(tok.pos)
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	if p.lookahead.IsEof() {
		return nil, unexpectedEofErr(p.lookahead.pos)
	}

	exp, err := p.Exp()
	if err != nil {
		return nil, err
	}

	return &RecordField{
		expr:  exp,
		ident: sym,
		pos:   pos,
	}, nil
}

func (p *Parser) createRecord(ty Symbol, pos Pos) (Exp, error) {
	if err := p.nextToken(); err != nil {
		return nil, err
	}

	if p.lookahead.IsEof() {
		return nil, unexpectedEofErr(p.lookahead.pos)
	}

	field, err := p.oneField()
	if err != nil {
		return nil, err
	}

	fields := []*RecordField{field}
	for true {
		tok := p.peekToken()
		if tok.tok != "," {
			if tok.tok == "}" {
				if err := p.nextToken(); err != nil {
					return nil, err
				}

				if p.lookahead.IsEof() {
					return nil, unexpectedEofErr(p.lookahead.pos)
				}

				break
			} else {
				return nil, unexpectedTokErr(tok.pos)
			}
		}

		if err := p.nextToken(); err != nil {
			return nil, err
		}

		if p.lookahead.IsEof() {
			return nil, unexpectedEofErr(p.lookahead.pos)
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
		pos:    pos,
	}, err
}

func (p *Parser) funcOrIdent() (Exp, error) {
	tok := p.peekToken()
	name := tok.value.(string)
	sym := p.strings.Symbol(name)

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	if p.lookahead.IsEof() {
		return &VarExp{v: &SimpleVar{
			symbol: sym,
			pos:    tok.pos,
		}}, nil
	}

	switch tok.tok {
	case "(":
		if err := p.nextToken(); err != nil {
			return nil, err
		}

		if p.lookahead.IsEof() {
			return nil, unexpectedEofErr(p.lookahead.pos)
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
			pos:      tok.pos,
		}, nil
	case "{":
		return p.createRecord(sym, tok.pos)
	default:
		return p.lvalueOrAssign(&SimpleVar{symbol: sym, pos: tok.pos})
	}
}

func (p *Parser) lvalueSubscript(v Var) (Var, error) {
	if err := p.nextToken(); err != nil {
		return nil, err
	}

	if p.lookahead.IsEof() {
		return nil, unexpectedEofErr(p.lookahead.pos)
	}

	exp, err := p.Exp()
	if err != nil {
		return nil, err
	}

	tok := p.peekToken()
	if tok.tok != "]" {
		return nil, unexpectedTokErr(tok.pos)
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	return &SubscriptionVar{
		variable: v,
		exp:      exp,
		pos:      exp.ExpPos(),
	}, nil
}

func (p *Parser) lvalueField(v Var) (Var, error) {
	if err := p.nextToken(); err != nil {
		return nil, err
	}

	if p.lookahead.IsEof() {
		return nil, unexpectedEofErr(p.lookahead.pos)
	}

	tok := p.peekToken()
	if tok.tok != "ident" {
		return nil, unexpectedTokErr(tok.pos)
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	return &FieldVar{
		variable: v,
		field:    p.strings.Symbol(tok.value.(string)),
		pos:      v.VarPos(),
	}, nil
}

func (p *Parser) lvalue(v Var) (Var, error) {
	switch p.peekToken().tok {
	case "[":
		return p.lvalueSubscript(v)
	case ".":
		return p.lvalueField(v)
	default:
		return v, nil
	}
}

func (p *Parser) array(t Symbol, size Exp, pos Pos) (Exp, error) {
	if err := p.nextToken(); err != nil {
		return nil, err
	}

	if p.lookahead.IsEof() {
		return nil, unexpectedEofErr(p.lookahead.pos)
	}

	init, err := p.Exp()
	if err != nil {
		return nil, err
	}

	return &ArrExp{
		init: init,
		size: size,
		typ:  t,
		pos:  pos,
	}, nil
}

func (p *Parser) lvalueOrAssign(v Var) (Exp, error) {
	v1, err := p.lvalue(v)
	if err != nil {
		return nil, err
	}

	tok := p.peekToken()
	switch tok.tok {
	case "of":
		v2, ok := v1.(*SubscriptionVar)
		if !ok {
			return nil, unexpectedTokErr(v2.pos)
		}

		v3, ok := v2.variable.(*SimpleVar)
		if !ok {
			return nil, unexpectedTokErr(v3.pos)
		}

		return p.array(v3.symbol, v2.exp, v.VarPos())
	case ":=":
		if err := p.nextToken(); err != nil {
			return nil, err
		}

		if p.lookahead.IsEof() {
			return nil, unexpectedEofErr(p.lookahead.pos)
		}

		exp, err := p.Exp()
		if err != nil {
			return nil, err
		}

		return &AssignExp{
			exp:      exp,
			variable: v1,
		}, nil
	default:
		return &VarExp{v: v1}, nil
	}
}

func (p *Parser) intConst() (Exp, error) {
	tok := p.peekToken()
	if err := p.nextToken(); err != nil {
		return nil, err
	}

	return &IntExp{
		val: tok.value.(int64),
		pos: tok.pos,
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
		return 0, unexpectedTokErr(tok.pos)
	}

	return p.strings.Symbol(tok.value.(string)), p.nextToken()
}

func (p *Parser) fieldDecl() (*Field, error) {
	tok := p.peekToken()
	if tok.tok != "ident" {
		return nil, unexpectedTokErr(tok.pos)
	}

	pos := tok.pos

	varSym := p.strings.Symbol(tok.value.(string))
	tok, err := p.peekNext()
	if err != nil {
		return nil, err
	}

	if tok.tok != ":" {
		return nil, unexpectedTokErr(tok.pos)
	}

	tok, err = p.peekNext()
	if err != nil {
		return nil, err
	}

	if tok.tok != "ident" {
		return nil, unexpectedTokErr(tok.pos)
	}

	typSym := p.strings.Symbol(tok.value.(string))
	return &Field{
		name:   varSym,
		escape: false,
		typ:    typSym,
		pos:    pos,
	}, p.nextToken()
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

			if p.lookahead.IsEof() {
				return nil, unexpectedEofErr(p.lookahead.pos)
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
		return nil, unexpectedTokErr(ident.pos)
	}

	functionNameSym := p.strings.Symbol(ident.value.(string))

	openCur, err := p.peekNext()
	if err != nil {
		return nil, err
	}

	if openCur.tok != "(" {
		return nil, unexpectedTokErr(openCur.pos)
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	if p.lookahead.IsEof() {
		return nil, unexpectedEofErr(p.lookahead.pos)
	}

	params, err := p.fields(")")
	if err != nil {
		return nil, err
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	if p.lookahead.IsEof() {
		return nil, unexpectedEofErr(p.lookahead.pos)
	}
	ty, err := p.optionalType()
	if err != nil {
		return nil, err
	}

	tok = p.peekToken()
	if tok.tok != "=" {
		return nil, unexpectedTokErr(tok.pos)
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	if p.lookahead.IsEof() {
		return nil, unexpectedEofErr(p.lookahead.pos)
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
		pos:      funcPos,
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
		return nil, unexpectedTokErr(tok.pos)
	}

	tok, err = p.peekNext()
	if err != nil {
		return nil, err
	}

	if tok.tok != "ident" {
		return nil, unexpectedTokErr(tok.pos)
	}

	sym := p.strings.Symbol(tok.value.(string))
	return &ArrayTy{
		ty:  sym,
		pos: pos,
	}, nil
}

func (p *Parser) recTy() (Ty, error) {
	pos := p.peekToken().pos
	if err := p.nextToken(); err != nil {
		return nil, err
	}

	if p.lookahead.IsEof() {
		return nil, unexpectedEofErr(p.lookahead.pos)
	}

	fields, err := p.fields("}")
	if err != nil {
		return nil, err
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	if p.lookahead.IsEof() {
		return nil, unexpectedEofErr(p.lookahead.pos)
	}

	return &RecordTy{
		ty:  fields,
		pos: pos,
	}, nil
}

func (p *Parser) nameTy() (Ty, error) {
	tyName := p.peekToken().value.(string)
	pos := p.peekToken().pos
	tySym := p.strings.Symbol(tyName)
	return &NameTy{
		ty:  tySym,
		pos: pos,
	}, nil
}

func (p *Parser) tyDecl() (Declaration, error) {
	pos := p.peekToken().pos
	if err := p.nextToken(); err != nil {
		return nil, err
	}

	if p.lookahead.IsEof() {
		return nil, unexpectedEofErr(p.lookahead.pos)
	}

	tok := p.peekToken()
	if tok.tok != "ident" {
		return nil, unexpectedTokErr(tok.pos)
	}

	tyName := p.strings.Symbol(tok.value.(string))

	tok, err := p.peekNext()
	if err != nil {
		return nil, err
	}

	if tok.tok != "=" {
		return nil, unexpectedTokErr(tok.pos)
	}

	tok, err = p.peekNext()
	if err != nil {
		return nil, err
	}

	var ty Ty

	switch tok.tok {
	case "array":
		ty, err = p.arrayTy()
	case "{":
		ty, err = p.recTy()
	case "ident":
		ty, err = p.nameTy()
	default:
		return nil, unexpectedTokErr(tok.pos)
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	if p.lookahead.IsEof() {
		return nil, unexpectedEofErr(p.lookahead.pos)
	}

	return &TypeDecl{
		tyName: tyName,
		ty:     ty,
		pos:    pos,
	}, nil
}

func (p *Parser) varDecl() (Declaration, error) {
	pos := p.peekToken().pos
	tok, err := p.peekNext()
	if err != nil {
		return nil, err
	}

	if tok.tok != "ident" {
		return nil, unexpectedTokErr(tok.pos)
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	if p.lookahead.IsEof() {
		return nil, unexpectedEofErr(p.lookahead.pos)
	}

	varName := p.strings.Symbol(tok.value.(string))
	ty, err := p.optionalType()
	if err != nil {
		return nil, err
	}

	tok = p.peekToken()
	if tok.tok != ":=" {
		return nil, unexpectedTokErr(tok.pos)
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	if p.lookahead.IsEof() {
		return nil, unexpectedEofErr(p.lookahead.pos)
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
		pos:    pos,
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
		return nil, unexpectedTokErr(tok.pos)
	}
}

func (p *Parser) letExp() (Exp, error) {
	pos := p.peekToken().pos
	if err := p.nextToken(); err != nil {
		return nil, err
	}

	if p.lookahead.IsEof() {
		return nil, unexpectedEofErr(p.lookahead.pos)
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

	peekToken := p.peekToken()
	if peekToken.tok != "in" {
		return nil, unexpectedTokErr(peekToken.pos)
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	if p.lookahead.IsEof() {
		return nil, unexpectedEofErr(p.lookahead.pos)
	}

	exp, err := p.Exp()
	if err != nil {
		return nil, err
	}

	firstExpPos := exp.ExpPos()

	exps := []Exp{exp}
	for {
		if p.peekToken().tok != ";" {
			break
		}

		if err := p.nextToken(); err != nil {
			return nil, err
		}

		if p.lookahead.IsEof() {
			return nil, unexpectedEofErr(p.lookahead.pos)
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

	peekToken = p.peekToken()
	if peekToken.tok != "end" {
		return nil, unexpectedTokErr(peekToken.pos)
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	return &LetExp{
		body:  seqExp,
		decls: decls,
		pos:   pos,
	}, nil
}

func (p *Parser) nilExp() (Exp, error) {
	pos := p.peekToken().pos
	if err := p.nextToken(); err != nil {
		return nil, err
	}

	return &NilExp{pos}, nil
}

func (p *Parser) seqExp() (Exp, error) {
	pos := p.peekToken().pos
	if err := p.nextToken(); err != nil {
		return nil, err
	}

	if p.lookahead.IsEof() {
		return nil, unexpectedEofErr(p.lookahead.pos)
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

		if p.lookahead.IsEof() {
			return nil, unexpectedEofErr(p.lookahead.pos)
		}
		exp, err := p.Exp()
		if err != nil {
			return nil, err
		}

		seqExp = append(seqExp, exp)
	}

	peekToken := p.peekToken()
	if peekToken.tok != ")" {
		return nil, unexpectedTokErr(peekToken.pos)
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	return &SequenceExp{
		seq: seqExp,
		pos: pos,
	}, nil
}

func (p *Parser) strExp() (Exp, error) {
	tok := p.peekToken()
	if err := p.nextToken(); err != nil {
		return nil, err
	}

	return &StrExp{
		str: tok.value.(string),
		pos: tok.pos,
	}, nil
}

func (p *Parser) whileExp() (Exp, error) {
	pos := p.peekToken().pos
	if err := p.nextToken(); err != nil {
		return nil, err
	}

	if p.lookahead.IsEof() {
		return nil, unexpectedEofErr(p.lookahead.pos)
	}
	test, err := p.Exp()
	if err != nil {
		return nil, err
	}

	if p.peekToken().tok != "do" {
		return nil, unexpectedTokErr(p.peekToken().pos)
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	if p.lookahead.IsEof() {
		return nil, unexpectedEofErr(p.lookahead.pos)
	}
	body, err := p.Exp()
	if err != nil {
		return nil, err
	}

	return &WhileExp{
		pred: test,
		body: body,
		pos:  pos,
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
		return nil, unexpectedTokErr(tok.pos)
	}
}

func (p *Parser) unaryExp() (Exp, error) {
	tok := p.peekToken()
	if tok.tok == "-" {
		pos := tok.pos
		if err := p.nextToken(); err != nil {
			return nil, err
		}

		if p.lookahead.IsEof() {
			return nil, unexpectedEofErr(p.lookahead.pos)
		}

		exp, err := p.unaryExp()
		if err != nil {
			return nil, err
		}

		return &OperExp{
			left: &IntExp{
				val: 0,
				pos: pos,
			},
			op:    Minus,
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

		if p.lookahead.IsEof() {
			return nil, unexpectedEofErr(p.lookahead.pos)
		}

		nextExp, err := p.unaryExp()
		if err != nil {
			return nil, err
		}

		if op.tok == "*" {
			exp = &OperExp{
				left:  exp,
				op:    Mul,
				right: nextExp,
			}
		} else {
			exp = &OperExp{
				left:  exp,
				op:    Div,
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

		if p.lookahead.IsEof() {
			return nil, unexpectedEofErr(p.lookahead.pos)
		}
		nextExp, err := p.mulExp()
		if err != nil {
			return nil, err
		}

		if op.tok == "+" {
			exp = &OperExp{
				left:  exp,
				op:    Plus,
				right: nextExp,
			}
		} else {
			exp = &OperExp{
				left:  exp,
				op:    Minus,
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

	if p.lookahead.IsEof() {
		return nil, unexpectedEofErr(p.lookahead.pos)
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
		left:  left,
		op:    op,
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

	if p.lookahead.IsEof() {
		return nil, unexpectedEofErr(p.lookahead.pos)
	}
	right, err := p.relationalExp()
	if err != nil {
		return nil, err
	}

	return &OperExp{
		left:  left,
		op:    And,
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

	if p.lookahead.IsEof() {
		return nil, unexpectedEofErr(p.lookahead.pos)
	}

	right, err := p.andExp()
	if err != nil {
		return nil, err
	}

	return &OperExp{
		left:  left,
		op:    Or,
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
