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

func New() *Interpreter {
	return &Interpreter{env: environment{variables: make([]map[string]runtime.Value, 0, 1), depth: 0}}
}

func (interpreter *Interpreter) Run(program *ast.Program) {
	// TODO: Might return a value to the caller of the interpreter run
	_ = interpreter.visitBlock(program.Body, true)
}

// --- Private ---
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

func (interpreter *Interpreter) report(msg string, args ...any) {
	res := fmt.Sprintf(msg, args...)
	shared.ReportErr("Runtime: " + res)
}

func (interpreter *Interpreter) visit(node ast.Node) runtime.Value {
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
	}

	interpreter.report("Unhandled node in visit '%s':'%s'", node.GetToken().Lexeme, node.GetToken().Kind.Name())
	return nil
}

func (interpreter *Interpreter) visitBlock(block *ast.Block, newEnv bool) runtime.Value {
	if newEnv {
		interpreter.push()
	}

	for _, stmt := range block.Statements {
		interpreter.visit(stmt)
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
	interpreter.insert(decl.GetToken().Lexeme, interpreter.visit(decl.Expr))
	return &runtime.UnitVal{}
}

func (interpreter *Interpreter) visitPrint(print *ast.Print) runtime.Value {
	var sb strings.Builder

	for _, expr := range print.Exprs {
		sb.WriteString(interpreter.visit(expr).Inspect())
	}

	fmt.Println(sb.String())
	return &runtime.UnitVal{}
}
