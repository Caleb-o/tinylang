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

type Literal struct {
	Token *lexer.Token
}

type Call struct {
	Token     *lexer.Token
	Callee    Node
	Arguments []Node
}

type Assign struct {
	Token *lexer.Token
	Expr  Node
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

type Self struct {
	Token *lexer.Token
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

func (lit *Literal) GetToken() *lexer.Token {
	return lit.Token
}

func (lit *Literal) AsSExp() string {
	return lit.Token.Lexeme
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
	sb.WriteString(assign.Token.Lexeme + " = ")
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

func (expr *Self) GetToken() *lexer.Token {
	return expr.Token
}

func (expr *Self) AsSExp() string {
	return "self"
}
