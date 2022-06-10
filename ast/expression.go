package ast

import (
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

func (bin *BinaryOp) GetToken() *lexer.Token {
	return bin.token
}

func (unary *UnaryOp) GetToken() *lexer.Token {
	return unary.token
}

func (id *Identifier) GetToken() *lexer.Token {
	return id.token
}

func (lit *Literal) GetToken() *lexer.Token {
	return lit.token
}
