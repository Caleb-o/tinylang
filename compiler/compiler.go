package compiler

import (
	"fmt"
	"reflect"
	"tiny/ast"
	"tiny/lexer"
	"tiny/runtime"
)

type Compiler struct {
	chunk *Chunk
}

func NewCompiler() *Compiler {
	return &Compiler{
		&Chunk{Constants: make([]runtime.Value, 0), Instructions: make([]byte, 0)},
	}
}

func (c *Compiler) Compile(program *ast.Program) *Chunk {
	c.compileProgram(c.chunk, program)
	return c.chunk
}

func (c *Compiler) addVariable(identifier string) byte {
	c.chunk.Constants = append(c.chunk.Constants, &runtime.StringVal{identifier})
	return byte(len(c.chunk.Constants) - 1)
}

func (c *Compiler) getVariable(identifier string) byte {
	for idx, value := range c.chunk.Constants {
		if id, ok := value.(*runtime.StringVal); ok {
			if id.Value == identifier {
				return byte(idx)
			}
		}
	}

	return 0
}

func (c *Compiler) compileProgram(chunk *Chunk, program *ast.Program) {
	for _, node := range program.Body.Statements {
		c.visit(chunk, node)
	}

	chunk.addOp(Halt)
}

func (c *Compiler) visit(chunk *Chunk, node ast.Node) {
	switch n := node.(type) {
	case *ast.BinaryOp:
		c.binaryOp(chunk, n)
	case *ast.UnaryOp:
		c.unaryOp(chunk, n)

	case *ast.Print:
		for _, value := range n.Exprs {
			c.visit(chunk, value)
		}
		chunk.addOps(Print, byte(len(n.Exprs)))

	case *ast.VariableDecl:
		c.variableDecl(chunk, n)
	case *ast.Literal:
		chunk.addOps(Push, chunk.addConstant(n))
	case *ast.Identifier:
		c.chunk.addOps(Get, c.getVariable(n.GetToken().Lexeme))

	default:
		fmt.Printf("Gen: Unimplemented node '%s'", reflect.TypeOf(node))
	}
}

func (c *Compiler) binaryOp(chunk *Chunk, binop *ast.BinaryOp) {
	c.visit(chunk, binop.Left)
	c.visit(chunk, binop.Right)

	switch binop.Token.Kind {
	case lexer.PLUS:
		chunk.addOp(Add)
	case lexer.MINUS:
		chunk.addOp(Sub)
	case lexer.STAR:
		chunk.addOp(Mul)
	case lexer.SLASH:
		chunk.addOp(Div)
	}
}

func (c *Compiler) unaryOp(chunk *Chunk, unary *ast.UnaryOp) {
	c.visit(chunk, unary.Right)
	chunk.addOp(Negate)
}

func (c *Compiler) variableDecl(chunk *Chunk, decl *ast.VariableDecl) {
	// Analyser should pickup clashes, so we don't need to check names
	c.visit(chunk, decl.Expr)
	c.chunk.addOps(Set, c.addVariable(decl.GetToken().Lexeme))
}
