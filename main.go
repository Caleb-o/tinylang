package main

import (
	"flag"
	"fmt"
	"os"
	"tiny/analysis"
	"tiny/parser"
	"tiny/runtime"
	"tiny/shared"
)

func main() {
	var (
		checkOnly bool
		script    string
	)

	flag.BoolVar(&checkOnly, "check", false, "Checks if code is valid and does not run")
	flag.StringVar(&script, "script", "", "Script to run")
	flag.Parse()

	if len(script) == 0 {
		fmt.Println("usage: tiny [-script][-check]")
		return
	}

	if source, ok := shared.ReadFileErr(script); ok {

		parser := parser.New(source)
		program := parser.Parse()
		analyser := analysis.NewAnalyser(false)

		if !analyser.Run(program.Body) {
			return
		}

		if !checkOnly {
			interpreter := runtime.New()
			interpreter.Run(program)
		} else {
			fmt.Println("Good!")
		}
	} else {
		shared.ReportErrFatal(fmt.Sprintf("File '%s' does not exist.", os.Args[1]))
	}
}
