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
	Token       *lexer.Token
	Base        Node
	Constructor *FunctionDef
	Fields      map[string]*VariableDecl
	Methods     map[string]*FunctionDef
}

type StructDef struct {
	Token       *lexer.Token
	Constructor *FunctionDef
	Fields      map[string]*VariableDecl
}

type Return struct {
	Token *lexer.Token
	Expr  Node
}

type If struct {
	Token     *lexer.Token
	VarDec    *VariableDecl
	Condition Node
	TrueBody  *Block
	FalseBody Node
}

type While struct {
	Token     *lexer.Token
	VarDec    *VariableDecl
	Increment Node
	Condition Node
	Body      *Block
}

type Throw struct {
	Token *lexer.Token
	Expr  Node
}

type Import struct {
	Token *lexer.Token
}

type NameSpace struct {
	Token *lexer.Token
	Body  *Block
}

type Test struct {
	Token *lexer.Token
	Body  *Block
}

type Break struct {
	Token *lexer.Token
}

type Continue struct {
	Token *lexer.Token
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

	if klass.Base != nil {
		sb.WriteString(" : ")
		sb.WriteString(klass.Base.AsSExp())
		sb.WriteByte(' ')
	}
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

func (stmt *StructDef) GetToken() *lexer.Token {
	return stmt.Token
}

func (stmt *StructDef) AsSExp() string {
	var sb strings.Builder

	sb.WriteByte('(')
	sb.WriteString(stmt.Token.Lexeme)
	sb.WriteByte('(')

	idx := 0
	for _, value := range stmt.Fields {
		sb.WriteString(value.AsSExp())

		if idx < len(stmt.Fields)-1 {
			sb.WriteString(", ")
		}

		idx += 1
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

func (stmt *If) GetToken() *lexer.Token {
	return stmt.Token
}

func (stmt *If) AsSExp() string {
	var sb strings.Builder

	sb.WriteByte('(')
	sb.WriteString("if ")

	if stmt.VarDec != nil {
		sb.WriteString(stmt.VarDec.AsSExp())
		sb.WriteString("; ")
	}

	sb.WriteString(stmt.TrueBody.AsSExp())

	if stmt.FalseBody != nil {
		sb.WriteString(stmt.FalseBody.AsSExp())
	}
	sb.WriteByte(')')

	return sb.String()
}

func (stmt *While) GetToken() *lexer.Token {
	return stmt.Token
}

func (stmt *While) AsSExp() string {
	var sb strings.Builder

	sb.WriteByte('(')
	sb.WriteString("while ")

	if stmt.VarDec != nil {
		sb.WriteString(stmt.VarDec.AsSExp())
		sb.WriteString("; ")
	}

	sb.WriteString(stmt.Body.AsSExp())

	if stmt.Increment != nil {
		sb.WriteString("; ")
		sb.WriteString(stmt.Increment.AsSExp())
	}
	sb.WriteByte(')')

	return sb.String()
}

func (stmt *Throw) GetToken() *lexer.Token {
	return stmt.Token
}

func (stmt *Throw) AsSExp() string {
	var sb strings.Builder

	sb.WriteByte('(')
	sb.WriteString("throw ")
	sb.WriteString(stmt.Expr.AsSExp())
	sb.WriteByte(')')

	return sb.String()
}

func (stmt *Import) GetToken() *lexer.Token {
	return stmt.Token
}

func (stmt *Import) AsSExp() string {
	var sb strings.Builder

	sb.WriteByte('(')
	sb.WriteString("import ")
	sb.WriteString(stmt.Token.Lexeme)
	sb.WriteByte(')')

	return sb.String()
}

func (stmt *NameSpace) GetToken() *lexer.Token {
	return stmt.Token
}

func (stmt *NameSpace) AsSExp() string {
	var sb strings.Builder

	sb.WriteByte('(')
	sb.WriteString("namespace ")
	sb.WriteString(stmt.Token.Lexeme)
	sb.WriteString(stmt.Body.AsSExp())
	sb.WriteByte(')')

	return sb.String()
}

func (stmt *Test) GetToken() *lexer.Token {
	return stmt.Token
}

func (stmt *Test) AsSExp() string {
	var sb strings.Builder

	sb.WriteByte('(')
	sb.WriteString("test ")
	sb.WriteString(stmt.Token.Lexeme)
	sb.WriteString(stmt.Body.AsSExp())
	sb.WriteByte(')')

	return sb.String()
}

func (stmt *Break) GetToken() *lexer.Token {
	return stmt.Token
}

func (stmt *Break) AsSExp() string {
	return "break"
}

func (stmt *Continue) GetToken() *lexer.Token {
	return stmt.Token
}

func (stmt *Continue) AsSExp() string {
	return "continue"
}
