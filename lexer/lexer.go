package lexer

import "fmt"

type Lexer struct {
	source            string
	line, column, pos int
}

func New(source string) *Lexer {
	return &Lexer{source, 1, 1, 0}
}

func (lexer *Lexer) Next() *Token {
	lexer.skipWhitespace()

	if lexer.isAtEnd() {
		return lexer.makeEof()
	}

	// Identifiers cannot start with digits
	if isAlpha(lexer.peek()) {
		return lexer.readIdentifier()
	}

	if isDigit(lexer.peek()) {
		return lexer.readDigit()
	}

	// This is a fallthrough, which will return an error *Token otherwise
	return lexer.readSingle()
}

// --- Private ---
func (lexer *Lexer) makeEof() *Token {
	return &Token{EOF, "EndOfFile", lexer.line, lexer.column}
}

func (lexer *Lexer) makeError(msg string, arg ...any) *Token {
	return &Token{ERROR, fmt.Sprintf(msg, arg...), lexer.line, lexer.column}
}

func (lexer *Lexer) makeToken(kind TokenKind, lexeme string, column int) *Token {
	return &Token{kind, lexeme, lexer.line, column}
}

func (lexer *Lexer) peek() byte {
	if lexer.pos >= len(lexer.source) {
		return 0
	}
	return lexer.source[lexer.pos]
}

func (lexer *Lexer) advance() {
	lexer.column++
	lexer.pos++
}

func (lexer *Lexer) isAtEnd() bool {
	return lexer.pos >= len(lexer.source)
}

func isDigit(char byte) bool {
	return char >= '0' && char <= '9'
}

func isAlpha(char byte) bool {
	return char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char == '_'
}

func isIdentifier(char byte) bool {
	return isAlpha(char) || isDigit(char)
}

func getKeyword(lexeme string) TokenKind {
	kind, ok := KeyWords[lexeme]

	if !ok {
		return IDENTIFIER
	} else {
		return kind
	}
}

func (lexer *Lexer) skipWhitespace() {
	for !lexer.isAtEnd() {
		switch lexer.peek() {
		case '#':
			{
				for !lexer.isAtEnd() && lexer.peek() != '\n' {
					lexer.advance()
				}
			}

		case '\n':
			lexer.line++
			lexer.column = 1
			lexer.pos++

		case ' ':
			fallthrough
		case '\t':
			fallthrough
		case '\b':
			fallthrough
		case '\r':
			lexer.advance()

		default:
			return
		}
	}
}

func (lexer *Lexer) readSingle() *Token {
	kind, ok := Characters[lexer.peek()]
	lexer.advance()

	if !ok {
		return lexer.makeError("Unknown character found '%q'", string(lexer.peek()))
	} else {
		return lexer.makeToken(kind, lexer.source[lexer.pos-1:lexer.pos], lexer.column-1)
	}
}

func (lexer *Lexer) readDigit() *Token {
	start := lexer.pos
	start_col := lexer.column
	kind := INT

	// TODO: Allow floats
	for !lexer.isAtEnd() && isDigit(lexer.peek()) {
		lexer.advance()

		if lexer.peek() == '.' {
			if kind == FLOAT {
				return lexer.makeError("Floating point number cannot have multiple decimals %d:%d", lexer.line, lexer.column)
			}

			kind = FLOAT
			lexer.advance()
		}
	}

	return lexer.makeToken(kind, lexer.source[start:lexer.pos], start_col)
}

func (lexer *Lexer) readIdentifier() *Token {
	start := lexer.pos
	start_col := lexer.column

	lexer.advance()

	for !lexer.isAtEnd() && isIdentifier(lexer.peek()) {
		lexer.advance()
	}

	lexeme := lexer.source[start:lexer.pos]
	return lexer.makeToken(getKeyword(lexeme), lexeme, start_col)
}
