using System;
using System.Collections.Generic;


namespace TinyLang {
	abstract class TinyType {
		public abstract string Inspect();
		public abstract TinyType Inner();


		public static TinyType TypeFromToken(Token token) {
			switch(token.Kind) {
				case TokenKind.Int:				return new TinyInt();
				case TokenKind.Float:			return new TinyFloat();
				case TokenKind.Bool:			return new TinyBool();
				case TokenKind.String:			return new TinyString();
			}

			throw new InvalidOperationException($"Unable to determine type from '{token.Lexeme}':{token.Kind}");
		}

		public static TinyType TypeFromLexeme(string identifier) {
			switch(identifier) {
				case "int":				return new TinyInt();
				case "float":			return new TinyFloat();
				case "bool":			return new TinyBool();
				case "string":			return new TinyString();

				// Assume a struct?
				default:				return new TinyStruct(identifier);
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
					// We must use match to compare the inner type, instead of just checking its 
					// C# class type (like I did prior)
					return TinyType.Matches(((TinyList)x).inner, ((TinyList)y).inner);
				}

				default: 		return sameClass;
			}
		}
	}

	sealed class TinyNone : TinyType {
		public override string Inspect() => "none";
		public override string ToString()  => "none";
		public override TinyType Inner() => new TinyNone();
	}

	sealed class TinyAny : TinyType {
		public override string Inspect() => "any";
		public override string ToString()  => "any";
		public override TinyType Inner() => new TinyNone();
	}

	sealed class TinyUnit : TinyType {
		public override string Inspect() => "unit";
		public override string ToString() => "unit";
		public override TinyType Inner() => new TinyNone();
	}
	
	sealed class TinyInt : TinyType {
		public override string Inspect() => "int";
		public override string ToString() => "int";
		public override TinyType Inner() => new TinyNone();
	}

	sealed class TinyFloat : TinyType {
		public override string Inspect() => "float";
		public override string ToString() => "float";
		public override TinyType Inner() => new TinyNone();
	}
	
	sealed class TinyBool : TinyType {
		public override string Inspect() => "bool";
		public override string ToString() => "bool";
		public override TinyType Inner() => new TinyNone();
	}
	
	sealed class TinyString : TinyType {
		public override string Inspect() => "string";
		public override string ToString() => "string";
		public override TinyType Inner() => new TinyNone();
	}

	sealed class TinyFunction : TinyType {
		public readonly string identifier;
		public readonly List<TinyType> parameters;
		public readonly TinyType returns;

		// For generic-like calls
		public TinyFunction() {}

		public TinyFunction(string identifier) {
			this.identifier = identifier;
		}

		public TinyFunction(string identifier, List<TinyType> parameters, TinyType returns) {
			this.identifier = identifier;
			this.parameters = parameters;
			this.returns = returns;
		}

		public override string Inspect() => identifier;
		public override string ToString() => $"{identifier}(...)";
		public override TinyType Inner() => new TinyNone();
	}
	
	sealed class TinyList : TinyType {
		public readonly TinyType inner;

		public TinyList() {}

		public TinyList(TinyType inner) {
			this.inner = inner;
		}

		public override string Inspect() => inner.Inspect();
		public override string ToString() => $"[{inner}]";
		public override TinyType Inner() => inner;
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
		public override TinyType Inner() => new TinyNone();
	}
}