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

		public void Insert(string identifier, Symbol symbol) {
			symbols[identifier] = symbol;
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
}