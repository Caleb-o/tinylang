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

		Node Factor(Block block) {
			Token ftoken = current;
			
			switch(current.Kind) {
				case TokenKind.OpenParen: {
					Consume(TokenKind.OpenParen);
					Node node = Arithmetic(block);
					Consume(TokenKind.CloseParen);
					
					return node;
				}

				case TokenKind.Int:
				case TokenKind.Float:
				case TokenKind.String:
				case TokenKind.Boolean: {
					Consume(current.Kind);
					return new Literal(ftoken);
				}
			}

			Error($"Unknown token kind found in expression {ftoken.Kind}");
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

		void Statement(Block block) {
			while(current.Kind != TokenKind.End) {
				switch(current.Kind) {
					case TokenKind.Print: {
						PrintStatement(block);
						break;
					}

					default:
						Error($"Unknown token kind found {current.Kind}");
						return;
				}

				Consume(TokenKind.SemiColon);
			}
		}

		public Application Parse() {
			Block programBlock = new Block(new List<Node>());
			Statement(programBlock);

			return new Application(programBlock);
		}
	}
}