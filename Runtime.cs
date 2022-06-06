using System;
using System.Collections.Generic;

namespace TinyLang {
	// Dummy exception for unwinding the stack
	public class EscapeException : Exception
	{
		public EscapeException() {}
	}

	public class AssertionException : Exception
	{
		public AssertionException(string message) : base(message) {}
	}

	enum RecordType {
		Program, Function,
	}

	enum TypeKind {
		Int, Float, Bool, String, Tuple, List, Untyped,
	}

	sealed class Type {
		public readonly TypeKind Kind;
		// For lists and tuples
		public TypeKind[] SubKind { get; private set; }

		public Type(TypeKind kind) {
			this.Kind = kind;
		}

		public Type(TypeKind kind, TypeKind subKind) {
			this.Kind = kind;
			this.SubKind = new TypeKind[] { subKind };
		}

		public Type(TypeKind kind, TypeKind[] subKind) {
			this.Kind = kind;
			this.SubKind = subKind;
		}

		public bool Matches(Type other) {
			if (Kind != other.Kind) {
				return false;
			}

			if (SubKind == null && other.SubKind == null) {
				return true;
			}

			if (SubKind.Length != other.SubKind.Length) {
				return false;
			}

			for(int i = 0; i < SubKind.Length; i++) {
				if (SubKind[i] != other.SubKind[i]) {
					return false;
				}
			}

			return true;
		}

		public void SetSubKind(TypeKind[] subKind) {
			this.SubKind = subKind;
		}

		public override string ToString()
		{
			string subStr = "";

			if (SubKind != null) {
				for(int i = 0; i < SubKind.Length; i++) {
					subStr += SubKind[i];

					if (i < SubKind.Length - 1) {
						subStr += ", ";
					}
				}
			}

			if (Kind == TypeKind.List) {
				return $"[{subStr}]";
			}
			else if (Kind == TypeKind.Tuple) {
				return $"({subStr})";
			} else {
				return Kind.ToString();
			}
		}
	}

	abstract class Value {
		public object Data { get; protected set; }
		public Type Kind { get; protected set; }

		public void SetSubKind(TypeKind[] subkind) {
			this.Kind.SetSubKind(subkind);
		}

		public static Value EqualityEqual(Value me, Value other) {
			switch(me.Kind.Kind) {
				case TypeKind.Int:			return new BoolValue((int)me.Data == (int)other.Data);
				case TypeKind.Float:		return new BoolValue((float)me.Data == (float)other.Data);
				case TypeKind.Bool:			return new BoolValue((bool)me.Data == (bool)other.Data);
				case TypeKind.String:		return new BoolValue((string)me.Data == (string)other.Data);

				case TypeKind.List:
				case TypeKind.Tuple: {
					List<Value> meValues = (List<Value>)me.Data;
					List<Value> otherValues = (List<Value>)other.Data;

					if (meValues.Count != otherValues.Count) {
						return new BoolValue(false);
					}

					for(int i = 0; i < meValues.Count; i++) {
						if ((bool)Value.EqualityNotEqual(meValues[i], otherValues[i]).Data) {
							return new BoolValue(false);
						}
					}
					
					return new BoolValue(true);
				}
			}

			throw new InvalidOperationException("Unknown value type in arithmetic operation");
		}

		public static Value EqualityNotEqual(Value me, Value other) {
			switch(me.Kind.Kind) {
				case TypeKind.Int:			return new BoolValue((int)me.Data != (int)other.Data);
				case TypeKind.Float:		return new BoolValue((float)me.Data != (float)other.Data);
				case TypeKind.Bool:			return new BoolValue((bool)me.Data != (bool)other.Data);
				case TypeKind.String:		return new BoolValue((string)me.Data != (string)other.Data);

				case TypeKind.List:
				case TypeKind.Tuple: {
					List<Value> meValues = (List<Value>)me.Data;
					List<Value> otherValues = (List<Value>)other.Data;

					if (meValues.Count == otherValues.Count) {
						return new BoolValue(false);
					}

					for(int i = 0; i < meValues.Count; i++) {
						if ((bool)Value.EqualityEqual(meValues[i], otherValues[i]).Data) {
							return new BoolValue(false);
						}
					}
					
					return new BoolValue(true);
				}
			}

			throw new InvalidOperationException("Unknown value type in arithmetic operation");
		}

		public static Value operator-(Value me) {
			switch(me.Kind.Kind) {
				case TypeKind.Int:			return new IntValue(-(int)me.Data);
				case TypeKind.Float:		return new FloatValue(-(float)me.Data);
				case TypeKind.Bool:			return new BoolValue(!(bool)me.Data);
			}

			throw new InvalidOperationException("Unknown value type in arithmetic operation");
		}

		public static Value operator+(Value me, Value other) {
			switch(me.Kind.Kind) {
				case TypeKind.Int:			return new IntValue((int)me.Data + (int)other.Data);
				case TypeKind.Float:		return new FloatValue((float)me.Data + (float)other.Data);
				case TypeKind.String:		return new StringValue((string)me.Data + (string)other.Data);

				case TypeKind.List: {
					List<Value> meValues = (List<Value>)me.Data;
					List<Value> otherValues = (List<Value>)other.Data;

					List<Value> combined = new List<Value>(meValues);
					combined.AddRange(otherValues);

					ListValue value = new ListValue(combined);
					value.SetSubKind(new TypeKind[1] { me.Kind.SubKind[0] });
					return value;
				}
			}

			throw new InvalidOperationException("Unknown value type in arithmetic operation");
		}

		public static Value operator-(Value me, Value other) {
			switch(me.Kind.Kind) {
				case TypeKind.Int:			return new IntValue((int)me.Data - (int)other.Data);
				case TypeKind.Float:		return new FloatValue((float)me.Data - (float)other.Data);
			}

			throw new InvalidOperationException("Unknown value type in arithmetic operation");
		}

		public static Value operator*(Value me, Value other) {
			switch(me.Kind.Kind) {
				case TypeKind.Int:			return new IntValue((int)me.Data * (int)other.Data);
				case TypeKind.Float:		return new FloatValue((float)me.Data * (float)other.Data);
			}

			throw new InvalidOperationException("Unknown value type in arithmetic operation");
		}

		public static Value operator/(Value me, Value other) {
			switch(me.Kind.Kind) {
				case TypeKind.Int:			return new IntValue((int)me.Data / (int)other.Data);
				case TypeKind.Float:		return new FloatValue((float)me.Data / (float)other.Data);
			}

			throw new InvalidOperationException("Unknown value type in arithmetic operation");
		}

		public static Value operator>(Value me, Value other) {
			switch(me.Kind.Kind) {
				case TypeKind.Int:			return new BoolValue((int)me.Data > (int)other.Data);
				case TypeKind.Float:		return new BoolValue((float)me.Data > (float)other.Data);
			}

			throw new InvalidOperationException("Unknown value type in arithmetic operation");
		}

		public static Value operator<(Value me, Value other) {
			switch(me.Kind.Kind) {
				case TypeKind.Int:			return new BoolValue((int)me.Data < (int)other.Data);
				case TypeKind.Float:		return new BoolValue((float)me.Data < (float)other.Data);
			}

			throw new InvalidOperationException("Unknown value type in arithmetic operation");
		}

		public static Value operator>=(Value me, Value other) {
			switch(me.Kind.Kind) {
				case TypeKind.Int:			return new BoolValue((int)me.Data >= (int)other.Data);
				case TypeKind.Float:		return new BoolValue((float)me.Data >= (float)other.Data);
			}

			throw new InvalidOperationException("Unknown value type in arithmetic operation");
		}

		public static Value operator<=(Value me, Value other) {
			switch(me.Kind.Kind) {
				case TypeKind.Int:			return new BoolValue((int)me.Data <= (int)other.Data);
				case TypeKind.Float:		return new BoolValue((float)me.Data <= (float)other.Data);
			}

			throw new InvalidOperationException("Unknown value type in arithmetic operation");
		}

		public static bool IsType(string typeName) {
			switch(typeName) {
				case "int":
				case "string":
				case "boolean":
				case "float":
					return true;
			}

			return false;
		}

		public static TypeKind TypeFromToken(TokenKind kind) {
			switch(kind) {
				case TokenKind.Int:			return TypeKind.Int;
				case TokenKind.Float:		return TypeKind.Float;
				case TokenKind.Boolean:		return TypeKind.Bool;
				case TokenKind.String:		return TypeKind.String;
			}
			
			return TypeKind.Untyped;
		}

		public static TypeKind TypeFromStr(string typeName) {
			switch(typeName) {
				case "int":			return TypeKind.Int;
				case "float":		return TypeKind.Float;
				case "boolean":		return TypeKind.Bool;
				case "string":		return TypeKind.String;
			}
			
			return TypeKind.Untyped;
		}
	}

	sealed class IntValue : Value {
		public IntValue(int value) {
			this.Data = value;
			this.Kind = new Type(TypeKind.Int);			
		}

		public override string ToString()
		{
			return ((int)Data).ToString();
		}
	}

	sealed class FloatValue : Value {
		public FloatValue(float value) {
			this.Data = value;
			this.Kind = new Type(TypeKind.Float);	
		}

		public override string ToString()
		{
			return ((float)Data).ToString();
		}
	}

	sealed class BoolValue : Value {
		public BoolValue(bool value) {
			this.Data = value;
			this.Kind = new Type(TypeKind.Bool);			
		}

		public override string ToString()
		{
			return ((bool)Data).ToString();
		}
	}

	sealed class StringValue : Value {
		public StringValue(string value) {
			this.Data = value;
			this.Kind = new Type(TypeKind.String);			
		}

		public override string ToString()
		{
			return (string)Data;
		}
	}

	sealed class ListValue : Value {
		public ListValue(List<Value> value) {
			this.Data = value;
			this.Kind = new Type(TypeKind.List);			
		}

		public override string ToString()
		{
			string outStr = "";
			List<Value> values = (List<Value>)Data;

			int i = 0;
			foreach(Value value in (List<Value>)Data) {
				outStr += value;

				if (i++ < values.Count - 1) {
					outStr += ", ";
				}
			}

			return $"[{outStr}]";
		}
	}

	sealed class TupleValue : Value {
		public TupleValue(List<Value> value) {
			this.Data = value;
			this.Kind = new Type(TypeKind.Tuple);			
		}

		public override string ToString()
		{
			string outStr = "";
			List<Value> values = (List<Value>)Data;

			int i = 0;
			foreach(Value value in values) {
				outStr += value;
				
				if (i++ < values.Count - 1) {
					outStr += ", ";
				}
			}

			return $"({outStr})";
		}
	}

	class ActivationRecord {
		public readonly string identifier; 
		public readonly RecordType type;
		public readonly int depth;
		public Dictionary<string, VarSym> members = new Dictionary<string, VarSym>();

		public ActivationRecord(string identifier, RecordType type, int depth) {
			this.identifier = identifier;
			this.type = type;
			this.depth = depth;
		}
	}

	class CallStack {
		public List<ActivationRecord> stack = new List<ActivationRecord>();
	}
}