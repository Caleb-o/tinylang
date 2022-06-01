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
		public BuiltinTypeSym(string identifier)
			: base(identifier, new Type(Application.GetTypeID(identifier))) {}
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
			Insert(new BuiltinTypeSym("int"));
			Insert(new BuiltinTypeSym("float"));
			Insert(new BuiltinTypeSym("string"));
			Insert(new BuiltinTypeSym("boolean"));
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

		public Symbol LookupType(Type type, bool local) {
			if (type.IsSingleType()) {
				int typeID = type.type[0];

				if (symbols.ContainsKey(identifier)) {
					return symbols[identifier];
				}

				if (local) {
					return null;
				}

				return parent.Lookup(identifier, false);
			}

			// Unsupported
			return null;
		}
	}

	sealed class Analyser {
		public SymbolScope scope;

		public void Run(Application app) {
			// Setup the global scope with builtins
			scope = new SymbolScope("global", 0, null);
			scope.InitBuiltins();

			Visit(app.block);
		}

		void Error(string msg) {
			throw new Exception($"Analyser: {msg} : [{scope.identifier}]");
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
					// FIXME: This should use the type
					return ((FunctionCall)node).sym.def.returnType.type;
				}

				case BinOp: {
					return FindType(((BinOp)node).left);
				}

				case Literal: {
					return new Type(Application.GetTypeID(((Literal)node).token.Kind.ToString().ToLower()));
				}

				case UnaryOp: {
					return FindType(((UnaryOp)node).right);
				}

				case ComplexLiteral: {
					ComplexLiteral literal = (ComplexLiteral)node;
					List<int> typeIDs = new List<int>();

					foreach(Node expr in literal.exprs) {
						// Note: This can probably be more than one in the future,
						// 		 so a better method may be required
						typeIDs.Add(FindType(expr).type[0]);
					}

					return new Type(typeIDs.ToArray());
				}

				default:
					Error($"Unknown node in find type: {node}");
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
				case DoWhile: VisitWhile((While)node); break;
				case ConditionalOp: VisitConditionalOp((ConditionalOp)node); break;

				case Literal: break;
				case ComplexLiteral: break;
				case UnaryOp: Visit(((UnaryOp)node).right); break;

				default:
					throw new Exception($"Unimplemented node in analyser: {node.GetType()}");
			}
		}

		void VisitBinOp(BinOp binop) {
			Visit(binop.left);
			Visit(binop.right);
		}

		void VisitConditionalOp(ConditionalOp conditional) {
			Visit(conditional.left);
			Visit(conditional.right);
		}

		void VisitVar(Var var) {
			if (scope.Lookup(var.token.Lexeme, false) == null) {
				Error($"Variable '{var.token.Lexeme}' does not exist in any scope");
			}
		}

		void VisitBlock(Block block) {
			foreach(Node node in block.statements) {
				Visit(node);
			}
		}

		void VisitEscape(Escape ret) {
			if (scope.identifier == "global") {
				// TODO: This can probably occur for early returns
				Error("Cannot escape from global scope");
			}
		}

		void VisitIfStatement(IfStmt ifstmt) {
			Visit(ifstmt.expr);
			Visit(ifstmt.trueBody);

			if (ifstmt.falseBody != null) {
				Visit(ifstmt.falseBody);
			}
		}

		void VisitWhile(While whilestmt) {
			Visit(whilestmt.expr);
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
			if (!func.def.returnType.type.IsVoid()) {
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
				Type expected = parameter.type;
				Type ntype = FindType(node);

				if (!ntype.Matches(expected)) {
					Error($"'{function.token.Lexeme}' argument at position {current + 1} expected type {expected} but received {ntype}");
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

			// TODO: Typecheck builtin function arguments
			foreach(Node node in builtin.arguments) {
				Visit(node);
			}
		}

		void VisitVariableDeclaration(VarDecl decl) {
			if (scope.HasSymbol(decl.identifier)) {
				Error($"Variable '{decl.identifier}' already exists in the current scope");
			}

			Visit(decl.expr);

			Type received = FindType(decl.expr);
			if (!received.Matches(decl.type)) {
				Error($"'{decl.identifier}' expected type {decl.type} but received {received}");
			}

			scope.Insert(new VarSym(
				decl.identifier,
				decl.type,
				decl.mutable
			));
		}

		void VisitAssignment(Assignment assign) {
			VarSym sym = (VarSym)scope.Lookup(assign.identifier, (assign.identifier == "result"));

			if (sym == null) {
				Error($"Variable '{assign.identifier}' does not exist in any scope");
			}

			Visit(assign.expr);

			Type received = FindType(assign.expr);
			if (!received.Matches(sym.type)) {
				Error($"'{assign.identifier}' expected type {sym.type} but received {received}");
			}

			if (!sym.mutable) {
				Error($"Cannot reassign to immutable variable '{assign.identifier}'");
			}
		}
	}
}