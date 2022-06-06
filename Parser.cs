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

		FunctionDef FunctionDefinition() {
			Consume(TokenKind.Function);

			Consume(TokenKind.OpenParen);
			List<Identifier> identifiers = new List<Identifier>();

			while(current.Kind == TokenKind.Identifier) {
				identifiers.Add(new Identifier(current));
				Consume(TokenKind.Identifier);

				ConsumeIfExists(TokenKind.Comma);
			}
			Consume(TokenKind.CloseParen);

			Block inner = Body();

			return new FunctionDef(null, identifiers, inner);
		}

		Node Factor(Block block) {
			Token ftoken = current;
			
			switch(current.Kind) {
				case TokenKind.OpenParen: {
					Consume(TokenKind.OpenParen);
					Node node = Arithmetic(block);
					Consume(TokenKind.CloseParen);
					
					return node;
				}

				case TokenKind.Identifier:
					Consume(TokenKind.Identifier);
					return new Identifier(ftoken);

				case TokenKind.Int:
				case TokenKind.Float:
				case TokenKind.String:
				case TokenKind.Boolean: {
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

		void PrintStatement(Block block) {
			Consume(TokenKind.Print);
			Consume(TokenKind.OpenParen);

			List<Node> exprs = new List<Node>();

			while(current.Kind != TokenKind.CloseParen) {
				exprs.Add(Arithmetic(block));
				ConsumeIfExists(TokenKind.Comma);
			}

			Consume(TokenKind.CloseParen);

			block.statements.Add(new Print(exprs));
		}

		void VariableDeclaration(Block block) {
			Consume(TokenKind.Let);

			Token identifier = current;
			Consume(TokenKind.Identifier);

			Consume(TokenKind.Equals);

			Node expr = Arithmetic(block);

			block.statements.Add(new Variable(identifier, expr));
		}

		void Statement(Block block) {
			switch(current.Kind) {
				case TokenKind.Print: {
					PrintStatement(block);
					break;
				}

				case TokenKind.Let: {
					VariableDeclaration(block);
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