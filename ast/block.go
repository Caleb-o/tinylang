package ast

import (
	"strings"
	"tiny/lexer"
)

type Block struct {
	Statements []*Node
	token      *lexer.Token
}

func (block *Block) GetToken() *lexer.Token {
	return block.token
}

func (block *Block) AsSExp() string {
	var sb strings.Builder

	for _, stmt := range block.Statements {
		sb.WriteString((*stmt).AsSExp())
	}

	return sb.String()
}
