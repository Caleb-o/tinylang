using System;
using System.Collections.Generic;

namespace TinyLang {
	enum TypeKind {
		Int, Float, Bool, String, Function, Struct,
		List, Dictionary, Error,
	}

	abstract class Value {
		public readonly TypeKind Kind;
		public readonly object Data;

		public Value(TypeKind kind, object data) {
			this.Kind = kind;
			this.Data = data;
		}

		public static TypeKind TypeFromToken(Token token) {
			switch(token.Kind) {
				case TokenKind.Int:				return TypeKind.Int;
				case TokenKind.Float:			return TypeKind.Float;
				case TokenKind.Boolean:			return TypeKind.Bool;
				case TokenKind.String:			return TypeKind.String;
			}

			throw new InvalidOperationException($"Unknown type to fetch from token '{token.Kind}'");
		}

		public static Value operator+(Value me, Value other) {
			switch(me.Kind) {
				case TypeKind.Int:		return new IntValue((int)me.Data + (int)other.Data);
			}

			throw new InvalidOperationException($"Invalid operation in Value+ {me.Kind}");
		}

		public override string ToString()
		{
			return Data.ToString();
		}
	}

	sealed class IntValue : Value {
		public IntValue(int value) : base(TypeKind.Int, value) {}
	}

	sealed class FloatValue : Value {
		public FloatValue(float value) : base(TypeKind.Float, value) {}
	}

	sealed class BoolValue : Value {
		public BoolValue(bool value) : base(TypeKind.Bool, value) {}
	}

	sealed class StringValue : Value {
		public StringValue(string value) : base(TypeKind.String, value) {}
	}

	sealed class FunctionValue : Value {
		public FunctionValue(FunctionDef value) : base(TypeKind.Function, value) {}
	}
}