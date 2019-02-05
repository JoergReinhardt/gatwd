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

type Property d.BitFlag

func (p Property) Lazy() Property {
	if Eager == 1 {
		return Eager
	}
	return Lazy
}
func (p Property) Bound() Property {
	if Right_Bound == 1 {
		return Right_Bound
	}
	return Left_Bound
}
func (p Property) Mutable() Property {
	if Mutable == 1 {
		return Mutable
	}
	return Imutable
}
func (p Property) Effected() Property {
	if Effected == 1 {
		return Effected
	}
	return Pure
}
func (p Property) AccessorType() Property {
	if PosiArgs == 1 {
		return PosiArgs
	}
	return NamedArgs
}
func (p Property) Fix() Property {
	switch p {
	case PostFix:
		return PostFix
	case InFix:
		return InFix
	case PreFix:
		return PreFix
	}
	return PreFix
}

func (p Property) TypePrim() d.BitFlag { return d.BitFlag(p) }

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
	PosiArgs
	NamedArgs
	////
	Static
	Dynamic
	///////////////

	Default = PostFix | Lazy | Left_Bound |
		Imutable | Pure | PosiArgs | Dynamic

	DefaultStatic = PostFix | Lazy | Left_Bound |
		Imutable | Pure | PosiArgs | Static
)

// data to parse
type Token interface {
	TypeTok() TyToken
	TypePrim() d.BitFlag
	String() string
}

// Ident interface{}
//
// the ident interface is implemented by everything providing unique identification.
type Ident interface {
	f.Value
	Ident() f.Callable // calls enclosed fnc, with enclosed parameters
}
