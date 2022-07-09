package vm

import "tiny/lexer"

type binaryOp byte

const (
	Add binaryOp = iota
	Sub
	Mul
	Div

	Less
	LessEq
	Greater
	GreaterEq
	EqualEqual
	NotEqual
)

func (b binaryOp) Operator() string {
	switch b {
	case Add:
		return "+"
	case Sub:
		return "-"
	case Mul:
		return "*"
	case Div:
		return "/"

	case Less:
		return "<"
	case LessEq:
		return "<="
	case Greater:
		return ">"
	case GreaterEq:
		return ">="
	case EqualEqual:
		return "=="
	case NotEqual:
		return "!="
	}

	return ""
}

func (b binaryOp) ToKind() lexer.TokenKind {
	switch b {
	case Add:
		return lexer.PLUS
	case Sub:
		return lexer.MINUS
	case Mul:
		return lexer.STAR
	case Div:
		return lexer.SLASH

	case Less:
		return lexer.LESS
	}

	return lexer.ERROR
}
