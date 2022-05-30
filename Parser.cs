namespace TinyLang {
	class Parser {
		readonly Lexer? lexer;
		Token? currentToken;
		Application app;


		public Parser(string source) {
			this.app = new Application(new Block(new List<Node>()));
			this.lexer = new Lexer(source);
			currentToken = this.lexer.Next();
		}

		public Application Parse() {
			Declaration();
			return app;
		}

		void Error(string message) {
			throw new Exception($"{message} :: [{currentToken?.Column}:{currentToken?.Line}]");
		}

		void Consume(TokenKind expects) {
			if (currentToken?.Kind == expects) {
				currentToken = lexer?.Next();
			} else {
				Error($"Expected token kind {expects} but received {currentToken?.Kind}");
			}
		}

		void ConsumeIfExists(TokenKind expects) {
			if (currentToken?.Kind == expects) {
				currentToken = lexer?.Next();
			}
		}

		List<Parameter> ParameterList() {
			List<Parameter> parameters = new List<Parameter>();

			Consume(TokenKind.OpenParen);

			while(currentToken?.Kind != TokenKind.CloseParen) {
				List<Token?> identifiers = new List<Token?>();

				while(currentToken?.Kind == TokenKind.Identifier) {
					identifiers.Add(currentToken);
					Consume(TokenKind.Identifier);
					ConsumeIfExists(TokenKind.Comma);
				}

				Consume(TokenKind.Colon);

				Token? type_identifier = currentToken;
				Consume(TokenKind.Identifier);

				foreach(Token? id in identifiers) {
					parameters.Add(new Parameter(id, type_identifier));
				}
				ConsumeIfExists(TokenKind.Comma);
			}

			Consume(TokenKind.CloseParen);
			return parameters;
		}

		Var Variable() {
			Token? token = currentToken;
			Consume(TokenKind.Identifier);

			return new Var(token);
		}

		Node? Factor() {
			Token? current = currentToken;

			switch(currentToken?.Kind) {
				case TokenKind.Minus: {
					Consume(TokenKind.Minus);
					return new UnaryOp(current, Factor());
				}

				case TokenKind.Int:
				case TokenKind.Float:
				case TokenKind.Boolean:
				case TokenKind.String: {
					Consume(currentToken.Kind);
					return new Literal(current);
				}

				case TokenKind.OpenParen: {
					Consume(TokenKind.OpenParen);
					Node? n = Expr();
					Consume(TokenKind.CloseParen);
					return n;
				}

				default: {
					// Defaults to an identifier
					return Variable();
				}
			}
		}

		Node? Term() {
			Node? n = Factor();

			while(currentToken?.Kind == TokenKind.Star || currentToken?.Kind == TokenKind.Slash) {
				Token? op = currentToken;
				Consume(currentToken.Kind);
				n = new BinOp(op, n, Factor());
			}

			return n;
		}

		Node? Expr() {
			Node? n = Term();

			while(currentToken?.Kind == TokenKind.Plus || currentToken?.Kind == TokenKind.Minus) {
				Token? op = currentToken;
				Consume(currentToken.Kind);
				n = new BinOp(op, n, Term());
			}

			return n;
		}

		List<Node> GetArguments(TokenKind closing) {
			List<Node> arguments = new List<Node>();
			while(currentToken?.Kind != closing) {
				arguments.Add(Expr());
				ConsumeIfExists(TokenKind.Comma);
			}
			return arguments;
		}

		void Assignment(Block block) {
			Consume(TokenKind.Var);

			Token? type_id = currentToken;
			Consume(TokenKind.Identifier);

			Token? identifier = currentToken;
			Consume(TokenKind.Identifier);

			Consume(TokenKind.Equals);
			Node? expr = Expr();

			block.statements.Add(new Assignment(identifier?.Lexeme, type_id, expr));
		}

		void FunctionCall(Block block) {
			Token? identifier = currentToken;
			Consume(TokenKind.Identifier);

			Consume(TokenKind.OpenParen);
			List<Node> arguments = GetArguments(TokenKind.CloseParen);
			Consume(TokenKind.CloseParen);

			block.statements.Add(new FunctionCall(identifier, arguments));
		}

		void BuiltinFnCall(Block block) {
			Consume(TokenKind.At);

			Token? identifier = currentToken;
			Consume(TokenKind.Identifier);

			Consume(TokenKind.OpenParen);
			List<Node> arguments = GetArguments(TokenKind.CloseParen);
			Consume(TokenKind.CloseParen);

			block.statements.Add(new BuiltinFunctionCall(identifier?.Lexeme, arguments, "void"));
		}

		Block Body() {
			Block block = new Block(new List<Node>());

			Consume(TokenKind.OpenCurly);
			StatementList(block, TokenKind.CloseCurly);
			Consume(TokenKind.CloseCurly);

			return block;
		}

		void FunctionDef() {
			Consume(TokenKind.Function);

			Token? identifier = currentToken;
			Consume(TokenKind.Identifier);

			List<Parameter> parameters = ParameterList();

			Identifier? return_type = null;
			if (currentToken?.Kind == TokenKind.Colon) {
				Consume(TokenKind.Colon);

				Token? return_identifier = currentToken;
				Consume(TokenKind.Identifier);

				return_type = new Identifier(return_identifier?.Lexeme);
			} else {
				// Default to void
				return_type = new Identifier("void");
			}

			app.block.statements.Add(new FunctionDef(identifier, parameters, return_type, Body()));
		}

		void StatementList(Block block, TokenKind closing) {
			while(currentToken?.Kind != closing) {
				Statement(block);
			}
		}

		void Statement(Block block) {
			switch(currentToken?.Kind) {
				case TokenKind.Var: {
					Assignment(block);
					break;
				}

				case TokenKind.At: {
					BuiltinFnCall(block);
					break;
				}

				case TokenKind.Identifier: {
					FunctionCall(block);
					break;
				}

				default:
					Error($"Unknown token in statement: {currentToken?.Kind}");
					break;
			}

			Consume(TokenKind.SemiColon);
		}

		void Declaration() {
			// Record defs
			// Functions

			// -- Shared with body statements
			// Variables
			// Statements
			while (currentToken?.Kind != TokenKind.End) {
				switch(currentToken?.Kind) {
					case TokenKind.Function: {
						FunctionDef();
						break;
					}

					case TokenKind.Identifier:
					case TokenKind.Var: {
						StatementList(app.block, TokenKind.End);
						break;
					}

					default:
						Error($"{currentToken?.Kind} is not implemented");
						break;
				}
			}
		}
	}
}