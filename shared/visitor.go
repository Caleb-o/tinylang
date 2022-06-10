package shared

// This may or may not be required, as the interpreter will need to
// return a value whereas the analyser does not
type AstVisitor interface {
	Visit()
}
