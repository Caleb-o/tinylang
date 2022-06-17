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
	methods     map[string]*FunctionValue
}

type ClassInstanceValue struct {
	def    *ClassDefValue
	fields map[string]Value
}

func (u *UnitVal) GetType() Type   { return &UnitType{} }
func (u *UnitVal) Inspect() string { return "()" }

func (i *IntVal) GetType() Type   { return &IntType{} }
func (i *IntVal) Inspect() string { return fmt.Sprintf("%d", i.Value) }

func (f *FloatVal) GetType() Type   { return &FloatType{} }
func (f *FloatVal) Inspect() string { return fmt.Sprintf("%f", f.Value) }

func (b *BoolVal) GetType() Type   { return &BoolType{} }
func (b *BoolVal) Inspect() string { return fmt.Sprintf("%t", b.Value) }

func (str *StringVal) GetType() Type   { return &StringType{} }
func (str *StringVal) Inspect() string { return str.Value }

func (fn *FunctionValue) GetType() Type { return &FunctionType{} }
func (fn *FunctionValue) Inspect() string {
	return fmt.Sprintf("<fn %s>", fn.definition.GetToken().Lexeme)
}
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

func (ret *ReturnValue) GetType() Type   { return &ReturnType{} }
func (ret *ReturnValue) Inspect() string { return ret.inner.Inspect() }

func (def *ClassDefValue) GetType() Type   { return &ClassDefType{} }
func (def *ClassDefValue) Inspect() string { return fmt.Sprintf("<class %s>", def.identifier) }

func (def *ClassDefValue) Arity() int { return 0 }
func (def *ClassDefValue) Call(interpreter *Interpreter, values []Value) Value {
	instance := &ClassInstanceValue{def: def, fields: make(map[string]Value)}

	// Run the constructor
	if def.constructor != nil {
		interpreter.push()
		interpreter.insert("self", instance)
		def.constructor.Call(interpreter, values)
		interpreter.pop()
	}
	return instance
}

func (instance *ClassInstanceValue) GetType() Type { return &ClassInstanceType{} }
func (instance *ClassInstanceValue) Inspect() string {
	return fmt.Sprintf("<instance %s>", instance.def.identifier)
}

func (instance *ClassInstanceValue) Get(identifier string) Value {
	if val, ok := instance.fields[identifier]; ok {
		return val
	}

	if fn, ok := instance.def.methods[identifier]; ok {
		fn.bound = instance
		return fn
	}

	instance.fields[identifier] = &UnitVal{}
	return instance.fields[identifier]
}

func (instance *ClassInstanceValue) Set(identifier string, value Value) Value {
	instance.fields[identifier] = value
	return value
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
