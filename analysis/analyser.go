package analysis

import (
	"fmt"
	"tiny/ast"
	"tiny/lexer"
	"tiny/shared"
)

type Analyser struct {
	hadErr bool
	table  []*SymbolTable
}

func New() *Analyser {
	table := make([]*SymbolTable, 0, 2)
	table = append(table, NewTable(nil))

	return &Analyser{hadErr: false, table: table}
}

func (an *Analyser) Run(root ast.Node) bool {
	an.visit(root)
	an.pop()
	return !an.hadErr
}

// --- Private ---
func (an *Analyser) report(msg string, args ...any) {
	an.hadErr = true

	res := fmt.Sprintf(msg, args...)
	shared.ReportErr(res)
}

func (an *Analyser) reportT(msg string, token *lexer.Token, args ...any) {
	an.hadErr = true

	res := fmt.Sprintf(msg, args...)
	res2 := fmt.Sprintf("%s [%d:%d]", res, token.Line, token.Column)
	shared.ReportErr(res2)
}

func (an *Analyser) top() *SymbolTable {
	return an.table[len(an.table)-1]
}

func (an *Analyser) pop() {
	an.table = an.table[:len(an.table)-1]
}

func (an *Analyser) lookup(identifier string, local bool) Symbol {
	return an.top().Lookup(identifier, local)
}

func (an *Analyser) visit(node ast.Node) {
	switch n := node.(type) {
	case *ast.FunctionDef:
		an.visitFunctionDef(n)
	case *ast.VariableDecl:
		an.visitVarDecl(n)
	case *ast.Identifier:
		an.visitIdentifier(n)
	case *ast.Block:
		an.visitBlock(n, true)
	case *ast.UnaryOp:
		an.visit(n.Right)
	case *ast.BinaryOp:
		an.visit(n.Left)
		an.visit(n.Right)

	// Ignore
	case *ast.Literal:

	default:
		an.reportT("Unhandled node in analysis '%s'", n.GetToken(), n.GetToken().Kind.Name())
	}
}

func (an *Analyser) visitFunctionDef(def *ast.FunctionDef) {
	if an.lookup(def.GetToken().Lexeme, true) != nil {
		an.reportT("Item with name '%s' already exists in its current scope", def.GetToken(), def.GetToken().Lexeme)
	}
	// Must implement a block ourselves, so we don't mess up the current scope's symbols with params
	an.table = append(an.table, NewTable(an.top()))

	for _, param := range def.Params {
		an.top().Insert(param.GetToken().Lexeme, &VarSymbol{identifier: param.GetToken().Lexeme})
	}

	an.visitBlock(def.Body, false)

	an.pop()

	an.top().Insert(def.GetToken().Lexeme, &FunctionSymbol{identifier: def.GetToken().Lexeme})
}

func (an *Analyser) visitVarDecl(decl *ast.VariableDecl) {
	if an.lookup(decl.GetToken().Lexeme, true) != nil {
		an.reportT("Item with name '%s' already exists in its current scope", decl.GetToken(), decl.GetToken().Lexeme)
	}

	an.visit(decl.Expr)
	an.top().Insert(decl.GetToken().Lexeme, &VarSymbol{identifier: decl.GetToken().Lexeme})
}

func (an *Analyser) visitIdentifier(id *ast.Identifier) {
	if an.lookup(id.GetToken().Lexeme, false) == nil {
		an.reportT("Variable with name '%s' does not exist in any scope", id.GetToken(), id.GetToken().Lexeme)
	}
}

func (an *Analyser) visitBlock(block *ast.Block, newTable bool) {
	if newTable {
		an.table = append(an.table, NewTable(an.top()))
	}

	for _, node := range block.Statements {
		an.visit(node)
	}

	if newTable {
		an.pop()
	}
}
