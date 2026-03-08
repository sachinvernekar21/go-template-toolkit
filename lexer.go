package tt

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Lexer struct {
	input    string
	startTag string
	endTag   string
	pos      int
	line     int
	col      int
	tokens   []Token
}

func NewLexer(input, startTag, endTag string) *Lexer {
	if startTag == "" {
		startTag = "[%"
	}
	if endTag == "" {
		endTag = "%]"
	}
	return &Lexer{
		input:    input,
		startTag: startTag,
		endTag:   endTag,
		pos:      0,
		line:     1,
		col:      1,
	}
}

func (l *Lexer) Tokenize() ([]Token, error) {
	l.tokens = nil
	for l.pos < len(l.input) {
		idx := strings.Index(l.input[l.pos:], l.startTag)
		if idx < 0 {
			text := l.input[l.pos:]
			if len(text) > 0 {
				l.emitText(text)
			}
			l.pos = len(l.input)
			break
		}

		if idx > 0 {
			text := l.input[l.pos : l.pos+idx]
			l.emitText(text)
		}

		l.pos += idx + len(l.startTag)
		l.advanceLineCol(l.startTag)

		preChomp := false
		if l.pos < len(l.input) && l.input[l.pos] == '-' {
			preChomp = true
			l.pos++
			l.col++
		}

		if preChomp && len(l.tokens) > 0 {
			last := &l.tokens[len(l.tokens)-1]
			if last.Type == TokenText {
				last.Value = strings.TrimRight(last.Value, " \t\n\r")
			}
		}

		if err := l.lexTag(preChomp); err != nil {
			return nil, err
		}
	}
	l.tokens = append(l.tokens, Token{Type: TokenEOF, Line: l.line, Col: l.col})
	return l.tokens, nil
}

func (l *Lexer) emitText(text string) {
	tok := Token{Type: TokenText, Value: text, Line: l.line, Col: l.col}
	l.tokens = append(l.tokens, tok)
	l.advanceLineCol(text)
}

func (l *Lexer) advanceLineCol(s string) {
	for _, ch := range s {
		if ch == '\n' {
			l.line++
			l.col = 1
		} else {
			l.col++
		}
	}
}

func (l *Lexer) lexTag(preChomp bool) error {
	l.skipWhitespace()

	for l.pos < len(l.input) {
		postChomp := false

		if l.pos < len(l.input) && l.input[l.pos] == '-' {
			rest := l.input[l.pos:]
			if strings.HasPrefix(rest[1:], l.endTag) {
				postChomp = true
				l.pos++
				l.col++
			}
		}

		if strings.HasPrefix(l.input[l.pos:], l.endTag) {
			l.pos += len(l.endTag)
			l.advanceLineCol(l.endTag)

			if postChomp {
				l.chompFollowingWhitespace()
			}
			return nil
		}

		if postChomp {
			l.pos--
			l.col--
		}

		tok, err := l.lexToken()
		if err != nil {
			return err
		}
		if tok.Type == TokenType(-1) {
			l.skipWhitespace()
			continue
		}
		tok.PreChomp = preChomp
		l.tokens = append(l.tokens, tok)
		l.skipWhitespace()
	}

	return fmt.Errorf("line %d: unclosed template tag, expected %q", l.line, l.endTag)
}

func (l *Lexer) chompFollowingWhitespace() {
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == ' ' || ch == '\t' || ch == '\r' {
			l.pos++
			l.col++
			continue
		}
		if ch == '\n' {
			l.pos++
			l.line++
			l.col = 1
			return
		}
		return
	}
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == ' ' || ch == '\t' || ch == '\r' {
			l.pos++
			l.col++
		} else if ch == '\n' {
			l.pos++
			l.line++
			l.col = 1
		} else {
			break
		}
	}
}

func (l *Lexer) peek() byte {
	if l.pos < len(l.input) {
		return l.input[l.pos]
	}
	return 0
}

func (l *Lexer) lexToken() (Token, error) {
	ch := l.peek()
	startLine, startCol := l.line, l.col

	switch {
	case ch == '#':
		for l.pos < len(l.input) {
			if strings.HasPrefix(l.input[l.pos:], l.endTag) || strings.HasPrefix(l.input[l.pos:], "-"+l.endTag) {
				return Token{Type: TokenType(-1)}, nil
			}
			if l.input[l.pos] == '\n' {
				l.line++
				l.col = 1
				l.pos++
			} else {
				l.pos++
				l.col++
			}
		}
		return Token{Type: TokenType(-1)}, nil

	case ch == '\'' || ch == '"':
		return l.lexString()

	case ch >= '0' && ch <= '9':
		return l.lexNumber()

	case ch == '.' && l.pos+1 < len(l.input) && l.input[l.pos+1] == '.':
		l.pos += 2
		l.col += 2
		return Token{Type: TokenRange, Value: "..", Line: startLine, Col: startCol}, nil

	case ch == '.':
		l.pos++
		l.col++
		return Token{Type: TokenDot, Value: ".", Line: startLine, Col: startCol}, nil

	case ch == '(':
		l.pos++
		l.col++
		return Token{Type: TokenLParen, Value: "(", Line: startLine, Col: startCol}, nil

	case ch == ')':
		l.pos++
		l.col++
		return Token{Type: TokenRParen, Value: ")", Line: startLine, Col: startCol}, nil

	case ch == '[':
		l.pos++
		l.col++
		return Token{Type: TokenLBracket, Value: "[", Line: startLine, Col: startCol}, nil

	case ch == ']':
		l.pos++
		l.col++
		return Token{Type: TokenRBracket, Value: "]", Line: startLine, Col: startCol}, nil

	case ch == '{':
		l.pos++
		l.col++
		return Token{Type: TokenLBrace, Value: "{", Line: startLine, Col: startCol}, nil

	case ch == '}':
		l.pos++
		l.col++
		return Token{Type: TokenRBrace, Value: "}", Line: startLine, Col: startCol}, nil

	case ch == ',':
		l.pos++
		l.col++
		return Token{Type: TokenComma, Value: ",", Line: startLine, Col: startCol}, nil

	case ch == ';':
		l.pos++
		l.col++
		return Token{Type: TokenSemicolon, Value: ";", Line: startLine, Col: startCol}, nil

	case ch == '+':
		l.pos++
		l.col++
		return Token{Type: TokenPlus, Value: "+", Line: startLine, Col: startCol}, nil

	case ch == '*':
		l.pos++
		l.col++
		return Token{Type: TokenMultiply, Value: "*", Line: startLine, Col: startCol}, nil

	case ch == '/':
		l.pos++
		l.col++
		return Token{Type: TokenDivide, Value: "/", Line: startLine, Col: startCol}, nil

	case ch == '%':
		l.pos++
		l.col++
		return Token{Type: TokenModulo, Value: "%", Line: startLine, Col: startCol}, nil

	case ch == '?':
		l.pos++
		l.col++
		return Token{Type: TokenQuestion, Value: "?", Line: startLine, Col: startCol}, nil

	case ch == ':':
		l.pos++
		l.col++
		return Token{Type: TokenColon, Value: ":", Line: startLine, Col: startCol}, nil

	case ch == '|':
		l.pos++
		l.col++
		if l.pos < len(l.input) && l.input[l.pos] == '|' {
			l.pos++
			l.col++
			return Token{Type: TokenOr, Value: "||", Line: startLine, Col: startCol}, nil
		}
		return Token{Type: TokenPipe, Value: "|", Line: startLine, Col: startCol}, nil

	case ch == '&' && l.pos+1 < len(l.input) && l.input[l.pos+1] == '&':
		l.pos += 2
		l.col += 2
		return Token{Type: TokenAnd, Value: "&&", Line: startLine, Col: startCol}, nil

	case ch == '=':
		l.pos++
		l.col++
		if l.pos < len(l.input) && l.input[l.pos] == '=' {
			l.pos++
			l.col++
			return Token{Type: TokenEquals, Value: "==", Line: startLine, Col: startCol}, nil
		}
		if l.pos < len(l.input) && l.input[l.pos] == '>' {
			l.pos++
			l.col++
			return Token{Type: TokenFatArrow, Value: "=>", Line: startLine, Col: startCol}, nil
		}
		return Token{Type: TokenAssign, Value: "=", Line: startLine, Col: startCol}, nil

	case ch == '!':
		l.pos++
		l.col++
		if l.pos < len(l.input) && l.input[l.pos] == '=' {
			l.pos++
			l.col++
			return Token{Type: TokenNotEquals, Value: "!=", Line: startLine, Col: startCol}, nil
		}
		return Token{Type: TokenNot, Value: "!", Line: startLine, Col: startCol}, nil

	case ch == '<':
		l.pos++
		l.col++
		if l.pos < len(l.input) && l.input[l.pos] == '=' {
			l.pos++
			l.col++
			return Token{Type: TokenLTE, Value: "<=", Line: startLine, Col: startCol}, nil
		}
		return Token{Type: TokenLT, Value: "<", Line: startLine, Col: startCol}, nil

	case ch == '>':
		l.pos++
		l.col++
		if l.pos < len(l.input) && l.input[l.pos] == '=' {
			l.pos++
			l.col++
			return Token{Type: TokenGTE, Value: ">=", Line: startLine, Col: startCol}, nil
		}
		return Token{Type: TokenGT, Value: ">", Line: startLine, Col: startCol}, nil

	case ch == '-':
		l.pos++
		l.col++
		return Token{Type: TokenMinus, Value: "-", Line: startLine, Col: startCol}, nil

	case ch == '_':
		if l.pos+1 < len(l.input) && isIdentChar(rune(l.input[l.pos+1])) {
			return l.lexIdent()
		}
		l.pos++
		l.col++
		return Token{Type: TokenConcat, Value: "_", Line: startLine, Col: startCol}, nil

	case isIdentStart(rune(ch)):
		return l.lexIdent()

	default:
		return Token{}, fmt.Errorf("line %d, col %d: unexpected character %q", l.line, l.col, ch)
	}
}

func (l *Lexer) lexString() (Token, error) {
	quote := l.input[l.pos]
	startLine, startCol := l.line, l.col
	l.pos++
	l.col++

	var buf strings.Builder
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == '\\' && l.pos+1 < len(l.input) {
			next := l.input[l.pos+1]
			l.pos += 2
			l.col += 2
			switch next {
			case 'n':
				buf.WriteByte('\n')
			case 't':
				buf.WriteByte('\t')
			case '\\':
				buf.WriteByte('\\')
			case '\'':
				buf.WriteByte('\'')
			case '"':
				buf.WriteByte('"')
			default:
				buf.WriteByte('\\')
				buf.WriteByte(next)
			}
			continue
		}
		if ch == quote {
			l.pos++
			l.col++
			return Token{Type: TokenString, Value: buf.String(), Line: startLine, Col: startCol}, nil
		}
		if ch == '\n' {
			l.line++
			l.col = 1
		} else {
			l.col++
		}
		buf.WriteByte(ch)
		l.pos++
	}
	return Token{}, fmt.Errorf("line %d, col %d: unterminated string", startLine, startCol)
}

func (l *Lexer) lexNumber() (Token, error) {
	startLine, startCol := l.line, l.col
	start := l.pos
	isFloat := false

	for l.pos < len(l.input) && l.input[l.pos] >= '0' && l.input[l.pos] <= '9' {
		l.pos++
		l.col++
	}

	if l.pos < len(l.input) && l.input[l.pos] == '.' {
		if l.pos+1 < len(l.input) && l.input[l.pos+1] >= '0' && l.input[l.pos+1] <= '9' {
			isFloat = true
			l.pos++
			l.col++
			for l.pos < len(l.input) && l.input[l.pos] >= '0' && l.input[l.pos] <= '9' {
				l.pos++
				l.col++
			}
		}
	}

	val := l.input[start:l.pos]
	typ := TokenInteger
	if isFloat {
		typ = TokenFloat
	}
	return Token{Type: typ, Value: val, Line: startLine, Col: startCol}, nil
}

func (l *Lexer) lexIdent() (Token, error) {
	startLine, startCol := l.line, l.col
	start := l.pos

	r, size := utf8.DecodeRuneInString(l.input[l.pos:])
	if !isIdentStart(r) {
		return Token{}, fmt.Errorf("line %d, col %d: invalid identifier start", l.line, l.col)
	}
	l.pos += size
	l.col++

	for l.pos < len(l.input) {
		r, size = utf8.DecodeRuneInString(l.input[l.pos:])
		if !isIdentChar(r) {
			break
		}
		l.pos += size
		l.col++
	}

	word := l.input[start:l.pos]
	if kwType, ok := keywords[word]; ok {
		return Token{Type: kwType, Value: word, Line: startLine, Col: startCol}, nil
	}
	return Token{Type: TokenIdent, Value: word, Line: startLine, Col: startCol}, nil
}

func isIdentStart(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isIdentChar(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
