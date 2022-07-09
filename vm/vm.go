package vm

import (
	"fmt"
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
	chunk  *compiler.Chunk
	ip     int
	stack  []runtime.Value
	scope  []Scope
	frames []Frame
}

func NewVM(debug bool, chunk *compiler.Chunk) *VM {
	scopes := make([]Scope, 0, 1)
	scopes = append(scopes, Scope{make(map[string]runtime.Value)})

	return &VM{debug, chunk, 0, make([]runtime.Value, 0), scopes, make([]Frame, 0)}
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
		switch vm.chunk.Instructions[vm.ip] {
		case compiler.OpenScope:
			vm.begin()
			vm.ip++

		case compiler.CloseScope:
			vm.end()
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
			identifier := vm.chunk.Constants[vm.chunk.Instructions[vm.ip+2]]
			vm.push(vm.scope[vm.chunk.Instructions[vm.ip+1]].variables[identifier.Inspect()])
			vm.ip += 3

		case compiler.Set:
			identifier := vm.chunk.Constants[vm.chunk.Instructions[vm.ip+2]].Inspect()
			vm.scope[vm.chunk.Instructions[vm.ip+1]].variables[identifier] = vm.pop()
			vm.ip += 3

		case compiler.NewFn:
			arity := vm.chunk.Instructions[vm.ip+1]
			start := vm.chunk.Instructions[vm.ip+2]
			identifier := vm.chunk.Constants[vm.chunk.Instructions[vm.ip+3]].Inspect()

			// Assume global scope
			vm.scope[0].variables[identifier] = &runtime.CompiledFunctionValue{int(start), arity, nil}
			vm.ip += 4

		case compiler.Call:
			// Push current IP to stack
			scope := vm.chunk.Instructions[vm.ip+1]
			identifier := vm.chunk.Constants[vm.chunk.Instructions[vm.ip+2]].Inspect()

			fn, _ := vm.scope[scope].variables[identifier].(*runtime.CompiledFunctionValue)
			vm.newFrame(vm.ip+3, int(fn.Arity))

			vm.ip = fn.Start_ip

		case compiler.Return:
			frame := vm.dropFrame()

			// Remove stack values
			vm.stack = vm.stack[:frame.stack_start]
			vm.ip = frame.ret_to

		case compiler.Print:
			vm.print()
			vm.ip += 2

		case compiler.Jump:
			vm.ip = int(vm.chunk.Instructions[vm.ip+1])

		case compiler.JumpFalse:
			condition := vm.pop()

			value, _ := condition.(*runtime.BoolVal)

			if !value.Value {
				vm.ip = int(vm.chunk.Instructions[vm.ip+1])
				break
			}
			vm.ip += 2

		case compiler.Halt:
			vm.ip += 1

		default:
			vm.Report("Unknown operation in loop %d at position %d", vm.chunk.Instructions[vm.ip], vm.ip)
		}
	}
}

func (vm *VM) newFrame(return_to int, arity int) {
	vm.frames = append(vm.frames, Frame{return_to, len(vm.stack) - arity})

	if arity > 0 {
		start := len(vm.stack) - arity
		end := len(vm.stack)
		for idx := end - 1; idx >= start; idx-- {
			vm.push(vm.stack[idx])
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
}

func (vm *VM) pop() runtime.Value {
	value := vm.stack[len(vm.stack)-1]
	vm.stack = vm.stack[:len(vm.stack)-1]
	return value
}

func (vm *VM) print() {
	count := vm.chunk.Instructions[vm.ip+1]
	values := make([]runtime.Value, count)

	for i := int(count - 1); i >= 0; i-- {
		values[i] = vm.pop()
	}

	var sb strings.Builder

	for _, value := range values {
		sb.WriteString(value.Inspect())
	}

	fmt.Println(sb.String())
}
