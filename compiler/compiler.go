package compiler

import (
	"fmt"
	"reflect"
	"tiny/ast"
	"tiny/lexer"
	"tiny/runtime"
)

type local struct {
	id    string
	depth int
}

type scope struct {
	locals      []local
	local_depth int
}

func (s *scope) findLocal(identifier string) int {
	// Must iterate in reverse, so we can find the local at the correct
	// scope. If we shadow in other scopes, we may get the wrong value.
	for idx := len(s.locals) - 1; idx >= 0; idx-- {
		local := s.locals[idx]

		if local.id == identifier && local.depth <= s.local_depth {
			return idx
		}
	}

	return -1
}

type Compiler struct {
	chunk *Chunk
	ids   []*scope
	depth int
}

func NewCompiler() *Compiler {
	ids := make([]*scope, 0, 8)
	ids = append(ids, &scope{make([]local, 0, 8), 0})

	return &Compiler{
		chunk: &Chunk{Constants: make([]runtime.Value, 0), Instructions: make([]byte, 0)},
		ids:   ids,
		depth: 0,
	}
}

func (c *Compiler) Compile(program *ast.Program) *Chunk {
	c.compileProgram(c.chunk, program)
	return c.chunk
}

func (c *Compiler) begin() {
	c.depth++
	c.ids = append(c.ids, &scope{make([]local, 0, 8), 0})
}

func (c *Compiler) end() {
	c.depth--
	c.ids = c.ids[:len(c.ids)-1]
}

func (c *Compiler) open() {
	c.ids[len(c.ids)-1].local_depth++
}

func (c *Compiler) close() {
	c.ids[len(c.ids)-1].local_depth--
}

func (c *Compiler) addVariable(identifier string) (byte, byte) {
	index := -1

	for idx, constant := range c.chunk.Constants {
		if constant.Inspect() == identifier {
			index = idx
			break
		}
	}

	if index == -1 {
		c.chunk.Constants = append(c.chunk.Constants, &runtime.StringVal{identifier})
		index = len(c.chunk.Constants) - 1
	}

	if c.depth == 0 {
		return 0, byte(index)
	}

	scope := c.ids[c.depth]
	scope.locals = append(scope.locals, local{identifier, scope.local_depth})
	return byte(len(scope.locals) - 1), byte(index)
}

func (c *Compiler) addVariableWithSlot(identifier string, slot byte) {
	found := false

	for _, constant := range c.chunk.Constants {
		if constant.Inspect() == identifier {
			found = true
			break
		}
	}

	if !found {
		c.chunk.Constants = append(c.chunk.Constants, &runtime.StringVal{identifier})
	}

	scope := c.ids[len(c.ids)-1]
	scope.locals = append(c.ids[len(c.ids)-1].locals, local{identifier, c.depth})
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

func (c *Compiler) findVariableSlot(identifier string) byte {
	for idx := len(c.ids) - 1; idx >= 0; idx-- {
		if index := c.ids[idx].findLocal(identifier); index > -1 {
			return byte(index)
		}
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
		c.open()
		c.body(chunk, n)
		c.close()

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
		c.chunk.addOp(Return)

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

func (c *Compiler) body(chunk *Chunk, block *ast.Block) {
	for _, stmt := range block.Statements {
		c.visit(chunk, stmt)
	}
}

func (c *Compiler) functionDef(chunk *Chunk, def *ast.FunctionDef) {
	_, name_id := c.addVariable(def.GetToken().Lexeme)
	defStart := c.chunk.addOps(Jump, 0)

	c.begin()
	for idx := len(def.Params) - 1; idx >= 0; idx-- {
		c.addVariableWithSlot(def.Params[idx].Token.Lexeme, byte(idx))
	}
	c.body(chunk, def.Body)
	c.end()

	c.chunk.addOp(Return)
	c.chunk.upateOpPosNext(defStart)
	c.chunk.addOps(NewFn, byte(len(def.Params)), byte(defStart+1), name_id)
}

func (c *Compiler) anonFunction(chunk *Chunk, anon *ast.AnonymousFunction) {
	defStart := c.chunk.addOps(Jump, 0)

	c.begin()
	for idx := len(anon.Params) - 1; idx >= 0; idx-- {
		c.addVariableWithSlot(anon.Params[idx].Token.Lexeme, byte(idx))
	}
	c.body(chunk, anon.Body)
	c.end()

	c.chunk.addOp(Return)

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
	slot, id := c.addVariable(decl.GetToken().Lexeme)

	if c.depth > 0 {
		c.chunk.addOps(SetLocal, slot)
	} else {
		c.chunk.addOps(Set, id)
	}
}

func (c *Compiler) variableAssign(chunk *Chunk, assign *ast.Assign) {
	c.visit(chunk, assign.Expr)

	if c.depth > 0 {
		c.chunk.addOps(SetLocal, c.findVariableSlot(assign.Token.Lexeme))
	} else {
		c.chunk.addOps(Set, c.getVariable(assign.Token.Lexeme))
	}

	c.chunk.addOp(Pop)
}

func (c *Compiler) getIdentifier(chunk *Chunk, identifier *ast.Identifier) {
	if c.depth > 0 {
		c.chunk.addOps(GetLocal, c.findVariableSlot(identifier.Token.Lexeme))
	} else {
		c.chunk.addOps(Get, c.getVariable(identifier.Token.Lexeme))
	}
}

func (c *Compiler) ifStmt(chunk *Chunk, stmt *ast.If) {
	c.open()

	if stmt.VarDec != nil {
		c.variableDecl(chunk, stmt.VarDec)
	}

	c.visit(chunk, stmt.Condition)

	condition := c.chunk.addOps(JumpFalse, 0)

	c.body(chunk, stmt.TrueBody)

	var end_of_stmt int
	if stmt.FalseBody != nil {
		end_of_stmt = c.chunk.addOps(Jump, 0)
	}

	c.close()
	c.chunk.upateOpPosNext(condition)

	if stmt.FalseBody != nil {
		c.visit(chunk, stmt.FalseBody)
		c.chunk.upateOpPosNext(end_of_stmt)
	}
}

func (c *Compiler) whileStmt(chunk *Chunk, stmt *ast.While) {
	c.open()

	if stmt.VarDec != nil {
		c.variableDecl(chunk, stmt.VarDec)
	}

	condition := len(c.chunk.Instructions)
	c.visit(chunk, stmt.Condition)

	false_expr := c.chunk.addOps(JumpFalse, 0)
	c.body(chunk, stmt.Body)

	if stmt.Increment != nil {
		c.visit(chunk, stmt.Increment)
	}

	c.chunk.addOps(Jump, byte(condition))
	c.close()
	c.chunk.upateOpPosNext(false_expr)

}
