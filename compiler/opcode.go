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

	Push
	Pop
	Negate

	Add
	Sub
	Mul
	Div

	Get
	Set

	Jump
	JumpFalse

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

		case Get:
			sb.WriteString(fmt.Sprintf("Get<ID '%s'>", c.Constants[c.Instructions[idx+1]].Inspect()))
			idx += 2

		case Set:
			sb.WriteString(fmt.Sprintf("Set<ID '%s'>", c.Constants[c.Instructions[idx+1]].Inspect()))
			idx += 2

		case Print:
			sb.WriteString(fmt.Sprintf("Print<Count %d>", c.Instructions[idx+1]))
			idx += 2

		default:
			sb.WriteString("Unknown")
			idx++
		}

		sb.WriteByte('\n')
	}

	fmt.Println(sb.String())
}

func (c *Chunk) addOp(code byte) {
	c.Instructions = append(c.Instructions, code)
}

func (c *Chunk) addOps(left byte, right byte) {
	c.Instructions = append(c.Instructions, left)
	c.Instructions = append(c.Instructions, right)
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
