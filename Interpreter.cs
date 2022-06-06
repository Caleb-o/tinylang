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

		public VarSym Resolve(string identifier) {
			for(int i = stack.Count - 1; i >= 0; i++) {
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
		
		Value Visit(Node node) {
			switch(node) {
				case Block:				return VisitBlock((Block)node);
				case BinaryOp: 			return VisitBinaryOp((BinaryOp)node);
				case Print: 			return VisitPrintStmt((Print)node);
				case Literal:			return VisitLiteral((Literal)node);
				case VariableDecl:		return VisitVariableDecl((VariableDecl)node);
				case FunctionCall:		return VisitFunctionCall((FunctionCall)node);
				case Identifier:		return VisitIdentifier((Identifier)node);

				// NoOp
				case FunctionDef: 		return null;
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

			return null;
		}
		
		Value VisitBinaryOp(BinaryOp binaryOp) {
			switch(binaryOp.token.Kind) {
				case TokenKind.Plus:		return Visit(binaryOp.left) + Visit(binaryOp.right);
			}

			Error($"Unhandled operator in binary operation {binaryOp.token.Kind}");
			return null;
		}

		Value VisitBlock(Block block) {
			foreach(Node node in block.statements) {
				Visit(node);
			}
			return null;
		}

		Value VisitLiteral(Literal literal) {
			switch(literal.token.Kind) {
				case TokenKind.Int:			return new IntValue(int.Parse(literal.token.Lexeme));
				case TokenKind.Float:		return new FloatValue(float.Parse(literal.token.Lexeme));
				case TokenKind.Boolean:		return new BoolValue(bool.Parse(literal.token.Lexeme));
				case TokenKind.String:		return new StringValue(literal.token.Lexeme);
			}

			Error($"Unknown liteal type {literal.token.Kind}");
			return null;
		}

		Value VisitVariableDecl(VariableDecl vardecl) {
			VarSym variable = new VarSym(vardecl.token.Lexeme, vardecl.kind);
			variable.value = Visit(vardecl.expr);
			callStack.Add(variable);

			return null;
		}

		Value VisitFunctionCall(FunctionCall fncall) {
			callStack.PushRecord(fncall.token.Lexeme);
			
			int idx = 0;
			foreach(Argument arg in fncall.arguments) {
				VarSym variable = new VarSym(fncall.def.parameters[idx].token.Lexeme, arg.kind);
				variable.value = Visit(arg.expr);

				callStack.Add(variable);
				idx++;
			}

			Visit(fncall.def.block);
			
			callStack.PopRecord();
			// FIXME: Allow returning value from calls
			return null;
		}

		Value VisitIdentifier(Identifier identifier) {
			return callStack.Resolve(identifier.token.Lexeme).value;
		}
	}	
}