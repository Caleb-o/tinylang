package runtime

import (
	"fmt"
	"reflect"
	"strings"
	"tiny/ast"
	"tiny/lexer"
)

type NativeFn func(interpreter *Interpreter, values []Value) Value

type Value interface {
	GetType() Type
	Inspect() string
	Copy() Value
	Modify(operation lexer.TokenKind, value Value) bool
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
	bound      Value
}

type NativeFunctionValue struct {
	Identifier string
	Params     []string
	Fn         NativeFn
}

type ListVal struct {
	Values []Value
}

func NewFnValue(identifier string, params []string, fn NativeFn) *NativeFunctionValue {
	return &NativeFunctionValue{identifier, params, fn}
}

type AnonFunctionValue struct {
	definition *ast.AnonymousFunction
}

type ReturnValue struct {
	inner Value
}

type ThrowValue struct {
	inner Value
}

func NewThrow(inner Value) *ThrowValue {
	return &ThrowValue{inner}
}

func (throw *ThrowValue) GetInner() Value {
	return throw.inner
}

type NativeClassDefValue struct {
	Identifier string
	Fields     []string
	Methods    map[string]Value
}

func NewClassDefValue(identifier string, fields []string, methods map[string]*NativeFunctionValue) *NativeClassDefValue {
	temp := make(map[string]Value, len(methods))

	for key, val := range methods {
		temp[key] = val
	}

	return &NativeClassDefValue{identifier, fields, temp}
}

func (def *NativeClassDefValue) HasField(field string) bool {
	for _, f := range def.Fields {
		if f == field {
			return true
		}
	}
	return false
}

// TODO: Parent class
type ClassDefValue struct {
	identifier  string
	base        *ClassDefValue
	constructor *FunctionValue
	fields      []string
	methods     map[string]Value
}

func (def *ClassDefValue) HasField(field string) bool {
	for _, f := range def.fields {
		if f == field {
			return true
		}
	}
	return false
}

type ClassInstanceValue struct {
	Def    Value
	base   *ClassInstanceValue
	fields map[string]Value
}

func (klass *ClassInstanceValue) Definition() string {
	return klass.Def.(*ClassDefValue).identifier
}

type StructDefValue struct {
	identifier  string
	constructor *FunctionValue
	fields      []string
}

func (str *StructDefValue) HasField(field string) bool {
	for _, f := range str.fields {
		if f == field {
			return true
		}
	}
	return false
}

type StructInstanceValue struct {
	def    *StructDefValue
	fields map[string]Value
}

func (str *StructInstanceValue) Definition() string {
	return str.def.identifier
}

func (str *StructInstanceValue) HasField(field string) bool {
	_, ok := str.fields[field]
	return ok
}

type NameSpaceValue struct {
	Identifier string
	Members    map[string]Value
}

type LoopFlow struct {
	exit bool // Exit true = break, false = continue
}

func (v *UnitVal) GetType() Type                                      { return &UnitType{} }
func (v *UnitVal) Inspect() string                                    { return "()" }
func (v *UnitVal) Copy() Value                                        { return &UnitVal{} }
func (v *UnitVal) Modify(operation lexer.TokenKind, other Value) bool { return false }

func (v *IntVal) GetType() Type   { return &IntType{} }
func (v *IntVal) Inspect() string { return fmt.Sprintf("%d", v.Value) }
func (v *IntVal) Copy() Value     { return &IntVal{Value: v.Value} }
func (v *IntVal) Modify(operation lexer.TokenKind, other Value) bool {
	if value, ok := other.(*IntVal); ok {
		switch operation {
		case lexer.PLUS_EQUAL:
			v.Value += value.Value
		case lexer.MINUS_EQUAL:
			v.Value -= value.Value
		case lexer.STAR_EQUAL:
			v.Value *= value.Value
		case lexer.SLASH_EQUAL:
			v.Value /= value.Value
		default:
			return false
		}

		return true
	}

	return false
}

func (v *FloatVal) GetType() Type   { return &FloatType{} }
func (v *FloatVal) Inspect() string { return fmt.Sprintf("%f", v.Value) }
func (v *FloatVal) Copy() Value     { return &FloatVal{Value: v.Value} }
func (v *FloatVal) Modify(operation lexer.TokenKind, other Value) bool {
	if value, ok := other.(*FloatVal); ok {
		switch operation {
		case lexer.PLUS_EQUAL:
			v.Value += value.Value
		case lexer.MINUS_EQUAL:
			v.Value -= value.Value
		case lexer.STAR_EQUAL:
			v.Value *= value.Value
		case lexer.SLASH_EQUAL:
			v.Value /= value.Value
		default:
			return false
		}

		return true
	}

	return false
}

func (v *BoolVal) GetType() Type                                      { return &BoolType{} }
func (v *BoolVal) Inspect() string                                    { return fmt.Sprintf("%t", v.Value) }
func (v *BoolVal) Copy() Value                                        { return &BoolVal{Value: v.Value} }
func (v *BoolVal) Modify(operation lexer.TokenKind, other Value) bool { return false }

func (v *StringVal) GetType() Type   { return &StringType{} }
func (v *StringVal) Inspect() string { return v.Value }
func (v *StringVal) Copy() Value     { return &StringVal{Value: v.Value} }
func (v *StringVal) Modify(operation lexer.TokenKind, other Value) bool {
	if value, ok := other.(*StringVal); ok {
		switch operation {
		case lexer.PLUS_EQUAL:
			v.Value += value.Value
			return true
		}
	}

	return false
}

func (v *FunctionValue) GetType() Type { return &FunctionType{} }
func (v *FunctionValue) Inspect() string {
	return fmt.Sprintf("<fn %s>", v.definition.GetToken().Lexeme)
}
func (v *FunctionValue) Copy() Value                                        { return v }
func (v *FunctionValue) Modify(operation lexer.TokenKind, other Value) bool { return false }

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

	if thrown, ok := value.(*ThrowValue); ok {
		value = thrown
	}

	interpreter.pop()
	return value
}

func (v *NativeFunctionValue) GetType() Type { return &NativeFunctionType{} }
func (v *NativeFunctionValue) Inspect() string {
	return fmt.Sprintf("<native fn %s>", v.Identifier)
}
func (v *NativeFunctionValue) Copy() Value                                        { return v }
func (v *NativeFunctionValue) Modify(operation lexer.TokenKind, other Value) bool { return false }

func (v *NativeFunctionValue) Arity() int { return len(v.Params) }

func (v *NativeFunctionValue) Call(interpreter *Interpreter, values []Value) Value {
	if v.Arity() != len(values) {
		interpreter.Report("Native function '%s' expected %d arguments but received %d.", v.Identifier, v.Arity(), len(values))
	}
	return v.Fn(interpreter, values)
}

func (v *AnonFunctionValue) GetType() Type                                      { return &FunctionType{} }
func (v *AnonFunctionValue) Inspect() string                                    { return "<anon fn>" }
func (v *AnonFunctionValue) Copy() Value                                        { return v }
func (v *AnonFunctionValue) Modify(operation lexer.TokenKind, other Value) bool { return false }

func (fn *AnonFunctionValue) Arity() int { return len(fn.definition.Params) }

func (fn *AnonFunctionValue) Call(interpreter *Interpreter, values []Value) Value {
	interpreter.push()

	for idx, arg := range values {
		interpreter.insert(fn.definition.Params[idx].Token.Lexeme, arg)
	}

	value := interpreter.Visit(fn.definition.Body)

	if ret, ok := value.(*ReturnValue); ok {
		value = ret.inner
	}

	if thrown, ok := value.(*ThrowValue); ok {
		value = thrown
	}

	interpreter.pop()
	return value
}

func (v *ReturnValue) GetType() Type                                      { return &ReturnType{} }
func (v *ReturnValue) Inspect() string                                    { return v.inner.Inspect() }
func (v *ReturnValue) Copy() Value                                        { return &ReturnValue{inner: v.inner} }
func (v *ReturnValue) Modify(operation lexer.TokenKind, other Value) bool { return false }

func (v *ThrowValue) GetType() Type                                      { return &ThrowableType{} }
func (v *ThrowValue) Inspect() string                                    { return v.inner.Inspect() }
func (v *ThrowValue) Copy() Value                                        { return &ThrowValue{inner: v.inner} }
func (v *ThrowValue) Modify(operation lexer.TokenKind, other Value) bool { return false }

func (v *NativeClassDefValue) GetType() Type                                      { return &ClassDefType{} }
func (v *NativeClassDefValue) Inspect() string                                    { return fmt.Sprintf("<native class %s>", v.Identifier) }
func (v *NativeClassDefValue) Copy() Value                                        { return v }
func (v *NativeClassDefValue) Modify(operation lexer.TokenKind, other Value) bool { return false }

func (def *NativeClassDefValue) Arity() int {
	return len(def.Fields)
}

func (def *NativeClassDefValue) Call(interpreter *Interpreter, values []Value) Value {
	instance := &ClassInstanceValue{Def: def, fields: make(map[string]Value)}

	if def.Arity() != len(values) {
		interpreter.Report("Native class constructor '%s' expected %d arguments but received %d.", def.Identifier, def.Arity(), len(values))
	}

	for idx, id := range def.Fields {
		instance.fields[id] = values[idx]
	}

	return instance
}

func (v *ClassDefValue) GetType() Type                                      { return &ClassDefType{} }
func (v *ClassDefValue) Inspect() string                                    { return fmt.Sprintf("<class %s>", v.identifier) }
func (v *ClassDefValue) Copy() Value                                        { return v }
func (v *ClassDefValue) Modify(operation lexer.TokenKind, other Value) bool { return false }

func (def *ClassDefValue) Arity() int {
	if def.constructor != nil {
		return len(def.constructor.definition.Params)
	}
	return 0
}

func (def *ClassDefValue) Call(interpreter *Interpreter, values []Value) Value {
	instance := &ClassInstanceValue{Def: def, fields: make(map[string]Value)}

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
	var id string

	switch t := v.Def.(type) {
	case *ClassDefValue:
		id = t.identifier
	case *NativeClassDefValue:
		id = t.Identifier
	}

	return fmt.Sprintf("<instance %s : %p>", id, v)
}
func (v *ClassInstanceValue) Copy() Value                                        { return v }
func (v *ClassInstanceValue) Modify(operation lexer.TokenKind, other Value) bool { return false }

func (instance *ClassInstanceValue) Get(identifier string) (Value, bool) {
	if val, ok := instance.fields[identifier]; ok {
		return val, true
	}

	var methods map[string]Value

	switch t := instance.Def.(type) {
	case *ClassDefValue:
		methods = t.methods
	case *NativeClassDefValue:
		methods = t.Methods
	}

	if fn, ok := methods[identifier]; ok {
		// FIXME: Allow bound in nativefn
		if f, ok := fn.(*FunctionValue); ok {
			f.bound = instance
		}
		return fn, true
	}

	if def, ok := instance.Def.(*ClassDefValue); ok {
		if def.base != nil {
			return instance.base.Get(identifier)
		}
	}

	return nil, false
}

func (instance *ClassInstanceValue) Set(identifier string, value Value) (Value, bool) {
	if _, ok := instance.fields[identifier]; ok {
		instance.fields[identifier] = value
		return value, true
	}

	if def, ok := instance.Def.(*ClassDefValue); ok {
		if def.base != nil {
			return instance.base.Set(identifier, value)
		}
	}

	return nil, false
}

func (v *StructDefValue) GetType() Type                                      { return &StructDefType{} }
func (v *StructDefValue) Inspect() string                                    { return fmt.Sprintf("<struct %s>", v.identifier) }
func (v *StructDefValue) Copy() Value                                        { return v }
func (v *StructDefValue) Modify(operation lexer.TokenKind, other Value) bool { return false }

func (def *StructDefValue) Arity() int {
	if def.constructor != nil {
		return len(def.constructor.definition.Params)
	}
	return 0
}

func (def *StructDefValue) Call(interpreter *Interpreter, values []Value) Value {
	instance := &StructInstanceValue{def: def, fields: make(map[string]Value, len(def.fields))}

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

func (v *StructInstanceValue) GetType() Type { return &StructInstanceType{} }
func (v *StructInstanceValue) Inspect() string {
	return fmt.Sprintf("<struct instance %s : %p>", v.def.identifier, v)
}

// Copy semantics on structs
func (v *StructInstanceValue) Copy() Value {
	new := &StructInstanceValue{def: v.def, fields: make(map[string]Value, len(v.fields))}

	for key, value := range v.fields {
		new.fields[key] = value.Copy()
	}

	return new
}
func (v *StructInstanceValue) Modify(operation lexer.TokenKind, other Value) bool { return false }

func (instance *StructInstanceValue) Get(identifier string) (Value, bool) {
	if val, ok := instance.fields[identifier]; ok {
		return val, true
	}

	return nil, false
}

func (instance *StructInstanceValue) Set(identifier string, value Value) (Value, bool) {
	if _, ok := instance.fields[identifier]; ok {
		instance.fields[identifier] = value
		return value, true
	}
	return nil, false
}

func (v *NameSpaceValue) GetType() Type { return &NameSpaceType{} }
func (v *NameSpaceValue) Inspect() string {
	return fmt.Sprintf("<namespace %s>", v.Identifier)
}
func (v *NameSpaceValue) Copy() Value                                        { return v }
func (v *NameSpaceValue) Modify(operation lexer.TokenKind, other Value) bool { return false }

func (v *NameSpaceValue) Get(identifier string) (Value, bool) {
	if val, ok := v.Members[identifier]; ok {
		return val.Copy(), true
	}

	return nil, false
}

func (v *NameSpaceValue) Set(identifier string, value Value) (Value, bool) {
	if _, ok := v.Members[identifier]; ok {
		v.Members[identifier] = value
		return value, true
	}
	return nil, false
}

func (v *ListVal) GetType() Type { return &ListType{} }
func (v *ListVal) Inspect() string {
	var sb strings.Builder

	sb.WriteByte('[')
	for idx, value := range v.Values {
		sb.WriteString(value.Inspect())

		if idx < len(v.Values)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteByte(']')

	return sb.String()
}
func (v *ListVal) Copy() Value { return v }
func (v *ListVal) Modify(operation lexer.TokenKind, other Value) bool {
	if value, ok := other.(*ListVal); ok {
		switch operation {
		case lexer.PLUS_EQUAL:
			for _, item := range value.Values {
				v.Values = append(v.Values, item.Copy())
			}
		}
	}

	return false
}

func (v *ListVal) Set(operation lexer.TokenKind, index int, other Value) (Value, bool) {
	switch operation {
	case lexer.EQUAL:
		v.Values[index] = other
		return other, true
	case lexer.PLUS_EQUAL:
		fallthrough
	case lexer.MINUS_EQUAL:
		fallthrough
	case lexer.STAR_EQUAL:
		fallthrough
	case lexer.SLASH_EQUAL:
		result := v.Values[index].Modify(operation, other)
		return v.Values[index], result
	}

	return nil, false
}

func (v *LoopFlow) GetType() Type { return &LoopFlowType{} }
func (v *LoopFlow) Inspect() string {
	var str string

	if v.exit {
		str = "break"
	} else {
		str = "continue"
	}

	return str
}
func (v *LoopFlow) Copy() Value                                        { return v }
func (v *LoopFlow) Modify(operation lexer.TokenKind, other Value) bool { return false }

func BinopL(operator lexer.TokenKind, a []Value, b []Value) (Value, bool) {
	switch operator {
	case lexer.PLUS:
		list := make([]Value, 0)

		for _, value := range a {
			list = append(list, value.Copy())
		}

		for _, value := range b {
			list = append(list, value.Copy())
		}

		return &ListVal{list}, true
	// FIXME: Better equality checks
	case lexer.EQUAL_EQUAL:
		return &BoolVal{Value: len(a) == len(b)}, true
	case lexer.NOT_EQUAL:
		return &BoolVal{Value: len(a) != len(b)}, true
	}

	// Unreachable
	return nil, false
}

func BinopS(operator lexer.TokenKind, a string, b string) (Value, bool) {
	switch operator {
	case lexer.PLUS:
		return &StringVal{Value: a + b}, true
	case lexer.EQUAL_EQUAL:
		return &BoolVal{Value: a == b}, true
	case lexer.NOT_EQUAL:
		return &BoolVal{Value: a != b}, true
	case lexer.GREATER:
		return &BoolVal{Value: a > b}, true
	case lexer.GREATER_EQUAL:
		return &BoolVal{Value: a >= b}, true
	case lexer.LESS:
		return &BoolVal{Value: a < b}, true
	case lexer.LESS_EQUAL:
		return &BoolVal{Value: a <= b}, true
	}

	// Unreachable
	return nil, false
}

func BinopB(operator lexer.TokenKind, a bool, b bool) (Value, bool) {
	switch operator {
	case lexer.EQUAL_EQUAL:
		return &BoolVal{Value: a == b}, true
	case lexer.NOT_EQUAL:
		return &BoolVal{Value: a != b}, true
	}

	// Unreachable
	return nil, false
}

func BinopI(operator lexer.TokenKind, a int, b int) (Value, bool) {
	switch operator {
	case lexer.PLUS:
		return &IntVal{Value: a + b}, true
	case lexer.MINUS:
		return &IntVal{Value: a - b}, true
	case lexer.STAR:
		return &IntVal{Value: a * b}, true
	case lexer.SLASH:
		return &IntVal{Value: a / b}, true
	case lexer.EQUAL_EQUAL:
		return &BoolVal{Value: a == b}, true
	case lexer.NOT_EQUAL:
		return &BoolVal{Value: a != b}, true
	case lexer.GREATER:
		return &BoolVal{Value: a > b}, true
	case lexer.GREATER_EQUAL:
		return &BoolVal{Value: a >= b}, true
	case lexer.LESS:
		return &BoolVal{Value: a < b}, true
	case lexer.LESS_EQUAL:
		return &BoolVal{Value: a <= b}, true
	}

	// Unreachable
	return nil, false
}

func BinopF(operator lexer.TokenKind, a float32, b float32) (Value, bool) {
	switch operator {
	case lexer.PLUS:
		return &FloatVal{Value: a + b}, true
	case lexer.MINUS:
		return &FloatVal{Value: a - b}, true
	case lexer.STAR:
		return &FloatVal{Value: a * b}, true
	case lexer.SLASH:
		return &FloatVal{Value: a / b}, true
	case lexer.EQUAL_EQUAL:
		return &BoolVal{Value: a == b}, true
	case lexer.NOT_EQUAL:
		return &BoolVal{Value: a != b}, true
	case lexer.GREATER:
		return &BoolVal{Value: a > b}, true
	case lexer.GREATER_EQUAL:
		return &BoolVal{Value: a >= b}, true
	case lexer.LESS:
		return &BoolVal{Value: a < b}, true
	case lexer.LESS_EQUAL:
		return &BoolVal{Value: a <= b}, true
	}

	// Unreachable
	return nil, false
}

func Equality(left Value, right Value) bool {
	if reflect.TypeOf(left) != reflect.TypeOf(right) {
		return false
	}

	switch t := left.(type) {
	case *IntVal:
		return t.Value == right.(*IntVal).Value
	case *FloatVal:
		return t.Value == right.(*FloatVal).Value
	case *BoolVal:
		return t.Value == right.(*BoolVal).Value
	case *StringVal:
		return t.Value == right.(*StringVal).Value
	case *FunctionValue:
		return t.definition != right.(*FunctionValue).definition
	case *ListVal:
		// FIXME: Better equality
		return len(t.Values) == len(right.(*ListVal).Values)
	case *NameSpaceValue:
		// FIXME: Better equality
		return t.Identifier == right.(*NameSpaceValue).Identifier
	case *ClassDefValue:
		return t.identifier == right.(*ClassDefValue).identifier
	case *ClassInstanceValue:
		return Equality(t.Def, right.(*ClassInstanceValue).Def)
	case *StructDefValue:
		return t.identifier == right.(*StructDefValue).identifier
	case *StructInstanceValue:
		return Equality(t.def, right.(*StructInstanceValue).def)
	}

	return false
}

// Helpers

func checkNumericOperand(interpreter *Interpreter, token *lexer.Token, operand Value) {
	switch operand.(type) {
	case *IntVal:
		return
	case *FloatVal:
		return
	default:
		interpreter.Report("Value '%s' is not a numeric value '%s':%s %d", token.Lexeme, operand.Inspect(), reflect.TypeOf(operand), token.Line)
	}
}

func checkBoolOperand(interpreter *Interpreter, token *lexer.Token, operand Value) {
	switch operand.(type) {
	case *BoolVal:
		return
	default:
		interpreter.Report("Value '%s' is not a boolean value '%s'", token.Lexeme, operand.GetType())
	}
}
