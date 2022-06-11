package parser

import (
	"testing"
	"tiny/shared"
)

func exprEq(expect string, receive string) bool {
	return expect == receive
}

func exprNeq(expect string, receive string) bool {
	return expect != receive
}

func TestSimpleExpression(t *testing.T) {
	source := shared.ReadFile("../tests/valid/parser/simple_expression.tiny")
	parser := New(source)

	result := parser.Parse().Body.AsSExp()
	if exprNeq(result, "(+ (+ 1 (* 2 3)) (* 1 2))") {
		t.Fatalf("Expression failed '%s'", result)
	}
}
