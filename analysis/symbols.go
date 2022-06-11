package analysis

type Symbol interface {
	GetName() string
}

type VarSymbol struct {
	identifier string
}

type FunctionSymbol struct {
	identifier string
}

func (v *VarSymbol) GetName() string {
	return v.identifier
}

func (fn *FunctionSymbol) GetName() string {
	return fn.identifier
}
