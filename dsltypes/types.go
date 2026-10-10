package dsltypes

// TyComp represents a Composed Type. 
// It is used to represent Product Types (Tuples, Records) and complex
// functional signatures (Arguments -> Return Type).
type TyComp struct {
	Flags []BitFlag
}

// NewTyComp creates a new composed type out of multiple BitFlags.
func NewTyComp(flags ...BitFlag) TyComp {
	return TyComp{Flags: flags}
}

// Match checks if another TyComp matches this one exactly or conceptually.
func (t TyComp) Match(other TyComp) bool {
	if len(t.Flags) != len(other.Flags) {
		return false
	}
	for i := range t.Flags {
		if !t.Flags[i].Match(other.Flags[i]) {
			return false
		}
	}
	return true
}

// ASTNode represents the boundary between the Host Engine and the Guest DSL.
// It pairs raw Go value (generically typed for the host) with its runtime DSL type.
type ASTNode[T any] struct {
	Value   T
	DSLType BitFlag
}

// NewASTNode creates a new dynamically typed DSL node.
func NewASTNode[T any](val T, dslType BitFlag) ASTNode[T] {
	return ASTNode[T]{
		Value:   val,
		DSLType: dslType,
	}
}

// Match provides a safe way to evaluate if the AST Node adheres to a specific quasi-type.
func (n ASTNode[T]) Match(expected BitFlag) bool {
	return n.DSLType.Match(expected)
}
