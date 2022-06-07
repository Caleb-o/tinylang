using System;
using System.Collections.Generic;

namespace TinyLang {
	enum RecordType {
		Application, Function,
	}
	
	abstract class Symbol {
		public readonly string identifier;
		public readonly RecordType type;

		public Symbol(string identifier, RecordType type) {
			this.identifier = identifier;
			this.type = type;
		}
	}
	
	sealed class VarSym : Symbol {
		public readonly TypeKind kind;
		public readonly bool mutable;
		public Value value;
		// If the type must be handled at run-time
		public bool validated;
		public VarSym references;


		public VarSym(string identifier, bool mutable, FunctionDef def, RecordType type = RecordType.Function) : base(identifier, type) {
			this.mutable = mutable;
			this.kind = TypeKind.Function;
			this.value = new FunctionValue(def);
		}

		public VarSym(string identifier, bool mutable, TypeKind kind, RecordType type = RecordType.Function) : base(identifier, type) {
			this.mutable = mutable;
			this.kind = kind;
		}
	}

	sealed class SymbolTable {
		public SymbolTable parent;
		public readonly string identifier;
		public readonly Dictionary<string, Symbol> symbols = new Dictionary<string, Symbol>();

		public SymbolTable(string identifier, SymbolTable parent) {
			this.identifier = identifier;
			this.parent = parent;
		}

		public void Insert(Symbol symbol) {
			symbols[symbol.identifier] = symbol;
		}

		public bool HasSymbol(string identifier) {
			return symbols.ContainsKey(identifier);
		}

		public Symbol Lookup(string identifier, bool local) {
			if (symbols.ContainsKey(identifier)) {
				return symbols[identifier];
			}

			if (local || parent == null) {
				return null;
			}

			return parent.Lookup(identifier, local);
		}
	}

	sealed class Analyser {
		SymbolTable table = new SymbolTable("global", null);


		public void Analyse(Application application) {
			Visit(application.block);
		}

		void Error(string message) {
			throw new Exception($"Analyser: {message}");
		}

		void Error(string message, Token token) {
			throw new Exception($"Analyser: {message} '{token.Lexeme}' [{token.Line}:{token.Column}]");
		}

		TypeKind FindType(Node node) {
			switch(node) {
				case BinaryOp:			return FindType(((BinaryOp)node).left);
				case FunctionDef:		return TypeKind.Function;
				case Literal:			return Value.TypeFromToken(((Literal)node).token);
				case Argument:			return FindType(((Argument)node).expr);
				case ConditionalOp:		return TypeKind.Bool;

				case Identifier: {
					VarSym variable = (VarSym)table.Lookup(((Identifier)node).token.Lexeme, false);

					if (variable == null) {
						Error($"Variable '{((Identifier)node).token.Lexeme}' does not exist in any scope");
					}

					return variable.kind;
				}
			}

			Error($"Cannot get type kind from node '{node}'");
			return TypeKind.Error;
		}

		TypeKind ExpectType(Node node, TypeKind expected) {
			switch(node) {
				case BinaryOp: {
					TypeKind left = FindType(((BinaryOp)node).left);
					TypeKind right = FindType(((BinaryOp)node).right);

					if (left != expected) {
						return left;
					}

					if (right != expected) {
						return right;
					}

					return expected;
				}

				case FunctionDef: {
					if (expected != TypeKind.Function) {
						return TypeKind.Function;
					}

					return expected;
				}

				case Literal: {
					TypeKind literal = Value.TypeFromToken(((Literal)node).token);

					if (literal != expected) {
						return literal;
					}

					return expected;
				}

				case Argument: {
					TypeKind arg = FindType(((Argument)node).expr);
					
					if (expected != arg) {
						return arg;
					}

					return expected;
				}

				case Identifier: {
					TypeKind identifier = FindType(node);

					if (expected != identifier) {
						return identifier;
					}

					return expected;
				}

				case ConditionalOp: {
					ConditionalOp cond = (ConditionalOp)node;
					TypeKind left = FindType(cond.left);
					TypeKind right = FindType(cond.right);

					// Left and Right should equal
					if (left != right) {
						return right;
					}

					// Expected should be a Bool
					if (expected != TypeKind.Bool) {
						return expected;
					}

					return TypeKind.Bool;
				}
			}

			Error($"Cannot expect type kind from node '{node}'");
			return TypeKind.Error;
		}

		void Visit(Node node) {
			switch(node) {
				case Block: 				VisitBlock((Block)node); break;
				case BinaryOp: 				VisitBinaryOp((BinaryOp)node); break;
				case VariableDecl:			VisitVariableDecl((VariableDecl)node); break;
				case VariableAssignment:	VisitVariableAssign((VariableAssignment)node); break;
				case FunctionDef:			VisitFunctionDef((FunctionDef)node); break;
				case FunctionCall:			VisitFunctionCall((FunctionCall)node); break;
				case Print: 				VisitPrint((Print)node); break;
				case Identifier:			VisitIdentifier((Identifier)node); break;
				case IfStmt:				VisitIfStatement((IfStmt)node); break;
				case WhileStmt:				VisitWhileStatement((WhileStmt)node); break;
				case ConditionalOp:			VisitConditionalOp((ConditionalOp)node); break;

				// NoOp
				case Literal: break;

				default:
					Error($"Unhandled node in analysis {node}");
					break;
			}
		}

		void VisitBlock(Block block) {
			SymbolTable blockTable = new SymbolTable("block", table);
			table = blockTable;

			foreach(Node node in block.statements) {
				Visit(node);
			}

			table = table.parent;
		}

		void VisitBinaryOp(BinaryOp binaryOp) {
			Visit(binaryOp.left);
			Visit(binaryOp.right);

			TypeKind left = FindType(binaryOp.left);
			TypeKind right = ExpectType(binaryOp.right, left);

			if (right != left) {
				Error($"Binary operation expected type {left} but received {right}");
			}
		}

		void Assign(string identifier, TypeKind kind, bool mutable, Node expr) {
			// FIXME: Once the left-most type is found, compare that against the rest of the expression
			TypeKind real = ExpectType(expr, kind);

			if (real != kind) {
				Error($"Trying to reassign '{identifier}' with type {real} but expected {kind}");
			}

			if (kind == TypeKind.Function) {
				if (mutable) {
					Error($"Function '{identifier}' cannot be mutable, use let instead.");
				}

				FunctionDef fndef = (FunctionDef)expr;
				fndef.identifier = identifier;

				Visit(expr);
				table.Insert(new VarSym(identifier, mutable, fndef));
			} else {
				Visit(expr);
				table.Insert(new VarSym(identifier, mutable, kind));
			}
		}

		void VisitVariableDecl(VariableDecl vardecl) {
			if (table.HasSymbol(vardecl.token.Lexeme)) {
				Error($"'{vardecl.token.Lexeme}' has already been defined in the current scope");
			}

			TypeKind kind = FindType(vardecl.expr);
			
			if (vardecl.kind != TypeKind.Unknown && vardecl.kind != kind) {
				Error($"Variable '{vardecl.token.Lexeme}' expected type {vardecl.kind} but received {kind}");
			}
			vardecl.kind = kind;

			Assign(vardecl.token.Lexeme, kind, vardecl.mutable, vardecl.expr);
		}

		void VisitVariableAssign(VariableAssignment assign) {
			VarSym variable = (VarSym)table.Lookup(assign.token.Lexeme, false);

			if (variable == null) {
				Error($"'{assign.token.Lexeme}' has not been defined in the current scope");
			}

			if (!variable.mutable) {
				Error($"'{assign.token.Lexeme}' is immutable and cannot be reassigned");
			}

			Assign(variable.identifier, variable.kind, variable.mutable, assign.expr);
		}

		void VisitFunctionDef(FunctionDef fndef) {
			SymbolTable blockTable = new SymbolTable(fndef.identifier, table);
			table = blockTable;

			foreach(Parameter param in fndef.parameters) {
				table.Insert(new VarSym(param.token.Lexeme, false, param.kind));
			}
			
			VarSym variable = new VarSym(fndef.identifier, false, fndef);
			variable.value = new FunctionValue(fndef);
			variable.validated = true;

			table.Insert(variable);
			Visit(fndef.block);

			table = table.parent;
		}

		void VisitFunctionCall(FunctionCall fncall) {
			VarSym fnsym = (VarSym)table.Lookup(fncall.token.Lexeme, false);

			if (fnsym == null) {
				Error($"Function '{fncall.token.Lexeme}' has not been defined in any scope");
			}

			foreach(Argument arg in fncall.arguments) {
				arg.kind = FindType(arg.expr);
				Visit(arg.expr);
			}
		}

		void VisitPrint(Print print) {
			foreach(Node node in print.arguments) {
				Visit(node);
			}
		}

		void VisitIdentifier(Identifier identifier) {
			if (table.Lookup(identifier.token.Lexeme, false) == null) {
				Error($"Identifier '{identifier.token.Lexeme}' does not exist in any scope", identifier.token);
			}
		}

		void VisitIfStatement(IfStmt stmt) {
			if (stmt.initStatement != null) {
				SymbolTable if_table = new SymbolTable("if_init", table);
				table = if_table;

				Visit(stmt.initStatement);
			}

			Visit(stmt.expr);
			Visit(stmt.trueBody);

			if (stmt.falseBody != null) {
				Visit(stmt.falseBody);
			}

			if (stmt.initStatement != null) {
				table = table.parent;
			}
		}

		void VisitWhileStatement(WhileStmt stmt) {
			if (stmt.initStatement != null) {
				SymbolTable while_table = new SymbolTable("while_init", table);
				table = while_table;

				Visit(stmt.initStatement);
			}

			Visit(stmt.expr);
			Visit(stmt.body);

			if (stmt.initStatement != null) {
				table = table.parent;
			}
		}

		void VisitConditionalOp(ConditionalOp cond) {
			Visit(cond.left);
			Visit(cond.right);

			TypeKind kind = ExpectType(cond, TypeKind.Bool);

			if (kind != TypeKind.Bool) {
				Error($"Conditional expression expected type {TypeKind.Bool} but received {kind}");
			}
		}
	}
}