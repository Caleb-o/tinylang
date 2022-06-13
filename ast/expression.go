package ast

import (
	"strings"
	"tiny/lexer"
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
}

type Literal struct {
	Token *lexer.Token
}

type Call struct {
	Token     *lexer.Token
	Callee    Node
	Arguments []Node
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
