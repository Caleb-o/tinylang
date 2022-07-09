package runtime

import (
	"fmt"
	"log"
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

type testStats struct {
	tests  int
	passed int
}

type Interpreter struct {
	env     environment
	in_test bool
	tests   testStats
}

type TinyCallable interface {
	Arity() int
	Call(*Interpreter, []Value) Value
}

func New() *Interpreter {
	interpreter := &Interpreter{env: environment{variables: make([]map[string]Value, 0, 1), depth: 0}, in_test: false, tests: testStats{0, 0}}
	interpreter.push()

	return interpreter
}

func (interpreter *Interpreter) Run(program *ast.Program) {
	// TODO: Might return a value to the caller of the interpreter run
	result := interpreter.visitBlock(program.Body, false)
	interpreter.pop()

	if res, ok := result.(*ThrowValue); ok {
		interpreter.Report("Uncaught value thrown '%s'", res.inner.Inspect())
	}

	if interpreter.tests.tests > 0 {
		shared.Info(fmt.Sprintf("Tests passed [%d/%d]", interpreter.tests.passed, interpreter.tests.tests))
	}
}

func (interpreter *Interpreter) Import(identifier string, value Value) {
	interpreter.env.variables[interpreter.env.depth-1][identifier] = value
}

func (interpreter *Interpreter) insert(identifier string, value Value) {
	interpreter.env.variables[interpreter.env.depth-1][identifier] = value
}

func (interpreter *Interpreter) set(identifier string, operator lexer.TokenKind, value Value) {
	for idx := interpreter.env.depth - 1; idx >= 0; idx -= 1 {
		if _, ok := interpreter.env.variables[idx][identifier]; ok {
			if operator == lexer.EQUAL {
				interpreter.env.variables[idx][identifier] = value.Copy()
			} else {
				if ok := interpreter.env.variables[idx][identifier].Modify(operator, value); !ok {
					interpreter.Report("Cannot use operation '%s' on '%s'", operator.Name(), identifier)
					return
				}
			}
			return
		}
	}

	// Should not happen, but just to be safe
	interpreter.Report("Unknown identifier name in lookup '%s'", identifier)
}

func (interpreter *Interpreter) lookup(identifier string) Value {
	for idx := len(interpreter.env.variables) - 1; idx >= 0; idx -= 1 {
		if value, ok := interpreter.env.variables[idx][identifier]; ok {
			return value
		}
	}

	// Should not happen, but just to be safe
	interpreter.Report("Unknown identifier name in lookup '%s'", identifier)
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
func (interpreter *Interpreter) Report(msg string, args ...any) {
	if interpreter.in_test {
		interpreter.ReportTest(msg, nil)
	} else {
		res := fmt.Sprintf(msg, args...)
		shared.ReportErrFatal("Runtime: " + res)
	}
}

func (interpreter *Interpreter) ReportT(msg string, token *lexer.Token, args ...any) {
	if interpreter.in_test {
		interpreter.ReportTest(msg, token)
	} else {
		res := fmt.Sprintf(msg, args...)
		res = fmt.Sprintf("Runtime: %s [%d:%d]", res, token.Line, token.Column)
		shared.ReportErrFatal(res)
	}
}

func (interpreter *Interpreter) ReportTest(msg string, token *lexer.Token) {
	if token == nil {
		log.Printf("Runtime - Test: %s\n", msg)
	} else {
		log.Printf("Runtime - Test: %s [%d:%d]\n", msg, token.Line, token.Column)
	}
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
	case *ast.Unit:
		return &UnitVal{}
	case *ast.ListLiteral:
		return interpreter.visitList(n)
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
	case *ast.StructDef:
		return interpreter.visitStructDef(n)
	case *ast.NameSpace:
		return interpreter.visitNamespace(n)
	case *ast.Argument:
		return interpreter.Visit(n.Expr)
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
	case *ast.Index:
		return interpreter.visitIndex(n)
	case *ast.IndexSet:
		return interpreter.visitIndexSet(n)
	case *ast.Self:
		return interpreter.lookup("self")
	case *ast.If:
		return interpreter.visitIfStmt(n)
	case *ast.While:
		return interpreter.visitWhileStmt(n)
	case *ast.Catch:
		return interpreter.visitCatch(n)
	case *ast.Test:
		return interpreter.visitTest(n)
	case *ast.Break:
		return interpreter.visitLoopBreak()
	case *ast.Continue:
		return interpreter.visitLoopContinue()
	case *ast.Match:
		return interpreter.visitMatchCase(n)

	// Ignore
	case *ast.Import:
		return nil
	}

	interpreter.ReportT("Unhandled node in Visit '%s'", node.GetToken(), node.GetToken().Lexeme)
	return nil
}

func (interpreter *Interpreter) visitBinaryOp(binop *ast.BinaryOp) Value {
	// FIXME: Analysis should make sure all expressions are of the same type
	left := interpreter.Visit(binop.Left)
	right := interpreter.Visit(binop.Right)

	if reflect.TypeOf(left) != reflect.TypeOf(right) {
		interpreter.ReportT("Invalid binary operation '%s %s %s'", binop.Left.GetToken(), binop.Left.GetToken().Lexeme, binop.Token.Lexeme, binop.Right.GetToken().Lexeme)
		return nil
	}

	switch left.(type) {
	case *IntVal:
		if value, ok := BinopI(binop.GetToken().Kind, left.(*IntVal).Value, right.(*IntVal).Value); ok {
			return value
		}

	case *FloatVal:
		if value, ok := BinopF(binop.GetToken().Kind, left.(*FloatVal).Value, right.(*FloatVal).Value); ok {
			return value
		}

	case *BoolVal:
		if value, ok := BinopB(binop.GetToken().Kind, left.(*BoolVal).Value, right.(*BoolVal).Value); ok {
			return value
		}

	case *StringVal:
		if value, ok := BinopS(binop.GetToken().Kind, left.(*StringVal).Value, right.(*StringVal).Value); ok {
			return value
		}

	case *ListVal:
		if value, ok := BinopL(binop.GetToken().Kind, left.(*ListVal).Values, right.(*ListVal).Values); ok {
			return value
		}
	}

	interpreter.ReportT("Invalid binary operation '%s %s %s'", binop.Left.GetToken(), binop.Left.GetToken().Lexeme, binop.Token.Lexeme, binop.Right.GetToken().Lexeme)
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

	interpreter.ReportT("Invalid unary operation '%s%s'", unary.GetToken(), unary.GetToken().Lexeme, unary.Right.GetToken().Lexeme)
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
		defer interpreter.pop()
	}

	for _, stmt := range block.Statements {
		switch value := interpreter.Visit(stmt).(type) {
		case *ReturnValue:
			return value
		case *ThrowValue:
			return value
		}
	}

	return &UnitVal{}
}

func (interpreter *Interpreter) visitList(lit *ast.ListLiteral) Value {
	values := make([]Value, 0, len(lit.Exprs))

	for _, value := range lit.Exprs {
		values = append(values, interpreter.Visit(value))
	}

	return &ListVal{values}
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

	interpreter.ReportT("Unknown literal type found '%s'", lit.GetToken(), lit.GetToken().Lexeme)
	return &UnitVal{}
}

func (interpreter *Interpreter) visitIdentifier(id *ast.Identifier) Value {
	return interpreter.lookup(id.GetToken().Lexeme)
}

func (interpreter *Interpreter) visitVarDecl(decl *ast.VariableDecl) Value {
	value := interpreter.Visit(decl.Expr).Copy()
	interpreter.insert(decl.GetToken().Lexeme, value)
	return value
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
	classDef := &ClassDefValue{identifier: def.GetToken().Lexeme, constructor: nil, fields: make([]string, 0, len(def.Fields)), methods: make(map[string]Value, len(def.Methods))}

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
	return classDef
}

func (interpreter *Interpreter) visitStructDef(def *ast.StructDef) Value {
	structDef := &StructDefValue{identifier: def.GetToken().Lexeme, constructor: nil, fields: make([]string, 0, len(def.Fields))}

	if def.Constructor != nil {
		structDef.constructor = interpreter.visitFunctionDef(def.Constructor, false).(*FunctionValue)
	}

	for id := range def.Fields {
		structDef.fields = append(structDef.fields, id)
	}

	interpreter.insert(def.GetToken().Lexeme, structDef)
	return structDef
}

func (interpreter *Interpreter) visitNamespace(ns *ast.NameSpace) Value {
	namespace := &NameSpaceValue{Identifier: ns.Token.Lexeme, Members: make(map[string]Value)}

	for _, stmt := range ns.Body.Statements {
		namespace.Members[stmt.GetToken().Lexeme] = interpreter.Visit(stmt)
	}

	interpreter.insert(ns.Token.Lexeme, namespace)
	return namespace
}

func (interpreter *Interpreter) visitCall(call *ast.Call) Value {
	caller := interpreter.Visit(call.Callee)

	if _, ok := caller.(TinyCallable); !ok {
		interpreter.ReportT("'%s' is not callable.", call.GetToken(), caller.Inspect())
	}

	callable := caller.(TinyCallable)

	if callable.Arity() != len(call.Arguments) {
		interpreter.ReportT(
			"Function '%s' expected %d arguments but received %d",
			call.Token,
			call.Token.Lexeme,
			callable.Arity(),
			len(call.Arguments),
		)
	}

	arguments := make([]Value, 0, len(call.Arguments))

	for _, arg := range call.Arguments {
		arguments = append(arguments, interpreter.Visit(arg).Copy())
	}

	return callable.Call(interpreter, arguments)
}

func (interpreter *Interpreter) visitAssign(assign *ast.Assign) Value {
	value := interpreter.Visit(assign.Expr).Copy()
	interpreter.set(assign.GetToken().Lexeme, assign.Operator.Kind, value)

	return value
}

func (interpreter *Interpreter) visitReturn(ret *ast.Return) Value {
	if ret.Expr != nil {
		return &ReturnValue{inner: interpreter.Visit(ret.Expr).Copy()}
	}

	return &ReturnValue{inner: &UnitVal{}}
}

func (interpreter *Interpreter) visitThrow(throw *ast.Throw) Value {
	innerValue := interpreter.Visit(throw.Expr)

	// Re-throw the value if it is already a throw
	if inner, ok := innerValue.(*ThrowValue); ok {
		return inner.Copy()
	}

	return &ThrowValue{inner: innerValue.Copy()}
}

func (interpreter *Interpreter) visitGet(get *ast.Get) Value {
	value := interpreter.Visit(get.Expr)

	switch inner := value.(type) {
	case *ClassInstanceValue:
		if ret, ok := inner.Get(get.GetToken().Lexeme); ok {
			return ret.Copy()
		}
	case *StructInstanceValue:
		if ret, ok := inner.Get(get.GetToken().Lexeme); ok {
			return ret.Copy()
		}
	case *NameSpaceValue:
		if ret, ok := inner.Get(get.GetToken().Lexeme); ok {
			return ret.Copy()
		}
	}

	interpreter.ReportT("Cannot use getter on non-instance values '%s':%s", get.Expr.GetToken(), get.Expr.GetToken().Lexeme, reflect.TypeOf(value))
	return nil
}

func (interpreter *Interpreter) visitSet(set *ast.Set) Value {
	caller := interpreter.Visit(set.Caller)
	obj := interpreter.Visit(set.Expr)

	switch t := caller.(type) {
	case *ClassInstanceValue:
		if ret, ok := t.Set(set.GetToken().Lexeme, obj); ok {
			return ret.Copy()
		}
	case *StructInstanceValue:
		if ret, ok := t.Set(set.GetToken().Lexeme, obj); ok {
			return ret.Copy()
		}
	case *NameSpaceValue:
		if ret, ok := t.Set(set.GetToken().Lexeme, obj); ok {
			return ret.Copy()
		}
	}

	interpreter.ReportT("Cannot use setter on non-instance values '%s':%s %s", set.Caller.GetToken(), set.Caller.GetToken().Lexeme, reflect.TypeOf(set.Caller), reflect.TypeOf(caller))
	return nil
}

func (interpreter *Interpreter) visitIndex(index *ast.Index) Value {
	caller := interpreter.Visit(index.Caller)
	indexer := interpreter.Visit(index.Expr)

	if _, ok := indexer.(*IntVal); !ok {
		interpreter.ReportT("Index must use an integer value but received '%d'", index.GetToken(), indexer.Inspect())
	}

	indexer_int := indexer.(*IntVal).Value

	switch t := caller.(type) {
	case *ListVal:
		if indexer_int < 0 || indexer_int >= len(t.Values) {
			interpreter.ReportT("Index %d is out of list range 0-%d", index.GetToken(), indexer_int, len(t.Values)-1)
		}
		return t.Values[indexer_int]
	case *StringVal:
		if indexer_int < 0 || indexer_int >= len(t.Value) {
			interpreter.ReportT("Index %d is out of string range 0-%d", index.GetToken(), indexer_int, len(t.Value)-1)
		}

		// FIXME: Replace with char type
		return &StringVal{Value: string(t.Value[indexer_int])}
	}

	interpreter.ReportT("Cannot use index on '%s':'%s'", index.Caller.GetToken(), index.Caller.GetToken().Lexeme, reflect.TypeOf(caller))
	return nil
}

func (interpreter *Interpreter) visitIndexSet(iset *ast.IndexSet) Value {
	caller := interpreter.Visit(iset.Idx.Caller)
	index := interpreter.Visit(iset.Idx.Expr)
	value := interpreter.Visit(iset.Expr)

	indexer_int := index.(*IntVal).Value

	switch t := caller.(type) {
	case *ListVal:
		if indexer_int < 0 || indexer_int >= len(t.Values) {
			interpreter.ReportT("Index %d is out of list range 0-%d", iset.GetToken(), indexer_int, len(t.Values)-1)
		}
		if ret, ok := t.Set(iset.Token.Kind, indexer_int, value); ok {
			return ret
		}
	case *StringVal:
		if indexer_int < 0 || indexer_int >= len(t.Value) {
			interpreter.ReportT("Index %d is out of string range 0-%d", iset.GetToken(), indexer_int, len(t.Value)-1)
		}

		// FIXME: Make this better and only allow chars
		new_str := []byte(t.Value)
		new_str[indexer_int] = byte(value.Inspect()[0])
		t.Value = string(new_str)

		return t
	}

	interpreter.ReportT("Cannot use index on '%s':'%s'", iset.GetToken(), iset.Idx.Caller.GetToken().Lexeme, reflect.TypeOf(caller), iset.Idx.Caller.GetToken().Line, iset.Idx.Caller.GetToken().Column)
	return nil
}

func (interpreter *Interpreter) visitIfStmt(stmt *ast.If) Value {
	interpreter.push()
	defer interpreter.pop()

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

	return value
}

func (interpreter *Interpreter) visitWhileStmt(stmt *ast.While) Value {
	interpreter.push()
	defer interpreter.pop()

	if stmt.VarDec != nil {
		interpreter.visitVarDecl(stmt.VarDec)
	}

	condition := interpreter.Visit(stmt.Condition)
	checkBoolOperand(interpreter, stmt.Condition.GetToken(), condition)

	for condition.(*BoolVal).Value {
		if throw, ok := interpreter.visitBlock(stmt.Body, false).(*ReturnValue); ok {
			if loop, ok := throw.inner.(*LoopFlow); ok {
				if loop.exit {
					break
				}
			} else {
				return throw
			}
		}

		if stmt.Increment != nil {
			interpreter.Visit(stmt.Increment)
		}

		condition = interpreter.Visit(stmt.Condition)
	}

	return &UnitVal{}
}

func (interpreter *Interpreter) visitCatch(catch *ast.Catch) Value {
	interpreter.push()
	defer interpreter.pop()

	value := interpreter.Visit(catch.Expr)

	if thrown, ok := value.(*ThrowValue); ok {
		interpreter.push()
		interpreter.insert(catch.Var.Lexeme, thrown.inner)

		value = interpreter.visitBlock(catch.Body, false)

		interpreter.pop()
	}

	return value
}

func (interpreter *Interpreter) visitTest(test *ast.Test) Value {
	enclosing := interpreter.in_test
	interpreter.in_test = true

	interpreter.tests.tests += 1

	switch value := interpreter.visitBlock(test.Body, true).(type) {
	case *ThrowValue:
		interpreter.ReportTest(fmt.Sprintf("'%s' failed with '%s'", test.Token.Lexeme, value.Inspect()), test.GetToken())
	default:
		interpreter.tests.passed += 1
	}

	interpreter.in_test = enclosing
	return &UnitVal{}
}

func (interpreter *Interpreter) visitLoopBreak() Value {
	return &ThrowValue{&LoopFlow{true}}
}

func (interpreter *Interpreter) visitLoopContinue() Value {
	return &ThrowValue{&LoopFlow{false}}
}

func (interpreter *Interpreter) visitMatchCase(match *ast.Match) Value {
	value := interpreter.Visit(match.Expr)

	for _, expr := range match.Cases {
		caseExpr := interpreter.Visit(expr.Expr)

		if Equality(value, caseExpr) {
			return interpreter.Visit(expr.Body)
		}
	}

	if match.CatchAll != nil {
		return interpreter.Visit(match.CatchAll)
	}

	return &UnitVal{}
}
