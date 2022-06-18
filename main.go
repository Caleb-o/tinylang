package main

import (
	"fmt"
	"os"
	"tiny/analysis"
	"tiny/parser"
	"tiny/runtime"
	"tiny/shared"
)

func main() {
	if len(os.Args) > 2 || len(os.Args) < 2 {
		fmt.Println("usage: tiny script")
		return
	}

	if source, ok := shared.ReadFileErr(os.Args[1]); ok {
		parser := parser.New(source)
		program := parser.Parse()
		analyser := analysis.NewAnalyser(false)

		if !analyser.Run(program.Body) {
			return
		}

		interpreter := runtime.New()
		interpreter.Run(program)

		fmt.Println(program.Body.AsSExp())
	} else {
		shared.ReportErrFatal(fmt.Sprintf("File '%s' does not exist.", os.Args[1]))
	}
}
