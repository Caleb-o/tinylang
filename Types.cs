using System;
using System.Collections.Generic;


namespace TinyLang {
	abstract class TinyType {
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

		public static bool Matches(TinyType x, TinyType y) {
			if (x is TinyAny || y is TinyAny) {
				return true;
			}

			switch(x) {
				case TinyList: {
					if (x.GetType() != y.GetType()) {
						return false;
					}

					return ((TinyList)x).inner.GetType() == ((TinyList)y).inner.GetType();
				}

				default:
					return x.GetType() == y.GetType();
			}
		}
	}

	sealed class TinyAny : TinyType {
		public override string ToString()
		{
			return "any";
		}
	}

	sealed class TinyUnit : TinyType {
		public override string ToString()
		{
			return "unit";
		}
	}
	
	sealed class TinyInt : TinyType {
		public override string ToString()
		{
			return "int";
		}
	}

	sealed class TinyFloat : TinyType {
		public override string ToString()
		{
			return "float";
		}
	}
	
	sealed class TinyBool : TinyType {
		public override string ToString()
		{
			return "bool";
		}
	}
	
	sealed class TinyString : TinyType {
		public override string ToString()
		{
			return "string";
		}
	}

	sealed class TinyFunction : TinyType {
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
	
	sealed class TinyList : TinyType {
		public readonly TinyType inner;

		public TinyList() {}

		public TinyList(TinyType inner) {
			this.inner = inner;
		}

		public override string ToString()
		{
			return $"[{inner}]";
		}
	}

	sealed class TinyStruct : TinyType {
		// Name of the struct
		public readonly string identifier;
		public readonly Dictionary<string, TinyType> members;

		public TinyStruct() {}

		public TinyStruct(StructDef def) {
			this.identifier = def.identifier;
			this.members = def.members;
		}
	}
}