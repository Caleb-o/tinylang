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
	source := shared.ReadFile("../tests/valid/parser/simple_expression.tiny")
	parser := New(source)

	result := parser.expr(ast.NewBlock(nil)).AsSExp()
	if !exprEq(result, "(+ (+ 1 (* 2 3)) (* 1 2))") {
		t.Fatalf("Expression failed '%s'", result)
	}
}

func TestFunctionDefinition(t *testing.T) {
	source := shared.ReadFile("../tests/valid/parser/function_definition_no_block.tiny")
	parser := New(source)

	result := parser.Parse().Body.AsSExp()
	if !exprEq(result, "((function foo (a: any, b: any, c: any): any)()))") {
		t.Fatalf("Expression failed '%s'", result)
	}
}

func TestFunctionDefinitionMutable(t *testing.T) {
	source := shared.ReadFile("../tests/valid/parser/function_definition_mutable_no_block.tiny")
	parser := New(source)

	result := parser.Parse().Body.AsSExp()
	if !exprEq(result, "((function foo (mut a: any, mut b: any, c: any): any)()))") {
		t.Fatalf("Expression failed '%s'", result)
	}
}

func TestFunctionDefinitionNested(t *testing.T) {
	source := shared.ReadFile("../tests/valid/parser/function_definition_nested.tiny")
	parser := New(source)

	result := parser.Parse().Body.AsSExp()
	if !exprEq(result, "((function foo (a: any): any)((function bar (b: any): any)()))))") {
		t.Fatalf("Expression failed '%s'", result)
	}
}

func TestImmutableVariableDeclaration(t *testing.T) {
	source := shared.ReadFile("../tests/valid/parser/immutable_variable_declaration.tiny")
	parser := New(source)

	result := parser.Parse().Body.AsSExp()
	if !exprEq(result, "((foo (+ 1 2)))") {
		t.Fatalf("Expression failed '%s'", result)
	}
}

func TestMutableVariableDeclaration(t *testing.T) {
	source := shared.ReadFile("../tests/valid/parser/mutable_variable_declaration.tiny")
	parser := New(source)

	result := parser.Parse().Body.AsSExp()
	if !exprEq(result, "((mut foo (+ 1 2)))") {
		t.Fatalf("Expression failed '%s'", result)
	}
}

func TestFunctionWithBlock(t *testing.T) {
	source := shared.ReadFile("../tests/valid/parser/function_def_with_block.tiny")
	parser := New(source)

	result := parser.Parse().Body.AsSExp()
	if !exprEq(result, "((mut a 1)(mut b a)(function foo (a: any, b: any, c: any): any)((mut d 1)(mut e a))))") {
		t.Fatalf("Expression failed '%s'", result)
	}
}
