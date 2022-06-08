using System;
using System.IO;
using System.Collections.Generic;

namespace TinyLang {
	class Program {
		static Value TestFn(List<Value> arguments) {
			Console.WriteLine("Hello from C#");
			return new UnitValue();
		}

		public static void Main(string[] args) {
			if (args.Length < 1) {
				Console.WriteLine($"Usage: tiny [script]");
				return;
			}

			BuiltinFn testingFn = new BuiltinFn(
				"testfn", TestFn,
				new Parameter[] {
					new Parameter("a_bool", new TinyBool())
				},
				new TinyUnit()
			);

			try {
				Application app = new Parser(File.ReadAllText(args[0])).Parse();

				Interpreter interpreter = new Interpreter();
				interpreter.ImportFunction(testingFn);
				Value result = interpreter.Run(app);

				if (result is not UnitValue) {
					Console.WriteLine(result);
				}
			} catch(Exception e) {
				Console.WriteLine($"Error: {e.Message}");
				Console.WriteLine($"Trace: {e.StackTrace}");
			}
		}
	}
}