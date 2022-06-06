using System;
using System.Text;

namespace TinyLang {
	class Interpreter {
		public void Run(Application app) {
			Visit(app.block);
		}

		void Error(string message) {
			throw new Exception($"Runtime: {message}");
		}
		
		Value Visit(Node node) {
			switch(node) {
				case Block:			return VisitBlock((Block)node);
				case BinaryOp: 		return VisitBinaryOp((BinaryOp)node);
				case Print: 		return VisitPrintStmt((Print)node);
				case Literal:		return VisitLiteral((Literal)node);
			}

			Error($"Unhandled node in interpreter {node}");
			return null;
		}

		Value VisitPrintStmt(Print print) {
			StringBuilder sb = new StringBuilder();

			foreach(Node node in print.arguments) {
				sb.Append(Visit(node));
			}

			Console.WriteLine(sb.ToString());

			return null;
		}
		
		Value VisitBinaryOp(BinaryOp binaryOp) {
			switch(binaryOp.token.Kind) {
				case TokenKind.Plus:		return Visit(binaryOp.left) + Visit(binaryOp.right);
			}

			Error($"Unhandled operator in binary operation {binaryOp.token.Kind}");
			return null;
		}

		Value VisitBlock(Block block) {
			foreach(Node node in block.statements) {
				Visit(node);
			}
			return null;
		}

		Value VisitLiteral(Literal literal) {
			switch(literal.token.Kind) {
				case TokenKind.Int:			return new IntValue(int.Parse(literal.token.Lexeme));
				case TokenKind.Float:		return new FloatValue(float.Parse(literal.token.Lexeme));
				case TokenKind.Boolean:		return new BoolValue(bool.Parse(literal.token.Lexeme));
				case TokenKind.String:		return new StringValue(literal.token.Lexeme);
			}

			Error($"Unknown liteal type {literal.token.Kind}");
			return null;
		}
	}	
}