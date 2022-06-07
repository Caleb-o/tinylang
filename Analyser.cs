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
		public readonly TinyType kind;
		public readonly bool mutable;
		public Value value;
		// If the type must be handled at run-time
		public bool validated;
		public VarSym references;


		public VarSym(string identifier, bool mutable, FunctionDef def, RecordType type = RecordType.Function) : base(identifier, type) {
			this.mutable = mutable;

			List<TinyType> types = new List<TinyType>();
			
			foreach(Parameter param in def.parameters) {
				types.Add(param.kind);
			}

			this.kind = new TinyFunction(def.identifier, types, def.returns);
			this.value = new FunctionValue(def);
		}

		public VarSym(string identifier, bool mutable, TinyType kind, RecordType type = RecordType.Function) : base(identifier, type) {
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
		Block currentBlock = null;


		public void Analyse(Application application) {
			Visit(application.block);
		}

		void Error(string message) {
			throw new Exception($"Analyser: {message}");
		}

		void Error(string message, Token token) {
			throw new Exception($"Analyser: {message} '{token.Lexeme}' [{token.Line}:{token.Column}]");
		}

		TinyType FindType(Node node) {
			switch(node) {
				case BinaryOp:			return FindType(((BinaryOp)node).left);
				case FunctionDef:		return new TinyFunction();
				case FunctionCall: {
					VarSym def = (VarSym)table.Lookup(((FunctionCall)node).token.Lexeme, false);

					while (def.references != null) {
						def = def.references;
					}

					if (!def.validated) {
						// Assume Any
						return new TinyAny();
					}

					return ((FunctionDef)def.value.Data).returns;
				}
				case Literal:			return TinyType.TypeFromToken(((Literal)node).token);
				case Argument:			return FindType(((Argument)node).expr);
				case ConditionalOp:		return new TinyBool();

				case Identifier: {
					VarSym variable = (VarSym)table.Lookup(((Identifier)node).token.Lexeme, false);

					if (variable == null) {
						Error($"Variable '{((Identifier)node).token.Lexeme}' does not exist in any scope");
					}

					return variable.kind;
				}

				case ListLiteral: {
					return new TinyList(FindType(((ListLiteral)node).exprs[0]));
				}
			}

			Error($"Cannot get type kind from node '{node}'");
			return null;
		}

		TinyType ExpectType(Node node, TinyType expected) {
			switch(node) {
				case BinaryOp: {
					TinyType left = FindType(((BinaryOp)node).left);
					TinyType right = FindType(((BinaryOp)node).right);

					if (!TinyType.Matches(left, expected)) {
						return left;
					}

					if (!TinyType.Matches(right, expected)) {
						return right;
					}

					return expected;
				}

				case FunctionDef: {
					if (expected is not TinyFunction) {
						return new TinyFunction();
					}

					return expected;
				}

				case FunctionCall: {
					VarSym def = (VarSym)table.Lookup(((FunctionCall)node).token.Lexeme, false);

					while (def.references != null) {
						def = def.references;
					}

					if (!def.validated) {
						// Assume Any
						return expected;
					}

					TinyType ret = ((FunctionDef)def.value.Data).returns;

					if (!TinyType.Matches(expected, ret)) {
						return ret;
					}

					return expected;
				}

				case Literal: {
					TinyType literal = TinyType.TypeFromToken(((Literal)node).token);

					if (!TinyType.Matches(literal, expected)) {
						return literal;
					}

					return expected;
				}

				case Argument: {
					TinyType arg = FindType(((Argument)node).expr);
					
					if (!TinyType.Matches(expected, arg)) {
						return arg;
					}

					return expected;
				}

				case Identifier: {
					TinyType identifier = FindType(node);

					if (!TinyType.Matches(expected, identifier)) {
						return identifier;
					}

					return expected;
				}

				case ConditionalOp: {
					ConditionalOp cond = (ConditionalOp)node;
					TinyType left = FindType(cond.left);
					TinyType right = FindType(cond.right);

					// Left and Right should equal
					if (!TinyType.Matches(left, right)) {
						return right;
					}

					// Expected should be a Bool
					if (expected is not TinyBool) {
						return expected;
					}

					return new TinyBool();
				}

				case ListLiteral: {
					if (!TinyType.Matches(expected, FindType(node))) {
						return new TinyList();
					}

					ListLiteral literal = (ListLiteral)node;

					TinyType inner = FindType(literal.exprs[0]);
					TinyType exInner = ((TinyList)expected).inner;

					if (!TinyType.Matches(exInner, inner)) {
						return new TinyList(inner);
					}

					foreach(Node expr in literal.exprs) {
						TinyType exKind = ExpectType(expr, exInner);
						if (!TinyType.Matches(exInner, exKind)) {
							return new TinyList(exKind);
						}
					}

					return expected;
				}
			}

			Error($"Cannot expect type kind from node '{node}'");
			return null;
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
				case DoWhileStmt:			VisitDoWhileStatement((DoWhileStmt)node); break;
				case ConditionalOp:			VisitConditionalOp((ConditionalOp)node); break;
				case ListLiteral:			VisitListLiteral((ListLiteral)node); break;
				case Return:				VisitReturn((Return)node); break;

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

			currentBlock = block;

			foreach(Node node in block.statements) {
				Visit(node);
			}

			table = table.parent;
		}

		void VisitBinaryOp(BinaryOp binaryOp) {
			Visit(binaryOp.left);
			Visit(binaryOp.right);

			TinyType left = FindType(binaryOp.left);
			TinyType right = ExpectType(binaryOp.right, left);

			if (!TinyType.Matches(right, left)) {
				Error($"Binary operation expected type {left} but received {right}");
			}
		}

		void Assign(string identifier, TinyType kind, bool mutable, Node expr) {
			// FIXME: Once the left-most type is found, compare that against the rest of the expression
			TinyType real = ExpectType(expr, kind);

			if (!TinyType.Matches(real, kind)) {
				Error($"Trying to reassign '{identifier}' with type {real} but expected {kind}");
			}

			if (kind is TinyFunction) {
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

			TinyType kind = FindType(vardecl.expr);
			
			if (vardecl.kind is not TinyAny && !TinyType.Matches(vardecl.kind, kind)) {
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

			// Add implicit return value
			VarSym result = new VarSym("result", true, fndef.returns);
			result.value = new UnitValue();
			result.validated = true;
			table.Insert(result);
			
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

		void VisitDoWhileStatement(DoWhileStmt stmt) {
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

			TinyType kind = ExpectType(cond, new TinyBool());

			if (kind is not TinyBool) {
				Error($"Conditional expression expected type {new TinyBool()} but received {kind}");
			}
		}

		void VisitListLiteral(ListLiteral literal) {
			if (literal.exprs.Count == 0) {
				// Fixme: Allow empty lists
				Error("List literal cannot be empty");
			}

			TinyType kind = FindType(literal);
			TinyType real = ExpectType(literal, kind);

			if (!TinyType.Matches(kind, real)) {
				Error($"List literal expected type {kind} but received {real}");
			}

			literal.kind = kind;

			foreach(Node expr in literal.exprs) {
				Visit(expr);
			}
		}

		void VisitReturn(Return ret) {
			if (currentBlock.returnstmt != null) {
				Error("Blocks can only contain a single return statement", ret.token);
			}

			currentBlock.returnstmt = ret;
		}
	}
}