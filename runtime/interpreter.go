package runtime

import (
	"fmt"
	"strconv"
	"strings"
	"tiny/ast"
	"tiny/lexer"
	"tiny/shared"
)

type environment struct {
	// FIXME: Make an abstraction for variables, to allow mutability
	variables []map[string]Value
	depth     int
}

type Interpreter struct {
	env environment
}

type TinyCallable interface {
	Arity() int
	Call(*Interpreter, []Value) Value
}

func New() *Interpreter {
	return &Interpreter{env: environment{variables: make([]map[string]Value, 0, 1), depth: 0}}
}

func (interpreter *Interpreter) Run(program *ast.Program) {
	// TODO: Might return a value to the caller of the interpreter run
	_ = interpreter.visitBlock(program.Body, true)
}

func (interpreter *Interpreter) insert(identifier string, value Value) {
	interpreter.env.variables[interpreter.env.depth-1][identifier] = value
}

func (interpreter *Interpreter) lookup(identifier string) Value {
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
	interpreter.env.variables = append(interpreter.env.variables, make(map[string]Value))
}

func (interpreter *Interpreter) pop() {
	interpreter.env.depth -= 1
	interpreter.env.variables = interpreter.env.variables[:len(interpreter.env.variables)-1]
}

// --- Private ---
func (interpreter *Interpreter) report(msg string, args ...any) {
	res := fmt.Sprintf(msg, args...)
	shared.ReportErrFatal("Runtime: " + res)
}

func (interpreter *Interpreter) checkNumericOperand(token *lexer.Token, operand Value) {
	switch operand.(type) {
	case *IntVal:
		return
	case *FloatVal:
		return
	default:
		interpreter.report("Value '%s' is not a numeric value '%s'", token.Lexeme, operand.GetType())
	}
}

func (interpreter *Interpreter) Visit(node ast.Node) Value {
	switch n := node.(type) {
	case *ast.BinaryOp:
		return interpreter.visitBinaryOp(n)
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
		return interpreter.visitFunctionDef(n, true)
	case *ast.ClassDef:
		return interpreter.visitClassDef(n)
	case *ast.Call:
		return interpreter.visitCall(n)
	case *ast.Assign:
		return interpreter.visitAssign(n)
	case *ast.Return:
		return interpreter.visitReturn(n)
	}

	interpreter.report("Unhandled node in Visit '%s':'%s'", node.GetToken().Lexeme, node.GetToken().Kind.Name())
	return nil
}

func (interpreter *Interpreter) visitBinaryOp(binop *ast.BinaryOp) Value {
	// FIXME: Analysis should make sure all expressions are of the same type
	left := interpreter.Visit(binop.Left)
	right := interpreter.Visit(binop.Right)

	interpreter.checkNumericOperand(binop.Left.GetToken(), left)
	interpreter.checkNumericOperand(binop.Right.GetToken(), right)

	switch left.(type) {
	case *IntVal:
		return IntBinop(binop.Token.Kind, left.(*IntVal), right.(*IntVal))
	case *FloatVal:
		return FloatBinop(binop.Token.Kind, left.(*FloatVal), right.(*FloatVal))
	}

	return &UnitVal{}
}

func (interpreter *Interpreter) visitBlock(block *ast.Block, newEnv bool) Value {
	if newEnv {
		interpreter.push()
	}

	for _, stmt := range block.Statements {
		if val, ok := interpreter.Visit(stmt).(*ReturnValue); ok {
			if newEnv {
				interpreter.pop()
			}

			return val.inner
		}
	}

	if newEnv {
		interpreter.pop()
	}

	return &UnitVal{}
}

func (interpreter *Interpreter) visitLiteral(lit *ast.Literal) Value {
	switch lit.Token.Kind {
	case lexer.INT:
		value, _ := strconv.ParseInt(lit.GetToken().Lexeme, 10, 32)
		return &IntVal{Value: int(value)}

	case lexer.FLOAT:
		value, _ := strconv.ParseFloat(lit.GetToken().Lexeme, 32)
		return &FloatVal{Value: float32(value)}

	case lexer.BOOL:
		value, _ := strconv.ParseBool(lit.GetToken().Lexeme)
		return &BoolVal{Value: value}

	case lexer.STRING:
		return &StringVal{Value: lit.GetToken().Lexeme}
	}

	interpreter.report("Unknown literal type found '%s'", lit.GetToken().Lexeme)
	return &UnitVal{}
}

func (interpreter *Interpreter) visitIdentifier(id *ast.Identifier) Value {
	return interpreter.lookup(id.GetToken().Lexeme)
}

func (interpreter *Interpreter) visitVarDecl(decl *ast.VariableDecl) Value {
	interpreter.insert(decl.GetToken().Lexeme, interpreter.Visit(decl.Expr))
	return &UnitVal{}
}

func (interpreter *Interpreter) visitPrint(print *ast.Print) Value {
	var sb strings.Builder

	for _, expr := range print.Exprs {
		sb.WriteString(interpreter.Visit(expr).Inspect())
	}

	fmt.Println(sb.String())
	return &UnitVal{}
}

func (interpreter *Interpreter) visitFunctionDef(fndef *ast.FunctionDef, insert bool) Value {
	def := &FunctionValue{definition: fndef}
	if insert {
		interpreter.insert(fndef.GetToken().Lexeme, def)
	}
	return def
}

func (interpreter *Interpreter) visitClassDef(def *ast.ClassDef) Value {
	methods := make(map[string]*FunctionValue, 0)

	for id, val := range def.Methods {
		methods[id] = interpreter.visitFunctionDef(val, false).(*FunctionValue)
	}

	interpreter.insert(def.GetToken().Lexeme, &ClassDefValue{identifier: def.GetToken().Lexeme, methods: methods})
	return &UnitVal{}
}

func (interpreter *Interpreter) visitCall(call *ast.Call) Value {
	var callable TinyCallable = nil
	sym := interpreter.lookup(call.Token.Lexeme)

	switch sym.(type) {
	case *FunctionValue:
	case *ClassDefValue:
		callable = sym.(TinyCallable)
	}

	arguments := make([]Value, 0, len(call.Arguments))

	for _, arg := range call.Arguments {
		arguments = append(arguments, interpreter.Visit(arg))
	}

	return callable.Call(interpreter, arguments)
}

func (interpreter *Interpreter) visitAssign(assign *ast.Assign) Value {
	value := interpreter.Visit(assign.Expr)
	interpreter.insert(assign.GetToken().Lexeme, value)

	return value
}

func (interpreter *Interpreter) visitReturn(ret *ast.Return) Value {
	if ret.Expr != nil {
		return &ReturnValue{inner: interpreter.Visit(ret.Expr)}
	}

	return &ReturnValue{inner: &UnitVal{}}
}
