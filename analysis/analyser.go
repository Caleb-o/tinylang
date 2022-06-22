package analysis

import (
	"fmt"
	"tiny/ast"
	"tiny/lexer"
	"tiny/shared"
)

type ClassType byte
type FunctionType byte

const (
	CLASS_NONE ClassType = iota
	CLASS_CLASS
	CLASS_SUBCLASS
	CLASS_STRUCT
)

const (
	FUNCTION_NONE FunctionType = iota
	FUNCTION_FUNCTION
	FUNCTION_CONSTRUCTOR
	FUNCTION_METHOD
	FUNCTION_CATCH
)

type Analyser struct {
	hadErr          bool
	quiet           bool
	currentClass    ClassType
	currentFunction FunctionType
	table           []*SymbolTable
}

func NewAnalyser(quiet bool) *Analyser {
	table := make([]*SymbolTable, 0, 2)
	table = append(table, NewTable(nil))

	return &Analyser{hadErr: false, quiet: quiet, currentClass: CLASS_NONE, currentFunction: FUNCTION_NONE, table: table}
}

func (an *Analyser) Run(root ast.Node) bool {
	an.visit(root)
	an.pop()
	return !an.hadErr
}

func (an *Analyser) DeclareNativeNs(identifier string) {
	if an.lookup(identifier, true) != nil {
		an.report(fmt.Sprintf("Item with name '%s' already exists in the current scope.", identifier))
		return
	}

	an.top().Insert(identifier, &NameSpaceSymbol{identifier})
}

func (an *Analyser) DeclareNativeClass(identifier string, fields []string) {
	if an.lookup(identifier, true) != nil {
		an.report("Item with name '%s' already exists in the current scope.", identifier)
		return
	}

	an.top().Insert(identifier, &NativeClassSymbol{identifier, fields})
}

func (an *Analyser) DeclareNativeFn(identifier string, params []string) {
	if an.lookup(identifier, true) != nil {
		an.report("Item with name '%s' already exists in the current scope.", identifier)
		return
	}

	an.top().Insert(identifier, &NativeFunctionSymbol{identifier, params})
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
		an.visitFunctionDef(n, FUNCTION_FUNCTION)
	case *ast.AnonymousFunction:
		an.visitAnonymousFn(n)
	case *ast.ClassDef:
		an.visitClassDef(n)
	case *ast.StructDef:
		an.visitStructDef(n)
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
	case *ast.LogicalOp:
		an.visit(n.Left)
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
	case *ast.Get:
		an.visitGet(n)
	case *ast.Set:
		an.visitSet(n)
	case *ast.Self:
		an.visitSelf(n)
	case *ast.If:
		an.visitIfStmt(n)
	case *ast.While:
		an.visitWhileStmt(n)
	case *ast.Throw:
		an.visitThrow(n)
	case *ast.Catch:
		an.visitCatch(n)
	case *ast.NameSpace:
		an.visitNamespace(n)

	// Ignore
	case *ast.Unit:
	case *ast.Literal:
	case *ast.Import:

	default:
		an.report("Unhandled node in analysis '%s':'%s'", n.GetToken().Lexeme, n.GetToken().Kind.Name())
	}
}

func (an *Analyser) visitFunctionDef(def *ast.FunctionDef, fnType FunctionType) {
	enclosing := an.currentFunction
	an.currentFunction = fnType

	an.declare(def.GetToken(), &FunctionSymbol{identifier: def.GetToken().Lexeme, def: def})

	// Must implement a block ourselves, so we don't mess up the current scope's symbols with params
	an.table = append(an.table, NewTable(an.top()))

	for _, param := range def.Params {
		an.define(param.GetToken(), &VarSymbol{identifier: param.GetToken().Lexeme})
	}

	an.visitBlock(def.Body, false)

	an.pop()
	an.currentFunction = enclosing
}

func (an *Analyser) visitAnonymousFn(anon *ast.AnonymousFunction) {
	enclosing := an.currentFunction
	an.currentFunction = FUNCTION_FUNCTION

	an.table = append(an.table, NewTable(an.top()))

	for _, param := range anon.Params {
		an.define(param.GetToken(), &VarSymbol{identifier: param.GetToken().Lexeme})
	}

	an.visitBlock(anon.Body, false)

	an.pop()

	an.currentFunction = enclosing
}

func (an *Analyser) visitClassDef(def *ast.ClassDef) {
	enclosing := an.currentClass
	an.currentClass = CLASS_CLASS

	an.declare(def.Token, &ClassDefSymbol{def: def})

	if def.Base != nil {
		switch def.Base.(type) {
		case *ast.Identifier:
		case *ast.Get:
		default:
			an.reportT("Invalid symbol in class base '%s'\n", def.Token, def.Token.Lexeme)
		}

		an.visit(def.Base)
	}

	an.table = append(an.table, NewTable(an.top()))
	an.top().Insert("self", &VarSymbol{identifier: "self", mutable: false})

	for id, fn := range def.Methods {
		declaration := FUNCTION_METHOD

		if id == def.Token.Lexeme {
			declaration = FUNCTION_CONSTRUCTOR
			// Assign constructor to remove resolution at run-time
			def.Constructor = fn
		}

		an.visitFunctionDef(fn, declaration)
	}

	an.pop()

	an.currentClass = enclosing
}

func (an *Analyser) visitStructDef(def *ast.StructDef) {
	enclosing := an.currentClass
	an.currentClass = CLASS_STRUCT

	an.declare(def.Token, &StructDefSymbol{def: def})

	if def.Constructor != nil {
		an.table = append(an.table, NewTable(an.top()))
		an.top().Insert("self", &VarSymbol{identifier: "self", mutable: false})

		an.visitFunctionDef(def.Constructor, FUNCTION_CONSTRUCTOR)

		an.pop()
	}

	an.currentClass = enclosing
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
		defer an.pop()
	}

	for _, node := range block.Statements {
		an.visit(node)
	}
}

func (an *Analyser) visitPrint(print *ast.Print) {
	for _, expr := range print.Exprs {
		an.visit(expr)
	}
}

func (an *Analyser) visitCall(call *ast.Call) {
	an.visit(call.Callee)

	symbol := an.lookup(call.Callee.GetToken().Lexeme, false)

	switch sym := symbol.(type) {
	case *NativeFunctionSymbol:
		if len(sym.params) != len(call.Arguments) {
			an.reportT("Native function '%s' expected %d argument(s) but received %d", call.Token,
				call.Token.Lexeme, len(sym.params), len(call.Arguments))
		}
	case *NativeClassSymbol:
		if len(sym.fields) != len(call.Arguments) {
			an.reportT("Constuctor '%s' expected %d argument(s) but received %d", call.Token,
				call.Token.Lexeme, len(sym.fields), len(call.Arguments))
		}
	case *FunctionSymbol:
		if len(sym.def.Params) != len(call.Arguments) {
			an.reportT("Function '%s' expected %d argument(s) but received %d", call.Token,
				call.Token.Lexeme, len(sym.def.Params), len(call.Arguments))
		}
	case *ClassDefSymbol:
		if sym.def.Constructor != nil {
			cons := sym.def.Constructor

			if len(cons.Params) != len(call.Arguments) {
				an.reportT("Constuctor '%s' expected %d argument(s) but received %d", call.Token,
					call.Token.Lexeme, len(cons.Params), len(call.Arguments))
			}
		}
	}

	for _, expr := range call.Arguments {
		an.visit(expr)
	}
}

func (an *Analyser) visitAssign(assign *ast.Assign) {
	an.resolve(assign.GetToken())

	if sym, ok := an.lookup(assign.GetToken().Lexeme, false).(*VarSymbol); ok {
		if !sym.mutable {
			an.reportT("Cannot assign to immutable value '%s'.", assign.GetToken(), assign.GetToken().Lexeme)
		}
	} else {
		an.reportT("'%s' is not a variable, you cannot assign to it.", assign.GetToken(), assign.GetToken().Lexeme)
	}

	an.visit(assign.Expr)
}

func (an *Analyser) visitReturn(ret *ast.Return) {
	if an.currentFunction == FUNCTION_NONE {
		an.reportT("Cannot return outside of a function", ret.Token)
		return
	}

	if ret.Expr != nil {
		an.visit(ret.Expr)
	}
}

func (an *Analyser) visitGet(get *ast.Get) {
	an.visit(get.Expr)
}

func (an *Analyser) visitSet(set *ast.Set) {
	an.visit(set.Caller)
	an.visit(set.Expr)
}

func (an *Analyser) visitSelf(self *ast.Self) {
	if an.currentClass == CLASS_NONE {
		an.report("Cannot use 'self' outside of a class.")
	}
}

func (an *Analyser) visitIfStmt(stmt *ast.If) {
	if stmt.VarDec != nil {
		an.visit(stmt.VarDec)
	}

	an.visit(stmt.Condition)
	an.visit(stmt.TrueBody)

	if stmt.FalseBody != nil {
		an.visit(stmt.FalseBody)
	}
}

func (an *Analyser) visitWhileStmt(stmt *ast.While) {
	if stmt.VarDec != nil {
		an.visit(stmt.VarDec)
	}

	an.visit(stmt.Condition)

	if stmt.Increment != nil {
		an.visit(stmt.Increment)
	}

	an.visit(stmt.Body)
}

func (an *Analyser) visitThrow(throw *ast.Throw) {
	if an.currentFunction == FUNCTION_NONE {
		an.report("Cannot use 'throw' outside of a function.")
	}

	an.visit(throw.Expr)
}

func (an *Analyser) visitCatch(catch *ast.Catch) {
	an.table = append(an.table, NewTable(an.top()))
	enclosing := an.currentFunction
	an.currentFunction = FUNCTION_CATCH

	// Give expression its own scope
	an.table = append(an.table, NewTable(an.top()))
	an.visit(catch.Expr)
	an.pop()

	an.top().Insert(catch.Var.Lexeme, &VarSymbol{identifier: catch.Var.Lexeme, mutable: false})

	an.visitBlock(catch.Body, false)

	an.pop()
	an.currentFunction = enclosing
}

func (an *Analyser) visitNamespace(ns *ast.NameSpace) {
	an.declare(ns.Token, &NameSpaceSymbol{identifier: ns.Token.Lexeme})
	an.visitBlock(ns.Body, true)
}
