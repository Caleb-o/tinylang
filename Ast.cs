using System;
using System.Collections.Generic;

namespace TinyLang {
	sealed class BuiltinFn {
		public readonly Builtins.Fn function;
		public readonly int parity;

		public BuiltinFn(Builtins.Fn function, int parity) {
			this.function = function;
			this.parity = parity;
		}
	}

	static class Builtins {
		public delegate Value Fn(List<Value> arguments);
		public static Dictionary<string, BuiltinFn> Functions = new Dictionary<string, BuiltinFn>() {
			{ "println", new BuiltinFn(PrintLn, -1) },
		};

		public static Value PrintLn(List<Value> arguments) {
			foreach(Value n in arguments) {
				Console.Write(((object)n != null) ? n.value : "NONE");
			}
			Console.WriteLine();
			return null;
		}
	}

	abstract class Node {
		public readonly Token token;

		public Node(Token token) {
			this.token = token;
		}
	}

	sealed class Type : Node {
		// type_id, id?
		// This also helps with tuples and unpacking
		// eg. let (boolean ok, int value) = get_value();
		public readonly List<string> type;

		public Type(string value) : base(null) {
			this.type = new List<string>() { value };
		}

		public Type(List<string> type) : base(null) {
			this.type = type;
		}

		public bool IsVoid() {
			return type == null;
		}

		public bool IsSingleType() {
			return type != null && type.Count == 1;
		}

		public bool Matches(Type other) {
			if (type.Count != other.type.Count) {
				return false;
			}

			for(int i = 0; i < type.Count; i++) {
				// Proof that types need to be represented better, with something
				// like an integer as an ID
				string type_me = type[i];
				string type_other = other.type[i];

				if (type_me != type_other) {
					return false;
				}
			}

			return true;
		}

		public override string ToString() {
			return "{" + string.Join(", ", type) + "}";
		}
	}

	sealed class BinOp : Node {
		public readonly Node left, right;

		public BinOp(Token op, Node left, Node right) : base(op) {
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

	sealed class TupleLiteral: Node {
		public readonly List<Node> exprs;

		public TupleLiteral(List<Node> exprs) : base(null) {
			this.exprs = exprs;
		}
	}

	sealed class Var : Node {
		public Var(Token token) : base(token) {}
	}

	sealed class VarDecl : Node {
		public readonly string identifier;
		public readonly bool mutable;
		public readonly Type type;
		public readonly Node expr;

		public VarDecl(string identifier, Type type, bool mutable, Node expr) : base(null) {
			this.identifier = identifier;
			this.type = type;
			this.mutable = mutable;
			this.expr = expr;
		}
	}

	sealed class IfStmt : Node {
		public readonly Node expr;
		public readonly Block trueBody;
		public readonly Node falseBody;

		public IfStmt(Node expr, Block trueBody, Node falseBody) : base(null) {
			this.expr = expr;
			this.trueBody = trueBody;
			this.falseBody = falseBody;
		}
	}

	sealed class While : Node {
		public readonly Node expr;
		public readonly Block body;

		public While(Node expr, Block body) : base(null) {
			this.expr = expr;
			this.body = body;
		}
	}

	sealed class DoWhile : Node {
		public readonly Node expr;
		public readonly Block body;

		public DoWhile(Node expr, Block body) : base(null) {
			this.expr = expr;
			this.body = body;
		}
	}

	sealed class Assignment : Node {
		public readonly string identifier;
		public readonly Node expr;

		public Assignment(string identifier, Node expr) : base(null) {
			this.identifier = identifier;
			this.expr = expr;
		}
	}

	sealed class Parameter : Node {
		public readonly Type type;
		public readonly bool mutable;

		public Parameter(Token identifier, Type type, bool mutable) : base(identifier) {
			this.type = type;
			this.mutable = mutable;
		}
	}

	sealed class Block : Node {
		public readonly List<Node> statements;
		public Escape escape;

		public Block(List<Node> statements) : base(null) {
			this.statements = statements;
		}
	}

	sealed class Return : Node {
		// Return is the default value and type info
		// rather than the statement
		public readonly Type type;
		public readonly Node expr;

		public Return(Type type, Node expr) : base(null) {
			this.type = type;
			this.expr = expr;
		}
	}

	sealed class Escape : Node {
		// Escape is similar to a return statement,
		// except it does not return a value but signals
		// an exit for the function
		public Escape() : base(null) {}
	}

	sealed class FunctionDef : Node {
		public readonly List<Parameter> parameters;
		public readonly Block block;
		// Return type is used for the implicit return type
		public readonly Return returnType;

		public FunctionDef(Token identifier, List<Parameter> parameters, Return returnType, Block block) : base(identifier) {
			this.parameters = parameters;
			this.returnType = returnType;
			this.block = block;
		}
	}

	sealed class BuiltinFunctionCall : Node {
		public readonly string identifier;
		public readonly BuiltinFn native;
		public readonly List<Node> arguments;
		public readonly Type returnType;

		public BuiltinFunctionCall(string identifier, List<Node> arguments, Type returnType) : base(null) {
			this.identifier = identifier;
			this.arguments = arguments;
			this.returnType = returnType;

			if (!Builtins.Functions.ContainsKey(identifier)) {
				throw new Exception($"Builtin function '{identifier}' does not exist");
			}

			this.native = Builtins.Functions[identifier];
		}
	}

	sealed class FunctionCall : Node {
		public readonly List<Node> arguments;
		public FunctionSym sym;

		public FunctionCall(Token identifier, List<Node> arguments) : base(identifier) {
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