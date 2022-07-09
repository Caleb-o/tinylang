package vm

import (
	"fmt"
	"reflect"
	"strings"
	"tiny/compiler"
	"tiny/lexer"
	"tiny/runtime"
	"tiny/shared"
)

type binaryOp byte

const (
	Add binaryOp = iota
	Sub
	Mul
	Div

	Less
	LessEq
	Greater
	GreaterEq
	EqualEqual
	NotEqual
)

func (b binaryOp) Operator() string {
	switch b {
	case Add:
		return "+"
	case Sub:
		return "-"
	case Mul:
		return "*"
	case Div:
		return "/"

	case Less:
		return "<"
	case LessEq:
		return "<="
	case Greater:
		return ">"
	case GreaterEq:
		return ">="
	case EqualEqual:
		return "=="
	case NotEqual:
		return "!="
	}

	return ""
}

func (b binaryOp) ToKind() lexer.TokenKind {
	switch b {
	case Add:
		return lexer.PLUS
	case Sub:
		return lexer.MINUS
	case Mul:
		return lexer.STAR
	case Div:
		return lexer.SLASH

	case Less:
		return lexer.LESS
	}

	return lexer.ERROR
}

type Scope struct {
	variables map[string]runtime.Value
}

type VM struct {
	chunk *compiler.Chunk
	ip    int
	stack []runtime.Value
	scope []Scope
}

func NewVM(chunk *compiler.Chunk) *VM {
	scopes := make([]Scope, 0, 1)
	scopes = append(scopes, Scope{make(map[string]runtime.Value)})

	return &VM{chunk, 0, make([]runtime.Value, 0), scopes}
}

func (vm *VM) Report(msg string, args ...any) {
	res := fmt.Sprintf(msg, args...)
	shared.ReportErrFatal("Runtime: " + res)
}

func (vm *VM) Run(debug bool) {
	if debug {
		vm.chunk.Debug()
	}

	for vm.ip < len(vm.chunk.Instructions) {
		switch vm.chunk.Instructions[vm.ip] {
		case compiler.OpenScope:
			vm.scope = append(vm.scope, Scope{make(map[string]runtime.Value)})
			vm.ip++

		case compiler.CloseScope:
			vm.scope = vm.scope[:len(vm.scope)-1]
			vm.ip++

		case compiler.Push:
			vm.stack = append(vm.stack, vm.chunk.Constants[vm.chunk.Instructions[vm.ip+1]])
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

	case *runtime.ListVal:
		if value, ok := runtime.BinopL(operation.ToKind(), left.(*runtime.ListVal).Values, right.(*runtime.ListVal).Values); ok {
			vm.push(value)
		}
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
