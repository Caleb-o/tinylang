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

	enum TypeKind {
		Int, Float, Bool, String, Tuple,
	}

	class Type {
		// Type IDs will be used for tuples etc
		public readonly int[] typeIDs;

		public Type() {} // For null type

		public Type(TypeKind kind) {
			this.typeIDs = new int[] { (int)kind };
		}

		public Type(int typeID) {
			this.typeIDs = new int[] { typeID };
		}

		public Type(int[] typeIDs) {
			this.typeIDs = typeIDs;
		}

		public bool IsSingleType() {
			return typeIDs == null;
		}

		public TypeKind GetKind() {
			if (typeIDs.Length > 1) {
				return TypeKind.Tuple;
			}

			switch(typeIDs[0]) {
				case (int)TypeKind.Int: 	return TypeKind.Int;
				case (int)TypeKind.Float: 	return TypeKind.Float;
				case (int)TypeKind.Bool: 	return TypeKind.Bool;
				case (int)TypeKind.String: 	return TypeKind.String;
			}

			throw new InvalidOperationException($"Unknown kind provided '{typeIDs[0]}'");
		}

		public static TypeKind FromToken(TokenKind kind) {
			switch(kind) {
				case TokenKind.Int: 		return TypeKind.Int;
				case TokenKind.Float: 		return TypeKind.Float;
				case TokenKind.Boolean: 	return TypeKind.Bool;
				case TokenKind.String: 		return TypeKind.String;
			}

			throw new InvalidOperationException($"Invalid token kind to type kind '{kind}' -> ?");
		}

		public bool Matches(Type other) {
			if (typeIDs == null || other.typeIDs == null) {
				return false;
			}

			for(int idx = 0; idx < typeIDs.Length; idx++) {
				if (typeIDs[idx] != other.typeIDs[idx]) {
					return false;
				}
			}

			return true;
		}

		public override string ToString() {
			string outstr = "";

			int idx = 0;
			foreach(int id in typeIDs) {
				outstr += Application.GetTypeName(id);

				if (idx < typeIDs.Length - 1) {
					outstr += ", ";
				}
				idx++;
			}

			if (typeIDs.Length > 1) {
				return $"({outstr})";
			} else {
				return outstr;
			}
		}
	}

	class Value {
		public readonly Type type;
		public object value;
		public Value references = null;

		public Value(Type type, object value) {
			this.type = type;
			this.value = value;
		}

		// Since there is type checking, it should be possible to operate on different
		// value types at runtime
		public static Value operator-(Value me) {
			switch(me.type.GetKind()) {
				case TypeKind.Int: 		return new Value(me.type, -(int)me.value);
				case TypeKind.Float: 	return new Value(me.type, -(float)me.value);
				case TypeKind.Bool: 	return new Value(me.type, !(bool)me.value);

				case TypeKind.String: 	throw new InvalidOperationException("Cannot use unary negation on strings");
				
				// This should be unreachable
				default: throw new InvalidOperationException("Unknown value type in arithmetic operation");
			}
		}

		public static Value operator==(Value me, Value other) {
			switch(me.type.GetKind()) {
				case TypeKind.Int: 		return new Value(me.type, (int)me.value == (int)other.value);
				case TypeKind.Float: 	return new Value(me.type, (float)me.value == (float)other.value);
				case TypeKind.Bool: 	return new Value(me.type, (bool)me.value == (bool)other.value);
				case TypeKind.String: 	return new Value(me.type, (string)me.value == (string)other.value);
				
				// This should be unreachable
				default: throw new InvalidOperationException("Unknown value type in arithmetic operation");
			}
		}

		public static Value operator!=(Value me, Value other) {
			switch(me.type.GetKind()) {
				case TypeKind.Int: 		return new Value(me.type, (int)me.value != (int)other.value);
				case TypeKind.Float: 	return new Value(me.type, (float)me.value != (float)other.value);
				case TypeKind.Bool: 	return new Value(me.type, (bool)me.value != (bool)other.value);
				case TypeKind.String: 	return new Value(me.type, (string)me.value != (string)other.value);
				
				// This should be unreachable
				default: throw new InvalidOperationException("Unknown value type in arithmetic operation");
			}
		}

		public static Value operator>(Value me, Value other) {
			switch(me.type.GetKind()) {
				case TypeKind.Int: 		return new Value(me.type, (int)me.value > (int)other.value);
				case TypeKind.Float: 	return new Value(me.type, (float)me.value > (float)other.value);
				
				// This should be unreachable
				default: throw new InvalidOperationException("Unknown value type in arithmetic operation");
			}
		}

		public static Value operator>=(Value me, Value other) {
			switch(me.type.GetKind()) {
				case TypeKind.Int: 		return new Value(me.type, (int)me.value >= (int)other.value);
				case TypeKind.Float: 	return new Value(me.type, (float)me.value >= (float)other.value);
				
				// This should be unreachable
				default: throw new InvalidOperationException("Unknown value type in arithmetic operation");
			}
		}

		public static Value operator<(Value me, Value other) {
			switch(me.type.GetKind()) {
				case TypeKind.Int: 		return new Value(me.type, (int)me.value < (int)other.value);
				case TypeKind.Float: 	return new Value(me.type, (float)me.value < (float)other.value);
				
				// This should be unreachable
				default: throw new InvalidOperationException("Unknown value type in arithmetic operation");
			}
		}

		public static Value operator<=(Value me, Value other) {
			switch(me.type.GetKind()) {
				case TypeKind.Int: 		return new Value(me.type, (int)me.value <= (int)other.value);
				case TypeKind.Float: 	return new Value(me.type, (float)me.value <= (float)other.value);
				
				// This should be unreachable
				default: throw new InvalidOperationException("Unknown value type in arithmetic operation");
			}
		}

		public static Value operator+(Value me, Value other) {
			switch(me.type.GetKind()) {
				case TypeKind.Int: 		return new Value(me.type, (int)me.value + (int)other.value);
				case TypeKind.Float: 	return new Value(me.type, (float)me.value + (float)other.value);
				case TypeKind.Bool: 	return other;
				case TypeKind.String: 	return new Value(me.type, (string)me.value + (string)other.value);
				
				// This should be unreachable
				default: throw new InvalidOperationException("Unknown value type in arithmetic operation");
			}
		}

		public static Value operator-(Value me, Value other) {
			switch(me.type.GetKind()) {
				case TypeKind.Int: 		return new Value(me.type, (int)me.value - (int)other.value);
				case TypeKind.Float: 	return new Value(me.type, (float)me.value - (float)other.value);
				case TypeKind.Bool: 	return other;
				case TypeKind.String: 	throw new InvalidOperationException("Cannot use minus on strings");
				
				// This should be unreachable
				default: throw new InvalidOperationException("Unknown value type in arithmetic operation");
			}
		}

		public static Value operator*(Value me, Value other) {
			switch(me.type.GetKind()) {
				case TypeKind.Int: 		return new Value(me.type, (int)me.value * (int)other.value);
				case TypeKind.Float: 	return new Value(me.type, (float)me.value * (float)other.value);
				
				case TypeKind.Bool:
				case TypeKind.String: 	throw new InvalidOperationException($"Cannot use multiply on {me.type.GetKind()}");
				
				// This should be unreachable
				default: throw new InvalidOperationException("Unknown value type in arithmetic operation");
			}
		}

		public static Value operator/(Value me, Value other) {
			switch(me.type.GetKind()) {
				case TypeKind.Int: 		return new Value(me.type, (int)me.value / (int)other.value);
				case TypeKind.Float: 	return new Value(me.type, (float)me.value / (float)other.value);
				
				case TypeKind.Bool:
				case TypeKind.String: 	throw new InvalidOperationException($"Cannot use divide on {me.type.GetKind()}");
				
				// This should be unreachable
				default: throw new InvalidOperationException("Unknown value type in arithmetic operation");
			}
		}

		public override string ToString()
		{
			switch (type.GetKind()) {
				case TypeKind.Tuple: {
					string outstr = "";

					List<Value> values = (List<Value>)value;

					int index = 0;
					foreach(Value value in values) {
						outstr += value;

						if (index++ < values.Count - 1) {
							outstr += ", ";
						}
					}

					return $"({outstr})";
				}

				default:
					return value.ToString();
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