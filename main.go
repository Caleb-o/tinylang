package main

import (
	"flag"
	"fmt"
	"os"
	"tiny/analysis"
	"tiny/compiler"
	"tiny/parser"
	"tiny/runtime"
	"tiny/shared"
	"tiny/vm"
)

func main() {
	var (
		checkOnly bool
		usevm     bool
		script    string
	)

	flag.BoolVar(&checkOnly, "check", false, "Checks if code is valid and does not run")
	flag.BoolVar(&usevm, "vm", false, "Use the new interpreter to run code")
	flag.StringVar(&script, "script", "", "Script to run")
	flag.Parse()

	if len(script) == 0 {
		fmt.Println("usage: tiny [-script][-check]")
		return
	}

	if source, ok := shared.ReadFileErr(script); ok {

		parser := parser.New(source, script)
		program := parser.Parse()
		analyser := analysis.NewAnalyser(false)

		if !analyser.Run(program.Body) {
			return
		}

		if !checkOnly {
			if !usevm {
				interpreter := runtime.New()
				interpreter.Run(program)
			} else {
				chunk := compiler.New().Compile(program)
				machine := vm.New(chunk)
				machine.Run()
			}
		} else {
			fmt.Println("Good!")
		}
	} else {
		shared.ReportErrFatal(fmt.Sprintf("File '%s' does not exist.", os.Args[1]))
	}
}
