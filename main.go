package main

import (
	"fmt"
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
		shared.ReportErrFatal("Failed to analyse")
	}

	interpreter := runtime.New()
	interpreter.Run(program)

	fmt.Println(program.Body.AsSExp())
}
