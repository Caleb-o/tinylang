package runtime

import (
	"fmt"
	"reflect"
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
	result := interpreter.visitBlock(program.Body, true)

	if res, ok := result.(*ThrowValue); ok {
		interpreter.report("Uncaught value thrown '%s'", res.inner.Inspect())
	}
}

func (interpreter *Interpreter) insert(identifier string, value Value) {
	interpreter.env.variables[interpreter.env.depth-1][identifier] = value
}

func (interpreter *Interpreter) set(identifier string, operator lexer.TokenKind, value Value) {
	for idx := interpreter.env.depth - 1; idx >= 0; idx -= 1 {
		if _, ok := interpreter.env.variables[idx][identifier]; ok {
			if operator == lexer.EQUAL {
				interpreter.env.variables[idx][identifier] = value
			} else {
				if ok := interpreter.env.variables[idx][identifier].Modify(operator, value); !ok {
					interpreter.report("Cannot use operation '%s' on '%s'", operator.Name(), identifier)
					return
				}
			}
			return
		}
	}

	// Should not happen, but just to be safe
	interpreter.report("Unknown identifier name in lookup '%s'", identifier)
}

func (interpreter *Interpreter) lookup(identifier string) Value {
	for idx := len(interpreter.env.variables) - 1; idx >= 0; idx -= 1 {
		if value, ok := interpreter.env.variables[idx][identifier]; ok {
			return value.Copy()
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

func (interpreter *Interpreter) Visit(node ast.Node) Value {
	switch n := node.(type) {
	case *ast.BinaryOp:
		return interpreter.visitBinaryOp(n)
	case *ast.UnaryOp:
		return interpreter.visitUnaryOp(n)
	case *ast.LogicalOp:
		return interpreter.visitLogicalOp(n)
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
	case *ast.AnonymousFunction:
		return interpreter.visitAnonymousFunction(n)
	case *ast.ClassDef:
		return interpreter.visitClassDef(n)
	case *ast.Call:
		return interpreter.visitCall(n)
	case *ast.Assign:
		return interpreter.visitAssign(n)
	case *ast.Return:
		return interpreter.visitReturn(n)
	case *ast.Throw:
		return interpreter.visitThrow(n)
	case *ast.Get:
		return interpreter.visitGet(n)
	case *ast.Set:
		return interpreter.visitSet(n)
	case *ast.Self:
		return interpreter.lookup("self")
	case *ast.If:
		return interpreter.visitIfStmt(n)
	case *ast.While:
		return interpreter.visitWhileStmt(n)
	case *ast.Catch:
		return interpreter.visitCatch(n)
	}

	interpreter.report("Unhandled node in Visit '%s':%d", node.GetToken().Lexeme, node.GetToken().Line)
	return nil
}

func (interpreter *Interpreter) visitBinaryOp(binop *ast.BinaryOp) Value {
	// FIXME: Analysis should make sure all expressions are of the same type
	left := interpreter.Visit(binop.Left)
	right := interpreter.Visit(binop.Right)

	checkNumericOperand(interpreter, binop.Left.GetToken(), left)
	checkNumericOperand(interpreter, binop.Right.GetToken(), right)

	// HACK
	var left_value float32 = 0.0
	var right_value float32 = 0.0

	switch ll := left.(type) {
	case *IntVal:
		left_value = float32(ll.Value)
	case *FloatVal:
		left_value = ll.Value
	}

	switch rr := right.(type) {
	case *IntVal:
		right_value = float32(rr.Value)
	case *FloatVal:
		right_value = rr.Value
	}

	if value, ok := Binop(binop.GetToken().Kind, left_value, right_value); ok {
		return value
	}

	interpreter.report("Invalid binary operation '%s %s %s'", binop.Left.GetToken().Lexeme, binop.Token.Lexeme, binop.Right.GetToken().Lexeme)
	return nil
}

func (interpreter *Interpreter) visitUnaryOp(unary *ast.UnaryOp) Value {
	right := interpreter.Visit(unary.Right)

	switch unary.Token.Kind {
	case lexer.BANG:
		checkBoolOperand(interpreter, unary.Right.GetToken(), right)
		return &BoolVal{Value: !right.(*BoolVal).Value}
	case lexer.MINUS:
		checkNumericOperand(interpreter, unary.Right.GetToken(), right)

		if _, ok := right.(*IntVal); ok {
			return &IntVal{Value: -right.(*IntVal).Value}
		}
		return &FloatVal{Value: -right.(*FloatVal).Value}
	}

	interpreter.report("Invalid unary operation '%s%s'", unary.GetToken().Lexeme, unary.Right.GetToken().Lexeme)
	return nil
}

func (interpreter *Interpreter) visitLogicalOp(logical *ast.LogicalOp) Value {
	left := interpreter.Visit(logical.Left)
	right := interpreter.Visit(logical.Right)

	checkBoolOperand(interpreter, logical.Left.GetToken(), left)
	checkBoolOperand(interpreter, logical.Right.GetToken(), right)

	if logical.Token.Kind == lexer.AND {
		return &BoolVal{Value: left.(*BoolVal).Value && right.(*BoolVal).Value}
	}

	return &BoolVal{Value: left.(*BoolVal).Value || right.(*BoolVal).Value}
}

func (interpreter *Interpreter) visitBlock(block *ast.Block, newEnv bool) Value {
	if newEnv {
		interpreter.push()
	}

	for _, stmt := range block.Statements {
		value := interpreter.Visit(stmt)

		if val, ok := value.(*ReturnValue); ok {
			if newEnv {
				interpreter.pop()
			}

			return val
		}

		if val, ok := value.(*ThrowValue); ok {
			if newEnv {
				interpreter.pop()
			}

			return val
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
	def := &FunctionValue{definition: fndef, bound: nil}
	if insert {
		interpreter.insert(fndef.GetToken().Lexeme, def)
	}
	return def
}

func (interpreter *Interpreter) visitAnonymousFunction(fndef *ast.AnonymousFunction) Value {
	return &AnonFunctionValue{definition: fndef}
}

func (interpreter *Interpreter) visitClassDef(def *ast.ClassDef) Value {
	classDef := &ClassDefValue{identifier: def.GetToken().Lexeme, constructor: nil, fields: make([]string, 0), methods: make(map[string]*FunctionValue, 0)}

	if def.Constructor != nil {
		classDef.constructor = interpreter.visitFunctionDef(def.Constructor, false).(*FunctionValue)
	}

	for id, val := range def.Methods {
		classDef.methods[id] = interpreter.visitFunctionDef(val, false).(*FunctionValue)
	}

	for id := range def.Fields {
		classDef.fields = append(classDef.fields, id)
	}

	interpreter.insert(def.GetToken().Lexeme, classDef)
	return nil
}

func (interpreter *Interpreter) visitCall(call *ast.Call) Value {
	caller := interpreter.Visit(call.Callee)

	if _, ok := caller.(TinyCallable); !ok {
		interpreter.report("'%s' is not callable.", caller.Inspect())
	}

	callable := caller.(TinyCallable)

	arguments := make([]Value, 0, len(call.Arguments))

	for _, arg := range call.Arguments {
		arguments = append(arguments, interpreter.Visit(arg))
	}

	return callable.Call(interpreter, arguments)
}

func (interpreter *Interpreter) visitAssign(assign *ast.Assign) Value {
	value := interpreter.Visit(assign.Expr)
	interpreter.set(assign.GetToken().Lexeme, assign.Operator.Kind, value.Copy())

	return value
}

func (interpreter *Interpreter) visitReturn(ret *ast.Return) Value {
	if ret.Expr != nil {
		return &ReturnValue{inner: interpreter.Visit(ret.Expr)}
	}

	return &ReturnValue{inner: &UnitVal{}}
}

func (interpreter *Interpreter) visitThrow(throw *ast.Throw) Value {
	return &ThrowValue{inner: interpreter.Visit(throw.Expr)}
}

func (interpreter *Interpreter) visitGet(get *ast.Get) Value {
	value := interpreter.Visit(get.Expr)

	switch value.(type) {
	case *ClassInstanceValue:
	default:
		interpreter.report("Cannot use getter on non-instance values '%s'", get.Expr.GetToken().Lexeme)
	}

	if ret, ok := value.(*ClassInstanceValue).Get(get.GetToken().Lexeme); ok {
		return ret
	}

	interpreter.report("'%s' does not contain field '%s'", value.(*ClassInstanceValue).def.identifier, get.GetToken().Lexeme)
	return nil
}

func (interpreter *Interpreter) visitSet(set *ast.Set) Value {
	value := interpreter.Visit(set.Caller)

	switch value.(type) {
	case *ClassInstanceValue:
	default:
		interpreter.report("Cannot use setter on non-instance values '%s':%s %s", set.Caller.GetToken().Lexeme, reflect.TypeOf(set.Caller), reflect.TypeOf(value))
	}

	if ret, ok := value.(*ClassInstanceValue).Set(set.GetToken().Lexeme, interpreter.Visit(set.Expr)); ok {
		return ret
	}

	interpreter.report("'%s' does not contain field '%s'", value.(*ClassInstanceValue).def.identifier, set.GetToken().Lexeme)
	return nil
}

func (interpreter *Interpreter) visitIfStmt(stmt *ast.If) Value {
	interpreter.push()

	if stmt.VarDec != nil {
		interpreter.visitVarDecl(stmt.VarDec)
	}

	condition := interpreter.Visit(stmt.Condition)
	checkBoolOperand(interpreter, stmt.Condition.GetToken(), condition)

	var value Value = nil
	if condition.(*BoolVal).Value {
		value = interpreter.visitBlock(stmt.TrueBody, false)
	} else if stmt.FalseBody != nil {
		value = interpreter.Visit(stmt.FalseBody)
	}

	interpreter.pop()
	return value
}

func (interpreter *Interpreter) visitWhileStmt(stmt *ast.While) Value {
	interpreter.push()

	if stmt.VarDec != nil {
		interpreter.visitVarDecl(stmt.VarDec)
	}

	condition := interpreter.Visit(stmt.Condition)
	checkBoolOperand(interpreter, stmt.Condition.GetToken(), condition)

	var value Value = &UnitVal{}
	for condition.(*BoolVal).Value {
		if ret, ok := interpreter.visitBlock(stmt.Body, false).(*ReturnValue); ok {
			value = ret
			break
		}

		condition = interpreter.Visit(stmt.Condition)

		if stmt.Increment != nil {
			interpreter.Visit(stmt.Increment)
		}
	}

	interpreter.pop()
	return value
}

func (interpreter *Interpreter) visitCatch(catch *ast.Catch) Value {
	interpreter.push()
	value := interpreter.Visit(catch.Expr)

	if thrown, ok := value.(*ThrowValue); ok {
		interpreter.push()
		interpreter.insert(catch.Var.Lexeme, thrown.inner)

		value = interpreter.visitBlock(catch.Body, false)

		interpreter.pop()
	}

	interpreter.pop()
	return value
}
