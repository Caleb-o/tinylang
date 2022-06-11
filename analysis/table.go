package analysis

type SymbolTable struct {
	parent  *SymbolTable
	symbols map[string]Symbol
}

func NewTable(parent *SymbolTable) *SymbolTable {
	return &SymbolTable{parent: parent, symbols: make(map[string]Symbol)}
}

func (st *SymbolTable) Contains(identifier string) bool {
	_, ok := st.symbols[identifier]
	return ok
}

func (st *SymbolTable) Insert(identifier string, sym Symbol) {
	st.symbols[identifier] = sym
}

func (st *SymbolTable) Lookup(identifier string, local bool) Symbol {
	if _, ok := st.symbols[identifier]; ok {
		return st.symbols[identifier]
	}

	if local || st.parent == nil {
		return nil
	}

	return st.parent.Lookup(identifier, false)
}
