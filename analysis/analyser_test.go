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
	path := "../tests/valid/analyser/identifier_lookup_assign.tiny"
	source := shared.ReadFile(path)
	program := parser.New(source, path, false).Parse()
	analyser := NewAnalyser(true)

	eq(t, analyser.Run(program.Body), true, "ID lookup failed")
}

func TestFunctionScope(t *testing.T) {
	path := "../tests/valid/analyser/function_scope.tiny"
	source := shared.ReadFile(path)
	program := parser.New(source, path, false).Parse()
	analyser := NewAnalyser(true)

	eq(t, analyser.Run(program.Body), true, "Could not resolve ID from function scope")
}

// --- Invalid ---
func TestInvalidIdentifierLookup(t *testing.T) {
	path := "../tests/invalid/analyser/identifier_lookup_assign.tiny"
	source := shared.ReadFile(path)
	program := parser.New(source, path, false).Parse()
	analyser := NewAnalyser(false)

	eq(t, analyser.Run(program.Body), false, "ID lookup failed")
}
