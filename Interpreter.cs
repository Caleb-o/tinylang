using System;
using System.Collections.Generic;

namespace TinyLang {
	class Interpreter {
		CallStack callStack;


		public Interpreter() {
			callStack = new CallStack();
		}

		public void Run(Application app) {
			callStack.stack.Add(new ActivationRecord(
				"global", RecordType.Program, 0
			));

			try {
				Visit(app.block);
			} catch (AssertionException assert) {
				Console.WriteLine(assert.Message);
				PrintCallStack();
			}
			
			callStack.stack.Remove(callStack.stack[^1]);
		}

		Value GetDefaultSingle(TypeKind typeID) {
			switch(typeID) {
				case TypeKind.Int: return new IntValue(0);
				case TypeKind.Float: return new FloatValue(0.0f);
				case TypeKind.Bool: return new BoolValue(false);
				case TypeKind.String: return new StringValue("");
			}
			
			throw new InvalidCastException($"TypeKind {(TypeKind)typeID} does not support a default value");
		}

		Value DefaultValue(Return ret) {
			// TODO: If return is a record/struct, then return a Visit on a mandatory
			// constructor for the type
			if (ret.type.SubKind == null) {
				return GetDefaultSingle(ret.type.Kind);
			}

			List<Value> values = new List<Value>();

			foreach(TypeKind typeID in ret.type.SubKind) {
				values.Add(GetDefaultSingle(typeID));
			}

			Value value = null;

			if (ret.type.Kind == TypeKind.Tuple) {
				value = new TupleValue(values);
			} else {
				value = new ListValue(values);
			}

			value.SetSubKind(ret.type.SubKind);
			return value;
		}

		void PrintCallStack() {
			Console.WriteLine("-- Call Stack --");
			for(int i = callStack.stack.Count - 1; i >= 0; i--) {
				Console.WriteLine($"[{i}] {callStack.stack[i].identifier}");

				foreach(var record in callStack.stack[i].members) {
					if (record.Value.value != null) {
						Console.WriteLine($"{record.Key.PadLeft(16)} [{record.Value.value.Kind}] = {record.Value.value}");
					} else {
						Console.WriteLine($"{record.Key.PadLeft(16)} = Unbound");
					}
				}
			}
			Console.WriteLine();
		}

		void Error(string msg) {
			PrintCallStack();

			string identifier = (callStack.stack.Count > 0) ? callStack.stack[^1].identifier : "None";
			throw new InvalidOperationException($"Runtime: {msg} [{identifier}]");
		}

		// Finds the closest AR that contains a member with the identifier
		ActivationRecord ResolveRecord(string identifier, int offset = 0) {
			for(int i = callStack.stack.Count - (offset + 1); i >= 0; i--) {
				if (callStack.stack[i].members.ContainsKey(identifier)) {
					return callStack.stack[i];
				}
			}

			Error($"unable to find record with variable '{identifier}'");
			return null;
		}

		VarSym ResolveVar(string identifier, int offset = 0) {
			ActivationRecord record = ResolveRecord(identifier);

			if (record != null) {
				return record.members[identifier];
			}

			Error($"unable to find variable '{identifier}'");
			return null;
		}

		public Value Visit(Node node) {
			switch(node) {
				case Block: VisitBlock((Block)node); return null;
				case BinOp: return VisitBinOp((BinOp)node);
				case ConditionalOp: return VisitConditionalOp((ConditionalOp)node);
				case UnaryOp: return VisitUnaryOp((UnaryOp)node);
				case IfStmt: VisitIfStmt((IfStmt)node); return null;
				case While: VisitWhile((While)node); return null;
				case DoWhile: VisitDoWhile((DoWhile)node); return null;

				case FunctionDef: return null;
				case FunctionCall: return VisitFunctionCall((FunctionCall)node);
				case BuiltinFunctionCall: VisitBuiltinFunctionCall((BuiltinFunctionCall)node); return null;
				case Escape: VisitEscape((Escape)node); return null;

				case Var: return VisitVar((Var)node);
				case Literal: return VisitLiteral((Literal)node);
				case ComplexLiteral: return VisitComplexLiteral((ComplexLiteral)node);
				case Index: return VisitIndex((Index)node);
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

		Value VisitConditionalOp(ConditionalOp conditional) {
			switch(conditional.token.Kind) {
				case TokenKind.EqualEqual: 		return Value.EqualityEqual(Visit(conditional.left), Visit(conditional.right));
				case TokenKind.NotEqual: 		return Value.EqualityNotEqual(Visit(conditional.left), Visit(conditional.right));

				case TokenKind.Greater: 		return Visit(conditional.left) > Visit(conditional.right);
				case TokenKind.GreaterEqual: 	return Visit(conditional.left) >= Visit(conditional.right);

				case TokenKind.Less: 			return Visit(conditional.left) < Visit(conditional.right);
				case TokenKind.LessEqual: 		return Visit(conditional.left) <= Visit(conditional.right);

				default: 
					Error($"Unknown conditional operation {conditional.token.Kind}");
					return null;
			}
		}

		Value VisitUnaryOp(UnaryOp unary) {
			return -Visit(unary.right);
		}

		void VisitBlock(Block block) {
			callStack.stack.Add(new ActivationRecord("anon_block", RecordType.Function, callStack.stack[^1].depth + 1));

			foreach(Node node in block.statements) {
				Visit(node);
			}

			callStack.stack.Remove(callStack.stack[^1]);
		}

		// http://craftinginterpreters.com/functions.html#return-statements
		void VisitEscape(Escape escape) {
			// Unwinds the stack
			throw new EscapeException();
		}

		void VisitIfStmt(IfStmt ifstmt) {
			if (ifstmt.initStatement != null) {
				VisitVarDecl(ifstmt.initStatement);
			}

			if ((bool)Visit(ifstmt.expr).Data) {
				VisitBlock(ifstmt.trueBody);
			} else if (ifstmt.falseBody != null) {
				Visit(ifstmt.falseBody);
			}
		}

		void VisitWhile(While whilestmt) {
			if (whilestmt.initStatement != null) {
				VisitVarDecl(whilestmt.initStatement);
			}

			while ((bool)Visit(whilestmt.expr).Data) {
				VisitBlock(whilestmt.body);
			}
		}

		void VisitDoWhile(DoWhile whilestmt) {
			VisitBlock(whilestmt.body);

			while ((bool)Visit(whilestmt.expr).Data) {
				VisitBlock(whilestmt.body);
			}
		}

		void VisitVarDecl(VarDecl decl) {
			callStack.stack[^1].members[decl.identifier] = new VarSym(decl.identifier, decl.type, decl.mutable);
			callStack.stack[^1].members[decl.identifier].value = Visit(decl.expr);
		}

		void VisitAssignment(Assignment assign) {
			string identifier = assign.identifier.token.Lexeme;
			VarSym variable = ResolveVar(identifier);

			if (assign.identifier is Var) {
				if (variable.references != null) {
					variable.references.value = Visit(assign.expr);
				} else {
					ResolveRecord(identifier).members[identifier].value = Visit(assign.expr);
				}
			} else {
				// Indexing
				Index index = (Index)assign.identifier;
				int indexValue = GetIndexValue((Index)index);
				List<Value> values = (List<Value>)variable.value.Data;

				values[indexValue] = Visit(assign.expr);
			}
		}

		Value VisitFunctionCall(FunctionCall function) {
			ActivationRecord fnscope = new ActivationRecord(
				function.token.Lexeme,
				RecordType.Function,
				callStack.stack[^1].depth + 1
			);

			int idx = 0;
			foreach(Node arg in function.arguments) {
				string identifier = function.sym.parameters[idx].identifier;
				VarSym parameter = function.sym.parameters[idx];
				// We must use a new VarSym in the scope, otherwise issues occur
				fnscope.members[identifier] = new VarSym(parameter.identifier, parameter.type, parameter.mutable);
				fnscope.members[identifier].value = Visit(arg);
				
				if (arg is Var) {
					VarSym variable = ResolveVar(arg.token.Lexeme);	

					while (variable.references != null){
						variable = variable.references;
					}

					fnscope.members[
						function.sym.parameters[idx].identifier
					].references = variable;
				}

				idx++;
			}
			
			// Insert an implicit return value
			if (function.sym.def.returnType != null) {
				fnscope.members["result"] = new VarSym("result", function.sym.def.returnType.type, true);

				if (function.sym.def.returnType.expr != null) {
					fnscope.members["result"].value = Visit(function.sym.def.returnType.expr);
				} else {
					fnscope.members["result"].value = DefaultValue(function.sym.def.returnType);
				}
			}

			callStack.stack.Add(fnscope);

			// A bit of a hack to allow for returning from functions
			try {
				VisitBlock(function.sym.def.block);
			} catch(EscapeException) {}

			Value result = null;

			if (function.sym.def.returnType != null) {
				result = fnscope.members["result"].value;
			}
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
			VarSym variable = ResolveVar(var.token.Lexeme);
			return (variable.references != null) ? variable.references.value : variable.value;
		}

		Value VisitLiteral(Literal literal) {
			switch(literal.token.Kind) {
				case TokenKind.Int:			return Application.GetOrInsertLiteral(literal.token.Lexeme, TypeKind.Int);
				case TokenKind.Float:		return Application.GetOrInsertLiteral(literal.token.Lexeme, TypeKind.Float);
				case TokenKind.Boolean:		return Application.GetOrInsertLiteral(literal.token.Lexeme, TypeKind.Bool);
				case TokenKind.String:		return Application.GetOrInsertLiteral(literal.token.Lexeme, TypeKind.String);

				default:
					Error($"Unknown literal type {literal.token.Kind}");
					return null;
			}
		}

		Value VisitComplexLiteral(ComplexLiteral literal) {
			List<Value> values = new List<Value>();
			List<TypeKind> typeIDs = new List<TypeKind>();

			foreach(Node expr in literal.exprs) {
				Value val = Visit(expr);
				values.Add(val);
				// FIXME: Allow nested types
				typeIDs.Add(val.Kind.Kind);
			}

			Value value;

			if (literal.kind.Kind == TypeKind.Tuple) {
				value = new TupleValue(values);
			} else {
				value = new ListValue(values);
			}

			value.SetSubKind(typeIDs.ToArray());
			return value;
		}

		int GetIndexValue(Index index) {
			return (int)Visit(index.exprs[0]).Data;
		}

		Value VisitIndex(Index index) {
			int indexValue = GetIndexValue(index);
			VarSym variable = ResolveVar(index.token.Lexeme);

			List<Value> values = (List<Value>)variable.value.Data;

			if (indexValue < 0 || indexValue > values.Count - 1) {
				Error($"Index out of bounds on '{index.token.Lexeme}'. Indexing with {indexValue} where the length is {values.Count}");
			}
			return values[indexValue];
		}
	}
}