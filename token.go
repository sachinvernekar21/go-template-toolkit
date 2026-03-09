package tt

import "fmt"

type TokenType int

const (
	// Structural
	TokenText TokenType = iota
	TokenTagOpen
	TokenTagClose
	TokenEOF

	// Literals
	TokenIdent
	TokenInteger
	TokenFloat
	TokenString
	TokenRegex

	// Operators
	TokenDot
	TokenAssign
	TokenPlus
	TokenMinus
	TokenMultiply
	TokenDivide
	TokenModulo
	TokenConcat // _
	TokenEquals
	TokenNotEquals
	TokenLT
	TokenGT
	TokenLTE
	TokenGTE
	TokenAnd
	TokenOr
	TokenNot
	TokenQuestion
	TokenColon
	TokenFatArrow // =>
	TokenRange    // ..
	TokenPipe     // |
	TokenDollar   // $

	// Punctuation
	TokenLParen
	TokenRParen
	TokenLBracket
	TokenRBracket
	TokenLBrace
	TokenRBrace
	TokenComma
	TokenSemicolon

	// Keywords — directives
	TokenGET
	TokenSET
	TokenDEFAULT
	TokenCALL
	TokenIF
	TokenELSIF
	TokenELSE
	TokenUNLESS
	TokenEND
	TokenFOREACH
	TokenFOR
	TokenIN
	TokenWHILE
	TokenSWITCH
	TokenCASE
	TokenINCLUDE
	TokenPROCESS
	TokenINSERT
	TokenBLOCK
	TokenFILTER
	TokenWRAPPER
	TokenMACRO
	TokenTRY
	TokenCATCH
	TokenTHROW
	TokenFINAL
	TokenNEXT
	TokenLAST
	TokenRETURN
	TokenSTOP
	TokenCLEAR
	TokenMETA
	TokenTAGS
	TokenDEBUG

	// Keyword operators
	TokenAND_WORD // and
	TokenOR_WORD  // or
	TokenNOT_WORD // not
	TokenMOD_WORD // mod
	TokenDIV_WORD // div
)

var tokenNames = map[TokenType]string{
	TokenText:      "TEXT",
	TokenTagOpen:   "TAG_OPEN",
	TokenTagClose:  "TAG_CLOSE",
	TokenEOF:       "EOF",
	TokenIdent:     "IDENT",
	TokenInteger:   "INTEGER",
	TokenFloat:     "FLOAT",
	TokenString:    "STRING",
	TokenRegex:     "REGEX",
	TokenDot:       ".",
	TokenAssign:    "=",
	TokenPlus:      "+",
	TokenMinus:     "-",
	TokenMultiply:  "*",
	TokenDivide:    "/",
	TokenModulo:    "%",
	TokenConcat:    "_",
	TokenEquals:    "==",
	TokenNotEquals: "!=",
	TokenLT:        "<",
	TokenGT:        ">",
	TokenLTE:       "<=",
	TokenGTE:       ">=",
	TokenAnd:       "&&",
	TokenOr:        "||",
	TokenNot:       "!",
	TokenQuestion:  "?",
	TokenColon:     ":",
	TokenFatArrow:  "=>",
	TokenRange:     "..",
	TokenPipe:      "|",
	TokenDollar:    "$",
	TokenLParen:    "(",
	TokenRParen:    ")",
	TokenLBracket:  "[",
	TokenRBracket:  "]",
	TokenLBrace:    "{",
	TokenRBrace:    "}",
	TokenComma:     ",",
	TokenSemicolon: ";",
	TokenGET:       "GET",
	TokenSET:       "SET",
	TokenDEFAULT:   "DEFAULT",
	TokenCALL:      "CALL",
	TokenIF:        "IF",
	TokenELSIF:     "ELSIF",
	TokenELSE:      "ELSE",
	TokenUNLESS:    "UNLESS",
	TokenEND:       "END",
	TokenFOREACH:   "FOREACH",
	TokenFOR:       "FOR",
	TokenIN:        "IN",
	TokenWHILE:     "WHILE",
	TokenSWITCH:    "SWITCH",
	TokenCASE:      "CASE",
	TokenINCLUDE:   "INCLUDE",
	TokenPROCESS:   "PROCESS",
	TokenINSERT:    "INSERT",
	TokenBLOCK:     "BLOCK",
	TokenFILTER:    "FILTER",
	TokenWRAPPER:   "WRAPPER",
	TokenMACRO:     "MACRO",
	TokenTRY:       "TRY",
	TokenCATCH:     "CATCH",
	TokenTHROW:     "THROW",
	TokenFINAL:     "FINAL",
	TokenNEXT:      "NEXT",
	TokenLAST:      "LAST",
	TokenRETURN:    "RETURN",
	TokenSTOP:      "STOP",
	TokenCLEAR:     "CLEAR",
	TokenMETA:      "META",
	TokenTAGS:      "TAGS",
	TokenDEBUG:     "DEBUG",
	TokenAND_WORD:  "AND",
	TokenOR_WORD:   "OR",
	TokenNOT_WORD:  "NOT",
	TokenMOD_WORD:  "MOD",
	TokenDIV_WORD:  "DIV",
}

var keywords = map[string]TokenType{
	"GET":     TokenGET,
	"SET":     TokenSET,
	"DEFAULT": TokenDEFAULT,
	"CALL":    TokenCALL,
	"IF":      TokenIF,
	"ELSIF":   TokenELSIF,
	"ELSE":    TokenELSE,
	"UNLESS":  TokenUNLESS,
	"END":     TokenEND,
	"FOREACH": TokenFOREACH,
	"FOR":     TokenFOR,
	"IN":      TokenIN,
	"WHILE":   TokenWHILE,
	"SWITCH":  TokenSWITCH,
	"CASE":    TokenCASE,
	"INCLUDE": TokenINCLUDE,
	"PROCESS": TokenPROCESS,
	"INSERT":  TokenINSERT,
	"BLOCK":   TokenBLOCK,
	"FILTER":  TokenFILTER,
	"WRAPPER": TokenWRAPPER,
	"MACRO":   TokenMACRO,
	"TRY":     TokenTRY,
	"CATCH":   TokenCATCH,
	"THROW":   TokenTHROW,
	"FINAL":   TokenFINAL,
	"NEXT":    TokenNEXT,
	"LAST":    TokenLAST,
	"RETURN":  TokenRETURN,
	"STOP":    TokenSTOP,
	"CLEAR":   TokenCLEAR,
	"META":    TokenMETA,
	"TAGS":    TokenTAGS,
	"DEBUG":   TokenDEBUG,
	"and":     TokenAND_WORD,
	"or":      TokenOR_WORD,
	"not":     TokenNOT_WORD,
	"mod":     TokenMOD_WORD,
	"div":     TokenDIV_WORD,
	"AND":     TokenAND_WORD,
	"OR":      TokenOR_WORD,
	"NOT":     TokenNOT_WORD,
	"MOD":     TokenMOD_WORD,
	"DIV":     TokenDIV_WORD,

	// Lowercase directive keywords
	"get":     TokenGET,
	"set":     TokenSET,
	"default": TokenDEFAULT,
	"call":    TokenCALL,
	"if":      TokenIF,
	"elsif":   TokenELSIF,
	"else":    TokenELSE,
	"unless":  TokenUNLESS,
	"end":     TokenEND,
	"foreach": TokenFOREACH,
	"for":     TokenFOR,
	"in":      TokenIN,
	"while":   TokenWHILE,
	"switch":  TokenSWITCH,
	"case":    TokenCASE,
	"include": TokenINCLUDE,
	"process": TokenPROCESS,
	"insert":  TokenINSERT,
	"block":   TokenBLOCK,
	"filter":  TokenFILTER,
	"wrapper": TokenWRAPPER,
	"macro":   TokenMACRO,
	"try":     TokenTRY,
	"catch":   TokenCATCH,
	"throw":   TokenTHROW,
	"final":   TokenFINAL,
	"next":    TokenNEXT,
	"last":    TokenLAST,
	"return":  TokenRETURN,
	"stop":    TokenSTOP,
	"clear":   TokenCLEAR,
	"meta":    TokenMETA,
	"tags":    TokenTAGS,
	"debug":   TokenDEBUG,
}

type Token struct {
	Type    TokenType
	Value   string
	Line    int
	Col     int
	PreChomp  bool
	PostChomp bool
}

func (t Token) String() string {
	name, ok := tokenNames[t.Type]
	if !ok {
		name = "UNKNOWN"
	}
	if t.Value != "" {
		return fmt.Sprintf("%s(%q)", name, t.Value)
	}
	return name
}
