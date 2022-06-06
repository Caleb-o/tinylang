using System.Collections.Generic;

namespace TinyLang {
	abstract class Node  {
		public readonly Token token;

		public Node(Token token) {
			this.token = token;
		}
	}

	sealed class VariableDecl : Node {
		public readonly Node expr;
		public TypeKind kind; // Resolved in Analysis

		public VariableDecl(Token identifier, Node expr) : base(identifier) {
			this.expr = expr;
		}
	}

	sealed class BinaryOp : Node {
		public readonly Node left, right;

		public BinaryOp(Token op, Node left, Node right) : base(op) {
			this.left = left;
			this.right = right;
		}
	}

	sealed class UnaryOp : Node {
		public readonly Node right;

		public UnaryOp(Token op, Node right) : base(op) {
			this.right = right;
		}
	}

	sealed class Literal : Node {
		public Literal(Token token) : base(token) {}
	}

	sealed class Identifier : Node {
		public Identifier(Token token) : base(token) {}
	}

	sealed class Block : Node {
		public readonly List<Node> statements;

		public Block(List<Node> statements) : base(null) {
			this.statements = statements;
		}
	}

	sealed class FunctionDef : Node {
		public readonly List<Identifier> parameters;
		public readonly Block block;

		public FunctionDef(Token identifier, List<Identifier> parameters, Block block) : base(identifier) {
			this.parameters = parameters;
			this.block = block;
		}
	}

	sealed class FunctionCall : Node {
		public readonly List<Node> arguments;

		public FunctionCall(Token identifier, List<Node> arguments) : base(identifier) {
			this.arguments = arguments;
		}
	}

	sealed class Print : Node {
		public readonly List<Node> arguments;

		public Print(List<Node> arguments) : base(null) {
			this.arguments = arguments;
		}		
	}

	sealed class Application : Node {
		public readonly Block block;

		public Application(Block block) : base(null) {
			this.block = block;
		}
	}
}