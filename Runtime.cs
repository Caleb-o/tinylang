using System;
using System.Collections.Generic;

namespace TinyLang {
	enum TypeKind {
		Int, Float, Bool, String, Function, Struct,
		List, Dictionary,
	}

	abstract class Value {
		public readonly TypeKind Kind;
		public readonly object Data;

		public Value(TypeKind kind, object data) {
			this.Kind = kind;
			this.Data = data;
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
		public IntValue(int value) : base(TypeKind.Int, (int)value) {}
	}

	sealed class FloatValue : Value {
		public FloatValue(float value) : base(TypeKind.Float, (float)value) {}
	}

	sealed class BoolValue : Value {
		public BoolValue(bool value) : base(TypeKind.Bool, (bool)value) {}
	}

	sealed class StringValue : Value {
		public StringValue(string value) : base(TypeKind.String, (string)value) {}
	}
}