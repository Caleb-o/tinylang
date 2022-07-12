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

const (
	STACK_MAX int = 256
)

type Frame struct {
	ret_to      int
	stack_start int
}

type VM struct {
	debug   bool
	step    bool
	chunk   *compiler.Chunk
	ip      int
	sp      int
	globals map[string]runtime.Value
	stack   []runtime.Value
	frames  []Frame
}

func NewVM(debug bool, step bool, chunk *compiler.Chunk) *VM {
	return &VM{debug, step, chunk, 0, 0, make(map[string]runtime.Value, 32), make([]runtime.Value, STACK_MAX), make([]Frame, 0, STACK_MAX)}
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
		case compiler.Push:
			vm.push(vm.chunk.Constants[vm.chunk.Instructions[vm.ip+1]])
			vm.ip += 2

		case compiler.Pop:
			vm.sp--
			vm.ip++

		case compiler.PopN:
			count := int(vm.chunk.Instructions[vm.ip+1])
			vm.stack = vm.stack[:vm.sp-count]
			vm.sp -= count
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
			vm.compare(Less)
			vm.ip++

		case compiler.LessEq:
			vm.compare(LessEq)
			vm.ip++

		case compiler.Greater:
			vm.compare(Greater)
			vm.ip++

		case compiler.GreaterEq:
			vm.compare(GreaterEq)
			vm.ip++

		case compiler.EqEq:
			vm.compare(EqualEqual)
			vm.ip++

		case compiler.NotEq:
			vm.compare(NotEqual)
			vm.ip++

		case compiler.Get:
			identifier := vm.chunk.Constants[vm.chunk.Instructions[vm.ip+1]].Inspect()
			vm.push(vm.globals[identifier])
			vm.ip += 2

		case compiler.Set:
			identifier := vm.chunk.Constants[vm.chunk.Instructions[vm.ip+1]].Inspect()
			vm.globals[identifier] = vm.pop()
			vm.ip += 2

		case compiler.GetLocal:
			vm.push(vm.stack[vm.frames[len(vm.frames)-1].stack_start+int(vm.chunk.Instructions[vm.ip+1])])
			vm.ip += 2

		case compiler.SetLocal:
			idx := vm.frames[len(vm.frames)-1].stack_start + int(vm.chunk.Instructions[vm.ip+1])
			if vm.sp != idx {
				vm.stack[idx] = vm.peek()
			}
			vm.ip += 2

		case compiler.NewFn:
			arity := vm.chunk.Instructions[vm.ip+1]
			start := vm.chunk.Instructions[vm.ip+2]
			identifier := vm.chunk.Constants[vm.chunk.Instructions[vm.ip+3]].Inspect()

			vm.globals[identifier] = &runtime.CompiledFunctionValue{int(start), arity, nil}
			vm.ip += 4

		case compiler.NewAnonFn:
			arity := vm.chunk.Instructions[vm.ip+1]
			start := vm.chunk.Instructions[vm.ip+2]

			vm.push(&runtime.CompiledFunctionValue{int(start), arity, nil})
			vm.ip += 3

		case compiler.Call:
			identifier := vm.chunk.Constants[vm.chunk.Instructions[vm.ip+1]].Inspect()
			fn := vm.globals[identifier].(*runtime.CompiledFunctionValue)

			// TODO: See if this works as intended in more complex cases
			if vm.sp-vm.frames[len(vm.frames)-1].stack_start < int(fn.Arity) {
				vm.Report("Function '%s' expected %d argument(s) but receieved %d", identifier, fn.Arity, vm.sp-vm.frames[len(vm.frames)-1].stack_start)
			}

			vm.newFrame(vm.ip+2, int(fn.Arity))
			vm.ip = fn.Start_ip

		case compiler.Return:
			frame := vm.dropFrame()

			var retValue runtime.Value = nil

			if vm.sp > frame.stack_start {
				retValue = vm.pop()
			}

			// Remove stack values
			vm.sp = frame.stack_start

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
				vm.frames = make([]Frame, STACK_MAX)
				vm.globals = make(map[string]runtime.Value, 32)
				vm.newFrame(-1, 0)
				vm.sp = 0
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

	sb.WriteString("=== Globals ===\n")
	idx := 0
	for id, global := range vm.globals {
		sb.WriteString(fmt.Sprintf("%s = '%s'\n", id, global.Inspect()))

		if idx >= 20 {
			sb.WriteString("...\n")
			break
		}
		idx += 1
	}
	sb.WriteByte('\n')

	sb.WriteString("=== Constants ===\n")
	for idx, constant := range vm.chunk.Constants {
		sb.WriteString(fmt.Sprintf("%d: '%s'\n", idx, constant.Inspect()))

		if idx >= 20 {
			sb.WriteString("...\n")
			break
		}
	}
	sb.WriteByte('\n')

	sb.WriteString("Stack [")

	if vm.sp > 20 {
		sb.WriteString("..., ")
	}

	for idx := 0; idx < vm.sp; idx++ {
		sb.WriteString(fmt.Sprintf("'%s'", vm.stack[idx].Inspect()))

		if idx < vm.sp-1 {
			sb.WriteString(", ")
		}
	}

	sb.WriteByte(']')
	sb.WriteByte('\n')

	fmt.Printf("\033c%s\n", sb.String())
}

func (vm *VM) newFrame(return_to int, arity int) {
	if len(vm.frames) >= STACK_MAX {
		vm.Report("Stack overflow")
	}
	vm.frames = append(vm.frames, Frame{return_to, vm.sp - arity})
}

func (vm *VM) dropFrame() Frame {
	frame := vm.frames[len(vm.frames)-1]
	vm.frames = vm.frames[:len(vm.frames)-1]
	return frame
}

func (vm *VM) binaryOp(operation binaryOp) {
	right := vm.pop()
	// This shouldn't need to be here, as below, but it's still faster
	index := vm.sp - 1
	left := vm.stack[index].Copy()

	if reflect.TypeOf(left) != reflect.TypeOf(right) {
		vm.Report("Invalid binary operation '%s:%s %s %s:%s'", left.Inspect(), reflect.TypeOf(left), operation.Operator(), right.Inspect(), reflect.TypeOf(right))
		return
	}

	if !left.Modify(operation.ToKind(), right) {
		vm.Report("Value of type '%s' cannot be modified", reflect.TypeOf(left))
	}

	// This shouldn't need to be here, but it works
	vm.stack[index] = left
}

func (vm *VM) compare(operation binaryOp) {
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
	}
}

func (vm *VM) push(value runtime.Value) {
	if vm.sp+1 >= STACK_MAX {
		vm.Report("Stack overflow")
	}

	vm.stack[vm.sp] = value
	vm.sp++
}

func (vm *VM) pop() runtime.Value {
	value := vm.stack[vm.sp-1]
	vm.sp--
	return value
}

func (vm *VM) peek() runtime.Value {
	return vm.stack[vm.sp-1].Copy()
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
