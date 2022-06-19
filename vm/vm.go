package vm

import (
	"fmt"
	"reflect"
	"tiny/compiler"
	"tiny/lexer"
	"tiny/runtime"
	"tiny/shared"
)

type Frame struct {
	data map[string]runtime.Value
}

type VM struct {
	stack  []runtime.Value
	frames []Frame
	ip     int
	code   *compiler.Chunk
}

func New(code *compiler.Chunk) *VM {
	return &VM{stack: make([]runtime.Value, 0), frames: make([]Frame, 0), ip: 0, code: code}
}

func (vm *VM) Run() {
	for {
		switch node := vm.code.Code[vm.ip].(type) {
		case *compiler.Literal:
			vm.push(node.Data)

		case *compiler.BinOp:
			vm.transform(node.Kind, vm.pop())

		case *compiler.Print:
			vm.visitPrint(node.Count)

		case *compiler.Halt:
			return
		}

		vm.ip++
	}
}

func (vm *VM) Print() {
	fmt.Println("--- Code ---")
	for _, c := range vm.code.Code {
		fmt.Print(c.Inspect() + " ")
	}
}

// --- Private ---
func (vm *VM) push(value runtime.Value) {
	vm.stack = append(vm.stack, value)
}

func (vm *VM) transform(operation lexer.TokenKind, value runtime.Value) {
	if len(vm.stack) == 0 {
		shared.ReportErrFatal("Trying to transform an empty stack.")
	}

	if !vm.stack[len(vm.stack)-1].Modify(operation, value) {
		left := vm.stack[len(vm.stack)-1]
		shared.ReportErrFatal(fmt.Sprintf("Could not transform value '%s':%s with '%s':%s.", left.Inspect(), reflect.TypeOf(left), value.Inspect(), reflect.TypeOf(value)))
	}
}

func (vm *VM) pop() runtime.Value {
	if len(vm.stack) == 0 {
		shared.ReportErrFatal("Trying to pop an empty stack.")
	}

	value := vm.stack[len(vm.stack)-1]
	vm.stack = vm.stack[:len(vm.stack)-1]
	return value
}

func (vm *VM) visitPrint(count int) {
	for idx := count - 1; idx >= 0; idx-- {
		fmt.Print(vm.stack[len(vm.stack)-1-idx].Inspect())
		defer vm.pop()
	}
}
