package ast

import "tiny/lexer"

type NameSpace struct {
	Variables []*Node
	Functions []*Node
	Body      *Block
	token     *lexer.Token
}

func (ns *NameSpace) GetToken() *lexer.Token {
	return ns.token
}
