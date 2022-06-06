using System;
using System.Collections.Generic;

namespace TinyLang {
	enum TokenKind {
		End, Error,

		String, Int, Float, Boolean,

		Plus, Minus, Star, Slash, At,

		Let, Var, Identifier, Struct, Function,
		Return, If, Else, While, Do, Auto,
		Dot, Comma, Colon, SemiColon, Equals,
		OpenParen, CloseParen, OpenCurly, CloseCurly,
		OpenSquare, CloseSquare, New,

		NotEqual, EqualEqual, Greater, Less, GreaterEqual, LessEqual,
	}
	
	class Token {
		public readonly string Lexeme;
		public readonly TokenKind Kind;
		public readonly int Column, Line;

		public Token(string lexeme, TokenKind kind, int column, int line) {
			this.Lexeme = lexeme;
			this.Kind = kind;
			this.Column = column;
			this.Line = line;
		}
	}

	class Lexer {
		readonly string source;
		int ip, line, column;

		static Dictionary<string, TokenKind> KeyWords = new Dictionary<string, TokenKind>() {
			{ "var", TokenKind.Var },
			{ "let", TokenKind.Let },
			{ "auto", TokenKind.Auto },
			{ "return", TokenKind.Return },
			{ "fn", TokenKind.Function },
			{ "new", TokenKind.New },
			{ "struct", TokenKind.Struct },
			{ "do", TokenKind.Do },
			{ "while", TokenKind.While },
			{ "if", TokenKind.If },
			{ "else", TokenKind.Else },
			{ "true", TokenKind.Boolean },
			{ "false", TokenKind.Boolean },
		};

		static Dictionary<char, TokenKind> Singles = new Dictionary<char, TokenKind>() {
			{ '+', TokenKind.Plus },
			{ '-', TokenKind.Minus },
			{ '*', TokenKind.Star },
			{ '/', TokenKind.Slash },
			{ '@', TokenKind.At },
			{ '.', TokenKind.Dot },
			{ ',', TokenKind.Comma },
			{ ':', TokenKind.Colon },
			{ ';', TokenKind.SemiColon },
			{ '=', TokenKind.Equals },
			{ '(', TokenKind.OpenParen },
			{ ')', TokenKind.CloseParen },
			{ '{', TokenKind.OpenCurly },
			{ '}', TokenKind.CloseCurly },
			{ '[', TokenKind.OpenSquare },
			{ ']', TokenKind.CloseSquare },
			{ '>', TokenKind.Greater },
			{ '<', TokenKind.Less },
		};

		public Lexer(string source) {
			this.source = source;
			this.line = 1;
			this.column = 1;
		}

		public Token Next() {
			SkipWhitespace();

			if (ip >= source.Length) {
				return new Token("EOF", TokenKind.End, column, line);
			}

			if (IsIdentifier(source[ip])) {
				return Identifier();
			}

			if (Char.IsDigit(source[ip])) {
				return Number();
			}

			if (source[ip] == '"') {
				return String();
			}

			if (Peek() == '!' && Peek(1) == '=') {
				return MakeDouble(TokenKind.NotEqual);
			}

			if (Peek() == '=' && Peek(1) == '=') {
				return MakeDouble(TokenKind.EqualEqual);
			}

			if (Peek() == '>' && Peek(1) == '=') {
				return MakeDouble(TokenKind.GreaterEqual);
			}

			if (Peek() == '<' && Peek(1) == '=') {
				return MakeDouble(TokenKind.LessEqual);
			}
			
			if (Singles.ContainsKey(source[ip])) {
				Token t = new Token(source[ip].ToString(), Singles[source[ip]], column, line);
				Advance();
				return t;
			}

			return new Token("ERROR", TokenKind.End, column, line);
		}

		Token MakeDouble(TokenKind kind) {
			Advance();
			Advance();
			return new Token(source[(ip - 2)..ip], kind, column - 2, line);
		}

		Token Identifier() {
			int start = ip;
			int col = column;
			TokenKind kind = TokenKind.Identifier;

			while (ip < source.Length && (IsIdentifier(source[ip]) || char.IsDigit(source[ip]))) {
				Advance();
			}

			string lexeme = source[start..ip];

			if (KeyWords.ContainsKey(lexeme)) {
				kind = KeyWords[lexeme];
			}

			return new Token(lexeme, kind, col, line);
		}

		Token String() {
			Advance();

			int start = ip;
			int col = column;

			while (ip < source.Length && source[ip] != '"') {
				Advance();

				if (ip >= source.Length) {
					throw new Exception("Unterminated string");
				}
			}
			Advance();

			return new Token(source[start..(ip-1)], TokenKind.String, col, line);
		}

		Token Number() {
			int start = ip;
			int col = column;
			TokenKind kind = TokenKind.Int;

			while (ip < source.Length && Char.IsDigit(source[ip])) {
				Advance();

				if (ip < source.Length && source[ip] == '.') {
					if (kind == TokenKind.Float) {
						throw new Exception("Floating point number contains multiple decimals");
					}

					kind = TokenKind.Float;
					Advance();
				}
			}

			return new Token(source[start..ip], kind, col, line);
		}

		bool IsIdentifier(char ch) {
			return Char.IsLetter(ch) || ch == '_';
		}

		void Advance() {
			ip++;
			column++;
		}

		char Peek(int depth = 0) {
			if (ip + depth >= source.Length) return '\0';
			return source[ip + depth];
		}

		void SkipWhitespace() {
			while (ip < source.Length) {
				switch(source[ip]) {
					case '#': {
						while (ip < source.Length && source[ip] != '\n') {
							Advance();
						}
						break;
					}

					case '\n': {
						line++;
						ip++;
						column = 1;
						break;
					}

					case ' ':
					case '\t':
					case '\b':
					case '\r':
						Advance();
						break;
					
					default:
						return;
				}
			}
		}
	}
}