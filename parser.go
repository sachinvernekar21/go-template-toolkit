package tt

import (
	"fmt"
)

type Parser struct {
	tokens  []Token
	pos     int
	startTag string
	endTag   string
}

func NewParser(tokens []Token, startTag, endTag string) *Parser {
	if startTag == "" {
		startTag = "[%"
	}
	if endTag == "" {
		endTag = "%]"
	}
	return &Parser{tokens: tokens, pos: 0, startTag: startTag, endTag: endTag}
}

func Parse(input, startTag, endTag string) (*TemplateNode, error) {
	lexer := NewLexer(input, startTag, endTag)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return nil, err
	}
	parser := NewParser(tokens, startTag, endTag)
	return parser.Parse()
}

func (p *Parser) Parse() (*TemplateNode, error) {
	nodes, err := p.parseBody()
	if err != nil {
		return nil, err
	}
	return &TemplateNode{Children: nodes}, nil
}

func (p *Parser) peek() Token {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return Token{Type: TokenEOF}
}

func (p *Parser) advance() Token {
	tok := p.peek()
	if p.pos < len(p.tokens) {
		p.pos++
	}
	return tok
}

func (p *Parser) expect(t TokenType) (Token, error) {
	tok := p.advance()
	if tok.Type != t {
		return tok, fmt.Errorf("line %d, col %d: expected %v, got %s", tok.Line, tok.Col, tokenNames[t], tok)
	}
	return tok, nil
}

func (p *Parser) at(types ...TokenType) bool {
	cur := p.peek().Type
	for _, t := range types {
		if cur == t {
			return true
		}
	}
	return false
}

func (p *Parser) parseBody(stopTokens ...TokenType) ([]Node, error) {
	var nodes []Node
	for {
		tok := p.peek()
		if tok.Type == TokenEOF {
			break
		}
		for _, st := range stopTokens {
			if tok.Type == st {
				return nodes, nil
			}
		}

		node, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if node != nil {
			nodes = append(nodes, node)
		}
	}
	return nodes, nil
}

func (p *Parser) parseStatement() (Node, error) {
	tok := p.peek()

	if tok.Type == TokenText {
		p.advance()
		return &TextNode{Text: tok.Value}, nil
	}

	if tok.Type == TokenSemicolon {
		p.advance()
		return nil, nil
	}

	return p.parseDirective()
}

func (p *Parser) parseDirective() (Node, error) {
	tok := p.peek()

	switch tok.Type {
	case TokenGET:
		return p.parseGet()
	case TokenSET:
		return p.parseSet()
	case TokenDEFAULT:
		return p.parseDefault()
	case TokenCALL:
		return p.parseCall()
	case TokenIF:
		return p.parseIf(false)
	case TokenUNLESS:
		return p.parseIf(true)
	case TokenFOREACH, TokenFOR:
		return p.parseForeach()
	case TokenWHILE:
		return p.parseWhile()
	case TokenSWITCH:
		return p.parseSwitch()
	case TokenINCLUDE:
		return p.parseInclude(true)
	case TokenPROCESS:
		return p.parseInclude(false)
	case TokenINSERT:
		return p.parseInsert()
	case TokenBLOCK:
		return p.parseBlock()
	case TokenFILTER:
		return p.parseFilterBlock()
	case TokenWRAPPER:
		return p.parseWrapper()
	case TokenMACRO:
		return p.parseMacro()
	case TokenTRY:
		return p.parseTry()
	case TokenTHROW:
		return p.parseThrow()
	case TokenNEXT:
		p.advance()
		return &FlowNode{Action: "NEXT"}, nil
	case TokenLAST:
		p.advance()
		return &FlowNode{Action: "LAST"}, nil
	case TokenRETURN:
		p.advance()
		return &FlowNode{Action: "RETURN"}, nil
	case TokenSTOP:
		p.advance()
		return &FlowNode{Action: "STOP"}, nil
	case TokenCLEAR:
		p.advance()
		return &FlowNode{Action: "CLEAR"}, nil

	default:
		return p.parseImplicitGetOrSet()
	}
}

func (p *Parser) parseGet() (Node, error) {
	p.advance() // consume GET
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	p.skipSemicolons()
	return &GetNode{Expr: expr}, nil
}

func (p *Parser) parseSet() (Node, error) {
	p.advance() // consume SET
	return p.parseAssignments()
}

func (p *Parser) parseDefault() (Node, error) {
	p.advance() // consume DEFAULT
	pairs, err := p.parseSetPairs()
	if err != nil {
		return nil, err
	}
	p.skipSemicolons()
	return &DefaultNode{Pairs: pairs}, nil
}

func (p *Parser) parseCall() (Node, error) {
	p.advance() // consume CALL
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	p.skipSemicolons()
	return &CallNode{Expr: expr}, nil
}

func (p *Parser) parseIf(negate bool) (Node, error) {
	p.advance() // consume IF or UNLESS
	cond, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	p.skipSemicolons()

	body, err := p.parseBody(TokenELSIF, TokenELSE, TokenEND)
	if err != nil {
		return nil, err
	}

	node := &IfNode{
		Branches: []IfBranch{{Condition: cond, Body: body, Negate: negate}},
	}

	for p.at(TokenELSIF) {
		p.advance()
		elsifCond, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		p.skipSemicolons()
		elsifBody, err := p.parseBody(TokenELSIF, TokenELSE, TokenEND)
		if err != nil {
			return nil, err
		}
		node.Branches = append(node.Branches, IfBranch{Condition: elsifCond, Body: elsifBody})
	}

	if p.at(TokenELSE) {
		p.advance()
		p.skipSemicolons()
		elseBody, err := p.parseBody(TokenEND)
		if err != nil {
			return nil, err
		}
		node.Else = elseBody
	}

	if _, err := p.expect(TokenEND); err != nil {
		return nil, err
	}
	p.skipSemicolons()
	return node, nil
}

func (p *Parser) parseForeach() (Node, error) {
	p.advance() // consume FOREACH/FOR

	var loopVar string
	if p.at(TokenIdent) {
		saved := p.pos
		ident := p.advance()
		if p.at(TokenIN) || p.at(TokenAssign) {
			loopVar = ident.Value
			p.advance() // consume IN or =
		} else {
			p.pos = saved
		}
	}

	listExpr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	p.skipSemicolons()

	body, err := p.parseBody(TokenEND)
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(TokenEND); err != nil {
		return nil, err
	}
	p.skipSemicolons()

	return &ForeachNode{LoopVar: loopVar, List: listExpr, Body: body}, nil
}

func (p *Parser) parseWhile() (Node, error) {
	p.advance() // consume WHILE
	cond, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	p.skipSemicolons()

	body, err := p.parseBody(TokenEND)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(TokenEND); err != nil {
		return nil, err
	}
	p.skipSemicolons()
	return &WhileNode{Condition: cond, Body: body}, nil
}

func (p *Parser) parseSwitch() (Node, error) {
	p.advance() // consume SWITCH
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	p.skipSemicolons()

	// Skip text between SWITCH and first CASE
	for p.at(TokenText) {
		p.advance()
	}

	var cases []CaseBranch
	for p.at(TokenCASE) {
		p.advance()
		var values []Expr
		if !p.at(TokenText) && !p.at(TokenCASE) && !p.at(TokenEND) {
			for {
				val, err := p.parseExpression()
				if err != nil {
					return nil, err
				}
				values = append(values, val)
				if !p.at(TokenComma) {
					break
				}
				p.advance()
			}
		}
		p.skipSemicolons()
		body, err := p.parseBody(TokenCASE, TokenEND)
		if err != nil {
			return nil, err
		}
		cases = append(cases, CaseBranch{Values: values, Body: body})
	}

	if _, err := p.expect(TokenEND); err != nil {
		return nil, err
	}
	p.skipSemicolons()
	return &SwitchNode{Expr: expr, Cases: cases}, nil
}

func (p *Parser) parseInclude(localize bool) (Node, error) {
	p.advance() // consume INCLUDE or PROCESS

	tmpl, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	var params []SetPair
	for p.at(TokenIdent) {
		saved := p.pos
		ident := p.advance()
		if p.at(TokenAssign) {
			p.advance()
			val, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			params = append(params, SetPair{
				Variable: &IdentExpr{Segments: []IdentSegment{{Name: ident.Value}}},
				Value:    val,
			})
		} else {
			p.pos = saved
			break
		}
	}

	p.skipSemicolons()
	return &IncludeNode{Template: tmpl, Params: params, Localize: localize}, nil
}

func (p *Parser) parseInsert() (Node, error) {
	p.advance() // consume INSERT
	filename, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	p.skipSemicolons()
	return &InsertNode{Filename: filename}, nil
}

func (p *Parser) parseBlock() (Node, error) {
	p.advance() // consume BLOCK
	nameTok, err := p.expect(TokenIdent)
	if err != nil {
		return nil, err
	}
	p.skipSemicolons()

	body, err := p.parseBody(TokenEND)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(TokenEND); err != nil {
		return nil, err
	}
	p.skipSemicolons()
	return &BlockDefNode{Name: nameTok.Value, Body: body}, nil
}

func (p *Parser) parseFilterBlock() (Node, error) {
	p.advance() // consume FILTER
	nameTok, err := p.expect(TokenIdent)
	if err != nil {
		return nil, err
	}

	var args []Expr
	if p.at(TokenLParen) {
		p.advance()
		args, err = p.parseArgList()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(TokenRParen); err != nil {
			return nil, err
		}
	}
	p.skipSemicolons()

	body, err := p.parseBody(TokenEND)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(TokenEND); err != nil {
		return nil, err
	}
	p.skipSemicolons()
	return &FilterBlockNode{Name: nameTok.Value, Args: args, Body: body}, nil
}

func (p *Parser) parseWrapper() (Node, error) {
	p.advance() // consume WRAPPER

	tmpl, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	var params []SetPair
	for p.at(TokenIdent) {
		saved := p.pos
		ident := p.advance()
		if p.at(TokenAssign) {
			p.advance()
			val, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			params = append(params, SetPair{
				Variable: &IdentExpr{Segments: []IdentSegment{{Name: ident.Value}}},
				Value:    val,
			})
		} else {
			p.pos = saved
			break
		}
	}
	p.skipSemicolons()

	body, err := p.parseBody(TokenEND)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(TokenEND); err != nil {
		return nil, err
	}
	p.skipSemicolons()
	return &WrapperNode{Template: tmpl, Params: params, Body: body}, nil
}

func (p *Parser) parseMacro() (Node, error) {
	p.advance() // consume MACRO
	nameTok, err := p.expect(TokenIdent)
	if err != nil {
		return nil, err
	}

	var args []string
	if p.at(TokenLParen) {
		p.advance()
		for !p.at(TokenRParen) && !p.at(TokenEOF) {
			arg, err := p.expect(TokenIdent)
			if err != nil {
				return nil, err
			}
			args = append(args, arg.Value)
			if p.at(TokenComma) {
				p.advance()
			}
		}
		if _, err := p.expect(TokenRParen); err != nil {
			return nil, err
		}
	}

	if p.at(TokenBLOCK) {
		p.advance()
	}
	p.skipSemicolons()

	body, err := p.parseBody(TokenEND)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(TokenEND); err != nil {
		return nil, err
	}
	p.skipSemicolons()
	return &MacroNode{Name: nameTok.Value, Args: args, Body: body}, nil
}

func (p *Parser) parseTry() (Node, error) {
	p.advance() // consume TRY
	p.skipSemicolons()

	body, err := p.parseBody(TokenCATCH, TokenFINAL, TokenEND)
	if err != nil {
		return nil, err
	}

	var catches []CatchBranch
	for p.at(TokenCATCH) {
		p.advance()
		var catchType string
		if p.at(TokenIdent) || p.at(TokenString) {
			tok := p.advance()
			catchType = tok.Value
		}
		p.skipSemicolons()

		catchBody, err := p.parseBody(TokenCATCH, TokenFINAL, TokenEND)
		if err != nil {
			return nil, err
		}
		catches = append(catches, CatchBranch{Type: catchType, Body: catchBody})
	}

	var finalBody []Node
	if p.at(TokenFINAL) {
		p.advance()
		p.skipSemicolons()
		finalBody, err = p.parseBody(TokenEND)
		if err != nil {
			return nil, err
		}
	}

	if _, err := p.expect(TokenEND); err != nil {
		return nil, err
	}
	p.skipSemicolons()
	return &TryNode{Body: body, Catches: catches, Final: finalBody}, nil
}

func (p *Parser) parseThrow() (Node, error) {
	p.advance() // consume THROW
	errType, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	var msg Expr
	if !p.at(TokenSemicolon) && !p.at(TokenEOF) && !p.at(TokenText) && !p.at(TokenEND) {
		msg, err = p.parseExpression()
		if err != nil {
			return nil, err
		}
	}
	p.skipSemicolons()
	return &ThrowNode{Type: errType, Message: msg}, nil
}

func (p *Parser) parseImplicitGetOrSet() (Node, error) {
	if !p.at(TokenIdent) {
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		p.skipSemicolons()
		return &GetNode{Expr: expr}, nil
	}

	saved := p.pos
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if p.at(TokenAssign) {
		p.advance()
		val, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		pairs := []SetPair{{Variable: expr, Value: val}}

		for p.at(TokenSemicolon) || p.at(TokenIdent) {
			if p.at(TokenSemicolon) {
				p.advance()
			}
			if !p.at(TokenIdent) {
				break
			}
			nextSaved := p.pos
			nextExpr, err := p.parseExpression()
			if err != nil {
				p.pos = nextSaved
				break
			}
			if !p.at(TokenAssign) {
				p.pos = nextSaved
				break
			}
			p.advance()
			nextVal, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			pairs = append(pairs, SetPair{Variable: nextExpr, Value: nextVal})
		}

		p.skipSemicolons()
		return &SetNode{Pairs: pairs}, nil
	}

	_ = saved
	p.skipSemicolons()
	return &GetNode{Expr: expr}, nil
}

func (p *Parser) parseAssignments() (Node, error) {
	pairs, err := p.parseSetPairs()
	if err != nil {
		return nil, err
	}
	p.skipSemicolons()
	return &SetNode{Pairs: pairs}, nil
}

func (p *Parser) parseSetPairs() ([]SetPair, error) {
	var pairs []SetPair
	for {
		varExpr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(TokenAssign); err != nil {
			return nil, err
		}
		val, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		pairs = append(pairs, SetPair{Variable: varExpr, Value: val})

		if p.at(TokenSemicolon) {
			p.advance()
		}
		if !p.at(TokenIdent) {
			break
		}
		if !p.looksLikeAssignment() {
			break
		}
	}
	return pairs, nil
}

// looksLikeAssignment peeks ahead to see if the current position starts a `var = expr` assignment.
func (p *Parser) looksLikeAssignment() bool {
	saved := p.pos
	defer func() { p.pos = saved }()

	// Skip the identifier and any dot chain
	for p.at(TokenIdent) {
		p.advance()
		if p.at(TokenDot) {
			p.advance()
		} else {
			break
		}
	}
	return p.at(TokenAssign)
}

// Expression parsing — precedence climbing

func (p *Parser) parseExpression() (Expr, error) {
	return p.parseTernary()
}

func (p *Parser) parseTernary() (Expr, error) {
	expr, err := p.parseOr()
	if err != nil {
		return nil, err
	}
	if p.at(TokenQuestion) {
		p.advance()
		then, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(TokenColon); err != nil {
			return nil, err
		}
		els, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		return &TernaryExpr{Condition: expr, Then: then, Else: els}, nil
	}
	return expr, nil
}

func (p *Parser) parseOr() (Expr, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for p.at(TokenOr, TokenOR_WORD) {
		p.advance()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &BinOpExpr{Op: "or", Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseAnd() (Expr, error) {
	left, err := p.parseNot()
	if err != nil {
		return nil, err
	}
	for p.at(TokenAnd, TokenAND_WORD) {
		p.advance()
		right, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		left = &BinOpExpr{Op: "and", Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseNot() (Expr, error) {
	if p.at(TokenNot, TokenNOT_WORD) {
		p.advance()
		operand, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{Op: "not", Operand: operand}, nil
	}
	return p.parseComparison()
}

func (p *Parser) parseComparison() (Expr, error) {
	left, err := p.parseConcat()
	if err != nil {
		return nil, err
	}
	for p.at(TokenEquals, TokenNotEquals, TokenLT, TokenGT, TokenLTE, TokenGTE) {
		tok := p.advance()
		right, err := p.parseConcat()
		if err != nil {
			return nil, err
		}
		left = &BinOpExpr{Op: tok.Value, Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseConcat() (Expr, error) {
	left, err := p.parseAddSub()
	if err != nil {
		return nil, err
	}
	for p.at(TokenConcat) {
		p.advance()
		right, err := p.parseAddSub()
		if err != nil {
			return nil, err
		}
		left = &BinOpExpr{Op: "_", Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseAddSub() (Expr, error) {
	left, err := p.parseMulDiv()
	if err != nil {
		return nil, err
	}
	for p.at(TokenPlus, TokenMinus) {
		tok := p.advance()
		right, err := p.parseMulDiv()
		if err != nil {
			return nil, err
		}
		left = &BinOpExpr{Op: tok.Value, Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseMulDiv() (Expr, error) {
	left, err := p.parseUnaryMinus()
	if err != nil {
		return nil, err
	}
	for p.at(TokenMultiply, TokenDivide, TokenModulo, TokenDIV_WORD, TokenMOD_WORD) {
		tok := p.advance()
		op := tok.Value
		if tok.Type == TokenDIV_WORD {
			op = "div"
		} else if tok.Type == TokenMOD_WORD {
			op = "mod"
		}
		right, err := p.parseUnaryMinus()
		if err != nil {
			return nil, err
		}
		left = &BinOpExpr{Op: op, Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseUnaryMinus() (Expr, error) {
	if p.at(TokenMinus) {
		p.advance()
		operand, err := p.parseUnaryMinus()
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{Op: "-", Operand: operand}, nil
	}
	return p.parseFilter()
}

func (p *Parser) parseFilter() (Expr, error) {
	expr, err := p.parsePostfix()
	if err != nil {
		return nil, err
	}
	for p.at(TokenPipe) {
		p.advance()
		nameTok, err := p.expect(TokenIdent)
		if err != nil {
			return nil, err
		}
		var args []Expr
		if p.at(TokenLParen) {
			p.advance()
			args, err = p.parseArgList()
			if err != nil {
				return nil, err
			}
			if _, err := p.expect(TokenRParen); err != nil {
				return nil, err
			}
		}
		expr = &FilterExpr{Input: expr, Name: nameTok.Value, Args: args}
	}
	return expr, nil
}

func (p *Parser) parsePostfix() (Expr, error) {
	return p.parsePrimary()
}

func (p *Parser) parsePrimary() (Expr, error) {
	tok := p.peek()

	switch tok.Type {
	case TokenString:
		p.advance()
		interpolated := false
		raw := tok.Value
		if len(p.tokens) > 0 {
			// Detect double-quoted strings by checking if the raw input was double-quoted
			// We mark interpolation based on the lexer context.
			// The lexer stores the string value after parsing escapes.
			// We need to check if the original quote was double.
			// Since the lexer doesn't store the quote type, we use a heuristic:
			// if the string came from a double-quoted context, it was stored.
			// For now, we set all strings to potentially interpolated and let
			// the evaluator check for $ signs.
		}
		expr := &StringExpr{Value: tok.Value, Interpolated: interpolated, Raw: raw}
		return p.parseDotChain(expr)

	case TokenInteger, TokenFloat:
		p.advance()
		expr := &NumberExpr{Value: tok.Value, IsFloat: tok.Type == TokenFloat}
		// Check for range operator
		if p.at(TokenRange) {
			p.advance()
			end, err := p.parsePrimary()
			if err != nil {
				return nil, err
			}
			return &RangeExpr{Start: expr, End: end}, nil
		}
		return expr, nil

	case TokenIdent:
		return p.parseIdentExpr()

	case TokenLParen:
		p.advance()
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(TokenRParen); err != nil {
			return nil, err
		}
		return expr, nil

	case TokenLBracket:
		return p.parseArrayLiteral()

	case TokenLBrace:
		return p.parseHashLiteral()

	case TokenMinus:
		p.advance()
		operand, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{Op: "-", Operand: operand}, nil

	default:
		return nil, fmt.Errorf("line %d, col %d: unexpected token %s in expression", tok.Line, tok.Col, tok)
	}
}

func (p *Parser) parseIdentExpr() (Expr, error) {
	var segments []IdentSegment

	// First segment: plain ident
	tok := p.peek()
	if tok.Type != TokenIdent {
		return nil, fmt.Errorf("line %d, col %d: expected identifier", tok.Line, tok.Col)
	}
	p.advance()
	seg := IdentSegment{Name: tok.Value}
	if p.at(TokenLParen) {
		p.advance()
		args, err := p.parseArgList()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(TokenRParen); err != nil {
			return nil, err
		}
		seg.Args = args
	}
	segments = append(segments, seg)

	// Subsequent segments after dots, supporting $var for dynamic keys
	for p.at(TokenDot) {
		p.advance() // consume dot

		dynamic := false
		if p.at(TokenDollar) {
			dynamic = true
			p.advance() // consume $
		}

		tok := p.peek()
		if tok.Type != TokenIdent && !isKeywordToken(tok.Type) {
			return nil, fmt.Errorf("line %d, col %d: expected identifier after '.'", tok.Line, tok.Col)
		}
		p.advance()
		seg := IdentSegment{Name: tok.Value, Dynamic: dynamic}

		if p.at(TokenLParen) {
			p.advance()
			args, err := p.parseArgList()
			if err != nil {
				return nil, err
			}
			if _, err := p.expect(TokenRParen); err != nil {
				return nil, err
			}
			seg.Args = args
		}

		segments = append(segments, seg)
	}

	expr := Expr(&IdentExpr{Segments: segments})

	// Check for range operator
	if p.at(TokenRange) {
		p.advance()
		end, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		return &RangeExpr{Start: expr, End: end}, nil
	}

	return expr, nil
}

func (p *Parser) parseDotChain(base Expr) (Expr, error) {
	// For string/number literals that might have vmethods called on them: "foo".length
	if !p.at(TokenDot) {
		return base, nil
	}
	// Not typical in TT2, but return as-is for simplicity
	return base, nil
}

func (p *Parser) parseArrayLiteral() (Expr, error) {
	p.advance() // consume [
	var elements []Expr
	for !p.at(TokenRBracket) && !p.at(TokenEOF) {
		elem, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		elements = append(elements, elem)
		if p.at(TokenComma) {
			p.advance()
		}
	}
	if _, err := p.expect(TokenRBracket); err != nil {
		return nil, err
	}
	return &ArrayExpr{Elements: elements}, nil
}

func (p *Parser) parseHashLiteral() (Expr, error) {
	p.advance() // consume {
	var pairs []HashPair
	for !p.at(TokenRBrace) && !p.at(TokenEOF) {
		var key Expr

		// Bare identifiers before => are literal string keys, not variable references
		if p.at(TokenIdent) {
			saved := p.pos
			tok := p.advance()
			if p.at(TokenFatArrow) || p.at(TokenAssign) {
				key = &StringExpr{Value: tok.Value}
			} else {
				p.pos = saved
				var err error
				key, err = p.parseExpression()
				if err != nil {
					return nil, err
				}
			}
		} else {
			var err error
			key, err = p.parseExpression()
			if err != nil {
				return nil, err
			}
		}

		if p.at(TokenFatArrow) {
			p.advance()
		} else if p.at(TokenAssign) {
			p.advance()
		} else if p.at(TokenComma) || p.at(TokenRBrace) {
			pairs = append(pairs, HashPair{Key: key, Value: key})
			if p.at(TokenComma) {
				p.advance()
			}
			continue
		}
		val, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		pairs = append(pairs, HashPair{Key: key, Value: val})
		if p.at(TokenComma) {
			p.advance()
		}
	}
	if _, err := p.expect(TokenRBrace); err != nil {
		return nil, err
	}
	return &HashExpr{Pairs: pairs}, nil
}

func (p *Parser) parseArgList() ([]Expr, error) {
	var args []Expr
	if p.at(TokenRParen) {
		return args, nil
	}
	for {
		arg, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
		if !p.at(TokenComma) {
			break
		}
		p.advance()
	}
	return args, nil
}

func (p *Parser) skipSemicolons() {
	for p.at(TokenSemicolon) {
		p.advance()
	}
}

// isKeywordToken returns true for directive/operator keyword tokens that can
// also appear as field names after a dot (e.g. loop.last, loop.size, item.default).
func isKeywordToken(t TokenType) bool {
	switch t {
	case TokenGET, TokenSET, TokenDEFAULT, TokenCALL,
		TokenIF, TokenELSIF, TokenELSE, TokenUNLESS, TokenEND,
		TokenFOREACH, TokenFOR, TokenIN, TokenWHILE,
		TokenSWITCH, TokenCASE,
		TokenINCLUDE, TokenPROCESS, TokenINSERT,
		TokenBLOCK, TokenFILTER, TokenWRAPPER, TokenMACRO,
		TokenTRY, TokenCATCH, TokenTHROW, TokenFINAL,
		TokenNEXT, TokenLAST, TokenRETURN, TokenSTOP, TokenCLEAR,
		TokenMETA, TokenTAGS, TokenDEBUG,
		TokenAND_WORD, TokenOR_WORD, TokenNOT_WORD,
		TokenMOD_WORD, TokenDIV_WORD:
		return true
	}
	return false
}
