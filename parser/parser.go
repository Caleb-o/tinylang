package parser

import (
	"fmt"
	"tiny/ast"
	"tiny/lexer"
	"tiny/runtime"
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

func (parser *Parser) consumeIfExists(expected lexer.TokenKind) {
	if parser.current.Kind == expected {
		parser.current = parser.lexer.Next()
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

	case lexer.IDENTIFIER:
		parser.consume(lexer.IDENTIFIER)
		return &ast.Identifier{Token: ftoken}

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

func (parser *Parser) collectParameters() []*ast.Parameter {
	params := make([]*ast.Parameter, 0, 2)
	parser.consume(lexer.OPENPAREN)

	for parser.current.Kind == lexer.IDENTIFIER || parser.current.Kind == lexer.VAR {
		mutable := false

		if parser.current.Kind == lexer.VAR {
			parser.consume(lexer.VAR)
			mutable = true
		}

		identifier := parser.current
		parser.consume(lexer.IDENTIFIER)

		params = append(params, &ast.Parameter{Token: identifier, Mutable: mutable, Type: &runtime.AnyType{}})
		parser.consumeIfExists(lexer.COMMA)
	}

	parser.consume(lexer.CLOSEPAREN)
	return params
}

func (parser *Parser) functionDef(outer *ast.Block) *ast.FunctionDef {
	parser.consume(lexer.FUNCTION)

	identifier := parser.current
	parser.consume(lexer.IDENTIFIER)

	// FIXME: Add function return type
	return ast.NewFnDef(identifier, parser.collectParameters(), parser.block())
}

func (parser *Parser) variableDecl(outer *ast.Block, mutable bool) {
	identifier := parser.current
	parser.consume(lexer.IDENTIFIER)

	parser.consume(lexer.EQUAL)
	expr := parser.expr(outer)

	outer.Statements = append(outer.Statements, ast.NewVarDecl(identifier, mutable, expr))
}

func (parser *Parser) block() *ast.Block {
	start_token := parser.current
	parser.consume(lexer.OPENCURLY)

	block := ast.NewBlock(start_token)
	parser.statementList(block, lexer.CLOSECURLY)

	parser.consume(lexer.CLOSECURLY)

	return block
}

func (parser *Parser) print(outer *ast.Block) {
	token := parser.current
	parser.consume(lexer.PRINT)
	parser.consume(lexer.OPENPAREN)

	exprs := make([]ast.Node, 0)
	for parser.current.Kind != lexer.CLOSEPAREN {
		exprs = append(exprs, parser.expr(outer))
		parser.consumeIfExists(lexer.COMMA)
	}

	parser.consume(lexer.CLOSEPAREN)

	outer.Statements = append(outer.Statements, &ast.Print{Token: token, Exprs: exprs})
}

func (parser *Parser) statement(outer *ast.Block) {
	switch parser.current.Kind {
	case lexer.FUNCTION:
		outer.Statements = append(outer.Statements, parser.functionDef(outer))
	case lexer.VAR:
		parser.consume(lexer.VAR)
		parser.variableDecl(outer, true)
	case lexer.LET:
		parser.consume(lexer.LET)
		parser.variableDecl(outer, false)
	case lexer.PRINT:
		parser.print(outer)
	default:
		report("Unknown token found in statement '%s'", parser.current.Lexeme)
	}

	parser.consume(lexer.SEMICOLON)
}

func (parser *Parser) statementList(outer *ast.Block, endType lexer.TokenKind) {
	for parser.current.Kind != endType {
		parser.statement(outer)
	}
}
