package compiler

import (
	"fmt"
	"reflect"
	"tiny/ast"
	"tiny/lexer"
	"tiny/runtime"
)

type Compiler struct {
	chunk       *Chunk
	ids         []map[string]interface{}
	scope_depth int
	depth       int
}

func NewCompiler() *Compiler {
	ids := make([]map[string]interface{}, 0)
	ids = append(ids, make(map[string]interface{}))

	return &Compiler{
		chunk:       &Chunk{Constants: make([]runtime.Value, 0), Instructions: make([]byte, 0)},
		ids:         ids,
		scope_depth: 0,
		depth:       0,
	}
}

func (c *Compiler) Compile(program *ast.Program) *Chunk {
	c.compileProgram(c.chunk, program)
	return c.chunk
}

func (c *Compiler) begin(open bool) {
	c.scope_depth += 1
	if open {
		c.chunk.addOp(OpenScope)
	}

	c.depth++
	c.ids = append(c.ids, make(map[string]interface{}))
}

func (c *Compiler) end(close bool) {
	c.scope_depth -= 1

	if close {
		c.chunk.addOp(CloseScope)
	}

	c.depth--
	c.ids = c.ids[:len(c.ids)-1]
}

func (c *Compiler) addVariable(identifier string) byte {
	c.chunk.Constants = append(c.chunk.Constants, &runtime.StringVal{identifier})
	c.ids[c.depth][identifier] = nil
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

	// TODO: Report error or unreachable
	return 0
}

func (c *Compiler) findVariableScope(identifier string) byte {
	idx := c.depth

	for idx >= 0 {
		if _, ok := c.ids[idx][identifier]; ok {
			return byte(idx)
		}
		idx--
	}

	// TODO: Report error or unreachable
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
	case *ast.Block:
		c.body(chunk, n, true)

	case *ast.FunctionDef:
		c.functionDef(chunk, n)
	case *ast.AnonymousFunction:
		c.anonFunction(chunk, n)
	case *ast.Call:
		c.call(chunk, n)

	case *ast.BinaryOp:
		c.binaryOp(chunk, n)
	case *ast.UnaryOp:
		c.unaryOp(chunk, n)

	case *ast.Print:
		for _, value := range n.Exprs {
			c.visit(chunk, value)
		}
		chunk.addOps(Print, byte(len(n.Exprs)))

	case *ast.Return:
		if n.Expr != nil {
			c.visit(chunk, n.Expr)
		}
		c.chunk.addOps(Return, byte(c.scope_depth))

	case *ast.VariableDecl:
		c.variableDecl(chunk, n)
	case *ast.Assign:
		c.variableAssign(chunk, n)
	case *ast.Literal:
		chunk.addOps(Push, chunk.addConstant(n))
	case *ast.If:
		c.ifStmt(chunk, n)
	case *ast.While:
		c.whileStmt(chunk, n)
	case *ast.Identifier:
		c.getIdentifier(chunk, n)

	default:
		fmt.Printf("Compiler: Unimplemented node '%s'", reflect.TypeOf(node))
	}
}

func (c *Compiler) body(chunk *Chunk, block *ast.Block, newEnv bool) {
	if newEnv {
		c.begin(true)
		defer c.end(true)
	}

	for _, stmt := range block.Statements {
		c.visit(chunk, stmt)
	}
}

func (c *Compiler) functionDef(chunk *Chunk, def *ast.FunctionDef) {
	name_id := c.addVariable(def.GetToken().Lexeme)
	defStart := c.chunk.addOps(Jump, 0)

	c.scope_depth = 0
	c.begin(false)
	for _, value := range def.Params {
		c.chunk.addOps(Define, c.addVariable(value.Token.Lexeme))
	}
	c.body(chunk, def.Body, false)
	c.end(false)

	c.chunk.addOps(Return, 0)
	c.chunk.upateOpPosNext(defStart)
	c.chunk.addOps(NewFn, byte(len(def.Params)), byte(defStart+1), name_id)
}

func (c *Compiler) anonFunction(chunk *Chunk, anon *ast.AnonymousFunction) {
	defStart := c.chunk.addOps(Jump, 0)

	c.scope_depth = 0
	c.begin(false)
	for _, value := range anon.Params {
		c.chunk.addOps(Define, c.addVariable(value.Token.Lexeme))
	}
	c.body(chunk, anon.Body, false)
	c.end(false)

	c.chunk.addOps(Return, 0)

	c.chunk.upateOpPosNext(defStart)
	c.chunk.addOps(NewAnonFn, byte(len(anon.Params)), byte(defStart+1))
}

func (c *Compiler) call(chunk *Chunk, call *ast.Call) {
	for _, value := range call.Arguments {
		c.visit(chunk, value)
	}
	c.chunk.addOps(Call, c.getVariable(call.Token.Lexeme))
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

	case lexer.LESS:
		chunk.addOp(Less)
	case lexer.LESS_EQUAL:
		chunk.addOp(LessEq)
	case lexer.GREATER:
		chunk.addOp(Greater)
	case lexer.GREATER_EQUAL:
		chunk.addOp(GreaterEq)
	case lexer.EQUAL_EQUAL:
		chunk.addOp(EqEq)
	case lexer.NOT_EQUAL:
		chunk.addOp(NotEq)
	}
}

func (c *Compiler) unaryOp(chunk *Chunk, unary *ast.UnaryOp) {
	c.visit(chunk, unary.Right)
	chunk.addOp(Negate)
}

func (c *Compiler) variableDecl(chunk *Chunk, decl *ast.VariableDecl) {
	// Analyser should pickup clashes, so we don't need to check names
	c.visit(chunk, decl.Expr)
	c.chunk.addOps(Define, c.addVariable(decl.GetToken().Lexeme))
}

func (c *Compiler) variableAssign(chunk *Chunk, assign *ast.Assign) {
	c.visit(chunk, assign.Expr)

	if c.findVariableScope(assign.Token.Lexeme) > 0 {
		c.chunk.addOps(SetLocal, c.getVariable(assign.Token.Lexeme))
	} else {
		c.chunk.addOps(Set, c.getVariable(assign.Token.Lexeme))
	}
}

func (c *Compiler) getIdentifier(chunk *Chunk, identifier *ast.Identifier) {
	if c.findVariableScope(identifier.Token.Lexeme) > 0 {
		c.chunk.addOps(GetLocal, c.getVariable(identifier.Token.Lexeme))
	} else {
		c.chunk.addOps(Get, c.getVariable(identifier.Token.Lexeme))
	}
}

func (c *Compiler) ifStmt(chunk *Chunk, stmt *ast.If) {
	c.begin(true)

	if stmt.VarDec != nil {
		c.variableDecl(chunk, stmt.VarDec)
	}

	c.visit(chunk, stmt.Condition)

	condition := c.chunk.addOps(JumpFalse, 0)

	c.body(chunk, stmt.TrueBody, false)

	var end_of_stmt int
	if stmt.FalseBody != nil {
		end_of_stmt = c.chunk.addOps(Jump, 0)
	}

	c.chunk.upateOpPosNext(condition)

	if stmt.FalseBody != nil {
		c.visit(chunk, stmt.FalseBody)
		c.chunk.upateOpPosNext(end_of_stmt)
	}

	c.end(true)
}

func (c *Compiler) whileStmt(chunk *Chunk, stmt *ast.While) {
	c.begin(true)

	if stmt.VarDec != nil {
		c.variableDecl(chunk, stmt.VarDec)
	}

	condition := len(c.chunk.Instructions)
	c.visit(chunk, stmt.Condition)

	false_expr := c.chunk.addOps(JumpFalse, 0)
	c.body(chunk, stmt.Body, false)

	if stmt.Increment != nil {
		c.visit(chunk, stmt.Increment)
	}

	c.chunk.addOps(Jump, byte(condition))
	c.end(true)

	c.chunk.upateOpPos(false_expr)
}
