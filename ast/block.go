package ast

import (
	"strings"
	"tiny/lexer"
)

type Block struct {
	Statements []Node
	token      *lexer.Token
}

func NewBlock(token *lexer.Token) *Block {
	return &Block{Statements: make([]Node, 0, 4), token: token}
}

func (block *Block) GetToken() *lexer.Token {
	return block.token
}

func (block *Block) AsSExp() string {
	var sb strings.Builder

	sb.WriteByte('(')
	for _, stmt := range block.Statements {
		sb.WriteString(stmt.AsSExp())
	}
	sb.WriteByte(')')

	return sb.String()
}
