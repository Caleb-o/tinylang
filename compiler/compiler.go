package compiler

import (
	"fmt"
	"reflect"
	"strconv"
	"tiny/ast"
	"tiny/lexer"
	"tiny/runtime"
	"tiny/shared"
)

type Chunk struct {
	Code []Code
}

type Compiler struct {
	chunk *Chunk
}

func New() *Compiler {
	return &Compiler{chunk: &Chunk{Code: make([]Code, 0)}}
}

func (compiler *Compiler) Compile(program *ast.Program) *Chunk {
	compiler.visit(program.Body)
	compiler.chunk.Code = append(compiler.chunk.Code, &Halt{})
	return compiler.chunk
}

// --- Private ---
func (compiler *Compiler) visit(node ast.Node) {
	switch n := node.(type) {
	case *ast.Block:
		compiler.visitBlock(n)
	case *ast.BinaryOp:
		compiler.visitBinaryOp(n)
	case *ast.Literal:
		compiler.visitLiteral(n)
	case *ast.Print:
		compiler.visitPrint(n)
	default:
		shared.ReportErrFatal(fmt.Sprintf("Unimplemented node in compiler '%s':%s", node.GetToken().Lexeme, reflect.TypeOf(node)))
	}
}

func (compiler *Compiler) visitBlock(block *ast.Block) {
	for _, node := range block.Statements {
		compiler.visit(node)
	}
}

func (compiler *Compiler) visitBinaryOp(binop *ast.BinaryOp) {
	compiler.visit(binop.Left)
	compiler.visit(binop.Right)

	newKind := binop.Token.Kind

	switch binop.Token.Kind {
	case lexer.PLUS:
		newKind = lexer.PLUS_EQUAL
	case lexer.MINUS:
		newKind = lexer.MINUS_EQUAL
	case lexer.STAR:
		newKind = lexer.STAR_EQUAL
	case lexer.SLASH:
		newKind = lexer.SLASH_EQUAL
	}

	compiler.chunk.Code = append(compiler.chunk.Code, &BinOp{Kind: newKind})
}

func (compiler *Compiler) visitLiteral(lit *ast.Literal) {
	var value runtime.Value = nil

	switch lit.Token.Kind {
	case lexer.INT:
		v, _ := strconv.ParseInt(lit.GetToken().Lexeme, 10, 32)
		value = &runtime.IntVal{Value: int(v)}

	case lexer.FLOAT:
		v, _ := strconv.ParseFloat(lit.GetToken().Lexeme, 32)
		value = &runtime.FloatVal{Value: float32(v)}

	case lexer.BOOL:
		v, _ := strconv.ParseBool(lit.GetToken().Lexeme)
		value = &runtime.BoolVal{Value: v}

	case lexer.STRING:
		value = &runtime.StringVal{Value: lit.GetToken().Lexeme}

	default:
		shared.ReportErrFatal(fmt.Sprintf("Unknown literal type found '%s'", lit.GetToken().Lexeme))
		return
	}

	compiler.chunk.Code = append(compiler.chunk.Code, &Literal{Data: value})
}

func (compiler *Compiler) visitPrint(print *ast.Print) {
	for _, node := range print.Exprs {
		compiler.visit(node)
	}

	compiler.chunk.Code = append(compiler.chunk.Code, &Print{Count: len(print.Exprs)})
}
