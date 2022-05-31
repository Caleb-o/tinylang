using System;
using System.Collections.Generic;

namespace TinyLang {
	class Parser {
		readonly Lexer lexer;
		Token currentToken;
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
			throw new Exception($"{message} :: [{currentToken.Column}:{currentToken.Line}]");
		}

		void Consume(TokenKind expects) {
			if (currentToken.Kind == expects) {
				currentToken = lexer.Next();
			} else {
				Error($"Expected token kind {expects} but received {currentToken.Kind}");
			}
		}

		void ConsumeIfExists(TokenKind expects) {
			if (currentToken.Kind == expects) {
				currentToken = lexer.Next();
			}
		}

		List<Parameter> ParameterList() {
			List<Parameter> parameters = new List<Parameter>();

			Consume(TokenKind.OpenParen);

			while(currentToken.Kind != TokenKind.CloseParen) {
				List<Token> identifiers = new List<Token>();

				while(currentToken.Kind == TokenKind.Identifier) {
					identifiers.Add(currentToken);
					Consume(TokenKind.Identifier);
					ConsumeIfExists(TokenKind.Comma);
				}

				Consume(TokenKind.Colon);

				// Parameters are immutable by default and require the var
				// keyword to make them mutable
				bool mutable = false;
				if (currentToken.Kind == TokenKind.Var) {
					Consume(TokenKind.Var);
					mutable = true;
				}

				Token type_identifier = currentToken;
				Consume(TokenKind.Identifier);

				foreach(Token id in identifiers) {
					parameters.Add(new Parameter(id, type_identifier, mutable));
				}
				ConsumeIfExists(TokenKind.Comma);
			}

			Consume(TokenKind.CloseParen);
			return parameters;
		}

		Var Variable(Token token) {
			return new Var(token);
		}

		Node Factor(Block block) {
			Token current = currentToken;

			switch(currentToken.Kind) {
				case TokenKind.Minus: {
					Consume(TokenKind.Minus);
					return new UnaryOp(current, Factor(block));
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
					Node n = Expr(block);
					Consume(TokenKind.CloseParen);
					return n;
				}

				case TokenKind.Identifier: {
					Token identifier = currentToken;
					Consume(TokenKind.Identifier);

					if (currentToken.Kind == TokenKind.OpenParen) {
						return FunctionCall(block, identifier);
					}
					else {
						return Variable(identifier);
					}
				}

				default: {
					Error($"Unknown token in expression {currentToken.Kind}");
					break;
				}
			}

			return null;
		}

		Node Term(Block block) {
			Node n = Factor(block);

			while(currentToken.Kind == TokenKind.Star || currentToken.Kind == TokenKind.Slash) {
				Token op = currentToken;
				Consume(currentToken.Kind);
				n = new BinOp(op, n, Factor(block));
			}

			return n;
		}

		Node Arithmetic(Block block) {
			Node n = Term(block);

			while(currentToken.Kind == TokenKind.Plus || currentToken.Kind == TokenKind.Minus) {
				Token op = currentToken;
				Consume(currentToken.Kind);
				n = new BinOp(op, n, Term(block));
			}

			return n;
		}

		Node Expr(Block block) {
			Node n = Arithmetic(block);

			while(currentToken.Kind == TokenKind.EqualEqual || currentToken.Kind == TokenKind.NotEqual ||
				currentToken.Kind == TokenKind.Greater || currentToken.Kind == TokenKind.GreaterEqual ||
				currentToken.Kind == TokenKind.Less || currentToken.Kind == TokenKind.LessEqual
				) {
				Token op = currentToken;
				Consume(currentToken.Kind);
				n = new ConditionalOp(op, n, Arithmetic(block));
			}

			return n;
		}

		List<Node> GetArguments(Block block, TokenKind closing) {
			List<Node> arguments = new List<Node>();
			while(currentToken.Kind != closing) {
				arguments.Add(Expr(block));
				ConsumeIfExists(TokenKind.Comma);
			}
			return arguments;
		}

		void VariableDeclaration(Block block, bool mutable) {
			Token type_id = currentToken;
			Consume(TokenKind.Identifier);

			while (currentToken.Kind == TokenKind.Identifier) {
				Token identifier = currentToken;
				Consume(TokenKind.Identifier);

				Consume(TokenKind.Equals);
				Node expr = Expr(block);

				block.statements.Add(new VarDecl(identifier.Lexeme, type_id, mutable, expr));
				ConsumeIfExists(TokenKind.Comma);
			}
		}

		void Assignment(Block block, Token identifier) {
			Consume(TokenKind.Equals);
			block.statements.Add(new Assignment(identifier.Lexeme, Expr(block)));
		}

		FunctionCall FunctionCall(Block block, Token identifier) {
			Consume(TokenKind.OpenParen);
			List<Node> arguments = GetArguments(block, TokenKind.CloseParen);
			Consume(TokenKind.CloseParen);

			return new FunctionCall(identifier, arguments);
		}

		void BuiltinFnCall(Block block) {
			Consume(TokenKind.At);

			Token identifier = currentToken;
			Consume(TokenKind.Identifier);

			Consume(TokenKind.OpenParen);
			List<Node> arguments = GetArguments(block, TokenKind.CloseParen);
			Consume(TokenKind.CloseParen);

			block.statements.Add(new BuiltinFunctionCall(identifier.Lexeme, arguments, "void"));
		}

		void Escape(Block block) {
			Consume(TokenKind.Escape);

			if (block.escape != null) {
				Error("Block already contains an escape statement");
			}

			Escape esc = new Escape();
			block.escape = esc;
			block.statements.Add(esc);
		}

		Block Body() {
			Block block = new Block(new List<Node>());

			Consume(TokenKind.OpenCurly);
			StatementList(block, TokenKind.CloseCurly);
			Consume(TokenKind.CloseCurly);

			return block;
		}

		void FunctionDef(Block block) {
			Consume(TokenKind.Function);

			Token identifier = currentToken;
			Consume(TokenKind.Identifier);

			List<Parameter> parameters = ParameterList();

			Return return_type = new Return("void", null);
			if (currentToken.Kind == TokenKind.Colon) {
				Consume(TokenKind.Colon);

				Token return_identifier = currentToken;
				Consume(TokenKind.Identifier);

				Node expr = null;
				if (currentToken.Kind == TokenKind.OpenParen) {
					Consume(TokenKind.OpenParen);
					expr = Expr(block);
					Consume(TokenKind.CloseParen);
				}
				return_type = new Return(return_identifier.Lexeme, expr);
			}

			block.statements.Add(new FunctionDef(identifier, parameters, return_type, Body()));
		}

		IfStmt IfStatement(Block block) {
			Consume(TokenKind.If);

			Node expr = Expr(block);
			Block true_body = Body();
			Node false_body = null;

			if (currentToken.Kind == TokenKind.Else) {
				Consume(TokenKind.Else);

				if (currentToken.Kind == TokenKind.If) {
					false_body = IfStatement(block);
				} else {
					false_body = Body();
				}
			}

			return new IfStmt(expr, true_body, false_body);
		}

		void StatementList(Block block, TokenKind closing) {
			while(currentToken.Kind != closing) {
				Statement(block);
			}
		}

		void Statement(Block block) {
			switch(currentToken.Kind) {
				case TokenKind.Var: {
					Consume(TokenKind.Var);
					VariableDeclaration(block, true);
					break;
				}

				case TokenKind.Let: {
					Consume(TokenKind.Let);
					VariableDeclaration(block, false);
					break;
				}

				case TokenKind.If: {
					block.statements.Add(IfStatement(block));
					return; // Avoid semicolon
				}

				case TokenKind.At: {
					BuiltinFnCall(block);
					break;
				}

				case TokenKind.Escape: {
					Escape(block);
					break;
				}

				case TokenKind.Identifier: {
					Token identifier = currentToken;
					Consume(TokenKind.Identifier);

					switch(currentToken.Kind) {
						case TokenKind.OpenParen:
							block.statements.Add(FunctionCall(block, identifier));
							break;

						// TODO: Support alternate assignment operators += -= *= /=
						case TokenKind.Equals:
							Assignment(block, identifier);
							break;

						default:
							Error($"Unknown token following identifier: {currentToken.Kind}");
							break;
					}
					break;
				}

				case TokenKind.Function: {
					FunctionDef(block);
					break;
				}

				default:
					Error($"Unknown token in statement: {currentToken.Kind}");
					break;
			}

			Consume(TokenKind.SemiColon);
		}

		void Declaration() {
			// Record defs
			// Functions
			while (currentToken.Kind != TokenKind.End) {
				switch(currentToken.Kind) {
					case TokenKind.Function: {
						FunctionDef(app.block);
						break;
					}

					default:
						Statement(app.block);
						break;
				}
			}
		}
	}
}