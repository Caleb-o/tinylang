using System;

namespace TinyLang {
	static class Reporter {
		public static void Assert(string message) {
			Console.WriteLine($"\u001b[33mAssertion:\u001b[0m {message}");
		}
		
		public static void Report(string message) {
			Console.WriteLine($"\u001b[31;1mError:\u001b[0m {message}");
		}

		public static void ReportSystem(string message) {
			Console.WriteLine($"\u001b[31;1mSystem Error:\u001b[0m {message}");
		}
	}
}