using System;
using System.Collections.Generic;


namespace TinyLang {
	abstract class TinyType {
		public abstract string Inspect();


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

				// Assume a struct?
				default:				return new TinyStruct(token.Lexeme);
			}

			// throw new InvalidOperationException($"Unable to determine type from '{token.Lexeme}'");
		}

		public static bool Matches(TinyType x, TinyType y) {
			if (x is TinyAny || y is TinyAny) {
				return true;
			}

			bool sameClass = x.GetType() == y.GetType();

			if (!sameClass) {
				return false;
			}

			switch(x) {
				case TinyStruct: {
					return ((TinyStruct)x).identifier == ((TinyStruct)y).identifier;
				}

				case TinyList: {
					if (x.GetType() != y.GetType()) {
						return false;
					}

					// We must use match to compare the inner type, instead of just checking its 
					// C# class type (like I did prior)
					return TinyType.Matches(((TinyList)x).inner, ((TinyList)y).inner);
				}

				default: 		return sameClass;
			}
		}
	}

	sealed class TinyAny : TinyType {
		public override string Inspect() => "any";
		public override string ToString()  => "any";
	}

	sealed class TinyUnit : TinyType {
		public override string Inspect() => "unit";
		public override string ToString() => "unit";
	}
	
	sealed class TinyInt : TinyType {
		public override string Inspect() => "int";
		public override string ToString() => "int";
	}

	sealed class TinyFloat : TinyType {
		public override string Inspect() => "float";
		public override string ToString() => "float";
	}
	
	sealed class TinyBool : TinyType {
		public override string Inspect() => "bool";
		public override string ToString() => "bool";
	}
	
	sealed class TinyString : TinyType {
		public override string Inspect() => "string";
		public override string ToString() => "string";
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

		public override string Inspect() => identifier;
		public override string ToString() => $"{identifier}(...)";
	}
	
	sealed class TinyList : TinyType {
		public readonly TinyType inner;

		public TinyList() {}

		public TinyList(TinyType inner) {
			this.inner = inner;
		}

		public override string Inspect() => inner.ToString();
		public override string ToString() => $"[{inner}]";
	}

	sealed class TinyStruct : TinyType {
		// Name of the struct
		public readonly string identifier;
		public StructDef def;
		public readonly Dictionary<string, TinyType> fields;

		public TinyStruct() {}

		public TinyStruct(string identifier) {
			this.identifier = identifier;
		}

		public TinyStruct(StructDef def) {
			this.def = def;
			this.identifier = def.identifier;
			this.fields = def.fields;
		}

		public override string Inspect() => identifier;
		public override string ToString() => identifier;
	}
}