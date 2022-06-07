using System;
using System.Text;
using System.Collections.Generic;

namespace TinyLang {
	enum TypeKind {
		Int, Float, Bool, String, Unit,
		Function, Struct, List, Dictionary,
		
		Unknown, Error,
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
				case TokenKind.Bool:			return TypeKind.Bool;
				case TokenKind.String:			return TypeKind.String;
			}

			throw new InvalidOperationException($"Unable to determine type from '{token.Lexeme}':{token.Kind}");
		}

		public static TypeKind TypeFromLexeme(Token token) {
			switch(token.Lexeme) {
				case "int":				return TypeKind.Int;
				case "float":			return TypeKind.Float;
				case "bool":			return TypeKind.Bool;
				case "string":			return TypeKind.String;
				case "func":			return TypeKind.Function;
			}

			throw new InvalidOperationException($"Unable to determine type from '{token.Lexeme}'");
		}

		public static Value EqualityEqual(Value me, Value other) {
			switch(me.Kind) {
				case TypeKind.Int:			return new BoolValue((int)me.Data == (int)other.Data);
				case TypeKind.Float:		return new BoolValue((float)me.Data == (float)other.Data);
				case TypeKind.Bool:			return new BoolValue((bool)me.Data == (bool)other.Data);
				case TypeKind.String:		return new BoolValue((string)me.Data == (string)other.Data);
			}

			throw new InvalidOperationException($"Invalid operation in Value== {me.Kind}");
		}

		public static Value EqualityNotEqual(Value me, Value other) {
			switch(me.Kind) {
				case TypeKind.Int:			return new BoolValue((int)me.Data != (int)other.Data);
				case TypeKind.Float:		return new BoolValue((float)me.Data != (float)other.Data);
				case TypeKind.Bool:			return new BoolValue((bool)me.Data != (bool)other.Data);
				case TypeKind.String:		return new BoolValue((string)me.Data != (string)other.Data);
			}

			throw new InvalidOperationException($"Invalid operation in Value!= {me.Kind}");
		}

		public static Value operator+(Value me, Value other) {
			switch(me.Kind) {
				case TypeKind.Int:			return new IntValue((int)me.Data + (int)other.Data);
				case TypeKind.Float:		return new FloatValue((float)me.Data + (float)other.Data);
				case TypeKind.String:		return new StringValue((string)me.Data + (string)other.Data);
			}

			throw new InvalidOperationException($"Invalid operation in Value+ {me.Kind}");
		}

		public static Value operator-(Value me, Value other) {
			switch(me.Kind) {
				case TypeKind.Int:			return new IntValue((int)me.Data - (int)other.Data);
				case TypeKind.Float:		return new FloatValue((float)me.Data - (float)other.Data);
			}

			throw new InvalidOperationException($"Invalid operation in Value- {me.Kind}");
		}

		public static Value operator*(Value me, Value other) {
			switch(me.Kind) {
				case TypeKind.Int:			return new IntValue((int)me.Data * (int)other.Data);
				case TypeKind.Float:		return new FloatValue((float)me.Data * (float)other.Data);
			}

			throw new InvalidOperationException($"Invalid operation in Value* {me.Kind}");
		}

		public static Value operator/(Value me, Value other) {
			switch(me.Kind) {
				case TypeKind.Int:			return new IntValue((int)me.Data / (int)other.Data);
				case TypeKind.Float:		return new FloatValue((float)me.Data / (float)other.Data);
			}

			throw new InvalidOperationException($"Invalid operation in Value/ {me.Kind}");
		}

		public override string ToString()
		{
			return Data.ToString();
		}
	}

	sealed class UnitValue : Value {
		public UnitValue() : base(TypeKind.Unit, (byte)0) {}
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

		public override string ToString()
		{
			FunctionDef def = (FunctionDef)Data;

			StringBuilder sb = new StringBuilder();
			for(int i = 0; i < def.parameters.Count; i++) {
				sb.Append(def.parameters[i].token.Lexeme);

				if (i < def.parameters.Count - 1) {
					sb.Append(", ");
				}
			}

			return $"{def.identifier}({sb.ToString()})";
		}
	}
}