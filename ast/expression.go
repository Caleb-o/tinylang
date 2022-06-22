package ast

import (
	"strings"
	"tiny/lexer"
)

type BinaryOp struct {
	Token       *lexer.Token
	Left, Right Node
}

type LogicalOp struct {
	Token       *lexer.Token
	Left, Right Node
}

type UnaryOp struct {
	Token *lexer.Token
	Right Node
}

type Identifier struct {
	Token *lexer.Token
}

type Parameter struct {
	Token   *lexer.Token
	Mutable bool
}

type Unit struct {
	Token *lexer.Token
}

type Literal struct {
	Token *lexer.Token
}

type ListLiteral struct {
	Token *lexer.Token
	Exprs []Node
}

type Call struct {
	Token     *lexer.Token
	Callee    Node
	Arguments []Node
}

type Assign struct {
	Token    *lexer.Token
	Operator *lexer.Token
	Expr     Node
}

type Get struct {
	Token *lexer.Token
	Expr  Node
}

type Set struct {
	Token  *lexer.Token
	Caller Node
	Expr   Node
}

type Index struct {
	Token  *lexer.Token
	Caller Node
	Expr   Node
}

type Self struct {
	Token *lexer.Token
}

type AnonymousFunction struct {
	token  *lexer.Token
	Params []*Parameter
	Body   *Block
}

type Catch struct {
	Token *lexer.Token
	Expr  Node
	Var   *lexer.Token
	Body  *Block
}

func NewAnonFn(token *lexer.Token, params []*Parameter, body *Block) *AnonymousFunction {
	return &AnonymousFunction{token: token, Params: params, Body: body}
}

func (expr *BinaryOp) GetToken() *lexer.Token {
	return expr.Token
}

func (expr *BinaryOp) AsSExp() string {
	var sb strings.Builder

	sb.WriteByte('(')
	sb.WriteString(expr.Token.Lexeme)
	sb.WriteByte(' ')
	sb.WriteString(expr.Left.AsSExp())
	sb.WriteByte(' ')
	sb.WriteString(expr.Right.AsSExp())
	sb.WriteByte(')')

	return sb.String()
}

func (expr *LogicalOp) GetToken() *lexer.Token {
	return expr.Token
}

func (expr *LogicalOp) AsSExp() string {
	var sb strings.Builder

	sb.WriteByte('(')
	sb.WriteString(expr.Token.Lexeme)
	sb.WriteByte(' ')
	sb.WriteString(expr.Left.AsSExp())
	sb.WriteByte(' ')
	sb.WriteString(expr.Right.AsSExp())
	sb.WriteByte(')')

	return sb.String()
}

func (unary *UnaryOp) GetToken() *lexer.Token {
	return unary.Token
}

func (unary *UnaryOp) AsSExp() string {
	var sb strings.Builder

	sb.WriteByte('(')
	sb.WriteString(unary.Token.Lexeme)
	sb.WriteByte(' ')
	sb.WriteString(unary.Right.AsSExp())
	sb.WriteByte(')')

	return sb.String()
}

func (id *Identifier) GetToken() *lexer.Token {
	return id.Token
}

func (id *Identifier) AsSExp() string {
	return id.Token.Lexeme
}

func (param *Parameter) GetToken() *lexer.Token {
	return param.Token
}

func (param *Parameter) AsSExp() string {
	return param.Token.Lexeme
}

func (unit *Unit) GetToken() *lexer.Token {
	return unit.Token
}

func (unit *Unit) AsSExp() string {
	return "(unit)"
}

func (lit *Literal) GetToken() *lexer.Token {
	return lit.Token
}

func (lit *Literal) AsSExp() string {
	return lit.Token.Lexeme
}

func (list *ListLiteral) GetToken() *lexer.Token {
	return list.Token
}

func (list *ListLiteral) AsSExp() string {
	var sb strings.Builder

	sb.WriteByte('[')
	for idx, value := range list.Exprs {
		sb.WriteString(value.AsSExp())

		if idx < len(list.Exprs)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteByte(']')

	return sb.String()
}

func (call *Call) GetToken() *lexer.Token {
	return call.Token
}

func (call *Call) AsSExp() string {
	var sb strings.Builder

	sb.WriteString(call.Token.Lexeme)
	sb.WriteByte('(')

	for idx, arg := range call.Arguments {
		sb.WriteString(arg.AsSExp())

		if idx < len(call.Arguments)-1 {
			sb.WriteString(", ")
		}
	}

	sb.WriteByte(')')

	return sb.String()
}

func (assign *Assign) GetToken() *lexer.Token {
	return assign.Token
}

func (assign *Assign) AsSExp() string {
	var sb strings.Builder

	sb.WriteByte('(')
	sb.WriteString(assign.Token.Lexeme + " " + assign.Operator.Lexeme + " ")
	sb.WriteString(assign.Expr.AsSExp())
	sb.WriteByte(')')

	return sb.String()
}

func (expr *Get) GetToken() *lexer.Token {
	return expr.Token
}

func (expr *Get) AsSExp() string {
	var sb strings.Builder

	sb.WriteString(expr.Token.Lexeme)
	sb.WriteByte('.')
	sb.WriteString(expr.Expr.AsSExp())

	return sb.String()
}

func (expr *Set) GetToken() *lexer.Token {
	return expr.Token
}

func (expr *Set) AsSExp() string {
	var sb strings.Builder

	sb.WriteString(expr.Token.Lexeme)
	sb.WriteByte('.')
	sb.WriteString(expr.Expr.AsSExp())

	return sb.String()
}

func (expr *Index) GetToken() *lexer.Token {
	return expr.Token
}

func (expr *Index) AsSExp() string {
	var sb strings.Builder

	sb.WriteString(expr.Caller.AsSExp())
	sb.WriteByte('[')
	sb.WriteString(expr.Expr.AsSExp())
	sb.WriteByte(']')

	return sb.String()
}

func (expr *Self) GetToken() *lexer.Token {
	return expr.Token
}

func (expr *Self) AsSExp() string {
	return "self"
}

func (expr *AnonymousFunction) GetToken() *lexer.Token {
	return expr.token
}

func (expr *AnonymousFunction) AsSExp() string {
	var sb strings.Builder

	sb.WriteByte('(')
	sb.WriteString("anon function ")
	sb.WriteByte('(')

	for idx, param := range expr.Params {
		if param.Mutable {
			sb.WriteString("mut ")
		}

		sb.WriteString(param.AsSExp())

		if idx < len(expr.Params)-1 {
			sb.WriteString(", ")
		}
	}

	sb.WriteByte(')')
	sb.WriteString(expr.Body.AsSExp())
	sb.WriteByte(')')

	return sb.String()
}

func (expr *Catch) GetToken() *lexer.Token {
	return expr.Token
}

func (expr *Catch) AsSExp() string {
	var sb strings.Builder

	sb.WriteByte('(')
	sb.WriteString("catch ")
	sb.WriteString(expr.Expr.AsSExp())
	sb.WriteByte(':')
	sb.WriteString(expr.Var.Lexeme)
	sb.WriteString(expr.Body.AsSExp())
	sb.WriteByte(')')

	return sb.String()
}
