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
	FAT_ARROW

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

	THROW
	CATCH

	IDENTIFIER
	PRINT
	VAR
	LET
	FUNCTION
	SELF
	CLASS
	STRUCT
	NAMESPACE
	RETURN
	IMPORT
	TEST
	BREAK
	CONTINUE
	MATCH

	EOF
	ERROR
)

type Token struct {
	Kind         TokenKind
	Lexeme       string
	Line, Column int
}

var KeyWords = map[string]TokenKind{
	"var":       VAR,
	"let":       LET,
	"print":     PRINT,
	"function":  FUNCTION,
	"self":      SELF,
	"class":     CLASS,
	"struct":    STRUCT,
	"return":    RETURN,
	"while":     WHILE,
	"if":        IF,
	"else":      ELSE,
	"throw":     THROW,
	"catch":     CATCH,
	"import":    IMPORT,
	"namespace": NAMESPACE,
	"test":      TEST,
	"break":     BREAK,
	"continue":  CONTINUE,
	"match":     MATCH,
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
	case VAR:
		return "var"
	case FUNCTION:
		return "function"
	case CLASS:
		return "class"
	case STRUCT:
		return "struct"
	case IDENTIFIER:
		return "identifier"
	case THROW:
		return "throw"
	case CATCH:
		return "catch"
	case IMPORT:
		return "import"
	case IF:
		return "if"
	case ELSE:
		return "else"
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
