using System;
using System.IO;


namespace TinyLang {
	static class Builtins {
		public delegate Value Fn(Value[] arguments);

		static Value TinyReadFile(Value[] arguments) {
			try {
				return new StringValue(File.ReadAllText((string)arguments[0].Data));
			} catch(Exception) {
				return new StringValue("");
			}
		}

		public static BuiltinFn[] BuiltinFunctions = new BuiltinFn[] {
			new BuiltinFn(
				"read_file", TinyReadFile,
				new Parameter[] {
					new Parameter("file_name", new TinyString())
				},
				new TinyString()
			)
		};
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