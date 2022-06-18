package runtime

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
