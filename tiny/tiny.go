package tiny

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
	imported map[string]*runtime.NameSpaceValue
}

func New() *Tiny {
	return &Tiny{builtins: &runtime.NameSpaceValue{"builtins", make(map[string]runtime.Value)}, imported: make(map[string]*runtime.NameSpaceValue)}
}

func (tiny *Tiny) AddNamespace(identifier string) {
	if _, ok := tiny.imported[identifier]; ok {
		shared.ReportErrFatal(fmt.Sprintf("Namespace '%s' already exists.", identifier))
	}

	tiny.imported[identifier] = &runtime.NameSpaceValue{identifier, make(map[string]runtime.Value)}
}

func (tiny *Tiny) AddClass(namespace string, identifier string, fields []string, methods map[string]*runtime.NativeFunctionValue) {
	tiny.checkId(namespace, identifier)
	tiny.imported[namespace].Members[identifier] = runtime.NewClassDefValue(identifier, fields, methods)
}

func (tiny *Tiny) AddFunction(namespace string, identifier string, params []string, fn runtime.NativeFn) {
	tiny.checkId(namespace, identifier)
	tiny.imported[namespace].Members[identifier] = runtime.NewFnValue(identifier, params, fn)
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
		analyser.DeclareNativeNs("builtins")

		for _, ns := range tiny.imported {
			analyser.DeclareNativeNs(ns.Identifier)
		}

		if !analyser.Run(program.Body) {
			return
		}

		if !checkOnly {
			interpreter := runtime.New()

			// Import native namespace into interpreter
			interpreter.Import("builtins", tiny.builtins)

			for _, ns := range tiny.imported {
				interpreter.Import(ns.Identifier, ns)
			}

			interpreter.Run(program)
		} else {
			fmt.Println("Good!")
		}
	} else {
		shared.ReportErrFatal(fmt.Sprintf("File '%s' does not exist.", os.Args[1]))
	}
}

// --- Private ---
func (tiny *Tiny) checkId(namespace string, identifier string) {
	if _, ok := tiny.imported[namespace]; !ok {
		shared.ReportErrFatal(fmt.Sprintf("Trying to add '%s' to unknown namespace '%s'", identifier, namespace))
	}

	if _, ok := tiny.imported[namespace].Members[identifier]; ok {
		shared.ReportErrFatal(fmt.Sprintf("Namespace '%s' already contains an item with identifier '%s'", namespace, identifier))
	}
}

func (tiny *Tiny) checkBuiltinId(identifier string) {
	if _, ok := tiny.builtins.Members[identifier]; ok {
		shared.ReportErrFatal(fmt.Sprintf("Identifier '%s' already exists in builtin namespace.", identifier))
	}
}

func (tiny *Tiny) addBuiltinClass(identifier string, fields []string, methods map[string]*runtime.NativeFunctionValue) {
	tiny.checkBuiltinId(identifier)
	tiny.builtins.Members[identifier] = runtime.NewClassDefValue(identifier, fields, methods)
}

func (tiny *Tiny) addBuiltinFn(identifier string, params []string, fn runtime.NativeFn) {
	tiny.checkBuiltinId(identifier)
	tiny.builtins.Members[identifier] = runtime.NewFnValue(identifier, params, fn)
}

func (tiny *Tiny) createBuiltins() {
	tiny.addBuiltinClass("test", []string{"x", "y"}, map[string]*runtime.NativeFunctionValue{
		"method": runtime.NewFnValue("method", []string{"a"}, func(interpreter *runtime.Interpreter, values []runtime.Value) runtime.Value {
			if len(values) == 1 {
				fmt.Printf("Test method %s\n", values[0].Inspect())
			}

			return &runtime.UnitVal{}
		}),
	})

	tiny.addBuiltinFn("read_file", []string{"fileName"}, func(interpreter *runtime.Interpreter, values []runtime.Value) runtime.Value {
		if fileName, ok := values[0].(*runtime.StringVal); !ok {
			interpreter.Report("Expected string as filename")
			return nil
		} else {
			return &runtime.StringVal{Value: shared.ReadFile(fileName.Value)}
		}
	})

	tiny.addBuiltinFn("write_file", []string{"fileName", "content"}, func(interpreter *runtime.Interpreter, values []runtime.Value) runtime.Value {
		if _, ok := values[0].(*runtime.StringVal); !ok {
			interpreter.Report("Expected string as filename")
			return nil
		}

		if _, ok := values[1].(*runtime.StringVal); !ok {
			interpreter.Report("Expected string as file contents")
			return nil
		}

		status := shared.WriteFile(values[0].(*runtime.StringVal).Value, values[1].(*runtime.StringVal).Value)
		return &runtime.BoolVal{Value: status}
	})
}
