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
		public Value value;

		public VarSym(string identifier, TypeKind kind, RecordType type = RecordType.Function) : base(identifier, type) {
			this.kind = kind;
		}
	}

	sealed class FunctionSym : Symbol {
		public FunctionDef def;

		public FunctionSym(string identifier, FunctionDef def, RecordType type = RecordType.Function) : base(identifier, type) {
			this.def = def;
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

			if (local) {
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

		TypeKind FindType(Node node) {
			switch(node) {
				case BinaryOp:			return FindType(((BinaryOp)node).left);
				case FunctionDef:		return TypeKind.Function;
				case Literal:			return Value.TypeFromToken(((Literal)node).token);
				case Argument:			return FindType(((Argument)node).expr);
			}

			Error($"Cannot get type kind from node '{node}'");
			return TypeKind.Error;
		}

		void Visit(Node node) {
			switch(node) {
				case Block: 			VisitBlock((Block)node); break;
				case BinaryOp: 			VisitBinaryOp((BinaryOp)node); break;
				case VariableDecl:		VisitVariableDecl((VariableDecl)node); break;
				case FunctionDef:		VisitFunctionDef((FunctionDef)node); break;
				case FunctionCall:		VisitFunctionCall((FunctionCall)node); break;

				// NoOp
				case Print: break;
				case Literal: break;

				default:
					Error($"Unhandled node in analysis {node}");
					break;
			}
		}

		void VisitBlock(Block block) {
			foreach(Node node in block.statements) {
				Visit(node);
			}
		}

		void VisitBinaryOp(BinaryOp binaryOp) {
			Visit(binaryOp.left);
			Visit(binaryOp.right);
		}

		void VisitVariableDecl(VariableDecl vardecl) {
			if (table.HasSymbol(vardecl.token.Lexeme)) {
				Error($"'{vardecl.token.Lexeme}' has already been defined in the current scope");
			}

			// FIXME: Once the left-most type is found, compare that against the rest of the expression
			TypeKind kind = FindType(vardecl.expr);

			if (kind == TypeKind.Function) {
				FunctionDef fndef = (FunctionDef)vardecl.expr;
				fndef.identifier = vardecl.token.Lexeme;

				table.Insert(new FunctionSym(vardecl.token.Lexeme, fndef));
				Visit(vardecl.expr);
			} else {
				Visit(vardecl.expr);
				table.Insert(new VarSym(vardecl.token.Lexeme, kind));
			}
		}

		void VisitFunctionDef(FunctionDef fndef) {
			Visit(fndef.block);
			
			foreach(Identifier id in fndef.parameters) {
				table.Insert(new VarSym(id.token.Lexeme, TypeKind.Unknown));
			}

			table.Insert(new FunctionSym(fndef.identifier, fndef));
		}

		void VisitFunctionCall(FunctionCall fncall) {
			FunctionSym fnsym = (FunctionSym)table.Lookup(fncall.token.Lexeme, false);

			if (fnsym == null) {
				Error($"Function '{fncall.token.Lexeme}' has not been defined in any scope");
			}

			fncall.def = fnsym.def;

			if (fncall.arguments.Count != fnsym.def.parameters.Count) {
				Error($"Function '{fncall.token.Lexeme}' expected {fnsym.def.parameters.Count} arguments but received {fncall.arguments.Count}");
			}

			foreach(Argument arg in fncall.arguments) {
				Visit(arg.expr);
				arg.kind = FindType(arg);
			}
		}
	}
}