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

		static Value TinyTypesMatch(Value[] arguments) {
			return new BoolValue(TinyType.Matches(arguments[0].Kind, arguments[1].Kind));
		}

		static Value TinyTypeOf(Value[] arguments) {
			return new StringValue(arguments[0].Kind.ToString());
		}

		static Value TinyEval(Value[] arguments) {
			return new Interpreter().Run(new Parser((string)arguments[0].Data).Parse());
		}

		static Value TinyAssert(Value[] arguments) {
			if (!(bool)arguments[0].Data) {
				throw new AssertException($"Assertion failed '{(string)arguments[1].Data}'");
			}
			
			return new UnitValue();
		}

		static Value TinyAssertEq(Value[] arguments) {
			if ((bool)arguments[0].Data != (bool)arguments[1].Data) {
				throw new AssertException($"Assertion failed '{(string)arguments[2].Data}' expected {(bool)arguments[1].Data} but received {(bool)arguments[0].Data}");
			}
			
			return new UnitValue();
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
			),
			new BuiltinFn(
				"types_match", TinyTypesMatch,
				new Parameter[] {
					new Parameter("left", new TinyAny()),
					new Parameter("right", new TinyAny())
				},
				new TinyBool()
			),
			new BuiltinFn(
				"type_of", TinyTypeOf,
				new Parameter[] {
					new Parameter("value", new TinyAny()),
				},
				new TinyString()
			),
			new BuiltinFn(
				"eval", TinyEval,
				new Parameter[] {
					new Parameter("source", new TinyString()),
				},
				new TinyAny()
			),
			new BuiltinFn(
				"assert", TinyAssert,
				new Parameter[] {
					new Parameter("condition", new TinyBool()),
					new Parameter("messsage", new TinyString()),
				},
				new TinyUnit()
			),
			new BuiltinFn(
				"assert_eq", TinyAssertEq,
				new Parameter[] {
					new Parameter("condition", new TinyBool()),
					new Parameter("expected", new TinyBool()),
					new Parameter("messsage", new TinyString()),
				},
				new TinyUnit()
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