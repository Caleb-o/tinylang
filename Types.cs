using System;
using System.Collections.Generic;


namespace TinyLang {
	class TinyType {
		public static TinyType TypeFromToken(Token token) {
			switch(token.Kind) {
				case TokenKind.Int:				return new TinyInt();
				case TokenKind.Float:			return new TinyFloat();
				case TokenKind.Bool:			return new TinyBool();
				case TokenKind.String:			return new TinyString();
			}

			throw new InvalidOperationException($"Unable to determine type from '{token.Lexeme}':{token.Kind}");
		}

		public static TinyType TypeFromLexeme(Token token) {
			switch(token.Lexeme) {
				case "int":				return new TinyInt();
				case "float":			return new TinyFloat();
				case "bool":			return new TinyBool();
				case "string":			return new TinyString();
			}

			throw new InvalidOperationException($"Unable to determine type from '{token.Lexeme}'");
		}
	}

	class TinyAny : TinyType {}
	class TinyUnit : TinyType {}
	class TinyInt : TinyType {}
	class TinyFloat : TinyType {}
	class TinyBool : TinyType {}
	class TinyString : TinyType {}

	class TinyFunction : TinyType {
		public readonly string identifier;
		public readonly List<TinyType> parameters;
		public readonly TinyType returns;

		// For generic-like calls
		public TinyFunction() {}

		public TinyFunction(string identifier, List<TinyType> parameters, TinyType returns) {
			this.identifier = identifier;
			this.parameters = parameters;
			this.returns = returns;
		}
	}
	
	class TinyList : TinyType {
		public readonly TinyType inner;

		public TinyList(TinyType inner) {
			this.inner = inner;
		}
	}
}