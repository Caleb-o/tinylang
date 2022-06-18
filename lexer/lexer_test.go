package lexer

import (
	"testing"
	"tiny/shared"
)

func TestValidTokens(t *testing.T) {
	source := shared.ReadFile("../tests/valid/lexer/tokens.tiny")
	lexer := New(source)

	for {
		token := lexer.Next()

		if token.Kind == EOF {
			break
		}

		if token.Kind == ERROR {
			t.Fatal(token.Lexeme)
		}
	}
}
