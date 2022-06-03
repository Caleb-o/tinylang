using System;
using System.Collections.Generic;

namespace TinyLang {
	// Dummy exception for unwinding the stack
	public class EscapeException : Exception
	{
		public EscapeException() {}
	}

	enum RecordType {
		Program, Function,
	}

	/*
		TODO: Transpile to Python or C++ instead of interpreting
		C++ might be a good primer for the main language
	*/
	enum ValueKind {
		Int, Float, String, Bool, Tuple,
	}

	class Value {
		public readonly ValueKind kind;
		public object value;
		public Value references = null;

		public Value(ValueKind kind, object value) {
			this.kind = kind;
			this.value = value;
		}

		// Since there is type checking, it should be possible to operate on different
		// value types at runtime
		public static Value operator-(Value me) {
			switch(me.kind) {
				case ValueKind.Int: 	return new Value(ValueKind.Int, -(int)me.value);
				case ValueKind.Float: 	return new Value(ValueKind.Float, -(float)me.value);
				case ValueKind.Bool: 	return new Value(ValueKind.Bool, !(bool)me.value);

				case ValueKind.String: 	throw new InvalidOperationException("Cannot use unary negation on strings");
				
				// This should be unreachable
				default: throw new InvalidOperationException("Unknown value type in arithmetic operation");
			}
		}

		public static Value operator==(Value me, Value other) {
			switch(me.kind) {
				case ValueKind.Int: 	return new Value(ValueKind.Int, (int)me.value == (int)other.value);
				case ValueKind.Float: 	return new Value(ValueKind.Float, (float)me.value == (float)other.value);
				case ValueKind.Bool: 	return new Value(ValueKind.Bool, (bool)me.value == (bool)other.value);
				case ValueKind.String: 	return new Value(ValueKind.String, (string)me.value == (string)other.value);
				
				// This should be unreachable
				default: throw new InvalidOperationException("Unknown value type in arithmetic operation");
			}
		}

		public static Value operator!=(Value me, Value other) {
			switch(me.kind) {
				case ValueKind.Int: 	return new Value(ValueKind.Int, (int)me.value != (int)other.value);
				case ValueKind.Float: 	return new Value(ValueKind.Float, (float)me.value != (float)other.value);
				case ValueKind.Bool: 	return new Value(ValueKind.Bool, (bool)me.value != (bool)other.value);
				case ValueKind.String: 	return new Value(ValueKind.String, (string)me.value != (string)other.value);
				
				// This should be unreachable
				default: throw new InvalidOperationException("Unknown value type in arithmetic operation");
			}
		}

		public static Value operator>(Value me, Value other) {
			switch(me.kind) {
				case ValueKind.Int: 	return new Value(ValueKind.Int, (int)me.value > (int)other.value);
				case ValueKind.Float: 	return new Value(ValueKind.Float, (float)me.value > (float)other.value);
				
				// This should be unreachable
				default: throw new InvalidOperationException("Unknown value type in arithmetic operation");
			}
		}

		public static Value operator>=(Value me, Value other) {
			switch(me.kind) {
				case ValueKind.Int: 	return new Value(ValueKind.Int, (int)me.value >= (int)other.value);
				case ValueKind.Float: 	return new Value(ValueKind.Float, (float)me.value >= (float)other.value);
				
				// This should be unreachable
				default: throw new InvalidOperationException("Unknown value type in arithmetic operation");
			}
		}

		public static Value operator<(Value me, Value other) {
			switch(me.kind) {
				case ValueKind.Int: 	return new Value(ValueKind.Int, (int)me.value < (int)other.value);
				case ValueKind.Float: 	return new Value(ValueKind.Float, (float)me.value < (float)other.value);
				
				// This should be unreachable
				default: throw new InvalidOperationException("Unknown value type in arithmetic operation");
			}
		}

		public static Value operator<=(Value me, Value other) {
			switch(me.kind) {
				case ValueKind.Int: 	return new Value(ValueKind.Int, (int)me.value <= (int)other.value);
				case ValueKind.Float: 	return new Value(ValueKind.Float, (float)me.value <= (float)other.value);
				
				// This should be unreachable
				default: throw new InvalidOperationException("Unknown value type in arithmetic operation");
			}
		}

		public static Value operator+(Value me, Value other) {
			switch(me.kind) {
				case ValueKind.Int: 	return new Value(ValueKind.Int, (int)me.value + (int)other.value);
				case ValueKind.Float: 	return new Value(ValueKind.Float, (float)me.value + (float)other.value);
				case ValueKind.Bool: 	return other;
				case ValueKind.String: 	return new Value(ValueKind.String, (string)me.value + (string)other.value);
				
				// This should be unreachable
				default: throw new InvalidOperationException("Unknown value type in arithmetic operation");
			}
		}

		public static Value operator-(Value me, Value other) {
			switch(me.kind) {
				case ValueKind.Int: 	return new Value(ValueKind.Int, (int)me.value - (int)other.value);
				case ValueKind.Float: 	return new Value(ValueKind.Float, (float)me.value - (float)other.value);
				case ValueKind.Bool: 	return other;
				case ValueKind.String: 	throw new InvalidOperationException("Cannot use minus on strings");
				
				// This should be unreachable
				default: throw new InvalidOperationException("Unknown value type in arithmetic operation");
			}
		}

		public static Value operator*(Value me, Value other) {
			switch(me.kind) {
				case ValueKind.Int: 	return new Value(ValueKind.Int, (int)me.value * (int)other.value);
				case ValueKind.Float: 	return new Value(ValueKind.Float, (float)me.value * (float)other.value);
				
				case ValueKind.Bool:
				case ValueKind.String: 	throw new InvalidOperationException($"Cannot use multiply on {me.kind}");
				
				// This should be unreachable
				default: throw new InvalidOperationException("Unknown value type in arithmetic operation");
			}
		}

		public static Value operator/(Value me, Value other) {
			switch(me.kind) {
				case ValueKind.Int: 	return new Value(ValueKind.Int, (int)me.value / (int)other.value);
				case ValueKind.Float: 	return new Value(ValueKind.Float, (float)me.value / (float)other.value);
				
				case ValueKind.Bool:
				case ValueKind.String: 	throw new InvalidOperationException($"Cannot use divide on {me.kind}");
				
				// This should be unreachable
				default: throw new InvalidOperationException("Unknown value type in arithmetic operation");
			}
		}
	}

	class ActivationRecord {
		public readonly string identifier; 
		public readonly RecordType type;
		public readonly int scopeLevel;
		public Dictionary<string, VarSym> members = new Dictionary<string, VarSym>();

		public ActivationRecord(string identifier, RecordType type, int scopeLevel) {
			this.identifier = identifier;
			this.type = type;
			this.scopeLevel = scopeLevel;
		}
	}

	class CallStack {
		public List<ActivationRecord> stack = new List<ActivationRecord>();
	}
}