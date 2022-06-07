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
		public readonly bool mutable;
		public TypeKind kind; // Resolved in Analysis

		public VariableDecl(Token identifier, bool mutable, Node expr) : base(identifier) {
			this.mutable = mutable;
			this.expr = expr;
		}
	}

	sealed class VariableAssignment : Node {
		public readonly Node expr;
		public TypeKind kind; // Resolved in Analysis

		public VariableAssignment(Token identifier, Node expr) : base(identifier) {
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

	sealed class ConditionalOp : Node {
		public readonly Node left, right;

		public ConditionalOp(Token op, Node left, Node right) : base(op) {
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

	sealed class Parameter : Node {
		public readonly TypeKind kind;

		public Parameter(Token token, TypeKind kind) : base(token) {
			this.kind = kind;
		}
	}

	sealed class Block : Node {
		public readonly List<Node> statements;

		public Block(List<Node> statements) : base(null) {
			this.statements = statements;
		}
	}

	sealed class Argument : Node {
		public readonly Node expr;
		public TypeKind kind;

		public Argument(Node expr) : base(null) {
			this.expr = expr;
		}
	}

	sealed class FunctionDef : Node {
		public string identifier;
		public readonly List<Parameter> parameters;
		public readonly Block block;

		public FunctionDef(Token identifier, List<Parameter> parameters, Block block) : base(identifier) {
			this.parameters = parameters;
			this.block = block;
		}
	}

	sealed class FunctionCall : Node {
		public readonly List<Argument> arguments;
		public FunctionDef def;

		public FunctionCall(Token identifier, List<Argument> arguments) : base(identifier) {
			this.arguments = arguments;
		}
	}

	sealed class IfStmt : Node {
		public readonly Node expr;
		public readonly VariableDecl initStatement;
		public readonly Block trueBody;
		public readonly Node falseBody;

		public IfStmt(Node expr, VariableDecl initStmt, Block trueBody, Node falseBody) : base(null) {
			this.expr = expr;
			this.initStatement = initStmt;
			this.trueBody = trueBody;
			this.falseBody = falseBody;
		}
	}

	sealed class WhileStmt : Node {
		public readonly Node expr;
		public readonly VariableDecl initStatement;
		public readonly Block body;

		public WhileStmt(Node expr, VariableDecl initStmt, Block body) : base(null) {
			this.expr = expr;
			this.initStatement = initStmt;
			this.body = body;
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