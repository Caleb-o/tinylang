package compiler

import (
	"fmt"
	"strconv"
	"strings"
	"tiny/ast"
	"tiny/lexer"
	"tiny/runtime"
)

const (
	Halt byte = iota

	Push // Push const_index
	Pop
	Negate

	OpenScope
	CloseScope

	Add
	Sub
	Mul
	Div

	Less
	LessEq
	Greater
	GreaterEq
	EqEq
	NotEq

	Get // Get scope_index const_index
	Set // Set scope_index const_index

	Jump      // Jump IP
	JumpFalse // JumpFalse IP

	Print
)

type Chunk struct {
	Constants    []runtime.Value
	Instructions []byte
}

func (c *Chunk) Debug() {
	var sb strings.Builder
	idx := 0

	sb.WriteString("=== ByteCode Debug ===\n")
	for idx < len(c.Instructions) {
		op := c.Instructions[idx]

		sb.WriteString(fmt.Sprintf("%04d: ", idx))

		switch op {
		case Halt:
			sb.WriteString("Halt")
			idx++

		case OpenScope:
			sb.WriteString("Open Scope")
			idx++

		case CloseScope:
			sb.WriteString("Close Scope")
			idx++

		case Negate:
			sb.WriteString("Negate")
			idx++

		case Push:
			sb.WriteString(fmt.Sprintf("Push<Value '%s'>", c.Constants[c.Instructions[idx+1]].Inspect()))
			idx += 2

		case Pop:
			sb.WriteString("Pop")
			idx++

		case Add:
			sb.WriteString("Add")
			idx++

		case Sub:
			sb.WriteString("Subtract")
			idx++

		case Mul:
			sb.WriteString("Multiply")
			idx++

		case Div:
			sb.WriteString("Divide")
			idx++

		case Less:
			sb.WriteString("Less")
			idx++

		case LessEq:
			sb.WriteString("Less Equal")
			idx++

		case Greater:
			sb.WriteString("Greater")
			idx++

		case GreaterEq:
			sb.WriteString("Greater Equal")
			idx++

		case EqEq:
			sb.WriteString("Equal Equal")
			idx++

		case NotEq:
			sb.WriteString("Not Equal")
			idx++

		case Get:
			sb.WriteString(fmt.Sprintf("Get<ID '%s'>", c.Constants[c.Instructions[idx+1]].Inspect()))
			idx += 2

		case Set:
			sb.WriteString(fmt.Sprintf("Set<ID '%s'>", c.Constants[c.Instructions[idx+1]].Inspect()))
			idx += 2

		case Print:
			sb.WriteString(fmt.Sprintf("Print<Count %d>", c.Instructions[idx+1]))
			idx += 2

		case Jump:
			sb.WriteString(fmt.Sprintf("Jump<Position %d>", c.Instructions[idx+1]))
			idx += 2

		case JumpFalse:
			sb.WriteString(fmt.Sprintf("Jump False<Position %d>", c.Instructions[idx+1]))
			idx += 2

		default:
			sb.WriteString(fmt.Sprintf("Unknown<%d>", op))
			idx++
		}

		sb.WriteByte('\n')
	}

	fmt.Println(sb.String())
}

func (c *Chunk) addOp(code byte) int {
	c.Instructions = append(c.Instructions, code)
	return len(c.Instructions) - 1
}

func (c *Chunk) addOps(operands ...byte) int {
	c.Instructions = append(c.Instructions, operands...)
	return len(c.Instructions) - 1
}

func (c *Chunk) upateOpPos(index int) {
	c.Instructions[index] = byte(len(c.Instructions) - 1)
}

func (c *Chunk) addConstant(node *ast.Literal) byte {
	lexeme := node.GetToken().Lexeme

	// TODO: Only add unique values

	switch node.GetToken().Kind {
	case lexer.INT:
		value, _ := strconv.ParseInt(lexeme, 10, 32)
		c.Constants = append(c.Constants, &runtime.IntVal{Value: int(value)})

	case lexer.FLOAT:
		value, _ := strconv.ParseFloat(lexeme, 32)
		c.Constants = append(c.Constants, &runtime.FloatVal{Value: float32(value)})

	case lexer.BOOL:
		value, _ := strconv.ParseBool(lexeme)
		c.Constants = append(c.Constants, &runtime.BoolVal{Value: value})

	case lexer.STRING:
		c.Constants = append(c.Constants, &runtime.StringVal{Value: lexeme})
	}

	return byte(len(c.Constants) - 1)
}
