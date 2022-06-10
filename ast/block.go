package ast

import "tiny/lexer"

type Block struct {
	Statements []*Node
	token      *lexer.Token
}

func (block *Block) GetToken() *lexer.Token {
	return block.token
}
