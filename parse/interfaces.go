package parse

import d "github.com/JoergReinhardt/godeep/data"
import f "github.com/JoergReinhardt/godeep/functions"

type Arity uint8

func (a Arity) Flag() d.BitFlag { return d.BitFlag(a) }

//go:generate stringer -type Arity
const (
	Nullary Arity = 0 + iota
	Unary
	Binary
	Ternary
	Quaternary
	Quinary
	Senary
	Septenary
	Octonary
	Novenary
	Denary
)

type FncDef interface {
	f.Functional
	UID() int
	Name() string
	Arity() Arity
	Fix() Property
	Lazy() Property
	Bound() Property
	Mutable() Property
	AccessType() Property
	RetType() d.BitFlag
}

type Property d.BitFlag

func (p Property) Flag() d.BitFlag { return d.BitFlag(p) }

//go:generate stringer -type Property
const (
	PostFix Property = 1
	InFix   Property = 1 << iota
	PreFix
	///
	Eager
	Lazy
	///
	Right_Bound
	Left_Bound
	///
	Mutable
	Imutable
	///
	Effected
	Pure
	////
	Positional
	NamedArgs
	////
	Lesser
	Equal
	Greater
)

// data to parse
type Token interface {
	TokType() TokType
	Flag() d.BitFlag
	String() string
}

// Ident interface{}
//
// the ident interface is implemented by everything providing unique identification.
type Ident interface {
	f.Functional
	Ident() f.Function // calls enclosed fnc, with enclosed parameters
}

// the tree function type contains a node and can be referenced by name as
// parent of the contained node.
type Tree interface {
	Token
	Parent() Node
}
