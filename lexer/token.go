package lexer

type TokenKind uint8

const (
	PLUS TokenKind = iota
	MINUS
	STAR
	SLASH
	EQUAL

	COLON
	SEMICOLON
	DOT

	OPENCURLY
	CLOSECURLY
	OPENPAREN
	CLOSEPAREN
	OPENSQUARE
	CLOSESQUARE

	INT
	FLOAT
	CHAR
	STRING
	BOOL

	IDENTIFIER
	VAR
	LET
	FUNCTION
	STRUCT
	NAMESPACE

	EOF
	ERROR
)

type Token struct {
	Kind         TokenKind
	Lexeme       string
	Line, Column int
}

var KeyWords = map[string]TokenKind{
	"let":      LET,
	"var":      VAR,
	"function": FUNCTION,
	"struct":   STRUCT,
	// These will be temporary, they will become a value later?
	"true":  BOOL,
	"false": BOOL,
}

var Characters = map[byte]TokenKind{
	'+': PLUS,
	'-': MINUS,
	'*': STAR,
	'/': SLASH,
	'=': EQUAL,
	':': COLON,
	';': SEMICOLON,
	'.': DOT,
	'(': OPENPAREN,
	')': CLOSEPAREN,
	'{': OPENCURLY,
	'}': CLOSECURLY,
	'[': OPENSQUARE,
	']': CLOSESQUARE,
}
