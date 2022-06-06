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
			}

			Error($"Cannot get type kind from node '{node}'");
			return TypeKind.Error;
		}

		void Visit(Node node) {
			switch(node) {
				case Block: 			VisitBlock((Block)node); break;
				case BinaryOp: 			VisitBinaryOp((BinaryOp)node); break;
				case VariableDecl:		VisitVariableDecl((VariableDecl)node); break;
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

			Visit(vardecl.expr);
			// FIXME: Once the left-most type is found, compare that against the rest of the expression
			table.Insert(new VarSym(vardecl.token.Lexeme, FindType(vardecl.expr)));
		}
	}
}