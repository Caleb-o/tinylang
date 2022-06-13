package interpreter

import (
	"fmt"
	"strconv"
	"strings"
	"tiny/ast"
	"tiny/lexer"
	"tiny/runtime"
	"tiny/shared"
)

type environment struct {
	// FIXME: Make an abstraction for variables, to allow mutability
	variables []map[string]runtime.Value
	depth     int
}

type Interpreter struct {
	env environment
}

type TinyCallable interface {
	Arity() int
	Call(*Interpreter, []runtime.Value) runtime.Value
}

type FunctionValue struct {
	definition *ast.FunctionDef
}

func (fn *FunctionValue) GetType() runtime.Type { return &runtime.FunctionType{} }
func (fn *FunctionValue) Inspect() string       { return fn.definition.GetToken().Lexeme }
func (fn *FunctionValue) Arity() int            { return len(fn.definition.Params) }

func (fn *FunctionValue) Call(interpreter *Interpreter, values []runtime.Value) runtime.Value {
	interpreter.push()

	for idx, arg := range values {
		interpreter.insert(fn.definition.Params[idx].Token.Lexeme, arg)
	}

	interpreter.Visit(fn.definition.Body)

	interpreter.pop()
	return fn
}

func New() *Interpreter {
	return &Interpreter{env: environment{variables: make([]map[string]runtime.Value, 0, 1), depth: 0}}
}

func (interpreter *Interpreter) Run(program *ast.Program) {
	// TODO: Might return a value to the caller of the interpreter run
	_ = interpreter.visitBlock(program.Body, true)
}

func (interpreter *Interpreter) insert(identifier string, value runtime.Value) {
	interpreter.env.variables[interpreter.env.depth-1][identifier] = value
}

func (interpreter *Interpreter) lookup(identifier string) runtime.Value {
	for idx := interpreter.env.depth - 1; idx >= 0; idx -= 1 {
		if value, ok := interpreter.env.variables[idx][identifier]; ok {
			return value
		}
	}

	// Should not happen, but just to be safe
	interpreter.report("Unknown identifier name in lookup '%s'", identifier)
	return nil
}

func (interpreter *Interpreter) push() {
	interpreter.env.depth += 1
	interpreter.env.variables = append(interpreter.env.variables, make(map[string]runtime.Value))
}

func (interpreter *Interpreter) pop() {
	interpreter.env.depth -= 1
	interpreter.env.variables = interpreter.env.variables[:len(interpreter.env.variables)-1]
}

// --- Private ---
func (interpreter *Interpreter) report(msg string, args ...any) {
	res := fmt.Sprintf(msg, args...)
	shared.ReportErr("Runtime: " + res)
}

func (interpreter *Interpreter) Visit(node ast.Node) runtime.Value {
	switch n := node.(type) {
	case *ast.Block:
		return interpreter.visitBlock(n, true)
	case *ast.Literal:
		return interpreter.visitLiteral(n)
	case *ast.Identifier:
		return interpreter.visitIdentifier(n)
	case *ast.VariableDecl:
		return interpreter.visitVarDecl(n)
	case *ast.Print:
		return interpreter.visitPrint(n)
	case *ast.FunctionDef:
		return interpreter.visitFunctionDef(n)
	case *ast.Call:
		return interpreter.visitCall(n)
	}

	interpreter.report("Unhandled node in Visit '%s':'%s'", node.GetToken().Lexeme, node.GetToken().Kind.Name())
	return nil
}

func (interpreter *Interpreter) visitBlock(block *ast.Block, newEnv bool) runtime.Value {
	if newEnv {
		interpreter.push()
	}

	for _, stmt := range block.Statements {
		interpreter.Visit(stmt)
	}

	if newEnv {
		interpreter.pop()
	}

	return &runtime.UnitVal{}
}

func (interpreter *Interpreter) visitLiteral(lit *ast.Literal) runtime.Value {
	switch lit.Token.Kind {
	case lexer.INT:
		value, _ := strconv.ParseInt(lit.GetToken().Lexeme, 10, 32)
		return &runtime.IntVal{Value: int(value)}

	case lexer.FLOAT:
		value, _ := strconv.ParseFloat(lit.GetToken().Lexeme, 32)
		return &runtime.FloatVal{Value: float32(value)}

	case lexer.BOOL:
		value, _ := strconv.ParseBool(lit.GetToken().Lexeme)
		return &runtime.BoolVal{Value: value}

	case lexer.STRING:
		return &runtime.StringVal{Value: lit.GetToken().Lexeme}
	}

	interpreter.report("Unknown literal type found '%s'", lit.GetToken().Lexeme)
	return &runtime.UnitVal{}
}

func (interpreter *Interpreter) visitIdentifier(id *ast.Identifier) runtime.Value {
	return interpreter.lookup(id.GetToken().Lexeme)
}

func (interpreter *Interpreter) visitVarDecl(decl *ast.VariableDecl) runtime.Value {
	interpreter.insert(decl.GetToken().Lexeme, interpreter.Visit(decl.Expr))
	return &runtime.UnitVal{}
}

func (interpreter *Interpreter) visitPrint(print *ast.Print) runtime.Value {
	var sb strings.Builder

	for _, expr := range print.Exprs {
		sb.WriteString(interpreter.Visit(expr).Inspect())
	}

	fmt.Println(sb.String())
	return &runtime.UnitVal{}
}

func (interpreter *Interpreter) visitFunctionDef(fndef *ast.FunctionDef) runtime.Value {
	interpreter.insert(fndef.GetToken().Lexeme, &FunctionValue{definition: fndef})
	return &runtime.UnitVal{}
}

func (interpreter *Interpreter) visitCall(call *ast.Call) runtime.Value {
	sym, _ := interpreter.lookup(call.Token.Lexeme).(*FunctionValue)
	callable := TinyCallable(sym)
	arguments := make([]runtime.Value, 0, len(call.Arguments))

	for _, arg := range call.Arguments {
		arguments = append(arguments, interpreter.Visit(arg))
	}

	return callable.Call(interpreter, arguments)
}
