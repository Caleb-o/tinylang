package lexer

type TokenKind uint8

const (
	PLUS TokenKind = iota
	MINUS
	STAR
	SLASH
	EQUAL
	BANG

	PLUS_EQUAL
	MINUS_EQUAL
	STAR_EQUAL
	SLASH_EQUAL

	NOT_EQUAL
	EQUAL_EQUAL
	GREATER
	GREATER_EQUAL
	LESS
	LESS_EQUAL
	AND
	OR

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

	WHILE
	IF
	ELSE
	AMPERSAND
	PIPE

	IDENTIFIER
	PRINT
	VAR
	LET
	FUNCTION
	SELF
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
	"self":     SELF,
	"class":    CLASS,
	"return":   RETURN,
	"while":    WHILE,
	"if":       IF,
	"else":     ELSE,
	// These will be temporary, they will become a value later?
	"true":  BOOL,
	"false": BOOL,
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
	case PLUS_EQUAL:
		return "+="
	case MINUS:
		return "-"
	case MINUS_EQUAL:
		return "-="
	case STAR:
		return "*"
	case STAR_EQUAL:
		return "*="
	case SLASH:
		return "/"
	case SLASH_EQUAL:
		return "/="
	case EQUAL:
		return "="
	case DOT:
		return "."
	case COLON:
		return ":"
	case SEMICOLON:
		return ";"
	case AMPERSAND:
		return "AMPERSAND"
	case PIPE:
		return "PIPE"
	case AND:
		return "AND"
	case OR:
		return "OR"
	default:
		return "Unknown"
	}
}
