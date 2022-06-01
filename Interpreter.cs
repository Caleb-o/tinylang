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

			Visit(app.block);
			callStack.stack.Remove(callStack.stack[^1]);
		}

		Value DefaultValue(Return ret) {
			// TODO: If return is a record/struct, then return a Visit on a mandatory
			// constructor for the type
			if (!ret.type.IsSingleType()) {
				Error($"Complex types aren't currently supported");
				return null;
			}

			switch(ret.type.type[0]) {
				// FIXME: This is bad becaue it *could* change
				case 0: return new Value(ValueKind.Int, 0);
				case 1: return new Value(ValueKind.Float, 0);
				case 2: return new Value(ValueKind.Bool, false);
				case 3: return new Value(ValueKind.Bool, "");
			}

			throw new InvalidCastException("Unreachable");
		}

		void PrintCallStack() {
			Console.WriteLine("-- Call Stack --");
			for(int i = callStack.stack.Count - 1; i >= 0; i--) {
				Console.WriteLine($"[{i}] {callStack.stack[i].identifier}");

				foreach(var record in callStack.stack[i].members) {
					Console.WriteLine($"{record.Key.PadLeft(16)} = {record.Value.value} [{record.Value.value.kind}]");
				}
			}
			Console.WriteLine();
		}

		void Error(string msg) {
			PrintCallStack();
			throw new InvalidOperationException($"Runtime: {msg} [{callStack.stack[^1].identifier}]");
		}

		// Finds the closest AR that contains a member with the identifier
		ActivationRecord ResolveRecord(string identifier, int offset = 0) {
			for(int i = callStack.stack.Count - (offset + 1); i >= 0; i--) {
				if (callStack.stack[i].members.ContainsKey(identifier)) {
					return callStack.stack[i];
				}
			}

			Error($"Unreachable: unable to find variable '{identifier}'");
			return null;
		}

		VarSym ResolveVar(string identifier, int offset = 0) {
			ActivationRecord record = ResolveRecord(identifier);

			if (record != null) {
				return record.members[identifier];
			}

			Error($"Unreachable: unable to find variable '{identifier}'");
			return null;
		}

		Value Visit(Node node) {
			switch(node) {
				case Block: VisitBlock((Block)node); return null;
				case Literal: return VisitLiteral((Literal)node);
				case BinOp: return VisitBinOp((BinOp)node);
				case ConditionalOp: return VisitConditionalOp((ConditionalOp)node);
				case UnaryOp: return VisitUnaryOp((UnaryOp)node);
				case IfStmt: VisitIfStmt((IfStmt)node); return null;
				case While: VisitWhile((While)node); return null;
				case DoWhile: VisitDoWhile((DoWhile)node); return null;

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

		Value VisitConditionalOp(ConditionalOp conditional) {
			switch(conditional.token.Kind) {
				case TokenKind.EqualEqual: 		return Visit(conditional.left) == Visit(conditional.right);
				case TokenKind.NotEqual: 		return Visit(conditional.left) != Visit(conditional.right);

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
			foreach(Node node in block.statements) {
				Visit(node);

				if (node is Escape) {
					break;
				}
			}
		}

		void VisitIfStmt(IfStmt ifstmt) {
			if ((bool)VisitConditionalOp((ConditionalOp)ifstmt.expr).value) {
				Visit(ifstmt.trueBody);
			} else if (ifstmt.falseBody != null) {
				Visit(ifstmt.falseBody);
			}
		}

		void VisitWhile(While whilestmt) {
			while ((bool)VisitConditionalOp((ConditionalOp)whilestmt.expr).value) {
				Visit(whilestmt.body);
			}
		}

		void VisitDoWhile(DoWhile whilestmt) {
			Visit(whilestmt.body);

			while ((bool)VisitConditionalOp((ConditionalOp)whilestmt.expr).value) {
				Visit(whilestmt.body);
			}
		}

		void VisitVarDecl(VarDecl decl) {
			callStack.stack[^1].members[decl.identifier] = new VarSym(decl.identifier, decl.type, decl.mutable);
			callStack.stack[^1].members[decl.identifier].value = Visit(decl.expr);
		}

		void VisitAssignment(Assignment assign) {
			VarSym variable = ResolveVar(assign.identifier);

			if ((object)variable.references != null) {
				variable.references.value = Visit(assign.expr);
			} else {
				ResolveRecord(assign.identifier).members[assign.identifier].value = Visit(assign.expr);
			}
		}

		Value VisitFunctionCall(FunctionCall function) {
			ActivationRecord fnscope = new ActivationRecord(
				function.token.Lexeme,
				RecordType.Function,
				callStack.stack[^1].scopeLevel + 1
			);

			int idx = 0;
			foreach(Node arg in function.arguments) {
				string identifier = function.sym.parameters[idx].identifier;
				fnscope.members[identifier] = function.sym.parameters[idx];
				fnscope.members[identifier].value = Visit(arg);
				
				if (arg is Var) {
					VarSym variable = ResolveVar(arg.token.Lexeme);	

					while ((object)variable.references != null){
						variable = variable.references;
					}

					fnscope.members[
						function.sym.parameters[idx].identifier
					].references = variable;
				}

				idx++;
			}
			
			// Insert an implicit return value
			if (!function.sym.def.returnType.type.Matches(new Type(Application.GetTypeID("void")))) {
				if (function.sym.def.returnType.expr != null) {
					fnscope.members["result"].value = Visit(function.sym.def.returnType.expr);
				} else {
					fnscope.members["result"].value = DefaultValue(function.sym.def.returnType);
				}
			}

			callStack.stack.Add(fnscope);
			VisitBlock(function.sym.def.block);

			Value result = null;

			if (!function.sym.def.returnType.type.Matches(new Type(Application.GetTypeID("void")))) {
				result = ResolveVar("result").value;
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
			return ((object)variable.references != null) ? variable.references.value : variable.value;
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