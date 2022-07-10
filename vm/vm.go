package vm

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strings"
	"tiny/compiler"
	"tiny/runtime"
	"tiny/shared"
)

type Scope struct {
	variables map[string]runtime.Value
}

type Frame struct {
	ret_to      int
	stack_start int
}

type VM struct {
	debug  bool
	step   bool
	chunk  *compiler.Chunk
	ip     int
	stack  []runtime.Value
	scope  []Scope
	frames []Frame
}

func NewVM(debug bool, step bool, chunk *compiler.Chunk) *VM {
	scopes := make([]Scope, 0, 1)
	scopes = append(scopes, Scope{make(map[string]runtime.Value)})

	return &VM{debug, step, chunk, 0, make([]runtime.Value, 0), scopes, make([]Frame, 0)}
}

func (vm *VM) Report(msg string, args ...any) {
	res := fmt.Sprintf(msg, args...)
	shared.ReportErrFatal("Runtime: " + res)
}

func (vm *VM) Run() {
	if vm.debug {
		vm.chunk.Debug()
	}

	vm.newFrame(-1, 0)
	defer vm.dropFrame()

	for vm.ip < len(vm.chunk.Instructions) {
		last := vm.ip

		switch vm.chunk.Instructions[vm.ip] {
		case compiler.OpenScope:
			// vm.begin()
			vm.ip++

		case compiler.CloseScope:
			// vm.end()
			vm.ip++

		case compiler.Push:
			vm.push(vm.chunk.Constants[vm.chunk.Instructions[vm.ip+1]])
			vm.ip += 2

		case compiler.Add:
			vm.binaryOp(Add)
			vm.ip++

		case compiler.Sub:
			vm.binaryOp(Sub)
			vm.ip++

		case compiler.Mul:
			vm.binaryOp(Mul)
			vm.ip++

		case compiler.Div:
			vm.binaryOp(Div)
			vm.ip++

		case compiler.Less:
			vm.binaryOp(Less)
			vm.ip++

		case compiler.LessEq:
			vm.binaryOp(LessEq)
			vm.ip++

		case compiler.Greater:
			vm.binaryOp(Greater)
			vm.ip++

		case compiler.GreaterEq:
			vm.binaryOp(GreaterEq)
			vm.ip++

		case compiler.EqEq:
			vm.binaryOp(EqualEqual)
			vm.ip++

		case compiler.NotEq:
			vm.binaryOp(NotEqual)
			vm.ip++

		case compiler.Get:
			identifier := vm.chunk.Constants[vm.chunk.Instructions[vm.ip+1]].Inspect()
			vm.push(vm.scope[0].variables[identifier])
			vm.ip += 2

		case compiler.Set:
			identifier := vm.chunk.Constants[vm.chunk.Instructions[vm.ip+1]].Inspect()
			vm.scope[0].variables[identifier] = vm.pop()
			vm.ip += 2

		case compiler.GetLocal:
			identifier := vm.chunk.Constants[vm.chunk.Instructions[vm.ip+1]].Inspect()

			for idx := len(vm.scope) - 1; idx >= 0; idx-- {
				if value, ok := vm.scope[idx].variables[identifier]; ok {
					vm.push(value)
					break
				}
			}
			vm.ip += 2

		case compiler.SetLocal:
			identifier := vm.chunk.Constants[vm.chunk.Instructions[vm.ip+1]].Inspect()
			for idx := len(vm.scope) - 1; idx >= 0; idx-- {
				if _, ok := vm.scope[idx].variables[identifier]; ok {
					vm.scope[idx].variables[identifier] = vm.pop()
					break
				}
			}
			vm.ip += 2

		case compiler.Define:
			identifier := vm.chunk.Constants[vm.chunk.Instructions[vm.ip+1]].Inspect()
			vm.scope[len(vm.scope)-1].variables[identifier] = vm.pop()
			vm.ip += 2

		case compiler.NewFn:
			arity := vm.chunk.Instructions[vm.ip+1]
			start := vm.chunk.Instructions[vm.ip+2]
			identifier := vm.chunk.Constants[vm.chunk.Instructions[vm.ip+3]].Inspect()

			// Assume global scope as functions cannot be defined in non-global scope
			vm.scope[0].variables[identifier] = &runtime.CompiledFunctionValue{int(start), arity, nil}
			vm.ip += 4

		case compiler.Call:
			// Push current IP to stack
			scope := vm.chunk.Instructions[vm.ip+1]
			identifier := vm.chunk.Constants[vm.chunk.Instructions[vm.ip+2]].Inspect()
			fn := vm.scope[scope].variables[identifier].(*runtime.CompiledFunctionValue)

			vm.begin()
			vm.newFrame(vm.ip+3, int(fn.Arity))

			vm.ip = fn.Start_ip

		case compiler.Return:
			frame := vm.dropFrame()

			var retValue runtime.Value = nil

			if len(vm.stack) > frame.stack_start {
				retValue = vm.pop()
			}

			// Remove stack values
			vm.stack = vm.stack[:frame.stack_start]
			vm.end()

			if retValue != nil {
				vm.push(retValue)
			}
			vm.ip = frame.ret_to

		case compiler.Print:
			vm.print()
			vm.ip += 2

		case compiler.Jump:
			vm.ip = int(vm.chunk.Instructions[vm.ip+1])

		case compiler.JumpFalse:
			condition := vm.pop()

			value := condition.(*runtime.BoolVal)

			if !value.Value {
				vm.ip = int(vm.chunk.Instructions[vm.ip+1])
			} else {
				vm.ip += 2
			}

		case compiler.Halt:
			vm.ip += 1

		default:
			vm.Report("Unknown operation in loop '%d' at position %d", vm.chunk.Instructions[vm.ip], vm.ip)
		}

		if vm.step {
			vm.printStepInfo(last)
			fmt.Print(">> ")
			raw, _ := bufio.NewReader(os.Stdin).ReadString('\n')

			text := strings.TrimSpace(raw)

			switch text {
			case "reset":
				vm.ip = 0
				vm.frames = make([]Frame, 1)
				vm.newFrame(-1, 0)
				vm.scope = make([]Scope, 0)
				vm.scope = append(vm.scope, Scope{make(map[string]runtime.Value)})
				vm.stack = make([]runtime.Value, 0)
			case "exit":
				return
			}

		}
	}
}

func (vm *VM) printStepInfo(last int) {
	var sb strings.Builder

	sb.WriteString("== General ==\n")
	sb.WriteString("[Now]  ")
	vm.chunk.PrintInstruction(&sb, last, vm.chunk.Instructions)

	if vm.ip < len(vm.chunk.Instructions) {
		sb.WriteString("[Next] ")
		vm.chunk.PrintInstruction(&sb, vm.ip, vm.chunk.Instructions)
		sb.WriteByte('\n')
	} else {
		sb.WriteString("-- End --\n")
	}

	count := 0
	for idx := len(vm.scope) - 1; idx >= 0; idx-- {
		sb.WriteString(fmt.Sprintf("== Scope %d ==\n", idx))
		for field, variable := range vm.scope[idx].variables {
			sb.WriteString(fmt.Sprintf("  '%s' = %s\n", field, variable.Inspect()))
		}

		if len(vm.scope[idx].variables) == 0 {
			sb.WriteString("EMPTY\n")
		}

		count += 1
		if count == 5 {
			if len(vm.scope) > 5 {
				sb.WriteString("...\n")
			}
			break
		}
	}
	sb.WriteByte('\n')
	sb.WriteByte('\n')

	sb.WriteString("Stack [")

	if len(vm.stack) > 20 {
		sb.WriteString("..., ")
	}

	count = 0

	if len(vm.stack) >= 20 {
		count = 10
	}

	for idx := count; idx < len(vm.stack); idx++ {
		sb.WriteString(vm.stack[idx].Inspect())
		count += 1

		if count == 20 {
			break
		}

		if idx < len(vm.stack)-1 {
			sb.WriteString(", ")
		}
	}

	sb.WriteByte(']')
	sb.WriteByte('\n')

	fmt.Printf("\033c%s\n", sb.String())
}

func (vm *VM) newFrame(return_to int, arity int) {
	vm.frames = append(vm.frames, Frame{return_to, len(vm.stack) - arity})

	if arity > 0 {
		new_stack := make([]runtime.Value, 0)

		for idx := 0; idx < arity; idx++ {
			new_stack = append(new_stack, vm.pop())
		}

		for idx := arity - 1; idx >= 0; idx-- {
			vm.push(new_stack[idx])
		}
	}
}

func (vm *VM) dropFrame() Frame {
	frame := vm.frames[len(vm.frames)-1]
	vm.frames = vm.frames[:len(vm.frames)-1]
	return frame
}

func (vm *VM) begin() {
	vm.scope = append(vm.scope, Scope{make(map[string]runtime.Value)})
}

func (vm *VM) end() {
	vm.scope = vm.scope[:len(vm.scope)-1]
}

func (vm *VM) binaryOp(operation binaryOp) {
	right := vm.pop()
	left := vm.pop()

	if reflect.TypeOf(left) != reflect.TypeOf(right) {
		fmt.Printf("Pos %d \n", vm.ip)
		vm.Report("Invalid binary operation '%s:%s %s %s:%s'", left.Inspect(), reflect.TypeOf(left), operation.Operator(), right.Inspect(), reflect.TypeOf(right))
		return
	}

	switch left.(type) {
	case *runtime.IntVal:
		if value, ok := runtime.BinopI(operation.ToKind(), left.(*runtime.IntVal).Value, right.(*runtime.IntVal).Value); ok {
			vm.push(value)
		}

	case *runtime.FloatVal:
		if value, ok := runtime.BinopF(operation.ToKind(), left.(*runtime.FloatVal).Value, right.(*runtime.FloatVal).Value); ok {
			vm.push(value)
		}

	case *runtime.BoolVal:
		if value, ok := runtime.BinopB(operation.ToKind(), left.(*runtime.BoolVal).Value, right.(*runtime.BoolVal).Value); ok {
			vm.push(value)
		}

	case *runtime.StringVal:
		if value, ok := runtime.BinopS(operation.ToKind(), left.(*runtime.StringVal).Value, right.(*runtime.StringVal).Value); ok {
			vm.push(value)
		}

		// case *ListVal:
		// 	if value, ok := BinopL(operation.ToKind(), left.(*ListVal).Values, right.(*ListVal).Values); ok {
		// 		vm.push(value)
		// 	}
	}
}

func (vm *VM) push(value runtime.Value) {
	vm.stack = append(vm.stack, value)

	// if vm.debug {
	// 	var sb strings.Builder

	// 	for idx, value := range vm.stack {
	// 		sb.WriteString(value.Inspect())

	// 		if idx < len(vm.stack)-1 {
	// 			sb.WriteString(", ")
	// 		}
	// 	}

	// 	fmt.Printf("%d: Push [%s]\n", vm.ip, sb.String())
	// 	bufio.NewReader(os.Stdin).ReadString('\n')
	// }
}

func (vm *VM) pop() runtime.Value {
	value := vm.stack[len(vm.stack)-1]
	vm.stack = vm.stack[:len(vm.stack)-1]

	// if vm.debug {
	// 	var sb strings.Builder

	// 	for idx, value := range vm.stack {
	// 		sb.WriteString(value.Inspect())

	// 		if idx < len(vm.stack)-1 {
	// 			sb.WriteString(", ")
	// 		}
	// 	}

	// 	fmt.Printf("%d: Pop [%s]\n", vm.ip, sb.String())
	// 	bufio.NewReader(os.Stdin).ReadString('\n')
	// }
	return value
}

func (vm *VM) print() {
	count := vm.chunk.Instructions[vm.ip+1]
	values := make([]runtime.Value, count)

	for idx := int(count - 1); idx >= 0; idx-- {
		values[idx] = vm.pop()
	}

	var sb strings.Builder

	for _, value := range values {
		sb.WriteString(value.Inspect())
	}

	fmt.Println(sb.String())
}
