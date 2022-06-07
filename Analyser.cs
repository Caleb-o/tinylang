using System;
using System.Linq;
using System.Collections.Generic;

namespace TinyLang {
	abstract class Symbol {
		public readonly string identifier;

		public Symbol(string identifier) {
			this.identifier = identifier;
		}
	}
	
	sealed class VarSym : Symbol {
		public readonly TinyType kind;
		public readonly bool mutable;
		public Value value;
		// If the type must be handled at run-time
		public bool validated;
		public VarSym references;


		// Function Def
		public VarSym(string identifier, FunctionDef def) : base(identifier) {
			this.mutable = false;

			List<TinyType> types = new List<TinyType>();
			
			foreach(Parameter param in def.parameters) {
				types.Add(param.kind);
			}

			this.kind = new TinyFunction(def.identifier, types, def.returns);
			this.value = new FunctionValue(def);
		}

		// Struct Def
		public VarSym(string identifier, StructDef def) : base(identifier) {
			this.mutable = false;

			this.kind = new TinyStruct(def);
			this.value = new StructValue(def, new List<Value>());
		}

		public VarSym(string identifier, bool mutable, TinyType kind) : base(identifier) {
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
			VarSym result = new VarSym("result", true, new TinyAny());
			result.value = new UnitValue();
			result.validated = true;
			table.Insert(result);

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
				case UnaryOp:			return FindType(((UnaryOp)node).right);
				case FunctionDef:		return new TinyFunction();
				case StructDef:			return new TinyStruct();
				
				case StructInstance: {
					return new TinyStruct(((StructInstance)node).identifier);
				}

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

				case UnaryOp: {
					TinyType right = FindType(((UnaryOp)node).right);

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

				case StructDef: {
					if (expected is not TinyStruct) {
						return new TinyStruct();
					}

					return expected;
				}

				case StructInstance: {
					StructInstance instance = (StructInstance)node;
					TinyStruct kind = new TinyStruct(instance.identifier);
					
					if (expected is not TinyStruct) {
						return kind;
					}

					if (!TinyType.Matches(expected, kind)) {
						return kind;
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
					TinyType listType = FindType(node);

					if (!TinyType.Matches(expected, listType)) {
						return listType;
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
				case UnaryOp: 				VisitUnaryOp((UnaryOp)node); break;
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
				case StructInstance:		VisitStructInstance((StructInstance)node); break;

				// NoOp
				case Literal: break;
				case StructDef: break;

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

		void VisitUnaryOp(UnaryOp uanry) {
			Visit(uanry.right);
		}

		void Assign(string identifier, TinyType kind, bool mutable, Node expr) {
			// FIXME: Once the left-most type is found, compare that against the rest of the expression
			TinyType real = ExpectType(expr, kind);

			if (!TinyType.Matches(real, kind)) {
				Error($"Trying to reassign '{identifier}' with type {real} but expected {kind}");
			}
			
			switch(expr) {
				case FunctionDef: {
					if (mutable) {
						Error($"Function definition '{identifier}' cannot be mutable, use let instead.");
					}

					FunctionDef fndef = (FunctionDef)expr;
					fndef.identifier = identifier;

					// Must be done here
					Visit(expr);
					table.Insert(new VarSym(identifier, fndef));
					break;
				}

				case StructDef: {
					if (mutable) {
						Error($"Struct definition '{identifier}' cannot be mutable, use let instead.");
					}

					StructDef def = (StructDef)expr;
					def.identifier = identifier;

					Visit(expr);
					table.Insert(new VarSym(identifier, def));
					break;
				}

				default: {
					Visit(expr);
					table.Insert(new VarSym(identifier, mutable, kind));
					break;
				}
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
				table.Insert(new VarSym(param.token.Lexeme, param.mutable, param.kind));
			}

			// Add implicit return value
			VarSym result = new VarSym("result", true, fndef.returns);
			result.value = new UnitValue();
			result.validated = true;
			table.Insert(result);
			
			VarSym variable = new VarSym(fndef.identifier, fndef);
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

			if (!TinyType.Matches(fnsym.kind, new TinyFunction())) {
				Error($"Identifier '{fncall.token.Lexeme}' is not of type function ({fnsym.kind})", fncall.token);
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

		void VisitStructInstance(StructInstance instance) {
			VarSym def = (VarSym)table.Lookup(instance.identifier, false);
			
			if (def == null) {
				Error($"Struct type '{instance.identifier}' does not exist in any scope", instance.token);
			}

			if (def.kind is not TinyStruct) {
				Error($"Identifier '{instance.identifier}' is not of type struct", instance.token);
			}

			TinyStruct sdef = (TinyStruct)def.kind;
			instance.def = sdef.def;

			if (sdef.fields.Count != instance.members.Count) {
				Error($"Struct initialiser expected {sdef.fields.Count} arguments but received {instance.members.Count}", instance.token);
			}

			foreach(var ((id, expr), field) in instance.members.Zip(sdef.fields.Keys)) {
				if (!sdef.fields.ContainsKey(id)) {
					Error($"Struct type {instance.identifier} does not contain a field {id}", instance.token);
				}

				if (id != field) {
					Error($"Struct {instance.identifier} initialiser expected field {field} but received {id}");
				}

				Visit(expr);

				TinyType expected = sdef.fields[id];
				TinyType kind = FindType(expr);

				if (!TinyType.Matches(expected, kind)) {
					Error($"Struct field {id} expected type {expected} but received {kind}", instance.token);
				}
			}
		}
	}
}