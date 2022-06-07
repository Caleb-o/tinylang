using System;
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

		public VarSym Resolve(string identifier, TypeKind kind = TypeKind.Unknown) {
			for(int i = stack.Count - 1; i >= 0; i--) {
				if (stack[i].members.ContainsKey(identifier)) {
					return stack[i].members[identifier];
				}
			}

			return null;
		}
	}

	class Interpreter {
		CallStack callStack = new CallStack();


		public void Run(Application app) {
			callStack.PushRecord("global");
			Visit(app.block);
			callStack.PopRecord();
		}

		void Error(string message) {
			throw new Exception($"Runtime: {message}");
		}

		void Error(string message, Token token) {
			throw new Exception($"Runtime: {message} ['{token.Lexeme}' {token.Line}:{token.Column}]");
		}
		
		Value Visit(Node node) {
			switch(node) {
				case Block:					return VisitBlock((Block)node);
				case BinaryOp: 				return VisitBinaryOp((BinaryOp)node);
				case Print: 				return VisitPrintStmt((Print)node);
				case Literal:				return VisitLiteral((Literal)node);
				case VariableDecl:			return VisitVariableDecl((VariableDecl)node);
				case VariableAssignment:	return VisitVariableAssign((VariableAssignment)node);
				case FunctionCall:			return VisitFunctionCall((FunctionCall)node);
				case Identifier:			return VisitIdentifier((Identifier)node);
				case FunctionDef: 			return VisitFunctionDef((FunctionDef)node);
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

		Value VisitBlock(Block block) {
			foreach(Node node in block.statements) {
				Visit(node);
			}
			return new UnitValue();
		}

		Value VisitLiteral(Literal literal) {
			switch(literal.token.Kind) {
				case TokenKind.Int:			return new IntValue(int.Parse(literal.token.Lexeme));
				case TokenKind.Float:		return new FloatValue(float.Parse(literal.token.Lexeme));
				case TokenKind.Boolean:		return new BoolValue(bool.Parse(literal.token.Lexeme));
				case TokenKind.String:		return new StringValue(literal.token.Lexeme);
			}

			Error($"Unknown literal type {literal.token.Kind}");
			return null;
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
			
			if (fnsym == null || fnsym.kind != TypeKind.Function) {
				Error($"Function '{fncall.token.Lexeme}' does not exist in any scope", fncall.token);
			}

			if (fnsym.validated) {
				fncall.def = (FunctionDef)fnsym.value.Data;
			
				// Check for required args
				if (fncall.arguments.Count != fncall.def.parameters.Count) {
					Error($"Function variable '{fncall.token.Lexeme}' expected {fncall.def.parameters.Count} arguments but received {fncall.arguments.Count}");
				}
			}

			callStack.PushRecord(fncall.token.Lexeme);
			
			int idx = 0;
			foreach(Argument arg in fncall.arguments) {
				Value value = Visit(arg.expr);

				if (arg.kind != TypeKind.Unknown && arg.kind != value.Kind) {
					Error($"Argument at position {idx + 1} in function '{fncall.token.Lexeme}' expected type {arg.kind} but received {value.Kind}");
				}

				VarSym variable = new VarSym(fncall.def.parameters[idx].token.Lexeme, false, arg.kind);
				variable.validated = true;
				variable.value = value;

				if (arg.expr is Identifier) {
					variable.references = callStack.Resolve(((Identifier)arg.expr).token.Lexeme);
				}

				callStack.Add(variable);
				idx++;
			}

			Visit(fncall.def.block);
			
			callStack.PopRecord();
			// FIXME: Allow returning value from calls
			return new UnitValue();
		}

		Value VisitIdentifier(Identifier identifier) {
			return callStack.Resolve(identifier.token.Lexeme).value;
		}

		Value VisitFunctionDef(FunctionDef fndef) {
			VarSym variable = new VarSym(fndef.identifier, false, fndef);
			variable.validated = true;
			callStack.Add(variable);

			return new FunctionValue(fndef);
		}
	}	
}