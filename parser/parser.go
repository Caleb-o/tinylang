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
		report("Expected token kind '%s' but received '%s':%s [%d:%d]", expected.Name(), parser.current.Lexeme, parser.current.Kind.Name(), parser.current.Line, parser.current.Column)
	}
}

// From crafting interpreters
func (parser *Parser) match(expected ...lexer.TokenKind) (*lexer.Token, bool) {
	for _, kind := range expected {
		if parser.current.Kind == kind {
			ftoken := parser.current
			parser.consume(kind)
			return ftoken, true
		}
	}

	return nil, false
}

func (parser *Parser) consumeIfExists(expected lexer.TokenKind) {
	if parser.current.Kind == expected {
		parser.current = parser.lexer.Next()
	}
}

func (parser *Parser) functionCall(outer *ast.Block, callee ast.Node) ast.Node {
	arguments := make([]ast.Node, 0)

	for parser.current.Kind != lexer.CLOSEPAREN {
		arguments = append(arguments, parser.expr(outer))
		parser.consumeIfExists(lexer.COMMA)
	}
	parser.consume(lexer.CLOSEPAREN)

	return &ast.Call{Token: callee.GetToken(), Callee: callee, Arguments: arguments}
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

	case lexer.SELF:
		parser.consume(ftoken.Kind)
		return &ast.Self{Token: ftoken}

	case lexer.IDENTIFIER:
		parser.consume(lexer.IDENTIFIER)
		return &ast.Identifier{Token: ftoken}

	case lexer.OPENPAREN:
		parser.consume(lexer.OPENPAREN)
		expr := parser.expr(outer)
		parser.consume(lexer.CLOSEPAREN)
		return expr

	case lexer.FUNCTION:
		return parser.anonymousFunction(outer)
	}

	report("Unknown token found in expression '%s'", parser.current.Lexeme)
	return nil
}

func (parser *Parser) call(outer *ast.Block) ast.Node {
	node := parser.primary(outer)

	for {
		if _, ok := parser.match(lexer.OPENPAREN); ok {
			node = parser.functionCall(outer, node)
		} else if _, ok := parser.match(lexer.DOT); ok {
			identifier := parser.current
			parser.consume(lexer.IDENTIFIER)

			node = &ast.Get{Token: identifier, Expr: node}
		} else {
			break
		}
	}

	return node
}

func (parser *Parser) unary(outer *ast.Block) ast.Node {
	if operator, ok := parser.match(lexer.BANG, lexer.MINUS); ok {
		return &ast.UnaryOp{Token: operator, Right: parser.unary(outer)}
	}

	return parser.call(outer)
}

func (parser *Parser) factor(outer *ast.Block) ast.Node {
	node := parser.unary(outer)

	for {
		if operator, ok := parser.match(lexer.STAR, lexer.SLASH); ok {
			node = &ast.BinaryOp{Token: operator, Left: node, Right: parser.unary(outer)}
		} else {
			break
		}
	}

	return node
}

func (parser *Parser) term(outer *ast.Block) ast.Node {
	node := parser.factor(outer)

	for {
		if operator, ok := parser.match(lexer.PLUS, lexer.MINUS); ok {
			node = &ast.BinaryOp{Token: operator, Left: node, Right: parser.factor(outer)}
		} else {
			break
		}
	}

	return node
}

func (parser *Parser) comparison(outer *ast.Block) ast.Node {
	node := parser.term(outer)

	for {
		if operator, ok := parser.match(lexer.LESS, lexer.LESS_EQUAL, lexer.GREATER, lexer.GREATER_EQUAL); ok {
			node = &ast.BinaryOp{Token: operator, Left: node, Right: parser.term(outer)}
		} else {
			break
		}
	}

	return node
}

func (parser *Parser) equality(outer *ast.Block) ast.Node {
	node := parser.comparison(outer)

	for {
		if operator, ok := parser.match(lexer.EQUAL_EQUAL, lexer.NOT_EQUAL); ok {
			node = &ast.BinaryOp{Token: operator, Left: node, Right: parser.comparison(outer)}
		} else {
			break
		}
	}

	return node
}

func (parser *Parser) and(outer *ast.Block) ast.Node {
	node := parser.equality(outer)

	for {
		if operator, ok := parser.match(lexer.AND); ok {
			node = &ast.LogicalOp{Token: operator, Left: node, Right: parser.equality(outer)}
		} else {
			break
		}
	}

	return node
}

func (parser *Parser) or(outer *ast.Block) ast.Node {
	node := parser.and(outer)

	for {
		if operator, ok := parser.match(lexer.OR); ok {
			node = &ast.LogicalOp{Token: operator, Left: node, Right: parser.and(outer)}
		} else {
			break
		}
	}

	return node
}

func (parser *Parser) assignment(outer *ast.Block) ast.Node {
	node := parser.or(outer)

	if operator, ok := parser.match(lexer.EQUAL, lexer.PLUS_EQUAL, lexer.MINUS_EQUAL, lexer.STAR_EQUAL, lexer.SLASH_EQUAL); ok {
		if get, ok := node.(*ast.Get); ok {
			return &ast.Set{Token: get.Token, Caller: get.Expr, Expr: parser.or(outer)}
		} else {
			return parser.variableAssign(outer, node.GetToken(), operator)
		}
	}

	return node
}

func (parser *Parser) expr(outer *ast.Block) ast.Node {
	return parser.assignment(outer)
}

func (parser *Parser) collectParameters() []*ast.Parameter {
	params := make([]*ast.Parameter, 0, 2)
	parser.consume(lexer.OPENPAREN)

	for parser.current.Kind == lexer.IDENTIFIER {
		identifier := parser.current
		parser.consume(lexer.IDENTIFIER)

		params = append(params, &ast.Parameter{Token: identifier, Mutable: false})
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

func (parser *Parser) anonymousFunction(outer *ast.Block) *ast.AnonymousFunction {
	ftoken := parser.current
	parser.consume(lexer.FUNCTION)

	// FIXME: Add function return type
	return ast.NewAnonFn(ftoken, parser.collectParameters(), parser.block())
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

			if _, ok := methods[fn.GetToken().Lexeme]; ok {
				shared.ReportErrFatal(fmt.Sprintf("Function with name '%s' already exists in class '%s'", fn.GetToken().Lexeme, identifier.Lexeme))
			}

			methods[fn.GetToken().Lexeme] = fn

		case lexer.VAR:
			parser.consume(lexer.VAR)
			variable := parser.variableDeclEmpty(true)

			if _, ok := fields[variable.GetToken().Lexeme]; ok {
				shared.ReportErrFatal(fmt.Sprintf("Field with name '%s' already exists in class '%s'", variable.GetToken().Lexeme, identifier.Lexeme))
			}

			fields[variable.GetToken().Lexeme] = variable
			parser.consume(lexer.SEMICOLON)

		default:
			shared.ReportErrFatal(fmt.Sprintf("Unknown item in class definition '%s'", parser.current.Lexeme))
		}
	}

	parser.consume(lexer.CLOSECURLY)

	return &ast.ClassDef{Token: identifier, Constructor: nil, Fields: fields, Methods: methods}
}

func (parser *Parser) variableAssign(outer *ast.Block, identifier *lexer.Token, operator *lexer.Token) *ast.Assign {
	return &ast.Assign{Token: identifier, Operator: operator, Expr: parser.expr(outer)}
}

func (parser *Parser) variableDecl(outer *ast.Block, mutable bool) *ast.VariableDecl {
	identifier := parser.current
	parser.consume(lexer.IDENTIFIER)

	parser.consume(lexer.EQUAL)
	expr := parser.expr(outer)

	return ast.NewVarDecl(identifier, mutable, expr)
}

func (parser *Parser) variableDeclEmpty(mutable bool) *ast.VariableDecl {
	identifier := parser.current
	parser.consume(lexer.IDENTIFIER)

	return ast.NewVarDecl(identifier, mutable, nil)
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

func (parser *Parser) returns(outer *ast.Block) {
	ret := parser.current
	parser.consume(lexer.RETURN)

	var expr ast.Node = nil
	if parser.current.Kind != lexer.SEMICOLON {
		expr = parser.expr(outer)
	}

	outer.Statements = append(outer.Statements, &ast.Return{Token: ret, Expr: expr})
}

func (parser *Parser) ifstmt(outer *ast.Block) *ast.If {
	ftoken := parser.current
	parser.consume(lexer.IF)

	var varDecl *ast.VariableDecl = nil
	if parser.current.Kind == lexer.LET || parser.current.Kind == lexer.VAR {
		mutable := parser.current.Kind == lexer.VAR
		parser.consume(parser.current.Kind)

		varDecl = parser.variableDecl(outer, mutable)
		parser.consume(lexer.SEMICOLON)
	}

	condition := parser.expr(outer)
	trueBody := parser.block()
	var falseBody ast.Node = nil

	if _, ok := parser.match(lexer.ELSE); ok {
		if parser.current.Kind == lexer.IF {
			falseBody = parser.ifstmt(outer)
		} else {
			falseBody = parser.block()
		}
	}

	return &ast.If{Token: ftoken, VarDec: varDecl, Condition: condition, TrueBody: trueBody, FalseBody: falseBody}
}

func (parser *Parser) whilestmt(outer *ast.Block) *ast.While {
	ftoken := parser.current
	parser.consume(lexer.WHILE)

	var varDecl *ast.VariableDecl = nil
	if parser.current.Kind == lexer.LET || parser.current.Kind == lexer.VAR {
		mutable := parser.current.Kind == lexer.VAR
		parser.consume(parser.current.Kind)

		varDecl = parser.variableDecl(outer, mutable)
		parser.consume(lexer.SEMICOLON)
	}

	condition := parser.expr(outer)

	var increment ast.Node = nil
	if _, ok := parser.match(lexer.SEMICOLON); ok {
		increment = parser.expr(outer)
	}

	return &ast.While{Token: ftoken, VarDec: varDecl, Condition: condition, Increment: increment, Body: parser.block()}
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
	case lexer.RETURN:
		parser.returns(outer)
	case lexer.IF:
		outer.Statements = append(outer.Statements, parser.ifstmt(outer))
		return
	case lexer.WHILE:
		outer.Statements = append(outer.Statements, parser.whilestmt(outer))
		return
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
