using System;
using System.Collections.Generic;

namespace TinyLang {
	class Parser {
		Lexer lexer;
		Token current;


		public Parser(string source) {
			lexer = new Lexer(source);
			current = lexer.Next();
		}

		void Error(string message) {
			throw new Exception($"Parser: {message}");
		}

		void Error(string message, Token token) {
			throw new Exception($"Parser: {message} ['{token.Lexeme}' {token.Line}:{token.Column}]");
		}

		void Consume(TokenKind expected) {
			if (current.Kind == expected) {
				current = lexer.Next();
			} else {
				Error($"Expected token {expected} but received {current.Kind}");
			}
		}

		void ConsumeIfExists(TokenKind expected) {
			if (current.Kind == expected) {
				current = lexer.Next();
			}
		}

		TinyType CollectType() {
			if (current.Kind == TokenKind.Colon) {
				Consume(TokenKind.Colon);

				Token type_id = current;
				Consume(TokenKind.Identifier);

				return TinyType.TypeFromLexeme(type_id);
			}

			return new TinyAny();
		}

		FunctionDef FunctionDefinition() {
			Consume(TokenKind.Function);

			Consume(TokenKind.OpenParen);
			List<Parameter> identifiers = new List<Parameter>();
			List<string> values = new List<string>();

			while(current.Kind == TokenKind.Identifier) {
				Token identifier = current;
				Consume(TokenKind.Identifier);

				if (values.Contains(identifier.Lexeme)) {
					Error($"Function already contains parameter '{identifier.Lexeme}'", identifier);
				}

				values.Add(identifier.Lexeme);

				identifiers.Add(new Parameter(identifier, CollectType()));
				ConsumeIfExists(TokenKind.Comma);
			}
			Consume(TokenKind.CloseParen);

			TinyType returns = CollectType();

			Block inner = Body();

			return new FunctionDef(null, identifiers, returns, inner);
		}

		FunctionCall FnCall(Block block, Token identifier) {
			Consume(TokenKind.OpenParen);
			List<Argument> arguments = new List<Argument>();

			while(current.Kind != TokenKind.CloseParen) {
				arguments.Add(new Argument(Expr(block)));
				ConsumeIfExists(TokenKind.Comma);
			}
			Consume(TokenKind.CloseParen);

			return new FunctionCall(identifier, arguments);
		}

		Node Factor(Block block) {
			Token ftoken = current;
			
			switch(current.Kind) {
				case TokenKind.OpenParen: {
					Consume(TokenKind.OpenParen);
					Node node = Expr(block);
					Consume(TokenKind.CloseParen);
					
					return node;
				}

				case TokenKind.Identifier: {
					Consume(TokenKind.Identifier);

					if (current.Kind == TokenKind.OpenParen) {
						return FnCall(block, ftoken);
					}

					return new Identifier(ftoken);
				}


				case TokenKind.Int:
				case TokenKind.Float:
				case TokenKind.String:
				case TokenKind.Bool: {
					Consume(current.Kind);
					return new Literal(ftoken);
				}

				case TokenKind.Function: {
					return FunctionDefinition();
				}
			}

			Error($"Unknown token kind found in expression '{ftoken.Lexeme}':{ftoken.Kind}");
			return null;
		}

		Node Term(Block block) {
			Node node = Factor(block);

			while(current.Kind == TokenKind.Star || current.Kind == TokenKind.Slash) {
				Token op = current;
				Consume(op.Kind);
				node = new BinaryOp(op, node, Factor(block));
			}

			return node;
		}

		Node Arithmetic(Block block) {
			Node node = Term(block);

			while(current.Kind == TokenKind.Plus || current.Kind == TokenKind.Minus) {
				Token op = current;
				Consume(op.Kind);
				node = new BinaryOp(op, node, Term(block));
			}

			return node;
		}

		Node Expr(Block block) {
			Node node = Arithmetic(block);

			while (current.Kind == TokenKind.EqualEqual || current.Kind == TokenKind.NotEqual ||
					current.Kind == TokenKind.Less || current.Kind == TokenKind.LessEqual ||
					current.Kind == TokenKind.Greater || current.Kind == TokenKind.GreaterEqual) {
				Token op = current;
				Consume(op.Kind);
				node = new ConditionalOp(op, node, Arithmetic(block));
			}

			return node;
		}

		void PrintStatement(Block block) {
			Consume(TokenKind.Print);
			Consume(TokenKind.OpenParen);

			List<Node> exprs = new List<Node>();

			while(current.Kind != TokenKind.CloseParen) {
				exprs.Add(Expr(block));
				ConsumeIfExists(TokenKind.Comma);
			}

			Consume(TokenKind.CloseParen);

			block.statements.Add(new Print(exprs));
		}

		VariableDecl VariableDeclaration(Block block, bool mutable) {
			Token identifier = current;
			Consume(TokenKind.Identifier);

			TinyType kind = CollectType();

			Consume(TokenKind.Equals);

			VariableDecl vardecl = new VariableDecl(identifier, mutable, Expr(block));
			vardecl.kind = kind;

			return vardecl;
		}

		void VariableDeclarations(Block block, bool mutable) {
			while(current.Kind == TokenKind.Identifier) {
				block.statements.Add(VariableDeclaration(block, mutable));
				ConsumeIfExists(TokenKind.Comma);
			}
		}

		void VariableAssign(Block block, Token identifier) {
			Consume(TokenKind.Equals);
			block.statements.Add(new VariableAssignment(identifier, Expr(block)));
		}

		IfStmt IfStatement(Block block) {
			Consume(TokenKind.If);

			VariableDecl vardecl = null;

			if (current.Kind == TokenKind.Var || current.Kind == TokenKind.Let) {
				Token varType = current;
				Consume(current.Kind);

				vardecl = VariableDeclaration(block, varType.Kind == TokenKind.Var);
				Consume(TokenKind.SemiColon);
			}

			Node expr = Expr(block);
			Block trueBody = Body();
			Node falseBody = null;

			if (current.Kind == TokenKind.Else) {
				Consume(TokenKind.Else);

				if (current.Kind == TokenKind.If) {
					falseBody = IfStatement(block);
				} else {
					falseBody = Body();
				}
			}	

			return new IfStmt(expr, vardecl, trueBody, falseBody);
		}

		void WhileStatement(Block block) {
			Consume(TokenKind.While);
			
			VariableDecl vardecl = null;

			if (current.Kind == TokenKind.Var || current.Kind == TokenKind.Let) {
				Token varType = current;
				Consume(current.Kind);
				vardecl = VariableDeclaration(block, varType.Kind == TokenKind.Var);
				Consume(TokenKind.SemiColon);
			}

			block.statements.Add(new WhileStmt(Expr(block), vardecl, Body()));
		}

		void DoWhileStatement(Block block) {
			Consume(TokenKind.Do);
			
			VariableDecl vardecl = null;

			if (current.Kind == TokenKind.Var || current.Kind == TokenKind.Let) {
				Token varType = current;
				Consume(current.Kind);
				vardecl = VariableDeclaration(block, varType.Kind == TokenKind.Var);
			}

			Block body = Body();
			Consume(TokenKind.While);
			Node expr = Expr(block);

			block.statements.Add(new DoWhileStmt(expr, vardecl, body));
		}

		void Statement(Block block) {
			switch(current.Kind) {
				case TokenKind.Print: {
					PrintStatement(block);
					break;
				}

				case TokenKind.Let: {
					Consume(TokenKind.Let);
					VariableDeclarations(block, false);
					break;
				}

				case TokenKind.Var: {
					Consume(TokenKind.Var);
					VariableDeclarations(block, true);
					break;
				}

				case TokenKind.Identifier: {
					Token identifier = current;
					Consume(TokenKind.Identifier);

					if (current.Kind == TokenKind.OpenParen) {
						block.statements.Add(FnCall(block, identifier));
					} else if (current.Kind == TokenKind.Equals) {
						VariableAssign(block, identifier);
					} else {
						Error($"Unknown token following identifier '{identifier.Lexeme}':{identifier.Kind}");
					}
					break;
				}

				case TokenKind.If: {
					block.statements.Add(IfStatement(block));
					break;
				}

				case TokenKind.While: {
					WhileStatement(block);
					break;
				}

				case TokenKind.Do: {
					DoWhileStatement(block);
					break;
				}

				default:
					Error($"Unknown token kind found {current.Kind}");
					return;
			}

			Consume(TokenKind.SemiColon);
		}

		void StatementList(Block block, TokenKind end = TokenKind.End) {
			while(current.Kind != end) {
				Statement(block);
			}
		}

		Block Body() {
			Consume(TokenKind.OpenCurly);
			Block inner = new Block(new List<Node>());
			StatementList(inner, TokenKind.CloseCurly);
			Consume(TokenKind.CloseCurly);

			return inner;
		}

		public Application Parse() {
			Block programBlock = new Block(new List<Node>());
			StatementList(programBlock);

			return new Application(programBlock);
		}
	}
}