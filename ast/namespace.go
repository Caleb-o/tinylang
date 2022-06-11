package ast

import (
	"strings"
	"tiny/lexer"
)

type NameSpace struct {
	Variables []*Node
	Functions []*Node
	Body      *Block
	token     *lexer.Token
}

func (ns *NameSpace) GetToken() *lexer.Token {
	return ns.token
}

func (ns *NameSpace) AsSExp() string {
	var sb strings.Builder

	for _, stmt := range ns.Body.Statements {
		sb.WriteString((*stmt).AsSExp())
	}

	return sb.String()
}
