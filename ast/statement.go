package ast

import (
	"strings"
	"tiny/lexer"
	"tiny/runtime"
)

type VariableDecl struct {
	token   *lexer.Token
	Mutable bool
	Expr    Node
}

type FunctionDef struct {
	token   *lexer.Token
	Params  []*Parameter
	ReturnT runtime.Type
	Body    *Block
}

func NewVarDecl(token *lexer.Token, mutable bool, expr Node) *VariableDecl {
	return &VariableDecl{token: token, Mutable: mutable, Expr: expr}
}

func (decl *VariableDecl) GetToken() *lexer.Token {
	return decl.token
}

func (decl *VariableDecl) AsSExp() string {
	var sb strings.Builder

	sb.WriteByte('(')
	if decl.Mutable {
		sb.WriteString("mut ")
	}
	sb.WriteString(decl.token.Lexeme + " ")
	sb.WriteString(decl.Expr.AsSExp())
	sb.WriteByte(')')

	return sb.String()
}

func NewFnDef(token *lexer.Token, params []*Parameter, body *Block) *FunctionDef {
	return &FunctionDef{token: token, Params: params, Body: body, ReturnT: &runtime.AnyType{}}
}

func (fndef *FunctionDef) GetToken() *lexer.Token {
	return fndef.token
}

func (fndef *FunctionDef) AsSExp() string {
	var sb strings.Builder

	sb.WriteByte('(')
	sb.WriteString(fndef.token.Lexeme)
	sb.WriteByte(' ')
	sb.WriteByte('(')

	for idx, param := range fndef.Params {
		if param.Mutable {
			sb.WriteString("mut ")
		}

		sb.WriteString(param.AsSExp())

		if idx < len(fndef.Params)-1 {
			sb.WriteString(", ")
		}
	}

	sb.WriteByte(')')
	sb.WriteString(": ")
	sb.WriteString(fndef.ReturnT.GetName())
	sb.WriteByte(')')
	sb.WriteString(fndef.Body.AsSExp())
	sb.WriteByte(')')

	return sb.String()
}
