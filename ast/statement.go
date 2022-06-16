package ast

import (
	"strings"
	"tiny/lexer"
)

type VariableDecl struct {
	token   *lexer.Token
	Mutable bool
	Expr    Node
}

type Assign struct {
	Token *lexer.Token
	Expr  Node
}

type FunctionDef struct {
	token  *lexer.Token
	Params []*Parameter
	Body   *Block
}

type Print struct {
	Token *lexer.Token
	Exprs []Node
}

type ClassDef struct {
	Token   *lexer.Token
	Fields  map[string]*VariableDecl
	Methods map[string]*FunctionDef
}

type Return struct {
	Token *lexer.Token
	Expr  Node
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

func (assign *Assign) GetToken() *lexer.Token {
	return assign.Token
}

func (assign *Assign) AsSExp() string {
	var sb strings.Builder

	sb.WriteByte('(')
	sb.WriteString(assign.Token.Lexeme + " = ")
	sb.WriteString(assign.Expr.AsSExp())
	sb.WriteByte(')')

	return sb.String()
}

func NewFnDef(token *lexer.Token, params []*Parameter, body *Block) *FunctionDef {
	return &FunctionDef{token: token, Params: params, Body: body}
}

func (fndef *FunctionDef) GetToken() *lexer.Token {
	return fndef.token
}

func (fndef *FunctionDef) AsSExp() string {
	var sb strings.Builder

	sb.WriteByte('(')
	sb.WriteString("function ")
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
	sb.WriteString(fndef.Body.AsSExp())
	sb.WriteByte(')')

	return sb.String()
}

func (p *Print) GetToken() *lexer.Token {
	return p.Token
}

func (p *Print) AsSExp() string {
	var sb strings.Builder

	sb.WriteByte('(')
	sb.WriteString("print")
	sb.WriteByte('(')

	for idx, n := range p.Exprs {
		sb.WriteString(n.AsSExp())

		if idx < len(p.Exprs)-1 {
			sb.WriteString(", ")
		}
	}

	sb.WriteByte(')')
	sb.WriteByte(')')

	return sb.String()
}

func (klass *ClassDef) GetToken() *lexer.Token {
	return klass.Token
}

func (klass *ClassDef) AsSExp() string {
	var sb strings.Builder

	sb.WriteByte('(')
	sb.WriteString(klass.Token.Lexeme)
	sb.WriteByte('(')

	idx := 0
	for _, value := range klass.Fields {
		sb.WriteString(value.AsSExp())

		if idx < len(klass.Fields)-1 {
			sb.WriteString(", ")
		}

		idx += 1
	}

	sb.WriteByte(')')
	sb.WriteByte('(')

	for _, value := range klass.Methods {
		sb.WriteString(value.AsSExp())
	}

	sb.WriteByte(')')
	sb.WriteByte(')')

	return sb.String()
}

func (ret *Return) GetToken() *lexer.Token {
	return ret.Token
}

func (ret *Return) AsSExp() string {
	var sb strings.Builder

	sb.WriteByte('(')
	sb.WriteString("return")

	if ret.Expr != nil {
		sb.WriteString(" " + ret.Expr.AsSExp())
	}
	sb.WriteByte(')')

	return sb.String()
}
