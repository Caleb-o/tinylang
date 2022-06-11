package parser

import (
	"testing"
	"tiny/shared"
)

func exprEq(expect string, receive string) bool {
	return expect == receive
}

func TestSimpleExpression(t *testing.T) {
	source := shared.ReadFile("../tests/valid/parser/simple_expression.tiny")
	parser := New(source)

	result := parser.Parse().Body.AsSExp()
	if !exprEq(result, "((+ (+ 1 (* 2 3)) (* 1 2)))") {
		t.Fatalf("Expression failed '%s'", result)
	}
}

func TestFunctionDefinition(t *testing.T) {
	source := shared.ReadFile("../tests/valid/parser/function_definition_no_block.tiny")
	parser := New(source)

	result := parser.Parse().Body.AsSExp()
	if !exprEq(result, "((foo (a: any, b: any, c: any): any)()))") {
		t.Fatalf("Expression failed '%s'", result)
	}
}

func TestFunctionDefinitionMutable(t *testing.T) {
	source := shared.ReadFile("../tests/valid/parser/function_definition_mutable_no_block.tiny")
	parser := New(source)

	result := parser.Parse().Body.AsSExp()
	if !exprEq(result, "((foo (mut a: any, mut b: any, c: any): any)()))") {
		t.Fatalf("Expression failed '%s'", result)
	}
}
