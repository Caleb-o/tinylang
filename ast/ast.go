package ast

import "tiny/lexer"

// It's a bit difficult to determine what every node shall have
// This not too generic, but should work in this case
type Node interface {
	GetToken() *lexer.Token
}

type Program struct {
	Body *Block
}
