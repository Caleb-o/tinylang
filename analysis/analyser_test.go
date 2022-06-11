package analysis

import (
	"testing"
	"tiny/parser"
	"tiny/shared"
)

func eq(t *testing.T, value bool, expected bool, msg string) {
	if value != expected {
		t.Fatalf("%s: Expected '%t' but received '%t'", msg, expected, value)
	}
}

func TestIdentifierLookup(t *testing.T) {
	source := shared.ReadFile("../tests/valid/analyser/identifier_lookup_assign.tiny")
	program := parser.New(source).Parse()
	analyser := NewAnalyser(true)

	eq(t, analyser.Run(program.Body), true, "ID lookup failed")
}

func TestFunctionScope(t *testing.T) {
	source := shared.ReadFile("../tests/valid/analyser/function_scope.tiny")
	program := parser.New(source).Parse()
	analyser := NewAnalyser(true)

	eq(t, analyser.Run(program.Body), true, "Could not resolve ID from function scope")
}

// --- Invalid ---
func TestInvalidIdentifierLookup(t *testing.T) {
	source := shared.ReadFile("../tests/invalid/analyser/identifier_lookup_assign.tiny")
	program := parser.New(source).Parse()
	analyser := NewAnalyser(true)

	eq(t, analyser.Run(program.Body), false, "ID lookup failed")
}
