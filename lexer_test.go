package tt

import (
	"testing"
)

func TestLexerBasicText(t *testing.T) {
	l := NewLexer("Hello World", "[%", "%]")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}
	if tokens[0].Type != TokenText || tokens[0].Value != "Hello World" {
		t.Errorf("expected TEXT(Hello World), got %s", tokens[0])
	}
	if tokens[1].Type != TokenEOF {
		t.Errorf("expected EOF, got %s", tokens[1])
	}
}

func TestLexerSimpleTag(t *testing.T) {
	l := NewLexer("[% foo %]", "[%", "%]")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	assertTokenTypes(t, tokens, []TokenType{TokenIdent, TokenEOF})
	if tokens[0].Value != "foo" {
		t.Errorf("expected ident 'foo', got %q", tokens[0].Value)
	}
}

func TestLexerTextAndTag(t *testing.T) {
	l := NewLexer("Hello [% name %]!", "[%", "%]")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	assertTokenTypes(t, tokens, []TokenType{TokenText, TokenIdent, TokenText, TokenEOF})
	if tokens[0].Value != "Hello " {
		t.Errorf("expected 'Hello ', got %q", tokens[0].Value)
	}
	if tokens[1].Value != "name" {
		t.Errorf("expected 'name', got %q", tokens[1].Value)
	}
	if tokens[2].Value != "!" {
		t.Errorf("expected '!', got %q", tokens[2].Value)
	}
}

func TestLexerKeywords(t *testing.T) {
	l := NewLexer("[% IF foo == 'bar' %]yes[% ELSE %]no[% END %]", "[%", "%]")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	expected := []TokenType{
		TokenIF, TokenIdent, TokenEquals, TokenString,
		TokenText,
		TokenELSE,
		TokenText,
		TokenEND,
		TokenEOF,
	}
	assertTokenTypes(t, tokens, expected)
}

func TestLexerOperators(t *testing.T) {
	l := NewLexer("[% a + b - c * d / e % f %]", "[%", "%]")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	expected := []TokenType{
		TokenIdent, TokenPlus, TokenIdent, TokenMinus,
		TokenIdent, TokenMultiply, TokenIdent, TokenDivide,
		TokenIdent, TokenModulo, TokenIdent, TokenEOF,
	}
	assertTokenTypes(t, tokens, expected)
}

func TestLexerDotNotation(t *testing.T) {
	l := NewLexer("[% foo.bar.baz %]", "[%", "%]")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	expected := []TokenType{
		TokenIdent, TokenDot, TokenIdent, TokenDot, TokenIdent, TokenEOF,
	}
	assertTokenTypes(t, tokens, expected)
}

func TestLexerForeach(t *testing.T) {
	l := NewLexer("[% FOREACH item IN list %][% item %][% END %]", "[%", "%]")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	expected := []TokenType{
		TokenFOREACH, TokenIdent, TokenIN, TokenIdent,
		TokenIdent,
		TokenEND,
		TokenEOF,
	}
	assertTokenTypes(t, tokens, expected)
}

func TestLexerSetAssignment(t *testing.T) {
	l := NewLexer("[% SET x = 42 %]", "[%", "%]")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	expected := []TokenType{TokenSET, TokenIdent, TokenAssign, TokenInteger, TokenEOF}
	assertTokenTypes(t, tokens, expected)
}

func TestLexerChompPre(t *testing.T) {
	l := NewLexer("Hello   \n  [%- foo %]", "[%", "%]")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	if tokens[0].Type != TokenText {
		t.Fatalf("expected TEXT, got %s", tokens[0])
	}
	if tokens[0].Value != "Hello" {
		t.Errorf("expected pre-chomp to trim trailing whitespace, got %q", tokens[0].Value)
	}
}

func TestLexerChompPost(t *testing.T) {
	l := NewLexer("[% foo -%]   \nWorld", "[%", "%]")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	if tokens[len(tokens)-2].Type != TokenText {
		t.Fatalf("expected last non-EOF token to be TEXT, got %s", tokens[len(tokens)-2])
	}
	if tokens[len(tokens)-2].Value != "World" {
		t.Errorf("expected post-chomp to trim following whitespace, got %q", tokens[len(tokens)-2].Value)
	}
}

func TestLexerNumbers(t *testing.T) {
	l := NewLexer("[% 42 3.14 %]", "[%", "%]")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	if tokens[0].Type != TokenInteger || tokens[0].Value != "42" {
		t.Errorf("expected INTEGER(42), got %s", tokens[0])
	}
	if tokens[1].Type != TokenFloat || tokens[1].Value != "3.14" {
		t.Errorf("expected FLOAT(3.14), got %s", tokens[1])
	}
}

func TestLexerStrings(t *testing.T) {
	l := NewLexer(`[% 'hello' "world" %]`, "[%", "%]")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	if tokens[0].Type != TokenString || tokens[0].Value != "hello" {
		t.Errorf("expected STRING(hello), got %s", tokens[0])
	}
	if tokens[1].Type != TokenString || tokens[1].Value != "world" {
		t.Errorf("expected STRING(world), got %s", tokens[1])
	}
}

func TestLexerComparison(t *testing.T) {
	l := NewLexer("[% a == b != c <= d >= e %]", "[%", "%]")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	expected := []TokenType{
		TokenIdent, TokenEquals, TokenIdent, TokenNotEquals,
		TokenIdent, TokenLTE, TokenIdent, TokenGTE, TokenIdent, TokenEOF,
	}
	assertTokenTypes(t, tokens, expected)
}

func TestLexerComment(t *testing.T) {
	l := NewLexer("[%# this is a comment %]hello", "[%", "%]")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	assertTokenTypes(t, tokens, []TokenType{TokenText, TokenEOF})
	if tokens[0].Value != "hello" {
		t.Errorf("expected 'hello', got %q", tokens[0].Value)
	}
}

func TestLexerFatArrow(t *testing.T) {
	l := NewLexer("[% { a => 1 } %]", "[%", "%]")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	expected := []TokenType{
		TokenLBrace, TokenIdent, TokenFatArrow, TokenInteger, TokenRBrace, TokenEOF,
	}
	assertTokenTypes(t, tokens, expected)
}

func TestLexerPipe(t *testing.T) {
	l := NewLexer("[% foo | upper %]", "[%", "%]")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	expected := []TokenType{TokenIdent, TokenPipe, TokenIdent, TokenEOF}
	assertTokenTypes(t, tokens, expected)
}

func assertTokenTypes(t *testing.T, got []Token, expected []TokenType) {
	t.Helper()
	if len(got) != len(expected) {
		names := make([]string, len(got))
		for i, tok := range got {
			names[i] = tok.String()
		}
		t.Fatalf("token count mismatch: expected %d, got %d: %v", len(expected), len(got), names)
	}
	for i, exp := range expected {
		if got[i].Type != exp {
			t.Errorf("token[%d]: expected %v, got %s", i, exp, got[i])
		}
	}
}
