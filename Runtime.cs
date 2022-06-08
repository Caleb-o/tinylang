using System;
using System.Linq;
using System.Text;
using System.Collections.Generic;

namespace TinyLang {
	sealed class ReturnException : Exception {
		public ReturnException() {}
	}

	sealed class AssertException : Exception {
		public AssertException(string message) : base(message) {}
	}

	abstract class Value {
		public readonly TinyType Kind;
		public readonly object Data;

		public Value(TinyType kind, object data) {
			this.Kind = kind;
			this.Data = data;
		}

		public static Value DefaultFrom(TinyType kind) {
			switch(kind) {
				case TinyInt:			return new IntValue(0);
				case TinyFloat:			return new FloatValue(0.0f);
				case TinyBool:			return new BoolValue(false);
				case TinyString:		return new StringValue("");
				
				default:
					return new UnitValue();
			}
		}

		public static Value EqualityEqual(Value me, Value other) {
			switch(me.Kind) {
				case TinyInt:			return new BoolValue((int)me.Data == (int)other.Data);
				case TinyFloat:			return new BoolValue((float)me.Data == (float)other.Data);
				case TinyBool:			return new BoolValue((bool)me.Data == (bool)other.Data);
				case TinyString:		return new BoolValue((string)me.Data == (string)other.Data);
			}

			throw new InvalidOperationException($"Invalid operation in Value== {me.Kind}");
		}

		public static Value EqualityNotEqual(Value me, Value other) {
			switch(me.Kind) {
				case TinyInt:			return new BoolValue((int)me.Data != (int)other.Data);
				case TinyFloat:			return new BoolValue((float)me.Data != (float)other.Data);
				case TinyBool:			return new BoolValue((bool)me.Data != (bool)other.Data);
				case TinyString:		return new BoolValue((string)me.Data != (string)other.Data);
			}

			throw new InvalidOperationException($"Invalid operation in Value!= {me.Kind}");
		}

		// Negate
		public static Value operator-(Value me) {
			switch(me.Kind) {
				case TinyInt:			return new IntValue(-(int)me.Data);
				case TinyFloat:			return new FloatValue(-(float)me.Data);
				case TinyBool:			return new BoolValue(!(bool)me.Data);
			}

			throw new InvalidOperationException($"Invalid operation in -Value {me.Kind}");
		}

		public static Value operator+(Value me, Value other) {
			switch(me.Kind) {
				case TinyInt:			return new IntValue((int)me.Data + (int)other.Data);
				case TinyFloat:			return new FloatValue((float)me.Data + (float)other.Data);
				case TinyString:		return new StringValue((string)me.Data + (string)other.Data);

				case TinyList: {
					List<Value> from_x = (List<Value>)((ListValue)me).Data;
					List<Value> from_y = (List<Value>)((ListValue)other).Data;

					return new ListValue(((ListValue)me).Kind, from_x.Concat(from_y).ToList());
				}
			}

			throw new InvalidOperationException($"Invalid operation in Value+ {me.Kind}");
		}

		public static Value operator-(Value me, Value other) {
			switch(me.Kind) {
				case TinyInt:			return new IntValue((int)me.Data - (int)other.Data);
				case TinyFloat:			return new FloatValue((float)me.Data - (float)other.Data);
			}

			throw new InvalidOperationException($"Invalid operation in Value- {me.Kind}");
		}

		public static Value operator*(Value me, Value other) {
			switch(me.Kind) {
				case TinyInt:			return new IntValue((int)me.Data * (int)other.Data);
				case TinyFloat:			return new FloatValue((float)me.Data * (float)other.Data);
			}

			throw new InvalidOperationException($"Invalid operation in Value* {me.Kind}");
		}

		public static Value operator/(Value me, Value other) {
			switch(me.Kind) {
				case TinyInt:			return new IntValue((int)me.Data / (int)other.Data);
				case TinyFloat:			return new FloatValue((float)me.Data / (float)other.Data);
			}

			throw new InvalidOperationException($"Invalid operation in Value/ {me.Kind}");
		}

		public static Value operator>(Value me, Value other) {
			switch(me.Kind) {
				case TinyInt:			return new BoolValue((int)me.Data > (int)other.Data);
				case TinyFloat:			return new BoolValue((float)me.Data > (float)other.Data);
				case TinyString:		return new BoolValue(((string)me.Data).Length > ((string)other.Data).Length);
			}

			throw new InvalidOperationException($"Invalid operation in Value> {me.Kind}");
		}

		public static Value operator<(Value me, Value other) {
			switch(me.Kind) {
				case TinyInt:			return new BoolValue((int)me.Data < (int)other.Data);
				case TinyFloat:			return new BoolValue((float)me.Data < (float)other.Data);
				case TinyString:		return new BoolValue(((string)me.Data).Length < ((string)other.Data).Length);
			}

			throw new InvalidOperationException($"Invalid operation in Value> {me.Kind}");
		}

		public static Value operator>=(Value me, Value other) {
			switch(me.Kind) {
				case TinyInt:			return new BoolValue((int)me.Data >= (int)other.Data);
				case TinyFloat:			return new BoolValue((float)me.Data >= (float)other.Data);
				case TinyString:		return new BoolValue(((string)me.Data).Length >= ((string)other.Data).Length);
			}

			throw new InvalidOperationException($"Invalid operation in Value> {me.Kind}");
		}

		public static Value operator<=(Value me, Value other) {
			switch(me.Kind) {
				case TinyInt:			return new BoolValue((int)me.Data <= (int)other.Data);
				case TinyFloat:			return new BoolValue((float)me.Data <= (float)other.Data);
				case TinyString:		return new BoolValue(((string)me.Data).Length <= ((string)other.Data).Length);
			}

			throw new InvalidOperationException($"Invalid operation in Value> {me.Kind}");
		}
	}

	sealed class UnitValue : Value {
		public UnitValue() : base(new TinyUnit(), (byte)0) {}

		public override string ToString() => "()";
	}

	sealed class IntValue : Value {
		public IntValue(int value) : base(new TinyInt(), value) {}

		public override string ToString() => ((int)Data).ToString();
	}

	sealed class FloatValue : Value {
		public FloatValue(float value) : base(new TinyFloat(), value) {}

		public override string ToString() => ((float)Data).ToString();
	}

	sealed class BoolValue : Value {
		public BoolValue(bool value) : base(new TinyBool(), value) {}

		public override string ToString() => ((bool)Data).ToString();
	}

	sealed class StringValue : Value {
		public StringValue(string value) : base(new TinyString(), value) {}

		public override string ToString() => (string)Data;
	}

	sealed class ListValue : Value {
		public ListValue(TinyType kind, List<Value> values) : base(kind, values) {}

		public override string ToString()
		{
			List<Value> values = (List<Value>)Data;

			StringBuilder sb = new StringBuilder();
			for(int i = 0; i < values.Count; i++) {
				sb.Append(values[i]);

				if (i < values.Count - 1) {
					sb.Append(", ");
				}
			}

			return $"[{sb.ToString()}]";
		}
	}

	sealed class FunctionValue : Value {
		public FunctionValue(FunctionDef value) : base(new TinyFunction(value.identifier), value) {}

		public override string ToString()
		{
			FunctionDef def = (FunctionDef)Data;

			StringBuilder sb = new StringBuilder();

			if (def.block is BuiltinFn) {
				sb.Append("[builtin] ");
			}

			sb.Append($"{def.identifier} (");

			for(int i = 0; i < def.parameters.Length; i++) {
				sb.Append(def.parameters[i].identifier);

				if (i < def.parameters.Length - 1) {
					sb.Append(", ");
				}
			}
			sb.Append(")");

			return $"{sb.ToString()}";
		}
	}

	sealed class StructValue : Value {
		public StructValue(StructDef def, List<Value> values) : base(new TinyStruct(def), values) {}

		public override string ToString()
		{
			StructDef def = ((TinyStruct)Kind).def;
			int idx = 0;

			StringBuilder sb = new StringBuilder(def.identifier);
			sb.Append(" { ");

			foreach(var ((id, kind), val) in def.fields.Zip((List<Value>)Data)) {
				sb.Append($"{id}:{kind} = {val}");

				if (idx < def.fields.Count - 1) {
					sb.Append(", ");
				}
				idx++;
			}

			sb.Append(" }");
			return sb.ToString();
		}
	}
}