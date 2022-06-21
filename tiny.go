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

type Tiny struct {
	builtins *runtime.NameSpaceValue
}

func New() *Tiny {
	return &Tiny{builtins: &runtime.NameSpaceValue{"builtins", make(map[string]runtime.Value)}}
}

func (tiny *Tiny) AddFn(identifier string, params []string, fn runtime.NativeFn) {
	tiny.builtins.Members[identifier] = runtime.NewFnValue(identifier, params, fn)
}

func (tiny *Tiny) createBuiltins() {
	tiny.AddFn("read_file", []string{"fileName"}, func(interpreter *runtime.Interpreter, values []runtime.Value) runtime.Value {
		if fileName, ok := values[0].(*runtime.StringVal); !ok {
			interpreter.Report("Expected string as filename")
			return nil
		} else {
			return &runtime.StringVal{Value: shared.ReadFile(fileName.Value)}
		}
	})
}

func (tiny *Tiny) Run() {
	var (
		checkOnly bool
		// usevm     bool
		script string
	)

	flag.BoolVar(&checkOnly, "check", false, "Checks if code is valid and does not run")
	// flag.BoolVar(&usevm, "vm", false, "Use the new interpreter to run code")
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

		tiny.createBuiltins()

		// Import functions into analyser
		analyser.DeclareNativeNs()

		for _, builtin := range tiny.builtins.Members {
			switch t := builtin.(type) {
			case *runtime.NativeFunctionValue:
				analyser.DeclareNativeFn(t.Identifier, t.Params)
			}
		}

		if !analyser.Run(program.Body) {
			return
		}

		if !checkOnly {
			interpreter := runtime.New()

			// Import native namespace into interpreter
			interpreter.Import("builtins", tiny.builtins)

			interpreter.Run(program)
		} else {
			fmt.Println("Good!")
		}
	} else {
		shared.ReportErrFatal(fmt.Sprintf("File '%s' does not exist.", os.Args[1]))
	}
}
