package analysis

import "tiny/ast"

type Symbol interface {
	GetName() string
}

type VarSymbol struct {
	identifier string
	mutable    bool
}

type FunctionSymbol struct {
	identifier string
	def        *ast.FunctionDef
}

type ClassDefSymbol struct {
	identifier string
	def        *ast.ClassDef
}

type StructDefSymbol struct {
	identifier string
	def        *ast.StructDef
}

type NameSpaceSymbol struct {
	identifier string
}

func (s *VarSymbol) GetName() string {
	return s.identifier
}

func (s *FunctionSymbol) GetName() string {
	return s.identifier
}

func (s *ClassDefSymbol) GetName() string {
	return s.identifier
}

func (s *StructDefSymbol) GetName() string {
	return s.identifier
}

func (s *NameSpaceSymbol) GetName() string {
	return s.identifier
}
