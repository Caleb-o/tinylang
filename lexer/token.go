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
	COMMA

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
	PRINT
	VAR
	LET
	FUNCTION
	CLASS
	NAMESPACE
	RETURN

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
	"print":    PRINT,
	"function": FUNCTION,
	"class":    CLASS,
	"return":   RETURN,
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
	',': COMMA,
	'(': OPENPAREN,
	')': CLOSEPAREN,
	'{': OPENCURLY,
	'}': CLOSECURLY,
	'[': OPENSQUARE,
	']': CLOSESQUARE,
}

// There has to be a nicer way of doing this
func (kind TokenKind) Name() string {
	switch kind {
	case INT:
		return "int"
	case FLOAT:
		return "float"
	case BOOL:
		return "bool"
	case CHAR:
		return "char"
	case STRING:
		return "string"
	case LET:
		return "let"
	case VAR:
		return "var"
	case FUNCTION:
		return "function"
	case CLASS:
		return "class"
	case IDENTIFIER:
		return "identifier"
	case OPENCURLY:
		return "{"
	case CLOSECURLY:
		return "}"
	case OPENPAREN:
		return "("
	case CLOSEPAREN:
		return ")"
	case OPENSQUARE:
		return "["
	case CLOSESQUARE:
		return "]"
	case PLUS:
		return "+"
	case MINUS:
		return "-"
	case STAR:
		return "*"
	case SLASH:
		return "/"
	case EQUAL:
		return "="
	case DOT:
		return "."
	case COLON:
		return ":"
	case SEMICOLON:
		return ";"
	default:
		return "Unknown"
	}
}
