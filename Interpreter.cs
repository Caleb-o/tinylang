using System;
using System.Linq;
using System.Collections.Generic;

namespace TinyLang {
	enum RecordType {
		Program, Function,
	}

	/*
		TODO: Transpile to Python or C++ instead of interpreting
		C++ might be a good primer for the main language
	*/
	enum ValueKind {
		Int, Float, String, Bool,
	}

	class Value {
		public readonly ValueKind kind;
		public readonly object value;
		// Which identifier it is referencing
		public string references = null;

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
		public Dictionary<string, Value> members = new Dictionary<string, Value>();

		public ActivationRecord(string identifier, RecordType type, int scopeLevel) {
			this.identifier = identifier;
			this.type = type;
			this.scopeLevel = scopeLevel;
		}
	}

	class CallStack {
		public List<ActivationRecord> stack = new List<ActivationRecord>();
	}

	class Interpreter {
		CallStack callStack;

		public Interpreter() {
			callStack = new CallStack();
		}

		public void Run(Application app) {
			callStack.stack.Add(new ActivationRecord(
				"global", RecordType.Program, 0
			));

			Visit(app.block);
			callStack.stack.Remove(callStack.stack[^1]);
		}

		Value DefaultValue(Return ret) {
			// TODO: If return is a record/struct, then return a Visit on a mandatory
			// constructor for the type
			switch(ret.type) {
				case "int": return new Value(ValueKind.Int, 0);
				case "float": return new Value(ValueKind.Float, 0);
				case "boolean": return new Value(ValueKind.Bool, false);
				case "string": return new Value(ValueKind.Bool, "");
			}

			throw new InvalidCastException("Unreachable");
		}

		void Error(string msg) {
			Console.WriteLine("-- Call Stack --");
			for(int i = callStack.stack.Count - 1; i >= 0; i--) {
				Console.WriteLine($"[{i}] {callStack.stack[i].identifier}");

				foreach(var record in callStack.stack[i].members) {
					Console.WriteLine($"{record.Key.PadLeft(16)} = {record.Value.value} [{record.Value.kind}]");
				}
			}
			Console.WriteLine();

			throw new InvalidOperationException($"Runtime: {msg} [{callStack.stack[^1].identifier}]");
		}

		ActivationRecord ResolveVar(string identifier, int offset = 0) {
			for(int i = callStack.stack.Count - (offset + 1); i >= 0; i--) {
				if (callStack.stack[i].members.ContainsKey(identifier)) {
					return callStack.stack[i];
				}
			}
			throw new InvalidOperationException($"Unreachable: unable to find variable '{identifier}'");
		}

		Value Visit(Node node) {
			switch(node) {
				case Block: VisitBlock((Block)node); return null;
				case Literal: return VisitLiteral((Literal)node);
				case BinOp: return VisitBinOp((BinOp)node);
				case UnaryOp: return VisitUnaryOp((UnaryOp)node);

				case FunctionDef: return null;
				case FunctionCall: return VisitFunctionCall((FunctionCall)node);
				case BuiltinFunctionCall: VisitBuiltinFunctionCall((BuiltinFunctionCall)node); return null;
				// FIXME: We need to somehow signal a return from the function
				case Escape: return null;

				case Var: return VisitVar((Var)node);
				case Assignment: VisitAssignment((Assignment)node); return null;
				case VarDecl: VisitVarDecl((VarDecl)node); return null;
			}

			Error($"Unknown node type {node}");
			return null;
		}

		Value VisitBinOp(BinOp binop) {
			switch(binop.token.Kind) {
				case TokenKind.Plus: 	return Visit(binop.left) + Visit(binop.right);
				case TokenKind.Minus: 	return Visit(binop.left) - Visit(binop.right);
				case TokenKind.Star: 	return Visit(binop.left) * Visit(binop.right);
				case TokenKind.Slash: 	return Visit(binop.left) / Visit(binop.right);

				default: 
					Error($"Unknown binary operation {binop.token.Kind}");
					return null;
			}
		}

		Value VisitUnaryOp(UnaryOp unary) {
			return -Visit(unary.right);
		}

		void VisitBlock(Block block) {
			foreach(Node node in block.statements) {
				Visit(node);

				if (node is Escape) {
					break;
				}
			}
		}

		void VisitVarDecl(VarDecl decl) {
			callStack.stack[^1].members[decl.identifier] = Visit(decl.expr);
		}

		void VisitAssignment(Assignment assign) {
			// Hack: This allows modifying values from other scopes
			string identifier = assign.identifier;
			ActivationRecord record = ResolveVar(identifier);

			if (record.members[assign.identifier].references != null) {
				identifier = record.members[assign.identifier].references;
				record = ResolveVar(identifier);
			}

			record.members[identifier] = Visit(assign.expr);
		}

		Value VisitFunctionCall(FunctionCall function) {
			ActivationRecord fnscope = new ActivationRecord(
				function.token.Lexeme,
				RecordType.Function,
				callStack.stack[^1].scopeLevel + 1
			);

			int idx = 0;
			foreach(Node arg in function.arguments) {
				fnscope.members[function.sym.parameters[idx].identifier] = Visit(arg);
				
				if (arg is Var) {
					// HACK:  Climb the ladder of references until we hit the uppermost variable
					//		  Obviously this is a terrible solution as it requires more work
					// FIXME: The reference should ideally never change, so a simple assignment
					//		  to the original name should be valid. Using a name is also not great,
					//		  as it may conflict with other scopes and will require more resolution
					//		  later on.
					Value val = ResolveVar(arg.token.Lexeme).members[arg.token.Lexeme];
					string upper_variable = arg.token.Lexeme;

					while (val.references != null){ 
						upper_variable = val.references;
						val = ResolveVar(val.references).members[val.references];
					}

					fnscope.members[
						function.sym.parameters[idx].identifier
					].references = upper_variable;
				}

				idx++;
			}
			
			// Insert an implicit return value
			if (function.sym.def.returnType.type != "void") {
				if (function.sym.def.returnType.expr != null) {
					fnscope.members["result"] = Visit(function.sym.def.returnType.expr);
				} else {
					fnscope.members["result"] = DefaultValue(function.sym.def.returnType);
				}
			}

			callStack.stack.Add(fnscope);
			VisitBlock(function.sym.def.block);

			Value result = ResolveVar("result").members["result"];
			callStack.stack.Remove(fnscope);

			return result;
		}

		void VisitBuiltinFunctionCall(BuiltinFunctionCall function) {
			List<Value> values = new List<Value>();
			foreach(Node n in function.arguments) {
				values.Add(Visit(n));
			}

			function.native.function(values);
		}

		Value VisitVar(Var var) {
			// Variable resolution through records
			for(int idx = callStack.stack.Count - 1; idx >= 0; idx--) {
				if (callStack.stack[idx].members.ContainsKey(var.token.Lexeme)) {
					Value val = callStack.stack[idx].members[var.token.Lexeme];
					return (val.references != null) ? ResolveVar(val.references).members[val.references] : val;
				}

			}
			Error($"Unknown variable read '{var.token.Lexeme}'");
			return null;
		}

		Value VisitLiteral(Literal literal) {
			switch(literal.token.Kind) {
				case TokenKind.Int:			return new Value(ValueKind.Int, int.Parse(literal.token.Lexeme));
				case TokenKind.Float:		return new Value(ValueKind.Float, float.Parse(literal.token.Lexeme));
				case TokenKind.Boolean:		return new Value(ValueKind.Bool, bool.Parse(literal.token.Lexeme));
				case TokenKind.String:		return new Value(ValueKind.String, literal.token.Lexeme);

				default:
					Error($"Unknown literal type {literal.token.Kind}");
					return null;
			}
		}
	}
}