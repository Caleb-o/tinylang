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
			{ "printobj", new BuiltinFn(PrintObj, -1) },
		};

		public static void InitTypes() {
			Application.GetTypeID("int");
			Application.GetTypeID("float");
			Application.GetTypeID("boolean");
			Application.GetTypeID("string");
		}

		public static Value PrintLn(List<Value> arguments) {
			foreach(Value n in arguments) {
				switch(n.type.GetKind()) {
					case TypeKind.Tuple: {
						string outStr = "";

						int idx = 0;
						List<Value> values = (List<Value>)n.value;
						foreach(Value value in values) {
							outStr += value.value;

							if (idx < values.Count - 1) {
								outStr += ", ";
							}
							idx++;
						}
						Console.Write($"[{outStr}]");
						break;
					}

					default: {
						Console.Write((n != null) ? n.value : "NONE");
						break;
					}
				}
			}
			Console.WriteLine();
			return null;
		}

		public static Value PrintObj(List<Value> arguments) {
			foreach(Value n in arguments) {
				switch(n.type.GetKind()) {
					case TypeKind.Tuple: {
						string outStr = "";

						int idx = 0;
						List<Value> values = (List<Value>)n.value;
						foreach(Value value in values) {
							outStr += value.value;

							if (idx < values.Count - 1) {
								outStr += ", ";
							}
							idx++;
						}
						Console.WriteLine($"Value [{outStr}] : {n.type.GetKind()}");
						break;
					}

					default: {
						object value = (n != null) ? n.value : "NONE";
						Console.WriteLine($"Value {value} : {n.type.GetKind()}");
						break;
					}
				}
			}
			return null;
		}
	}

	abstract class Node {
		public readonly Token token;

		public Node(Token token) {
			this.token = token;
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
		public Type type;
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
		public readonly VarDecl initStatement;
		public readonly Block trueBody;
		public readonly Node falseBody;

		public IfStmt(Node expr, VarDecl initStatement, Block trueBody, Node falseBody) : base(null) {
			this.expr = expr;
			this.initStatement = initStatement;
			this.trueBody = trueBody;
			this.falseBody = falseBody;
		}
	}

	sealed class While : Node {
		public readonly Node expr;
		public readonly VarDecl initStatement;
		public readonly Block body;

		public While(Node expr, Block body, VarDecl initStatement) : base(null) {
			this.expr = expr;
			this.body = body;
			this.initStatement = initStatement;
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
		public readonly Node identifier;
		public readonly Node expr;

		public Assignment(Node identifier, Node expr) : base(null) {
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

	// Used for literal tuples, lists and dictionaries
	sealed class ComplexLiteral : Node {
		public readonly List<Node> exprs;

		public ComplexLiteral(List<Node> exprs) : base(null) {
			this.exprs = exprs;
		}
	}

	// Eg. tuple[0][1]
	sealed class Index : Node {
		public readonly List<Node> exprs;
		public Type type;

		public Index(Token identifier, List<Node> exprs) : base(identifier) {
			this.exprs = exprs;
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

		static int TypeIdCounter = 0;
		static Dictionary<string, int> typeIDs = new Dictionary<string, int>();
		static Dictionary<int, string> typeNames = new Dictionary<int, string>();
		public static Dictionary<string, Value> literals = new Dictionary<string, Value>();

		public Application(Block block) : base(null) {
			this.block = block;
		}

		public static Value GetOrInsertLiteral(string lexeme, TypeKind kind) {
			if (literals.ContainsKey(lexeme)) {
				return literals[lexeme];
			}

			object value = null;

			switch(kind) {
				case TypeKind.Int:			value = int.Parse(lexeme); break;
				case TypeKind.Float:		value = float.Parse(lexeme); break;
				case TypeKind.Bool:			value = bool.Parse(lexeme); break;
				case TypeKind.String:		value = lexeme; break;

				default:
					throw new InvalidOperationException($"Unknown literal type {kind}");
			}

			literals[lexeme] = new Value(new Type(kind), value);
			return literals[lexeme];
		}

		public static int GetTypeID(string identifier) {
			if (typeIDs.ContainsKey(identifier)) {
				return typeIDs[identifier];
			}

			typeNames[TypeIdCounter] = identifier;
			typeIDs[identifier] = TypeIdCounter;
			return TypeIdCounter++;
		}

		public static string GetTypeName(int typeID) {
			return typeNames.GetValueOrDefault(typeID, "None");
		}
	}
}