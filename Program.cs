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
				Parser parser = new Parser(File.ReadAllText(args[0]));
				Application app = parser.Parse();

				Interpreter interpreter = new Interpreter();
				interpreter.Run(app);
			} catch(Exception e) {
				Console.WriteLine($"Error: {e.Message}");
				Console.WriteLine($"Trace: {e.StackTrace}");
			}
		}
	}
}