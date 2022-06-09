using System;
using System.Linq;
using System.Text;
using System.Collections.Generic;


namespace TinyLang {
	class ActivationRecord {
		public readonly string identifier;
		public readonly int depth;
		public readonly int line;
		public readonly Dictionary<string, VarSym> members;

		public ActivationRecord(string identifier, int depth, Dictionary<string, VarSym> members, int line) {
			this.identifier = identifier;
			this.depth = depth;
			this.members = members;
			this.line = line;
		}

		public override string ToString()
		{
			StringBuilder sb = new StringBuilder();

			int idx = 0;
			foreach(VarSym variable in members.Values) {
				sb.Append($"  [{idx}] '{variable.identifier}' {variable.value} => {variable.kind}\n");
				idx++;
			}

			return sb.ToString();
		}
	}

	class CallStack {
		public List<ActivationRecord> stack = new List<ActivationRecord>();

		public void Add(VarSym variable) {
			stack[^1].members[variable.identifier] = variable;
		}

		public void PushRecord(string identifier, Block block) {
			int line = (block != null && block.token != null) ? block.token.Line : 1;
			stack.Add(new ActivationRecord(identifier, stack.Count, new Dictionary<string, VarSym>(), line));
		}

		public void PushRecord(string identifier, int line) {
			stack.Add(new ActivationRecord(identifier, stack.Count, new Dictionary<string, VarSym>(), line));
		}

		public void PopRecord() {
			stack.Remove(stack[^1]);
		}

		public VarSym Resolve(string identifier) {
			for(int i = stack.Count - 1; i >= 0; i--) {
				if (stack[i].members.ContainsKey(identifier)) {
					return stack[i].members[identifier];
				}
			}

			return null;
		}

		public void Print() {
			Console.WriteLine("--- CallStack ---");
			for(int i = stack.Count - 1; i >= 0; i--) {
				ActivationRecord record = stack[i];

				Console.WriteLine($"[{i}] '{record.identifier}' : line {record.line}\n{record}");
			}
		}
	}

	class Interpreter {
		readonly CallStack callStack = new CallStack();
		readonly Analyser analyser = new Analyser();

		public Interpreter() {
			callStack.PushRecord("global", 1);
			AddBuiltinFns();
		}

		public void ImportFunction(BuiltinFn func) {
			analyser.ImportFunction(func);
			callStack.Add(new VarSym(func.identifier, new FunctionDef(func.identifier, func)));
		}

		public Value Run(Application app) {
			analyser.Run(app);

			if (Analyser.HasError) {
				return new UnitValue();
			}

			Value result = new UnitValue();

			VarSym resultvar = new VarSym("result", true, new TinyAny());
			resultvar.value = new UnitValue();
			callStack.Add(resultvar);

			try {
				result = Visit(app.block);
			} catch(AssertException assert) {
				callStack.Print();
				throw assert;
			} catch (ReturnException) {}

			result = resultvar.value;

			callStack.PopRecord();

			return result;
		}

		void AddBuiltinFns() {
			foreach(BuiltinFn func in Builtins.BuiltinFunctions) {
				analyser.ImportFunction(func);
				callStack.Add(new VarSym(func.identifier, new FunctionDef(func.identifier, func)));
			}
		}

		void Error(string message) {
			callStack.Print();
			throw new Exception($"Runtime: {message}");
		}

		void Error(string message, Token token) {
			callStack.Print();
			throw new Exception($"Runtime: {message} ['{token.Lexeme}' {token.Line}:{token.Column}]");
		}
		
		Value Visit(Node node) {
			switch(node) {
				case Block:					return VisitBlock((Block)node);
				case BinaryOp: 				return VisitBinaryOp((BinaryOp)node);
				case UnaryOp: 				return VisitUnaryOp((UnaryOp)node);
				case Print: 				return VisitPrintStmt((Print)node);
				case Literal:				return VisitLiteral((Literal)node);
				case ListLiteral:			return VisitListLiteral((ListLiteral)node);
				case VariableDecl:			return VisitVariableDecl((VariableDecl)node);
				case VariableAssignment:	return VisitVariableAssign((VariableAssignment)node);
				case FunctionDef: 			return VisitFunctionDef((FunctionDef)node);
				case FunctionCall:			return VisitFunctionCall((FunctionCall)node);
				case Identifier:			return VisitIdentifier((Identifier)node);
				case StructDef: 			return VisitStructDef((StructDef)node);
				case StructInstance:		return VisitStructInstance((StructInstance)node);
				case ConditionalOp:			return VisitConditionalOp((ConditionalOp)node);
				case IfStmt:				return VisitIfStatement((IfStmt)node);
				case WhileStmt:				return VisitWhileStatement((WhileStmt)node);
				case DoWhileStmt:			return VisitDoWhileStatement((DoWhileStmt)node);
				case Return:				return VisitReturn((Return)node);
				case Index:					return VisitIndex((Index)node);
				case MemberAccess:			return VisitMemberAccess((MemberAccess)node);
			}

			Error($"Unhandled node in interpreter {node}");
			return null;
		}

		Value VisitPrintStmt(Print print) {
			StringBuilder sb = new StringBuilder();

			foreach(Node node in print.arguments) {
				sb.Append(Visit(node));
			}

			Console.WriteLine(sb.ToString());

			return new UnitValue();
		}
		
		Value VisitBinaryOp(BinaryOp binaryOp) {
			switch(binaryOp.token.Kind) {
				case TokenKind.Plus:		return Visit(binaryOp.left) + Visit(binaryOp.right);
				case TokenKind.Minus:		return Visit(binaryOp.left) - Visit(binaryOp.right);
				case TokenKind.Star:		return Visit(binaryOp.left) * Visit(binaryOp.right);
				case TokenKind.Slash:		return Visit(binaryOp.left) / Visit(binaryOp.right);
			}

			Error($"Unhandled operator in binary operation {binaryOp.token.Kind}");
			return null;
		}

		Value VisitUnaryOp(UnaryOp unary) {
			return -Visit(unary.right);
		}

		Value VisitBlock(Block block) {
			callStack.PushRecord("block", block);

			foreach(Node node in block.statements) {
				Visit(node);
			}
			
			callStack.PopRecord();
			return new UnitValue();
		}

		Value VisitLiteral(Literal literal) {
			switch(literal.token.Kind) {
				case TokenKind.Int:			return new IntValue(int.Parse(literal.token.Lexeme));
				case TokenKind.Float:		return new FloatValue(float.Parse(literal.token.Lexeme));
				case TokenKind.Bool:		return new BoolValue(bool.Parse(literal.token.Lexeme));
				case TokenKind.String:		return new StringValue(literal.token.Lexeme);
			}

			Error($"Unknown literal type {literal.token.Kind}");
			return null;
		}

		Value VisitListLiteral(ListLiteral literal) {
			List<Value> values = new List<Value>();
			foreach(Node expr in literal.exprs) {
				values.Add(Visit(expr));
			}

			return new ListValue(literal.kind, values);
		}

		Value VisitVariableDecl(VariableDecl vardecl) {
			VarSym variable = new VarSym(vardecl.token.Lexeme, vardecl.mutable, vardecl.kind);
			variable.value = Visit(vardecl.expr);
			variable.validated = true;
			callStack.Add(variable);

			return new UnitValue();
		}

		Value VisitVariableAssign(VariableAssignment assign) {
			VarSym variable = callStack.Resolve(assign.token.Lexeme);

			while(variable.references != null) {
				variable = variable.references;
			}

			Value id = Visit(assign.identifier);
			id.Data = Visit(assign.expr).Data;

			return new UnitValue();
		}

		Value VisitFunctionCall(FunctionCall fncall) {
			VarSym fnsym = (VarSym)callStack.Resolve(fncall.token.Lexeme);

			// Follow the reference chain
			if (fnsym != null && fnsym.references != null) {
				fnsym = fnsym.references;

				while (fnsym.references != null) {
					fnsym = fnsym.references;
				}
			}
			
			if (fnsym == null || fnsym.kind is not TinyFunction) {
				Error($"Function '{fncall.token.Lexeme}' does not exist in any scope", fncall.token);
			}

			if (fnsym.validated) {
				fncall.def = (FunctionDef)fnsym.value.Data;
			
				// Check for required args
				if (fncall.arguments.Length != fncall.def.parameters.Length) {
					Error($"Function variable '{fncall.token.Lexeme}' expected {fncall.def.parameters.Length} argument(s) but received {fncall.arguments.Length}");
				}

				for(int i = 0; i < fncall.arguments.Length; i++) {
					TinyType param = fncall.def.parameters[i].kind;

					if (param is not TinyAny && !TinyType.Matches(fncall.arguments[i].kind, param)) {
						Error($"Argument '{fncall.def.parameters[i].identifier}' in function '{fncall.def.identifier}' expected type {param} but received {fncall.arguments[i].kind}");
					}
				}
			}

			if (fncall.def.block is BuiltinFn) {
				BuiltinFn fn = (BuiltinFn)fncall.def.block;

				Value[] arguments = new Value[fncall.arguments.Length];

				for(int idx = 0; idx < fncall.arguments.Length; idx++) {
					arguments[idx] = Visit(fncall.arguments[idx].expr);
				}

				callStack.PushRecord($"builtin_{fn.identifier}", fncall.token.Line);
				Value returns = fn.function(arguments);
				callStack.PopRecord();

				return returns;
			} else {
				callStack.PushRecord(fncall.token.Lexeme, (Block)fncall.def.block);

				VarSym result = new VarSym("result", true, fncall.def.returns);
				result.value = Value.DefaultFrom(fncall.def.returns);
				callStack.Add(result);
				
				int idx = 0;
				foreach(Argument arg in fncall.arguments) {
					Value value = Visit(arg.expr);

					// if (arg.kind is not TinyAny && !TinyType.Matches(arg.kind, value.Kind)) {
					// 	Error($"Argument at position {idx + 1} in function '{fncall.token.Lexeme}' expected type {arg.kind} but received {value.Kind}");
					// }

					Parameter param = fncall.def.parameters[idx];

					VarSym variable = new VarSym(param.identifier, param.mutable, arg.kind);
					variable.validated = true;
					variable.value = value;

					// FIXME: Move more analysis to the analyser
					if (arg.expr is Identifier) {
						variable.references = callStack.Resolve(((Identifier)arg.expr).token.Lexeme);
						
						if (param.mutable && !variable.references.mutable) {
							Error($"Trying to pass immutable argument '{variable.references.identifier}' to a mutable parameter '{param.identifier}'");
						}
					} else if (param.mutable) {
						Error($"Mutable parameter '{param.identifier}' must receive an identifier, but received '{value}'", fncall.token);
					}


					callStack.Add(variable);
					idx++;
				}
				
				try {
					Visit((Block)fncall.def.block);
				} catch (ReturnException) {}

				callStack.PopRecord();

				// FIXME: Allow returning value from calls
				return result.value;
			}
		}

		Value VisitIdentifier(Identifier identifier) {
			return callStack.Resolve(identifier.token.Lexeme).value;
		}

		Value VisitFunctionDef(FunctionDef fndef) {
			VarSym variable = new VarSym(fndef.identifier, fndef);
			variable.validated = true;
			callStack.Add(variable);

			return new FunctionValue(fndef);
		}

		Value VisitStructDef(StructDef sdef) {
			VarSym variable = new VarSym(sdef.identifier, sdef);
			variable.validated = true;
			callStack.Add(variable);

			return new UnitValue();
		}

		Value VisitStructInstance(StructInstance instance) {
			Dictionary<string, Value> values = new Dictionary<string, Value>();

			foreach(var (id, expr) in instance.members) {
				values[id] = Visit(expr);
			}

			return new StructValue(instance.def, values);
		}

		Value VisitConditionalOp(ConditionalOp cond) {
			Value left = Visit(cond.left);
			Value right = Visit(cond.right);

			switch(cond.token.Kind) {
				case TokenKind.EqualEqual:		return Value.EqualityEqual(left, right);
				case TokenKind.NotEqual:		return Value.EqualityNotEqual(left, right);

				case TokenKind.Less:			return left < right;
				case TokenKind.LessEqual:		return left <= right;
				case TokenKind.Greater:			return left > right;
				case TokenKind.GreaterEqual:	return left >= right;
			}

			throw new InvalidOperationException($"Invalid conditional operation '{cond.token.Kind}'");
		}

		Value VisitIfStatement(IfStmt stmt) {
			if (stmt.initStatement != null) {
				callStack.PushRecord("if_init", stmt.trueBody);
				Visit(stmt.initStatement);
			}

			if ((bool)Visit(stmt.expr).Data) {
				Visit(stmt.trueBody);
			} else if (stmt.falseBody != null) {
				Visit(stmt.falseBody);
			}

			if (stmt.initStatement != null) {
				callStack.PopRecord();
			}

			return new UnitValue();
		}

		Value VisitWhileStatement(WhileStmt stmt) {
			if (stmt.initStatement != null) {
				callStack.PushRecord("while_init", stmt.body);
				Visit(stmt.initStatement);
			}

			while ((bool)Visit(stmt.expr).Data) {
				Visit(stmt.body);
			}

			if (stmt.initStatement != null) {
				callStack.PopRecord();
			}

			return new UnitValue();
		}

		Value VisitDoWhileStatement(DoWhileStmt stmt) {
			if (stmt.initStatement != null) {
				callStack.PushRecord("dowhile_init", stmt.body);
				Visit(stmt.initStatement);
			}

			Visit(stmt.body);

			while ((bool)Visit(stmt.expr).Data) {
				Visit(stmt.body);
			}

			if (stmt.initStatement != null) {
				callStack.PopRecord();
			}

			return new UnitValue();
		}

		Value VisitReturn(Return ret) {
			throw new ReturnException();
		}

		Value VisitIndex(Index index) {
			VarSym variable = callStack.Resolve(index.token.Lexeme);
			ListValue list = (ListValue)variable.value;
			int lindex = -1;

			int idx = 0;
			foreach(Node expr in index.expr) {
				lindex = (int)Visit(expr).Data;
				List<Value> values = (List<Value>)list.Data;

				if (lindex < 0 || lindex >= values.Count) {
					Error($"Index out of bounds: {lindex} of range 0..{values.Count}", index.token);
				}

				idx++;

				if (idx < index.expr.Length) {
					list = (ListValue)values[lindex];
				}
			}

			return ((List<Value>)list.Data)[lindex];
		}

		Value VisitMemberAccess(MemberAccess access) {
			VarSym variable = callStack.Resolve(access.token.Lexeme);
			StructValue structv = (StructValue)variable.value;

			int idx = 0;
			Identifier id = null;

			foreach(Node expr in access.members) {
				id = (Identifier)expr;

				if (idx++ < access.members.Length - 1) {
					structv = (StructValue)((Dictionary<string, Value>)structv.Data)[id.token.Lexeme];
				}
			}

			return ((Dictionary<string, Value>)structv.Data)[id.token.Lexeme];
		}
	}	
}