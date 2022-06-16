package main

import (
	"tiny/analysis"
	"tiny/parser"
	"tiny/runtime"
	"tiny/shared"
)

func main() {
	parser := parser.New(shared.ReadFile("./experimenting.tiny"))
	program := parser.Parse()
	analyser := analysis.NewAnalyser(false)

	if !analyser.Run(program.Body) {
		return
	}

	interpreter := runtime.New()
	interpreter.Run(program)
}
