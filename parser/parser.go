package parser

import (
	"fmt"
	"tiny/ast"
	"tiny/lexer"
	"tiny/shared"
)

type Parser struct {
	lexer   *lexer.Lexer
	current *lexer.Token
}

func New(source string) *Parser {
	lexer := lexer.New(source)
	return &Parser{lexer, lexer.Next()}
}

func (parser *Parser) Parse() *ast.Program {
	program := ast.New()
	parser.statementList(program.Body, lexer.EOF)
	return program
}

// --- Private ---
func report(msg string, args ...any) {
	res := fmt.Sprintf(msg, args...)
	shared.ReportErrFatal(res)
}

func (parser *Parser) consume(expected lexer.TokenKind) {
	if parser.current.Kind == expected {
		parser.current = parser.lexer.Next()
	} else {
		report("Expected token kind '%s' but received '%s'", expected.Name(), parser.current.Lexeme)
	}
}

func (parser *Parser) factor(outer *ast.Block) ast.Node {
	ftoken := parser.current

	switch parser.current.Kind {
	case lexer.INT:
		fallthrough
	case lexer.FLOAT:
		fallthrough
	case lexer.BOOL:
		fallthrough
	case lexer.CHAR:
		fallthrough
	case lexer.STRING:
		parser.consume(ftoken.Kind)
		return &ast.Literal{Token: ftoken, Value: nil}

	case lexer.OPENPAREN:
		parser.consume(lexer.OPENPAREN)
		expr := parser.expr(outer)
		parser.consume(lexer.CLOSEPAREN)
		return expr
	}

	report("Unknown token found in expression '%s'", parser.current.Lexeme)
	return nil
}

func (parser *Parser) term(outer *ast.Block) ast.Node {
	node := parser.factor(outer)

	for parser.current.Kind == lexer.STAR || parser.current.Kind == lexer.SLASH {
		operator := parser.current
		parser.consume(operator.Kind)
		node = &ast.BinaryOp{Token: operator, Left: node, Right: parser.factor(outer)}
	}

	return node
}

func (parser *Parser) expr(outer *ast.Block) ast.Node {
	node := parser.term(outer)

	for parser.current.Kind == lexer.PLUS || parser.current.Kind == lexer.MINUS {
		operator := parser.current
		parser.consume(operator.Kind)
		node = &ast.BinaryOp{Token: operator, Left: node, Right: parser.term(outer)}
	}

	return node
}

func (parser *Parser) statement(outer *ast.Block) {
	switch parser.current.Kind {
	default:
		outer.Statements = append(outer.Statements, parser.expr(outer))
	}
}

func (parser *Parser) statementList(outer *ast.Block, endType lexer.TokenKind) {
	for parser.current.Kind != endType {
		parser.statement(outer)
	}
}
