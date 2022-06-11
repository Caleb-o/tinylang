package ast

import (
	"strings"
	"tiny/lexer"
	"tiny/runtime"
)

type BinaryOp struct {
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
	Type    runtime.Type
}

type Literal struct {
	Token *lexer.Token
	Value *runtime.Value
}

func (bin *BinaryOp) GetToken() *lexer.Token {
	return bin.Token
}

func (bin *BinaryOp) AsSExp() string {
	var sb strings.Builder

	sb.WriteByte('(')
	sb.WriteString(bin.Token.Lexeme)
	sb.WriteByte(' ')
	sb.WriteString(bin.Left.AsSExp())
	sb.WriteByte(' ')
	sb.WriteString(bin.Right.AsSExp())
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
	var sb strings.Builder

	sb.WriteString(param.Token.Lexeme)
	sb.WriteByte(':')
	sb.WriteByte(' ')
	sb.WriteString(param.Type.GetName())

	return sb.String()
}

func (lit *Literal) GetToken() *lexer.Token {
	return lit.Token
}

func (lit *Literal) AsSExp() string {
	return lit.Token.Lexeme
}
