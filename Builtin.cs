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

		static Value TinyWriteFile(Value[] arguments) {
			File.WriteAllText((string)arguments[0].Data, (string)arguments[1].Data);
			return new UnitValue();
		}

		static Value TinyFileExists(Value[] arguments) {
			return new BoolValue(File.Exists((string)arguments[0].Data));
		}

		public static BuiltinFn[] BuiltinFunctions = new BuiltinFn[] {
			new BuiltinFn(
				"read_file", TinyReadFile,
				new Parameter[] {
					new Parameter("file_name", new TinyString())
				},
				new TinyString()
			),
			new BuiltinFn(
				"write_file", TinyWriteFile,
				new Parameter[] {
					new Parameter("file_name", new TinyString()),
					new Parameter("text", new TinyString())
				},
				new TinyUnit()
			),
			new BuiltinFn(
				"file_exists", TinyFileExists,
				new Parameter[] {
					new Parameter("file_name", new TinyString())
				},
				new TinyBool()
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