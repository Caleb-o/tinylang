package compiler

import (
	"fmt"
	"strings"
	"tiny/lexer"
	"tiny/runtime"
)

type Code interface {
	Inspect() string
}

type Halt struct{}
type BinOp struct {
	Kind lexer.TokenKind
}

type Literal struct {
	Data runtime.Value
}

type VariableSet struct {
	Identifier string
	Data       runtime.Value
}

type VariableGet struct {
	Identifier string
	Data       runtime.Value
}

type Print struct {
	Count int
}

func (code *Halt) Inspect() string {
	return "HALT"
}

func (code *BinOp) Inspect() string {
	return code.Kind.Name()
}

func (code *Literal) Inspect() string {
	return code.Data.Inspect()
}

func (code *VariableSet) Inspect() string {
	var sb strings.Builder

	sb.WriteString(code.Identifier + " ")
	sb.WriteString(code.Data.Inspect())

	return sb.String()
}

func (code *VariableGet) Inspect() string {
	var sb strings.Builder

	sb.WriteString(code.Identifier + " ")
	sb.WriteString(code.Data.Inspect())

	return sb.String()
}

func (code *Print) Inspect() string {
	return fmt.Sprintf("print[%d]", code.Count)
}
