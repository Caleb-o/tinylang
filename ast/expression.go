package ast

import (
	"strings"
	"tiny/lexer"
	"tiny/runtime"
)

type BinaryOp struct {
	token       *lexer.Token
	Left, Right *Node
}

type UnaryOp struct {
	token *lexer.Token
	Right *Node
}

type Identifier struct {
	token *lexer.Token
}

type Literal struct {
	token *lexer.Token
	Value *runtime.Value
}

type GroupedExpr struct {
	Expr *Node
}

func (bin *BinaryOp) GetToken() *lexer.Token {
	return bin.token
}

func (bin *BinaryOp) AsSExp() string {
	var sb strings.Builder

	sb.WriteString(bin.token.Lexeme)
	sb.WriteString((*bin.Left).AsSExp())
	sb.WriteString((*bin.Right).AsSExp())

	return sb.String()
}

func (unary *UnaryOp) GetToken() *lexer.Token {
	return unary.token
}

func (unary *UnaryOp) AsSExp() string {
	var sb strings.Builder

	sb.WriteString(unary.token.Lexeme)
	sb.WriteString((*unary.Right).AsSExp())

	return sb.String()
}

func (id *Identifier) GetToken() *lexer.Token {
	return id.token
}

func (id *Identifier) AsSExp() string {
	return id.token.Lexeme
}

func (lit *Literal) GetToken() *lexer.Token {
	return lit.token
}

func (lit *Literal) AsSExp() string {
	return lit.token.Lexeme
}

func (gr *GroupedExpr) GetToken() *lexer.Token {
	return (*gr.Expr).GetToken()
}

func (gr *GroupedExpr) AsSExp() string {
	var sb strings.Builder

	sb.WriteByte('(')
	sb.WriteString((*gr.Expr).AsSExp())
	sb.WriteByte(')')

	return sb.String()
}
