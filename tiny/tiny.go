package tiny

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
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
	return &Tiny{builtins: &runtime.NameSpaceValue{"builtin", make(map[string]runtime.Value)}, imported: make(map[string]*runtime.NameSpaceValue)}
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
		analyser.DeclareNativeNs(tiny.builtins.Identifier)

		for _, ns := range tiny.imported {
			analyser.DeclareNativeNs(ns.Identifier)
		}

		if !analyser.Run(program.Body) {
			return
		}

		if !checkOnly {
			interpreter := runtime.New()

			// Import native namespace into interpreter
			interpreter.Import(tiny.builtins.Identifier, tiny.builtins)

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
	// --- Error Handling
	tiny.addBuiltinFn("assert", []string{"expr"}, func(interpreter *runtime.Interpreter, values []runtime.Value) runtime.Value {
		if value, ok := values[0].(*runtime.BoolVal); ok {
			if !value.Value {
				interpreter.Report("Assertion failed")
			}
		} else {
			interpreter.Report("assert expected an expression resulting in a boolean, as the first argument.")
		}

		return &runtime.UnitVal{}
	})

	tiny.addBuiltinFn("assertm", []string{"expr", "message"}, func(interpreter *runtime.Interpreter, values []runtime.Value) runtime.Value {
		if _, ok := values[1].(*runtime.StringVal); !ok {
			interpreter.Report("assertm expected a message as the second argument.")
		}

		if value, ok := values[0].(*runtime.BoolVal); ok {
			if !value.Value {
				interpreter.Report("Assertion failed: '%s'", values[1].(*runtime.StringVal).Value)
			}
		} else {
			interpreter.Report("assertm expected an expression resulting in a boolean, as the first argument.")
		}

		return &runtime.UnitVal{}
	})

	// --- Inspection
	tiny.addBuiltinFn("arg_count", []string{"object"}, func(interpreter *runtime.Interpreter, values []runtime.Value) runtime.Value {
		if value, ok := values[0].(runtime.TinyCallable); ok {
			return &runtime.IntVal{Value: value.Arity()}
		}

		return &runtime.IntVal{Value: 0}
	})

	tiny.addBuiltinFn("type_name", []string{"object"}, func(interpreter *runtime.Interpreter, values []runtime.Value) runtime.Value {
		switch obj := values[0].(type) {
		case *runtime.UnitVal:
			return &runtime.StringVal{Value: "unit"}
		case *runtime.IntVal:
			return &runtime.StringVal{Value: "int"}
		case *runtime.FloatVal:
			return &runtime.StringVal{Value: "float"}
		case *runtime.BoolVal:
			return &runtime.StringVal{Value: "bool"}
		case *runtime.StringVal:
			return &runtime.StringVal{Value: "string"}
		case *runtime.FunctionValue:
			return &runtime.StringVal{Value: "function"}
		case *runtime.AnonFunctionValue:
			return &runtime.StringVal{Value: "anon fn"}
		case *runtime.NativeFunctionValue:
			return &runtime.StringVal{Value: "native fn"}
		case *runtime.NativeClassDefValue:
			return &runtime.StringVal{Value: "native class"}
		case *runtime.ClassDefValue:
			return &runtime.StringVal{Value: "class"}
		case *runtime.ClassInstanceValue:
			return &runtime.StringVal{Value: obj.Definition()}
		case *runtime.StructDefValue:
			return &runtime.StringVal{Value: "struct"}
		case *runtime.StructInstanceValue:
			return &runtime.StringVal{Value: obj.Definition()}
		case *runtime.NameSpaceValue:
			return &runtime.StringVal{Value: obj.Identifier}
		case *runtime.ListVal:
			return &runtime.StringVal{Value: "list"}
		}

		return &runtime.StringVal{Value: "unknown"}
	})

	tiny.addBuiltinFn("is_callable", []string{"object"}, func(interpreter *runtime.Interpreter, values []runtime.Value) runtime.Value {
		if _, ok := values[0].(runtime.TinyCallable); ok {
			return &runtime.BoolVal{Value: true}
		}
		return &runtime.BoolVal{Value: false}
	})

	tiny.addBuiltinFn("has_field", []string{"object", "fieldName"}, func(interpreter *runtime.Interpreter, values []runtime.Value) runtime.Value {
		var str string

		if value, ok := values[1].(*runtime.StringVal); !ok {
			interpreter.Report("Expected property name to be string")
			return nil
		} else {
			str = value.Value
		}

		switch obj := values[0].(type) {
		case *runtime.ClassDefValue:
			return &runtime.BoolVal{Value: obj.HasField(str)}
		case *runtime.StructDefValue:
			return &runtime.BoolVal{Value: obj.HasField(str)}
		case *runtime.ClassInstanceValue:
			if def, ok := obj.Def.(*runtime.ClassDefValue); ok {
				return &runtime.BoolVal{Value: def.HasField(str)}
			} else if def, ok := obj.Def.(*runtime.NativeClassDefValue); ok {
				return &runtime.BoolVal{Value: def.HasField(str)}
			}
		case *runtime.StructInstanceValue:
			return &runtime.BoolVal{Value: obj.HasField(str)}
		}

		return &runtime.BoolVal{Value: false}
	})

	// --- IO
	tiny.addBuiltinFn("read_line", []string{}, func(interpreter *runtime.Interpreter, values []runtime.Value) runtime.Value {
		txt, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		return &runtime.StringVal{Value: strings.TrimSpace(txt)}
	})

	tiny.addBuiltinFn("prompt_read_line", []string{"prompt"}, func(interpreter *runtime.Interpreter, values []runtime.Value) runtime.Value {
		fmt.Print(values[0].Inspect())
		txt, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		return &runtime.StringVal{Value: strings.TrimSpace(txt)}
	})

	tiny.addBuiltinFn("read_file", []string{"fileName"}, func(interpreter *runtime.Interpreter, values []runtime.Value) runtime.Value {
		if fileName, ok := values[0].(*runtime.StringVal); !ok {
			interpreter.Report("Expected string as filename")
			return nil
		} else {
			return &runtime.StringVal{Value: shared.ReadFile(fileName.Value)}
		}
	})

	tiny.addBuiltinFn("delete_file", []string{"fileName"}, func(interpreter *runtime.Interpreter, values []runtime.Value) runtime.Value {
		if fileName, ok := values[0].(*runtime.StringVal); !ok {
			interpreter.Report("Expected string as filename")
			return nil
		} else {
			return &runtime.BoolVal{Value: shared.DeleteFile(fileName.Value)}
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

	// --- Converters
	tiny.addBuiltinFn("to_int", []string{"value"}, func(interpreter *runtime.Interpreter, values []runtime.Value) runtime.Value {
		switch value := values[0].(type) {
		case *runtime.IntVal:
			return value
		case *runtime.FloatVal:
			return &runtime.IntVal{Value: int(value.Value)}
		case *runtime.BoolVal:
			out := 0

			if value.Value {
				out = 1
			}

			return &runtime.IntVal{Value: out}

		case *runtime.StringVal:
			number, err := strconv.ParseInt(value.Value, 10, 32)

			if err != nil {
				return runtime.NewThrow(&runtime.StringVal{Value: fmt.Sprintf("Could not convert '%s' to int", values[0].Inspect())})
			}

			return &runtime.IntVal{Value: int(number)}
		}

		return runtime.NewThrow(&runtime.StringVal{Value: fmt.Sprintf("Could not convert '%s' to int", values[0].Inspect())})
	})

	tiny.addBuiltinFn("to_float", []string{"value"}, func(interpreter *runtime.Interpreter, values []runtime.Value) runtime.Value {
		switch value := values[0].(type) {
		case *runtime.IntVal:
			return &runtime.FloatVal{Value: float32(value.Value)}
		case *runtime.FloatVal:
			return value
		case *runtime.BoolVal:
			var out float32 = 0.0

			if value.Value {
				out = 1.0
			}

			return &runtime.FloatVal{Value: out}

		case *runtime.StringVal:
			number, err := strconv.ParseFloat(value.Value, 32)
			if err != nil {
				return runtime.NewThrow(&runtime.StringVal{Value: fmt.Sprintf("Could not convert '%s' to float", values[0].Inspect())})
			}
			return &runtime.FloatVal{Value: float32(number)}
		}

		return runtime.NewThrow(&runtime.StringVal{Value: fmt.Sprintf("Could not convert '%s' to float", values[0].Inspect())})
	})

	tiny.addBuiltinFn("to_string", []string{"value"}, func(interpreter *runtime.Interpreter, values []runtime.Value) runtime.Value {
		switch value := values[0].(type) {
		case *runtime.IntVal:
			return &runtime.StringVal{Value: value.Inspect()}
		case *runtime.FloatVal:
			return &runtime.StringVal{Value: value.Inspect()}
		case *runtime.BoolVal:
			return &runtime.StringVal{Value: value.Inspect()}
		case *runtime.StringVal:
			return value
		}

		return runtime.NewThrow(&runtime.StringVal{Value: fmt.Sprintf("Could not convert '%s' to string", values[0].Inspect())})
	})

	// --- Misc
	tiny.addBuiltinFn("append", []string{"list", "value"}, func(interpreter *runtime.Interpreter, values []runtime.Value) runtime.Value {
		if _, ok := values[0].(*runtime.ListVal); !ok {
			interpreter.Report("Cannot append to non-list")
			return nil
		}

		list := values[0].(*runtime.ListVal)
		list.Values = append(list.Values, values[1].Copy())

		return &runtime.UnitVal{}
	})

	tiny.addBuiltinFn("pop", []string{"list"}, func(interpreter *runtime.Interpreter, values []runtime.Value) runtime.Value {
		if _, ok := values[0].(*runtime.ListVal); !ok {
			interpreter.Report("Cannot pop non-list")
			return nil
		}

		list := values[0].(*runtime.ListVal)

		if len(list.Values) == 0 {
			return &runtime.UnitVal{}
		}

		value := list.Values[len(list.Values)-1].Copy()
		list.Values = list.Values[:len(list.Values)-1]

		return value
	})

	tiny.addBuiltinFn("is_err", []string{"value"}, func(interpreter *runtime.Interpreter, values []runtime.Value) runtime.Value {
		_, ok := values[0].(*runtime.ThrowValue)
		return &runtime.BoolVal{Value: ok}
	})

	tiny.addBuiltinFn("get_err", []string{"value"}, func(interpreter *runtime.Interpreter, values []runtime.Value) runtime.Value {
		if err, ok := values[0].(*runtime.ThrowValue); ok {
			return err.GetInner()
		}
		return &runtime.UnitVal{}
	})

	tiny.addBuiltinFn("len", []string{"value"}, func(interpreter *runtime.Interpreter, values []runtime.Value) runtime.Value {
		switch value := values[0].(type) {
		case *runtime.StringVal:
			return &runtime.IntVal{Value: len(value.Value)}
		case *runtime.ListVal:
			return &runtime.IntVal{Value: len(value.Values)}
		}

		return &runtime.IntVal{Value: 0}
	})

	tiny.addBuiltinFn("mod", []string{"x", "y"}, func(interpreter *runtime.Interpreter, values []runtime.Value) runtime.Value {
		if _, ok := values[0].(*runtime.IntVal); !ok {
			interpreter.Report("Expected int as dividend")
			return nil
		}

		if _, ok := values[1].(*runtime.IntVal); !ok {
			interpreter.Report("Expected int as divisor")
			return nil
		}

		return &runtime.IntVal{Value: values[0].(*runtime.IntVal).Value % values[1].(*runtime.IntVal).Value}
	})

	// Set the seed
	rand.Seed(time.Now().UnixNano())

	tiny.addBuiltinFn("rand", []string{"max"}, func(interpreter *runtime.Interpreter, values []runtime.Value) runtime.Value {
		if _, ok := values[0].(*runtime.IntVal); !ok {
			interpreter.Report("Expected int as max")
			return nil
		}

		return &runtime.IntVal{Value: rand.Intn(values[0].(*runtime.IntVal).Value)}
	})

	tiny.addBuiltinFn("rand_range", []string{"min", "max"}, func(interpreter *runtime.Interpreter, values []runtime.Value) runtime.Value {
		if _, ok := values[0].(*runtime.IntVal); !ok {
			interpreter.Report("Expected int as min")
			return nil
		}

		if _, ok := values[1].(*runtime.IntVal); !ok {
			interpreter.Report("Expected int as max")
			return nil
		}

		min := values[0].(*runtime.IntVal).Value
		max := values[1].(*runtime.IntVal).Value

		return &runtime.IntVal{Value: rand.Intn(max-min+1) + min}
	})
}
