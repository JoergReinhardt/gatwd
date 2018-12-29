package types

//go:generate stringer -type FixType
type FixType BitFlag

//go:generate stringer -type Arity
type Arity int8 // <-- this sounds so terribly wrong, for anyone born in germany

const (
	Nullary Arity = 0 + iota
	Unary
	Binary
	Trinary
	Ternary
	Quaternary
	Quinary
	Senary
	Septenary
	Octonary
	Novenary
	Denary
	Polyadic

	PreFix FixType = 0 + iota
	InFix
	PostFix
	ConFix
)

/////// RECURSIVE COLLECTION ///////
type recol func(p ...Attribute) (Data, recol)

// construt new recursive from n data elements
func conRecursive(d ...Data) recol {
	var head Data
	var tail recol
	switch len(d) {
	case 0:
		head = nilVal{}
		tail = nil
	case 1:
		head = d[0]
		tail = nil
	default:
		head = d[0]
		tail = conRecursive(d[1:]...)
	}
	return func(a ...Attribute) (Data, recol) {
		h := head
		t := tail
		return h, t
	}

}

//// standalone functions to implement recursive behaviour on recursively nested closures
func recolHead(r recol) Data  { h, _ := r(); return h }
func recolTail(r recol) recol { _, t := r(); return t }
func recolEmpty(r recol) bool {
	if r != nil {
		if !elemEmpty(recolHead(r)) {
			return false
		}
		if !recolEmpty(recolTail(r)) {
			return false
		}
	}
	return true
}

// recursive behaviour as method to implement interfaces
func (r recol) Eval() Data      { return r }
func (r recol) Flag() BitFlag   { return (List.Flag() | r.Head().Flag()) }
func (r recol) Head() Data      { return recolHead(r) } // --> current head
func (r recol) Tail() Recursive { return recolTail(r) } // --> current tail
func (r recol) Empty() bool     { return recolEmpty(r) }

////////////////////////////////////////////////////////

type tuple chain
