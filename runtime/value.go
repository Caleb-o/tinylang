package runtime

import (
	"fmt"
	"tiny/ast"
	"tiny/lexer"
)

type TypeKind uint8

const (
	TYPE_ANY TypeKind = iota
	TYPE_UNIT
	TYPE_INT
	TYPE_FLOAT
	TYPE_BOOL
	TYPE_CHAR
	TYPE_STRING
	TYPE_LIST
	TYPE_CLASS
	TYPE_CLASS_INSTANCE
	TYPE_FUNCTION
	TYPE_NAMESPACE
	TYPE_RETURN
)

type Type interface {
	GetKind() TypeKind
	GetName() string
}

type AnyType struct{}
type UnitType struct{}
type IntType struct{}
type FloatType struct{}
type CharType struct{}
type BoolType struct{}
type StringType struct{}
type FunctionType struct{}
type ReturnType struct{}
type ClassDefType struct{}
type ClassInstanceType struct{}

func (t *AnyType) GetKind() TypeKind { return TYPE_ANY }
func (t *AnyType) GetName() string   { return "any" }

func (t *UnitType) GetKind() TypeKind { return TYPE_UNIT }
func (t *UnitType) GetName() string   { return "unit" }

func (t *IntType) GetKind() TypeKind { return TYPE_INT }
func (t *IntType) GetName() string   { return "int" }

func (t *FloatType) GetKind() TypeKind { return TYPE_FLOAT }
func (t *FloatType) GetName() string   { return "float" }

func (t *CharType) GetKind() TypeKind { return TYPE_CHAR }
func (t *CharType) GetName() string   { return "char" }

func (t *BoolType) GetKind() TypeKind { return TYPE_BOOL }
func (t *BoolType) GetName() string   { return "bool" }

func (t *StringType) GetKind() TypeKind { return TYPE_STRING }
func (t *StringType) GetName() string   { return "string" }

func (t *FunctionType) GetKind() TypeKind { return TYPE_FUNCTION }
func (t *FunctionType) GetName() string   { return "function" }

func (t *ReturnType) GetKind() TypeKind { return TYPE_RETURN }
func (t *ReturnType) GetName() string   { return "return" }

func (t *ClassDefType) GetKind() TypeKind { return TYPE_CLASS }
func (t *ClassDefType) GetName() string   { return "class" }

func (t *ClassInstanceType) GetKind() TypeKind { return TYPE_CLASS_INSTANCE }
func (t *ClassInstanceType) GetName() string   { return "instance" }

type Value interface {
	GetType() Type
	Inspect() string
	Copy() Value
}

type UnitVal struct{}

type IntVal struct {
	Value int
}

type FloatVal struct {
	Value float32
}

type BoolVal struct {
	Value bool
}

type StringVal struct {
	Value string
}

type FunctionValue struct {
	definition *ast.FunctionDef
	bound      *ClassInstanceValue
}

type ReturnValue struct {
	inner Value
}

// TODO: Parent class
type ClassDefValue struct {
	identifier  string
	constructor *FunctionValue
	fields      []string
	methods     map[string]*FunctionValue
}

type ClassInstanceValue struct {
	def    *ClassDefValue
	fields map[string]Value
}

func (v *UnitVal) GetType() Type   { return &UnitType{} }
func (v *UnitVal) Inspect() string { return "()" }
func (v *UnitVal) Copy() Value     { return &UnitVal{} }

func (v *IntVal) GetType() Type   { return &IntType{} }
func (v *IntVal) Inspect() string { return fmt.Sprintf("%d", v.Value) }
func (v *IntVal) Copy() Value     { return &IntVal{Value: v.Value} }

func (v *FloatVal) GetType() Type   { return &FloatType{} }
func (v *FloatVal) Inspect() string { return fmt.Sprintf("%f", v.Value) }
func (v *FloatVal) Copy() Value     { return &FloatVal{Value: v.Value} }

func (v *BoolVal) GetType() Type   { return &BoolType{} }
func (v *BoolVal) Inspect() string { return fmt.Sprintf("%t", v.Value) }
func (v *BoolVal) Copy() Value     { return &BoolVal{Value: v.Value} }

func (v *StringVal) GetType() Type   { return &StringType{} }
func (v *StringVal) Inspect() string { return v.Value }
func (v *StringVal) Copy() Value     { return &StringVal{Value: v.Value} }

func (v *FunctionValue) GetType() Type { return &FunctionType{} }
func (v *FunctionValue) Inspect() string {
	return fmt.Sprintf("<fn %s>", v.definition.GetToken().Lexeme)
}
func (v *FunctionValue) Copy() Value { return v }

func (fn *FunctionValue) Arity() int { return len(fn.definition.Params) }

func (fn *FunctionValue) Call(interpreter *Interpreter, values []Value) Value {
	interpreter.push()

	for idx, arg := range values {
		interpreter.insert(fn.definition.Params[idx].Token.Lexeme, arg)
	}

	if fn.bound != nil {
		interpreter.insert("self", fn.bound)
	}

	value := interpreter.Visit(fn.definition.Body)

	if ret, ok := value.(*ReturnValue); ok {
		value = ret.inner
	}

	interpreter.pop()
	return value
}

func (v *ReturnValue) GetType() Type   { return &ReturnType{} }
func (v *ReturnValue) Inspect() string { return v.inner.Inspect() }
func (v *ReturnValue) Copy() Value     { return &ReturnValue{inner: v.inner} }

func (v *ClassDefValue) GetType() Type   { return &ClassDefType{} }
func (v *ClassDefValue) Inspect() string { return fmt.Sprintf("<class %s>", v.identifier) }
func (v *ClassDefValue) Copy() Value     { return v }

func (def *ClassDefValue) Arity() int {
	if def.constructor != nil {
		return len(def.constructor.definition.Params)
	}
	return 0
}

func (def *ClassDefValue) Call(interpreter *Interpreter, values []Value) Value {
	instance := &ClassInstanceValue{def: def, fields: make(map[string]Value)}

	for _, id := range def.fields {
		instance.fields[id] = &UnitVal{}
	}

	// Run the constructor
	if def.constructor != nil {
		def.constructor.bound = instance
		def.constructor.Call(interpreter, values)
	}
	return instance
}

func (v *ClassInstanceValue) GetType() Type { return &ClassInstanceType{} }
func (v *ClassInstanceValue) Inspect() string {
	return fmt.Sprintf("<instance %s : %p>", v.def.identifier, v)
}
func (v *ClassInstanceValue) Copy() Value { return v }

func (instance *ClassInstanceValue) Get(identifier string) (Value, bool) {
	if val, ok := instance.fields[identifier]; ok {
		return val.Copy(), true
	}

	if fn, ok := instance.def.methods[identifier]; ok {
		fn.bound = instance
		return fn, true
	}

	return nil, false
}

func (instance *ClassInstanceValue) Set(identifier string, value Value) (Value, bool) {
	if _, ok := instance.fields[identifier]; ok {
		instance.fields[identifier] = value
		return value, true
	}
	return nil, false
}

func IntBinop(operator lexer.TokenKind, a *IntVal, b *IntVal) (Value, bool) {
	switch operator {
	case lexer.PLUS:
		return &IntVal{Value: a.Value + b.Value}, true
	case lexer.MINUS:
		return &IntVal{Value: a.Value - b.Value}, true
	case lexer.STAR:
		return &IntVal{Value: a.Value * b.Value}, true
	case lexer.SLASH:
		return &IntVal{Value: a.Value / b.Value}, true
	case lexer.EQUAL_EQUAL:
		return &BoolVal{Value: a.Value == b.Value}, true
	case lexer.NOT_EQUAL:
		return &BoolVal{Value: a.Value != b.Value}, true
	case lexer.GREATER:
		return &BoolVal{Value: a.Value > b.Value}, true
	case lexer.GREATER_EQUAL:
		return &BoolVal{Value: a.Value >= b.Value}, true
	case lexer.LESS:
		return &BoolVal{Value: a.Value < b.Value}, true
	case lexer.LESS_EQUAL:
		return &BoolVal{Value: a.Value <= b.Value}, true
	}

	// Unreachable
	return nil, false
}

func FloatBinop(operator lexer.TokenKind, a *FloatVal, b *FloatVal) (Value, bool) {
	switch operator {
	case lexer.PLUS:
		return &FloatVal{Value: a.Value + b.Value}, true
	case lexer.MINUS:
		return &FloatVal{Value: a.Value - b.Value}, true
	case lexer.STAR:
		return &FloatVal{Value: a.Value * b.Value}, true
	case lexer.SLASH:
		return &FloatVal{Value: a.Value / b.Value}, true
	case lexer.EQUAL_EQUAL:
		return &BoolVal{Value: a.Value == b.Value}, true
	case lexer.NOT_EQUAL:
		return &BoolVal{Value: a.Value != b.Value}, true
	case lexer.GREATER:
		return &BoolVal{Value: a.Value > b.Value}, true
	case lexer.GREATER_EQUAL:
		return &BoolVal{Value: a.Value >= b.Value}, true
	case lexer.LESS:
		return &BoolVal{Value: a.Value < b.Value}, true
	case lexer.LESS_EQUAL:
		return &BoolVal{Value: a.Value <= b.Value}, true
	}

	// Unreachable
	return nil, false
}
