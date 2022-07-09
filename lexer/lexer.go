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

	if lexer.peek() == '"' {
		return lexer.readString()
	}

	// This is a fallthrough, which will return an error *Token otherwise
	return lexer.readChars()
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

func (lexer *Lexer) match(expected byte) bool {
	if lexer.isAtEnd() {
		return false
	}
	if lexer.peek() != expected {
		return false
	}

	lexer.advance()
	return true
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

		case ' ', '\t', '\b', '\r':
			lexer.advance()

		default:
			return
		}
	}
}

func (lexer *Lexer) readChars() *Token {
	current := lexer.peek()
	lexer.advance()

	size := 1
	kind := ERROR

	switch current {
	case '+':
		if lexer.match('=') {
			kind = PLUS_EQUAL
			size = 2
			break
		}
		kind = PLUS
	case '-':
		if lexer.match('=') {
			kind = MINUS_EQUAL
			size = 2
			break
		}
		kind = MINUS
	case '*':
		if lexer.match('=') {
			kind = STAR_EQUAL
			size = 2
			break
		}
		kind = STAR
	case '/':
		if lexer.match('=') {
			kind = SLASH_EQUAL
			size = 2
			break
		}
		kind = SLASH
	case '(':
		kind = OPENPAREN
	case ')':
		kind = CLOSEPAREN
	case '{':
		kind = OPENCURLY
	case '}':
		kind = CLOSECURLY
	case '[':
		kind = OPENSQUARE
	case ']':
		kind = CLOSESQUARE
	case ':':
		kind = COLON
	case ';':
		kind = SEMICOLON
	case '.':
		kind = DOT
	case ',':
		kind = COMMA

	case '&':
		if lexer.match('&') {
			kind = AND
			size = 2
			break
		}
		kind = AMPERSAND

	case '|':
		if lexer.match('|') {
			kind = OR
			size = 2
			break
		}
		kind = PIPE

	case '=':
		if lexer.match('=') {
			kind = EQUAL_EQUAL
			size = 2
			break
		} else if lexer.match('>') {
			kind = FAT_ARROW
			size = 2
			break
		}
		kind = EQUAL

	case '!':
		if lexer.match('=') {
			kind = NOT_EQUAL
			size = 2
			break
		}
		kind = BANG

	case '>':
		if lexer.match('=') {
			kind = GREATER_EQUAL
			size = 2
			break
		}
		kind = GREATER

	case '<':
		if lexer.match('=') {
			kind = LESS_EQUAL
			size = 2
			break
		}
		kind = LESS

	default:
		return lexer.makeError("Unknown character found '%q'", string(current))
	}

	return lexer.makeToken(kind, lexer.source[lexer.pos-size:lexer.pos], lexer.column-size)
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

func (lexer *Lexer) readString() *Token {
	lexer.advance()

	start := lexer.pos
	start_col := lexer.column

	for !lexer.isAtEnd() && lexer.peek() != '"' {
		lexer.advance()
	}

	lexer.advance()
	return lexer.makeToken(STRING, lexer.source[start:lexer.pos-1], start_col)
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
