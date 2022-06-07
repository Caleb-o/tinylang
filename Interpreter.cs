using System;
using System.Linq;
using System.Text;
using System.Collections.Generic;


namespace TinyLang {
	class ActivationRecord {
		public readonly string identifier;
		public readonly int depth;
		public readonly Dictionary<string, VarSym> members;

		public ActivationRecord(string identifier, int depth, Dictionary<string, VarSym> members) {
			this.identifier = identifier;
			this.depth = depth;
			this.members = members;
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

		public void PushRecord(string identifier) {
			stack.Add(new ActivationRecord(identifier, stack.Count, new Dictionary<string, VarSym>()));
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

				Console.WriteLine($"[{i}] '{record.identifier}'\n{record}");
			}
		}
	}

	class Interpreter {
		CallStack callStack = new CallStack();


		public Value Run(Application app) {
			Value result = new UnitValue();

			callStack.PushRecord("global");

			VarSym resultvar = new VarSym("result", true, new TinyAny());
			resultvar.value = new UnitValue();
			callStack.Add(resultvar);

			try {
				result = Visit(app.block);
			} catch (ReturnException) {}

			result = resultvar.value;

			callStack.PopRecord();

			return result;
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
				case FunctionCall:			return VisitFunctionCall((FunctionCall)node);
				case Identifier:			return VisitIdentifier((Identifier)node);
				case FunctionDef: 			return VisitFunctionDef((FunctionDef)node);
				case StructDef: 			return VisitStructDef((StructDef)node);
				case StructInstance:		return VisitStructInstance((StructInstance)node);
				case ConditionalOp:			return VisitConditionalOp((ConditionalOp)node);
				case IfStmt:				return VisitIfStatement((IfStmt)node);
				case WhileStmt:				return VisitWhileStatement((WhileStmt)node);
				case DoWhileStmt:			return VisitDoWhileStatement((DoWhileStmt)node);
				case Return:				return VisitReturn((Return)node);
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
			callStack.PushRecord("block");

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

			variable.value = Visit(assign.expr);

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
				if (fncall.arguments.Count != fncall.def.parameters.Count) {
					Error($"Function variable '{fncall.token.Lexeme}' expected {fncall.def.parameters.Count} argument(s) but received {fncall.arguments.Count}");
				}

				for(int i = 0; i < fncall.arguments.Count; i++) {
					TinyType param = fncall.def.parameters[i].kind;

					if (param is not TinyAny && !TinyType.Matches(fncall.arguments[i].kind, param)) {
						Error($"Argument '{fncall.def.parameters[i].token.Lexeme}' in function '{fncall.def.identifier}' expected type {param} but received {fncall.arguments[i].kind}");
					}
				}
			}

			callStack.PushRecord(fncall.token.Lexeme);

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

				VarSym variable = new VarSym(param.token.Lexeme, param.mutable, arg.kind);
				variable.validated = true;
				variable.value = value;

				if (arg.expr is Identifier) {
					variable.references = callStack.Resolve(((Identifier)arg.expr).token.Lexeme);
				}

				callStack.Add(variable);
				idx++;
			}

			try {
				Visit(fncall.def.block);
			} catch (ReturnException) {}

			callStack.PopRecord();
			// FIXME: Allow returning value from calls
			return result.value;
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
			List<Value> values = new List<Value>();

			foreach(Node node in instance.members.Values) {
				values.Add(Visit(node));
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
				callStack.PushRecord("if_init");
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
				callStack.PushRecord("while_init");
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
				callStack.PushRecord("dowhile_init");
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
	}	
}