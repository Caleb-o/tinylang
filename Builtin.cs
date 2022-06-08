using System.Collections.Generic;

namespace TinyLang {
	static class Builtins {
		public delegate Value Fn(List<Value> arguments);
	}

	sealed class BuiltinFn {
		public readonly string identifier;
		public readonly Builtins.Fn function;
		public readonly Parameter[] parameters;
		public readonly TinyType returns;

		public BuiltinFn(string identifier, Builtins.Fn function, Parameter[] parameters, TinyType returns) {
			this.identifier = identifier;
			this.function = function;
			this.parameters = parameters;
			this.returns = returns;
		}
	}
}