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
	inFunc bool
	table  []*SymbolTable
}

func NewAnalyser(quiet bool) *Analyser {
	table := make([]*SymbolTable, 0, 2)
	table = append(table, NewTable(nil))

	return &Analyser{hadErr: false, quiet: quiet, inFunc: false, table: table}
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

func (an *Analyser) resolve(identifier *lexer.Token) {
	if an.lookup(identifier.Lexeme, false) == nil {
		an.reportT("Item with name '%s' does not exist in any scope.", identifier, identifier.Lexeme)
	}
}

func (an *Analyser) resolveLocal(identifier *lexer.Token) {
	if an.lookup(identifier.Lexeme, true) == nil {
		an.reportT("Item with name '%s' does not exist in the current scope.", identifier, identifier.Lexeme)
	}
}

func (an *Analyser) define(identifier *lexer.Token, sym Symbol) {
	an.top().Insert(identifier.Lexeme, sym)
}

func (an *Analyser) declare(identifier *lexer.Token, sym Symbol) {
	if an.lookup(identifier.Lexeme, true) != nil {
		an.reportT("Item with name '%s' already exists in the current scope.", identifier, identifier.Lexeme)
		return
	}

	an.top().Insert(identifier.Lexeme, sym)
}

func (an *Analyser) visit(node ast.Node) {
	switch n := node.(type) {
	case *ast.FunctionDef:
		an.visitFunctionDef(n)
	case *ast.ClassDef:
		an.visitClassDef(n)
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
	case *ast.Return:
		an.visitReturn(n)

	// Ignore
	case *ast.Literal:

	default:
		an.report("Unhandled node in analysis '%s':'%s'", n.GetToken().Lexeme, n.GetToken().Kind.Name())
	}
}

func (an *Analyser) visitFunctionDef(def *ast.FunctionDef) {
	inFunc := an.inFunc
	an.inFunc = true

	an.declare(def.GetToken(), &FunctionSymbol{identifier: def.GetToken().Lexeme, def: def})

	// Must implement a block ourselves, so we don't mess up the current scope's symbols with params
	an.table = append(an.table, NewTable(an.top()))

	for _, param := range def.Params {
		an.define(param.GetToken(), &VarSymbol{identifier: param.GetToken().Lexeme})
	}

	an.visitBlock(def.Body, false)

	an.pop()

	an.inFunc = inFunc
}

func (an *Analyser) visitClassDef(def *ast.ClassDef) {
	an.declare(def.Token, &ClassDefSymbol{def: def})

	an.table = append(an.table, NewTable(an.top()))

	for _, fn := range def.Methods {
		an.visit(fn)
	}

	an.pop()
}

func (an *Analyser) visitVarDecl(decl *ast.VariableDecl) {
	an.visit(decl.Expr)
	an.declare(decl.GetToken(), &VarSymbol{identifier: decl.GetToken().Lexeme, mutable: decl.Mutable})
}

func (an *Analyser) visitIdentifier(id *ast.Identifier) {
	an.resolve(id.GetToken())
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
	// This should help fix weird chains like func()()()()();
	if _, ok := call.Callee.(*ast.Identifier); !ok {
		an.reportT("Cannot call non-identifier '%s'.", call.GetToken(), call.GetToken().Lexeme)
		return
	}

	an.visit(call.Callee)

	symbol := an.lookup(call.GetToken().Lexeme, false)
	if symbol == nil {
		an.reportT("Item '%s' does not exist in any scope.", call.Token, call.Token.Lexeme)
		return
	}

	switch sym := symbol.(type) {
	case *FunctionSymbol:
		if len(sym.def.Params) != len(call.Arguments) {
			an.reportT("Function '%s' expected %d arguments but received %d", call.Token,
				call.Token.Lexeme, len(sym.def.Params), len(call.Arguments))
		}
	case *ClassDefSymbol:
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

func (an *Analyser) visitReturn(ret *ast.Return) {
	if !an.inFunc {
		an.reportT("Cannot return outside of a function", ret.Token)
		return
	}

	if ret.Expr != nil {
		an.visit(ret.Expr)
	}
}
