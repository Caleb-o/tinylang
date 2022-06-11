package main

import (
	"fmt"
	"tiny/analysis"
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

	fmt.Println(program.Body.AsSExp())
}
