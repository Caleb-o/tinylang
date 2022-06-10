package runtime

import (
	"fmt"
)

type Type uint8

const (
	TYPE_INT Type = iota
	TYPE_FLOAT
	TYPE_BOOL
	TYPE_CHAR
	TYPE_STRING
	TYPE_LIST
	TYPE_STRUCT
	TYPE_NAMESPACE
)

type Value interface {
	GetType() Type
	Inspect() string
}

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

func (i *IntVal) GetType() Type   { return TYPE_INT }
func (i *IntVal) Inspect() string { return fmt.Sprintf("%d", i.Value) }

func (f *FloatVal) GetType() Type   { return TYPE_FLOAT }
func (f *FloatVal) Inspect() string { return fmt.Sprintf("%f", f.Value) }

func (b *BoolVal) GetType() Type   { return TYPE_BOOL }
func (b *BoolVal) Inspect() string { return fmt.Sprintf("%t", b.Value) }

func (str *StringVal) GetType() Type   { return TYPE_STRING }
func (str *StringVal) Inspect() string { return str.Value }
