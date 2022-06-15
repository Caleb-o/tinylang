package analysis

import (
	"fmt"
	"tiny/ast"
	"tiny/lexer"
	"tiny/shared"
)

type Analyser struct {
	hadErr bool
	quiet  bool
	table  []*SymbolTable
}

func NewAnalyser(quiet bool) *Analyser {
	table := make([]*SymbolTable, 0, 2)
	table = append(table, NewTable(nil))

	return &Analyser{hadErr: false, quiet: quiet, table: table}
}

func (an *Analyser) Run(root ast.Node) bool {
	an.visit(root)
	an.pop()
	return !an.hadErr
}

// --- Private ---
func (an *Analyser) report(msg string, args ...any) {
	an.hadErr = true

	if !an.quiet {
		res := fmt.Sprintf(msg, args...)
		shared.ReportErr(res)
	}
}

func (an *Analyser) reportT(msg string, token *lexer.Token, args ...any) {
	an.hadErr = true

	if !an.quiet {
		res := fmt.Sprintf(msg, args...)
		res2 := fmt.Sprintf("%s [%d:%d]", res, token.Line, token.Column)
		shared.ReportErr(res2)
	}
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
	case *ast.Print:
		an.visitPrint(n)
	case *ast.UnaryOp:
		an.visit(n.Right)
	case *ast.BinaryOp:
		an.visit(n.Left)
		an.visit(n.Right)
	case *ast.Call:
		an.visitCall(n)
	case *ast.Assign:
		an.visitAssign(n)

	// Ignore
	case *ast.Literal:

	default:
		an.report("Unhandled node in analysis '%s':'%s'", n.GetToken().Lexeme, n.GetToken().Kind.Name())
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

	an.top().Insert(def.GetToken().Lexeme, &FunctionSymbol{identifier: def.GetToken().Lexeme, def: def})
}

func (an *Analyser) visitVarDecl(decl *ast.VariableDecl) {
	if an.lookup(decl.GetToken().Lexeme, true) != nil {
		an.reportT("Item with name '%s' already exists in its current scope", decl.GetToken(), decl.GetToken().Lexeme)
	}

	an.visit(decl.Expr)
	an.top().Insert(decl.GetToken().Lexeme, &VarSymbol{identifier: decl.GetToken().Lexeme, mutable: decl.Mutable})
}

func (an *Analyser) visitIdentifier(id *ast.Identifier) {
	if an.lookup(id.GetToken().Lexeme, false) == nil {
		an.reportT("Variable '%s' does not exist in any scope", id.GetToken(), id.GetToken().Lexeme)
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

func (an *Analyser) visitPrint(print *ast.Print) {
	for _, expr := range print.Exprs {
		an.visit(expr)
	}
}

func (an *Analyser) visitCall(call *ast.Call) {
	symbol := an.lookup(call.GetToken().Lexeme, false)
	if symbol == nil {
		an.reportT("Function '%s' does not exist", call.Token, call.Token.Lexeme)
		return
	}

	if fnsym, ok := symbol.(*FunctionSymbol); !ok {
		an.reportT("Identifier '%s' is not a function", call.Token, call.Token.Lexeme)
		return
	} else {
		if len(fnsym.def.Params) != len(call.Arguments) {
			an.reportT("Function '%s' expected %d arguments but received %d", call.Token,
				call.Token.Lexeme, len(fnsym.def.Params), len(call.Arguments))
		}
	}

	for _, expr := range call.Arguments {
		an.visit(expr)
	}
}

func (an *Analyser) visitAssign(assign *ast.Assign) {
	symbol := an.lookup(assign.GetToken().Lexeme, false)
	if symbol == nil {
		an.reportT("Variable '%s' does not exist", assign.Token, assign.Token.Lexeme)
		return
	} else {
		if varsym, ok := symbol.(*VarSymbol); !ok {
			an.reportT("Identifier '%s' is not a variable", assign.Token, assign.Token.Lexeme)
			return
		} else {
			if !varsym.mutable {
				an.reportT("Variable '%s' is immutable", assign.Token, assign.Token.Lexeme)
				return
			}
		}
	}

	an.visit(assign.Expr)
}
