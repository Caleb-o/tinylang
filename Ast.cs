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
		public TinyType kind; // Resolved in Analysis

		public VariableDecl(Token identifier, bool mutable, Node expr) : base(identifier) {
			this.mutable = mutable;
			this.expr = expr;
		}
	}

	sealed class VariableAssignment : Node {
		public readonly Node expr;

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

	sealed class ListLiteral : Node {
		public readonly List<Node> exprs;
		public TinyType kind;

		public ListLiteral(Token token, List<Node> exprs) : base(token) {
			this.exprs = exprs;
		}
	}

	sealed class Identifier : Node {
		public Identifier(Token token) : base(token) {}
	}

	sealed class Parameter : Node {
		public readonly bool mutable;
		public readonly TinyType kind;

		public Parameter(Token token, bool mutable, TinyType kind) : base(token) {
			this.mutable = mutable;
			this.kind = kind;
		}
	}

	sealed class Return : Node {
		public Return(Token token) : base(token) {}
	}

	sealed class Block : Node {
		public readonly List<Node> statements;
		public Return returnstmt;

		public Block(List<Node> statements) : base(null) {
			this.statements = statements;
		}
	}

	sealed class Argument : Node {
		public readonly Node expr;
		public TinyType kind;

		public Argument(Node expr) : base(null) {
			this.expr = expr;
		}
	}

	sealed class FunctionDef : Node {
		public string identifier;
		public readonly List<Parameter> parameters;
		public readonly Block block;
		public readonly TinyType returns;

		public FunctionDef(Token token, List<Parameter> parameters, TinyType returns, Block block) : base(token) {
			this.parameters = parameters;
			this.returns = returns;
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

	sealed class StructDef : Node {
		public string identifier;
		public readonly Dictionary<string, TinyType> fields;

		public StructDef(Token token, Dictionary<string, TinyType> fields) : base(token) {
			this.fields = fields;
		}
	}

	sealed class StructInstance : Node {
		public string identifier;
		public StructDef def;
		public readonly Dictionary<string, Node> members;

		public StructInstance(Token token, Dictionary<string, Node> members) : base(token) {
			this.identifier = token.Lexeme;
			this.members = members;
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

	sealed class DoWhileStmt : Node {
		public readonly Node expr;
		public readonly VariableDecl initStatement;
		public readonly Block body;

		public DoWhileStmt(Node expr, VariableDecl initStmt, Block body) : base(null) {
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