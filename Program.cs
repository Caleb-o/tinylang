using System;
using System.IO;

namespace TinyLang {
	class Program {
		public static void Main(string[] args) {
			if (args.Length < 1) {
				Console.WriteLine($"Usage: tiny [script]");
				return;
			}

			try {
				Application app = new Parser(File.ReadAllText(args[0])).Parse();

				Interpreter interpreter = new Interpreter();
				Value result = interpreter.Run(app);

				if (result is not UnitValue) {
					Console.WriteLine(result);
				}
			} catch(AssertException assert) {
				Reporter.Assert(assert.Message);
			} catch(Exception e) {
				Reporter.Report($"Error: {e.Message}");
				Console.WriteLine($"Trace: {e.StackTrace}");
			}
		}
	}
}