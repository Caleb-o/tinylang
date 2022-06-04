using System;
using System.Collections.Generic;

namespace TinyLang {
	abstract class Symbol {
		public readonly string identifier;
		public int scopeLevel;
		public readonly Type type;

		public Symbol(string identifier, Type type) {
			this.identifier = identifier;
			this.type = type;
			this.scopeLevel = 0;
		}
	}

	sealed class VarSym : Symbol {
		public readonly bool mutable;
		public Value value;
		public VarSym references;

		public VarSym(string identifier, Type type, bool mutable)
			: base(identifier, type) {
			this.mutable = mutable;
		}
	}

	sealed class BuiltinTypeSym : Symbol {
		public BuiltinTypeSym(TypeKind kind, string identifier)
			: base(identifier, new Type(kind)) {}
	}

	sealed class FunctionSym : Symbol {
		public readonly List<VarSym> parameters;
		public readonly FunctionDef def;

		public FunctionSym(string identifier, List<VarSym> parameters, FunctionDef def)
			: base(identifier, null) {
			this.parameters = parameters;
			this.def = def;
		}
	}

	sealed class SymbolScope {
		public readonly Dictionary<string, Symbol> symbols = new Dictionary<string, Symbol>();
		public readonly string identifier;
		public readonly SymbolScope parent;
		public int scopeLevel { get; private set; }

		public SymbolScope(string identifier, int scopeLevel, SymbolScope parent) {
			this.identifier = identifier;
			this.scopeLevel = scopeLevel;
			this.parent = parent;
		}

		public void InitBuiltins() {
			Insert(new BuiltinTypeSym(TypeKind.Int, "int"));
			Insert(new BuiltinTypeSym(TypeKind.Float, "float"));
			Insert(new BuiltinTypeSym(TypeKind.Bool, "boolean"));
			Insert(new BuiltinTypeSym(TypeKind.String, "string"));
		}

		public void Insert(Symbol sym) {
			sym.scopeLevel = scopeLevel;
			symbols[sym.identifier] = sym;
		}

		public bool HasSymbol(string identifier) {
			return symbols.ContainsKey(identifier);
		}

		public Symbol Lookup(string identifier, bool local) {
			if (symbols.ContainsKey(identifier)) {
				return symbols[identifier];
			}

			if (local) {
				return null;
			}

			return parent.Lookup(identifier, false);
		}
	}

	sealed class Analyser {
		public SymbolScope scope;
		Block currentBlock;

		public void Run(Application app) {
			// Setup the global scope with builtins
			scope = new SymbolScope("global", 0, null);
			scope.InitBuiltins();

			Visit(app.block);
		}

		void Error(string msg) {
			throw new Exception($"Analyser: {msg} : [{scope.identifier}]");
		}

		void ErrorWith(string msg, Node node) {
			Token token = node.token;

			if (token == null) {
				Error(msg);
			}

			throw new Exception($"Analyser: {msg} : [{scope.identifier} {token.Column}:{token.Line}]");
		}

		void ClimbScope() {
			scope = scope.parent;
		}

		Type FindType(Node node) {
			switch(node) {
				case Var: {
					return scope.Lookup(((Var)node).token.Lexeme, false).type;
				}

				case FunctionCall: {
					FunctionDef def = ((FunctionCall)node).sym.def;

					if (def.returnType != null) {
						return def.returnType.type;
					} else {
						return new Type();
					}
				}

				case BinOp: {
					return FindType(((BinOp)node).left);
				}

				case Literal: {
					return new Type(Type.FromToken(node.token.Kind));
				}

				case UnaryOp: {
					return FindType(((UnaryOp)node).right);
				}

				case ConditionalOp: {
					return FindType(((ConditionalOp)node).left);
				}

				case Index: {
					// Hack: Trying to evaluate an index by interpreting it
					if (((Index)node).exprs[0] is not Literal) {
						ErrorWith("Indexing can only use literals currently", ((Index)node).exprs[0]);
					}

					Interpreter interpreter = new Interpreter();
					Value index = interpreter.Visit(((Index)node).exprs[0]);

					if (!index.type.Matches(new Type(TypeKind.Int))) {
						Error($"Index expected an integer but received {index.type}");
					}

					int indexValue = (int)index.value;

					if (indexValue < 0 || indexValue >= ((Index)node).type.typeIDs.Length) {
						Error($"Index out of range: {indexValue}");
					}

					return new Type(((Index)node).type.typeIDs[indexValue]);
				}

				case ComplexLiteral: {
					ComplexLiteral literal = (ComplexLiteral)node;
					List<int> typeIDs = new List<int>() { (int)literal.kind };

					foreach(Node expr in literal.exprs) {
						// Note: This can probably be more than one in the future,
						// 		 so a better method may be required
						typeIDs.Add((int)FindType(expr).GetKind());
					}

					return new Type(typeIDs.ToArray());
				}

				default:
					Error($"Unknown node in find type: {node}");
					return null;
			}
		}

		Type ExpectType(Node node, Type expects) {
			switch(node) {
				case BinOp: {
					BinOp binop = (BinOp)node;
					Type left = FindType(binop.left);
					Type right = FindType(binop.right);

					if (!expects.Matches(left)) {
						return left;
					}

					if (!expects.Matches(right)) {
						return right;
					}

					return null;
				}

				case ConditionalOp: {
					ConditionalOp conditional = (ConditionalOp)node;
					Type left = FindType(conditional.left);
					Type right = FindType(conditional.right);
					
					if (!expects.Matches(left)) {
						return left;
					}

					if (!expects.Matches(right)) {
						return right;
					}

					return null;
				}

				case Var:
				case Index:
				case UnaryOp:
				case Literal:
				case FunctionCall: {
					Type type = FindType(node);

					if (!expects.Matches(type)) {
						return type;
					}

					return null;
				}

				case ComplexLiteral: {
					ComplexLiteral literal = (ComplexLiteral)node;

					if (literal.kind == TypeKind.List) {
						// Only having the list as an ID is an untyped list
						// Which means it can be bound to anything
						if (expects.typeIDs.Length > 1) {
							int listType = expects.typeIDs[1];

							foreach(Node expr in literal.exprs) {
								Type realType = FindType(expr);

								if (!realType.Matches(new Type(listType))) {
									return realType;
								}
							}
						}
					} else {
						// Tuple
						Type type = FindType(node);

						if (!expects.Matches(type)) {
							return type;
						}
					}
					return null;
				}

				default:
					Error($"Unknown node in ExpectType {node}");
					return null;
			}
		}

		void Visit(Node node) {
			switch(node) {
				case Block: VisitBlock((Block)node); break;
				case BinOp: VisitBinOp((BinOp)node); break;
				case Assignment: VisitAssignment((Assignment)node); break;
				case VarDecl: VisitVariableDeclaration((VarDecl)node); break;
				case FunctionDef: VisitFunctionDef((FunctionDef)node); break;
				case FunctionCall: VisitFunctionCall((FunctionCall)node); break;
				case BuiltinFunctionCall: VisitBuiltinCall((BuiltinFunctionCall)node); break;
				case Var: VisitVar((Var)node); break;
				case Escape: VisitEscape((Escape)node); break;
				case IfStmt: VisitIfStatement((IfStmt)node); break;
				case While: VisitWhile((While)node); break;
				case DoWhile: VisitDoWhile((DoWhile)node); break;
				case ConditionalOp: VisitConditionalOp((ConditionalOp)node); break;

				case Literal: break;
				case ComplexLiteral: VisitComplexLiteral((ComplexLiteral)node); break;
				case Index: VisitIndex((Index)node); break;
				case UnaryOp: Visit(((UnaryOp)node).right); break;

				default:
					throw new Exception($"Unimplemented node in analyser: {node.GetType()}");
			}
		}

		void VisitComplexLiteral(ComplexLiteral literal) {
			if (literal.kind == TypeKind.List) {
				Type first = FindType(literal);
				Type realType = ExpectType(literal, first);

				if (realType != null) {
					Error($"List expected type {first} but received {realType}");
				}
			}
		}

		void VisitBinOp(BinOp binop) {
			Visit(binop.left);
			Visit(binop.right);
		}

		void VisitConditionalOp(ConditionalOp conditional) {
			Visit(conditional.left);
			Visit(conditional.right);
			
			Type left = FindType(conditional.left);
			Type right = ExpectType(conditional.right, left);

			if (right != null) {
				Error($"Conditional expected type {left} but received {right}");
			}
		}

		void VisitVar(Var var) {
			if (scope.Lookup(var.token.Lexeme, false) == null) {
				Error($"Variable '{var.token.Lexeme}' does not exist in any scope");
			}
		}

		void VisitIndex(Index index) {
			VarSym variable = (VarSym)scope.Lookup(index.token.Lexeme, false);

			if (variable == null) {
				Error($"Variable '{index.token.Lexeme}' does not exist in any scope");
			}

			if (index.exprs.Count > variable.type.typeIDs.Length) {
				Error($"Variable '{index.token.Lexeme}' is trying to access {index.exprs.Count} levels, when only {variable.type.typeIDs.Length} exist");
			}

			foreach(Node expr in index.exprs) {
				// Hack: This is the easiest way to check for a type, but it isn't reliable
				if (FindType(expr).typeIDs[0] != Application.GetTypeID("int")) {
					Error("Index expected an integer");
				}

				Visit(expr);
			}

			// Assign the type
			index.type = variable.type;
		}

		void VisitBlock(Block block) {
			// Hack for returns
			currentBlock = block;

			foreach(Node node in block.statements) {
				Visit(node);
			}
		}

		void VisitEscape(Escape ret) {
			if (scope.identifier == "global") {
				// TODO: This can probably occur for early returns
				Error("Cannot escape from global scope");
			}

			if (currentBlock.escape != null) {
				Error("Block cannot contain more than one escape statement!");
			}
			currentBlock.escape = ret;
		}

		void VisitIfStatement(IfStmt ifstmt) {
			if (ifstmt.initStatement != null) {
				VisitVariableDeclaration(ifstmt.initStatement);
			}

			Visit(ifstmt.expr);

			Type expected = FindType(ifstmt.expr);
			Type realType = ExpectType(ifstmt.expr, expected);

			if (realType != null) {
				ErrorWith($"If expr expected type {expected} but received {realType}", ifstmt.expr);
			}

			Visit(ifstmt.trueBody);

			if (ifstmt.falseBody != null) {
				Visit(ifstmt.falseBody);
			}
		}

		void VisitWhile(While whilestmt) {
			if (whilestmt.initStatement != null) {
				VisitVariableDeclaration(whilestmt.initStatement);
			}

			Visit(whilestmt.expr);

			Type expected = FindType(whilestmt.expr);
			Type realType = ExpectType(whilestmt.expr, expected);

			if (realType != null) {
				Error($"While expr expected type {expected} but received {realType}");
			}

			Visit(whilestmt.body);
		}

		void VisitDoWhile(DoWhile whilestmt) {
			Visit(whilestmt.expr);

			Type expected = FindType(whilestmt.expr);
			Type realType = ExpectType(whilestmt.expr, expected);

			if (realType != null) {
				Error($"Do While expr expected type {expected} but received {realType}");
			}

			Visit(whilestmt.body);
		}

		void VisitFunctionDef(FunctionDef function) {
			if (scope.HasSymbol(function.token.Lexeme)) {
				Error($"Function '{function.token.Lexeme}' has already been defined");
			}

			FunctionSym func = new FunctionSym(
				function.token.Lexeme,
				new List<VarSym>(),
				function
			);

			scope.Insert(func);
			SymbolScope func_scope = new SymbolScope(function.token.Lexeme, scope.scopeLevel + 1, scope);
			scope = func_scope;

			// Add paramaters
			foreach(Parameter param in function.parameters) {
				VarSym variable = new VarSym(
					param.token.Lexeme,
					param.type,
					param.mutable
				);
				scope.Insert(variable);
				func.parameters.Add(variable);
			}

			// Insert implicit return value
			if (func.def.returnType != null) {
				scope.Insert(new VarSym(
					"result",
					func.def.returnType.type,
					true
				));
			}

			VisitBlock(function.block);

			ClimbScope();
		}

		void VisitFunctionCall(FunctionCall function) {
			FunctionSym func = (FunctionSym)scope.Lookup(function.token.Lexeme, false);

			if (func == null) {
				Error($"Function '{function.token.Lexeme}' does not exist");
			}

			if (function.arguments.Count != func.parameters.Count) {
				Error($"Function '{function.token.Lexeme}' expected {func.parameters.Count} argument(s) but received {function.arguments.Count}");
			}

			// Assign the definition
			function.sym = func;

			int current = 0;
			foreach(Node node in function.arguments) {
				VarSym parameter = func.parameters[current];
				Type realType = ExpectType(node, parameter.type);

				if (realType != null) {
					Error($"'{parameter.identifier}' expected type {parameter.type} but received {realType} at position {current+1}");
				}

				// Incompatible mutability
				if (parameter.mutable && node is Var) {
					VarSym variable = (VarSym)scope.Lookup(node.token.Lexeme, true);

					if (!variable.mutable) {
						Error($"Trying to pass an immutable variable '{node.token.Lexeme}' to parameter '{parameter.identifier}' in function '{function.token.Lexeme}'");
					}
				}

				Visit(node);
				current++;
			}
		}

		void VisitBuiltinCall(BuiltinFunctionCall builtin) {
			if (builtin.native.parity >= 0 && builtin.arguments.Count != builtin.native.parity) {
				Error($"Builtin function '{builtin.identifier}' expected {builtin.native.parity} argument(s) but received {builtin.arguments.Count}");
			}

			int idx = 0;
			foreach(Node node in builtin.arguments) {
				Visit(node);

				if (idx < builtin.native.parity) {
					Type realType = ExpectType(node, builtin.native.required[idx]);

					if (realType != null) {
						ErrorWith($"'{builtin.identifier}' expected type {builtin.native.required[idx]} but received {realType} at position {idx+1}", node);
					}
					idx++;
				}
			}
		}

		void VisitVariableDeclaration(VarDecl decl) {
			if (scope.HasSymbol(decl.identifier)) {
				Error($"Variable '{decl.identifier}' already exists in the current scope");
			}

			Visit(decl.expr);

			Type received = FindType(decl.expr);
			if (decl.type.IsUntyped()) {
				// Assign the type from the RHS
				decl.type = received;
			}

			Type realType = ExpectType(decl.expr, received);

			if (realType != null) {
				Error($"'{decl.identifier}' expected type {decl.type} but received {realType}");
			}

			scope.Insert(new VarSym(
				decl.identifier,
				decl.type,
				decl.mutable
			));
		}

		void VisitAssignment(Assignment assign) {
			string identifier = assign.identifier.token.Lexeme;

			if (assign.identifier is not Var && assign.identifier is not Index) {
				Error($"'{identifier}' Assignment must use an indentifier or index");
			}

			VarSym sym = (VarSym)scope.Lookup(identifier, identifier == "result");

			if (sym == null) {
				Error($"Variable '{assign.identifier}' does not exist in any scope");
			}

			// This is required for some nodes to fetch symbols, which are used in FindType
			Visit(assign.identifier);
			Visit(assign.expr);

			Type realType = ExpectType(assign.expr, FindType(assign.identifier));

			if (realType != null) {
				Error($"Variable '{identifier}' expected type {sym.type} but received {realType}");
			}

			if (!sym.mutable) {
				Error($"Cannot reassign to immutable variable '{identifier}'");
			}
		}
	}
}