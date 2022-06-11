package ast

import (
	"tiny/lexer"
)

type Variable struct {
	Expr  *Node
	token *lexer.Token
}

func (va *Variable) GetToken() *lexer.Token {
	return va.token
}

func (va *Variable) AsSExp() string {
	return (*va.Expr).AsSExp()
}
