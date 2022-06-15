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

func (parser *Parser) consumeIfExists(expected lexer.TokenKind) {
	if parser.current.Kind == expected {
		parser.current = parser.lexer.Next()
	}
}

func (parser *Parser) functionCall(outer *ast.Block, identifier *lexer.Token) ast.Node {
	arguments := make([]ast.Node, 0)

	parser.consume(lexer.OPENPAREN)
	for parser.current.Kind != lexer.CLOSEPAREN {
		arguments = append(arguments, parser.expr(outer))
		parser.consumeIfExists(lexer.COMMA)
	}
	parser.consume(lexer.CLOSEPAREN)

	return &ast.Call{Token: identifier, Callee: &ast.Identifier{Token: identifier}, Arguments: arguments}
}

func (parser *Parser) primary(outer *ast.Block) ast.Node {
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
		return &ast.Literal{Token: ftoken}

	case lexer.IDENTIFIER:
		parser.consume(lexer.IDENTIFIER)

		if parser.current.Kind == lexer.OPENPAREN {
			return parser.functionCall(outer, ftoken)
		} else if parser.current.Kind == lexer.EQUAL {
			return parser.variableAssign(outer, ftoken)
		}

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

func (parser *Parser) factor(outer *ast.Block) ast.Node {
	node := parser.primary(outer)

	for parser.current.Kind == lexer.STAR || parser.current.Kind == lexer.SLASH {
		operator := parser.current
		parser.consume(operator.Kind)
		node = &ast.BinaryOp{Token: operator, Left: node, Right: parser.primary(outer)}
	}

	return node
}

func (parser *Parser) term(outer *ast.Block) ast.Node {
	node := parser.factor(outer)

	for parser.current.Kind == lexer.PLUS || parser.current.Kind == lexer.MINUS {
		operator := parser.current
		parser.consume(operator.Kind)
		node = &ast.BinaryOp{Token: operator, Left: node, Right: parser.factor(outer)}
	}

	return node
}

func (parser *Parser) expr(outer *ast.Block) ast.Node {
	return parser.term(outer)
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

		params = append(params, &ast.Parameter{Token: identifier, Mutable: mutable})
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

func (parser *Parser) classDef(outer *ast.Block) *ast.ClassDef {
	parser.consume(lexer.CLASS)

	identifier := parser.current
	parser.consume(lexer.IDENTIFIER)

	// TODO: identifier for inheritance
	curly := parser.current
	parser.consume(lexer.OPENCURLY)

	fields := make(map[string]*ast.VariableDecl, 0)
	methods := make(map[string]*ast.FunctionDef, 0)

	block := ast.NewBlock(curly)

	for parser.current.Kind != lexer.CLOSECURLY {
		switch parser.current.Kind {
		case lexer.FUNCTION:
			fn := parser.functionDef(block)
			methods[fn.GetToken().Lexeme] = fn

		case lexer.LET:
			variable := parser.variableDecl(block, true)
			fields[variable.GetToken().Lexeme] = variable
			parser.consume(lexer.SEMICOLON)

		default:
			shared.ReportErrFatal(fmt.Sprintf("Unknown item in class definition '%s'", parser.current.Lexeme))
		}
	}

	parser.consume(lexer.CLOSECURLY)

	return &ast.ClassDef{Token: identifier, Fields: fields, Methods: methods}
}

func (parser *Parser) variableAssign(outer *ast.Block, identifier *lexer.Token) *ast.Assign {
	parser.consume(lexer.EQUAL)
	return &ast.Assign{Token: identifier, Expr: parser.expr(outer)}
}

func (parser *Parser) variableDecl(outer *ast.Block, mutable bool) *ast.VariableDecl {
	identifier := parser.current
	parser.consume(lexer.IDENTIFIER)

	parser.consume(lexer.EQUAL)
	expr := parser.expr(outer)

	return ast.NewVarDecl(identifier, mutable, expr)
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

// FIXME: Use a system similar to Lox so that parsing expression statements are simplified
func (parser *Parser) statement(outer *ast.Block) {
	switch parser.current.Kind {
	case lexer.FUNCTION:
		outer.Statements = append(outer.Statements, parser.functionDef(outer))
		return
	case lexer.CLASS:
		outer.Statements = append(outer.Statements, parser.classDef(outer))
		return
	case lexer.VAR:
		parser.consume(lexer.VAR)
		outer.Statements = append(outer.Statements, parser.variableDecl(outer, true))
	case lexer.LET:
		parser.consume(lexer.LET)
		outer.Statements = append(outer.Statements, parser.variableDecl(outer, false))
	case lexer.PRINT:
		parser.print(outer)
	default:
		// Expression assignment
		outer.Statements = append(outer.Statements, parser.expr(outer))
	}

	parser.consume(lexer.SEMICOLON)
}

func (parser *Parser) statementList(outer *ast.Block, endType lexer.TokenKind) {
	for parser.current.Kind != endType {
		parser.statement(outer)
	}
}
