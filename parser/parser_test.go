package parser

import (
	"testing"
	"tiny/ast"
	"tiny/shared"
)

func exprEq(expect string, receive string) bool {
	return expect == receive
}

func TestSimpleExpression(t *testing.T) {
	path := "../tests/valid/parser/simple_expression.tiny"
	source := shared.ReadFile(path)
	parser := New(source, path, false)

	result := parser.expr(ast.NewBlock(nil)).AsSExp()
	if !exprEq(result, "(+ (+ 1 (* 2 3)) (* 1 2))") {
		t.Fatalf("Expression failed '%s'", result)
	}
}

func TestFunctionDefinition(t *testing.T) {
	path := "../tests/valid/parser/function_definition_no_block.tiny"
	source := shared.ReadFile(path)
	parser := New(source, path, false)

	result := parser.Parse().Body.AsSExp()
	if !exprEq(result, "((function foo (a, b, c)()))") {
		t.Fatalf("Expression failed '%s'", result)
	}
}

func TestFunctionDefinitionMutable(t *testing.T) {
	path := "../tests/valid/parser/function_definition_mutable_no_block.tiny"
	source := shared.ReadFile(path)
	parser := New(source, path, false)

	result := parser.Parse().Body.AsSExp()
	if !exprEq(result, "((function foo (a, b, c)()))") {
		t.Fatalf("Expression failed '%s'", result)
	}
}

func TestVariableDeclaration(t *testing.T) {
	path := "../tests/valid/parser/variable_declaration.tiny"
	source := shared.ReadFile(path)
	parser := New(source, path, false)

	result := parser.Parse().Body.AsSExp()
	if !exprEq(result, "((mut foo (+ 1 2)))") {
		t.Fatalf("Expression failed '%s'", result)
	}
}

func TestFunctionWithBlock(t *testing.T) {
	path := "../tests/valid/parser/function_def_with_block.tiny"
	source := shared.ReadFile(path)
	parser := New(source, path, false)

	result := parser.Parse().Body.AsSExp()
	if !exprEq(result, "((mut a 1)(mut b a)(function foo (a, b, c)((mut d 1)(mut e a))))") {
		t.Fatalf("Expression failed '%s'", result)
	}
}

func TestAssignment(t *testing.T) {
	path := "../tests/valid/parser/assignment.tiny"
	source := shared.ReadFile(path)
	parser := New(source, path, false)

	result := parser.Parse().Body.AsSExp()
	if !exprEq(result, "((mut a 10)(mut b (a = (+ 3 (/ (* 20 3) 2))))(b = 2))") {
		t.Fatalf("Expression failed '%s'", result)
	}
}

func TestAnonymousFunction(t *testing.T) {
	path := "../tests/valid/parser/anonymous_functions.tiny"
	source := shared.ReadFile(path)
	parser := New(source, path, false)

	result := parser.Parse().Body.AsSExp()
	if !exprEq(result, "((function call (x, y, fn)((return fn(x, y))))(mut fn (anon function (a, b)((return (+ a b)))))(mut value1 call(10, 20, fn))(print(value1))(mut value2 call(30, 40, (anon function (a, b)((return (+ a b))))))(print(value2)))") {
		t.Fatalf("Expression failed '%s'", result)
	}
}
