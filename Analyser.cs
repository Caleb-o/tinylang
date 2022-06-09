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
			this.validated = true;

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

	sealed class TypeSym : Symbol {
		public TypeSym(string identifier) : base (identifier) {}
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
		public static bool HasError { get; private set; } = false;

		SymbolTable table = new SymbolTable("global", null);
		Block currentBlock = null;


		public Analyser() {
			VarSym result = new VarSym("result", true, new TinyAny());
			result.value = new UnitValue();
			result.validated = true;
			table.Insert(result);

			AddTypes();
		}

		public void Run(Application application) {
			Visit(application.block);
		}

		public void ImportFunction(BuiltinFn func) {
			table.Insert(new VarSym(func.identifier, new FunctionDef(func.identifier, func)));
		}

		void AddTypes() {
			table.Insert(new TypeSym("int"));
			table.Insert(new TypeSym("float"));
			table.Insert(new TypeSym("bool"));
			table.Insert(new TypeSym("string"));
		}

		void Error(string message) {
			HasError = true;
			Reporter.Report(message);
		}

		void Error(string message, Token token) {
			Error($"{message} '{token.Lexeme}' [{token.Line}:{token.Column}]");
		}

		TinyType FindType(Node node) {
			switch(node) {
				case BinaryOp:			return FindType(((BinaryOp)node).left);
				case UnaryOp:			return FindType(((UnaryOp)node).right);
				case FunctionDef:		return new TinyFunction();
				case StructDef:			return new TinyStruct();
				
				case StructInstance: {
					StructInstance instance = (StructInstance)node;

					if (instance.def != null) {
						return new TinyStruct(instance.def);
					}

					return new TinyStruct(instance.identifier);
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
					if (((ListLiteral)node).exprs.Count == 0) {
						return new TinyAny();
					}
					return new TinyList(FindType(((ListLiteral)node).exprs[0]));
				}

				case Index: {
					return ((Index)node).kind;
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
					TinyType id = ((VarSym)table.Lookup(node.token.Lexeme, false)).kind;

					if (!TinyType.Matches(expected, id)) {
						return id;
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

					if (literal.exprs.Count == 0) {
						return new TinyAny();
					}

					TinyType inner = FindType(literal.exprs[0]);
					TinyType exInner = ((TinyList)expected).inner;

					if (!TinyType.Matches(exInner, inner)) {
						return new TinyList(inner);
					}

					foreach(Node expr in literal.exprs) {
						TinyType exKind = FindType(expr);
						if (!TinyType.Matches(exInner, exKind)) {
							// We can't just return a list literal, as that will be inferred
							// If we're failing at matching the contents, then we can't continue
							Error($"List literal expected inner type {exInner} but received {exKind}");
							return null;
						}
					}

					return expected;
				}

				case Index: {
					Index index = (Index)node;

					if (!TinyType.Matches(expected, index.kind)) {
						return index.kind;
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
				case StructDef:				VisitStructDef((StructDef)node); break;
				case StructInstance:		VisitStructInstance((StructInstance)node); break;
				case Index:					VisitIndex((Index)node); break;

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
				Error($"Binary operation expected type {left} but received {right}", binaryOp.left.token);
			}
		}

		void VisitUnaryOp(UnaryOp uanry) {
			Visit(uanry.right);
		}

		void Assign(string identifier, TinyType kind, bool mutable, Node expr, bool reassign) {
			if (expr is not FunctionDef) Visit(expr);
			
			TinyType real = FindType(expr);
			
			if (!TinyType.Matches(real, kind)) {
				Error($"Trying to reassign '{identifier}' with type {real} but expected {kind}", expr.token);
			}

			switch(expr) {
				case ListLiteral: {
					ListLiteral literal = (ListLiteral)expr;
					bool isany = kind is TinyAny;

					if (kind is TinyAny) {
						if (isany) {
							Error($"Cannot assign empty list to untyped variable", literal.token);
						}
					} else {
						if (isany) {
							literal.kind = kind;
						}
					}

					if (!reassign) table.Insert(new VarSym(identifier, mutable, kind));
					break;
				}

				case StructInstance: {
					table.Insert(new VarSym(identifier, mutable, real));
					break;
				}

				default: {
					if (!reassign) table.Insert(new VarSym(identifier, mutable, kind));
					break;
				}
			}
		}

		void VisitVariableDecl(VariableDecl vardecl) {
			if (table.HasSymbol(vardecl.token.Lexeme)) {
				Error($"'{vardecl.token.Lexeme}' has already been defined in the current scope");
			}

			if (vardecl.kind is not TinyAny && table.Lookup(vardecl.kind.Inspect(), false) == null) {
				Error($"Type {vardecl.kind.Inspect()} does not exist in any scope");
			}

			// Special case
			if (vardecl.expr is Index) {
				Visit(vardecl.expr);
			}

			TinyType kind = FindType(vardecl.expr);
			
			if (vardecl.kind is TinyAny) {
				vardecl.kind = kind;
			} else {
				if (!TinyType.Matches(vardecl.kind, kind)) {
					Error($"Variable '{vardecl.token.Lexeme}' expected type {vardecl.kind} but received {kind}");
				}
			}

			Assign(vardecl.token.Lexeme, vardecl.kind, vardecl.mutable, vardecl.expr, false);
		}

		void VisitVariableAssign(VariableAssignment assign) {
			VarSym variable = (VarSym)table.Lookup(assign.token.Lexeme, false);

			if (variable == null) {
				Error($"'{assign.token.Lexeme}' has not been defined in the current scope");
			}

			if (!variable.mutable) {
				Error($"'{assign.token.Lexeme}' is immutable and cannot be reassigned");
			}

			Visit(assign.identifier);

			Assign(variable.identifier, FindType(assign.identifier), variable.mutable, assign.expr, true);
		}

		void VisitFunctionDef(FunctionDef fndef) {
			VarSym variable = new VarSym(fndef.identifier, fndef);
			variable.value = new FunctionValue(fndef);
			variable.validated = true;
			table.Insert(variable);

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

			Visit((Block)fndef.block);

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

		void VisitStructDef(StructDef def) {
			table.Insert(new VarSym(def.identifier, def));
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
			instance.def = new StructDef(sdef.def);

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

				TinyType expected = instance.def.fields[id];
				TinyType kind = FindType(expr);

				if (!TinyType.Matches(expected, kind)) {
					Error($"Struct field {id} expected type {expected} but received {kind}", expr.token);
				}

				instance.def.fields[id] = kind;
			}
		}

		void VisitIndex(Index index) {
			VarSym caller = (VarSym)table.Lookup(index.token.Lexeme, false);

			if (caller == null) {
				Error($"Variable '{index.token.Lexeme}' does not exist in any scope");
			}

			if (caller.kind is not TinyList) {
				Error($"Cannot index non-list '{caller.identifier}'");
			}

			TinyType intType = new TinyInt();
			TinyType inner = caller.kind;
			int idx = 0;

			foreach(Node expr in index.expr) {
				Visit(expr);
				
				TinyType kind = FindType(expr);
				if (!TinyType.Matches(kind, intType)) {
					Error($"Value at position {idx + 1} in list index expected {intType} but received {kind}", index.token);
				}

				// Cannot index a non-list
				if (inner is not TinyList) {
					Error($"Variable '{caller.identifier}' cannot index type {inner}", index.token);
				}

				TinyType last = inner;
				inner = inner.Inner();

				if (inner is TinyNone) {
					Error($"Variable '{caller.identifier}' in {last} does not contain a type", index.token);
				}

				idx++;
			}

			index.kind = inner;
		}
	}
}