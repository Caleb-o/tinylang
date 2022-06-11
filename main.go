package main

import (
	"tiny/analysis"
	"tiny/interpreter"
	"tiny/parser"
	"tiny/shared"
)

func main() {
	parser := parser.New(shared.ReadFile("./experimenting.tiny"))
	program := parser.Parse()
	analyser := analysis.NewAnalyser(false)

	if !analyser.Run(program.Body) {
		shared.ReportErrFatal("Failed to analyse")
	}

	interpreter := interpreter.New()
	interpreter.Run(program)
}
