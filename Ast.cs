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
		public readonly Node identifier;
		public readonly Node expr;

		public VariableAssignment(Node identifier, Node expr) : base(identifier.token) {
			this.identifier = identifier;
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

	// eg. list[0]
	// eg. list[a + 1]
	// eg. list[a + 1][0]
	sealed class Index : Node {
		public readonly Node[] expr;
		// Kind of the indexed type. eg. list[0] might be an int
		public TinyType kind;

		public Index(Token caller, Node[] expr) : base(caller) {
			this.expr = expr;
		}
	}

	// eg. my_instance.x
	// eg. my_instance.obj[0]
	// eg. my_instance.obj[0].x
	sealed class MemberAccess : Node {
		public readonly Node[] members;

		// The caller is used to lookup the symbol to check types and mutability
		public MemberAccess(Token caller, Node[] members) : base(caller) {
			this.members = members;
		}
	}

	sealed class Parameter : Node {
		public readonly string identifier;
		public readonly bool mutable;
		public readonly TinyType kind;

		public Parameter(string identifier, TinyType kind) : base(null) {
			this.identifier = identifier;
			this.mutable = false;
			this.kind = kind;
		}

		public Parameter(Token token, bool mutable, TinyType kind) : base(token) {
			this.identifier = token.Lexeme;
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

		public Block(Token token, List<Node> statements) : base(token) {
			this.statements = statements;
		}

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
		public readonly Parameter[] parameters;
		// Block will be reused as a delegate
		public readonly object block;
		public readonly TinyType returns;

		// Imported/Builtin function
		public FunctionDef(string identifier, BuiltinFn fn) : base(null) {
			this.identifier = identifier;
			this.parameters = fn.parameters;
			this.returns = fn.returns;
			this.block = fn;
		}

		// User function
		public FunctionDef(Token token, Parameter[] parameters, TinyType returns, Block block) : base(token) {
			this.parameters = parameters;
			this.returns = returns;
			this.block = block;
		}
	}

	sealed class FunctionCall : Node {
		public readonly Argument[] arguments;
		public FunctionDef def;

		public FunctionCall(Token identifier, Argument[] arguments) : base(identifier) {
			this.arguments = arguments;
		}
	}

	sealed class StructDef : Node {
		public string identifier;
		public readonly Dictionary<string, TinyType> fields;

		public StructDef(StructDef other) : base(other.token) {
			this.identifier = new string(other.identifier);
			this.fields = new Dictionary<string, TinyType>(other.fields);
		}

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