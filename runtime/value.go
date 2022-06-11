package runtime

import (
	"fmt"
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
	TYPE_STRUCT
	TYPE_NAMESPACE
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
