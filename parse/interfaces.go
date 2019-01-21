package parse

import d "github.com/JoergReinhardt/godeep/data"
import f "github.com/JoergReinhardt/godeep/functions"
import l "github.com/JoergReinhardt/godeep/lang"

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
	f.Function
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

////////// STACK ////////////
//// LAST IN FIRST OUT //////
type Stacked interface {
	f.Collected
	Push(f.Function)
	Pop() f.Function
	Add(...f.Function)
}

///////// QUEUE /////////////
//// FIRST IN FIRST OUT /////
type Queued interface {
	f.Collected
	Put(f.Function)
	Pull() f.Function
	Append(...f.Function)
}

// interfaces dealing with instances of input items
type Queue interface {
	Next()
	Peek() l.Item
	Current() l.Item
	Idx() int
}

// data to parse
type Token interface {
	Type() d.BitFlag
	Flag() d.BitFlag
	String() string
}
