package gatw

import (
	"math/bits"
)

// FUNCTION DEFINITION{{{
// ARITY CONSTANTS
//go:generate stringer -type Ari
type Ari int8

const (
	Variadic Ari = -1

	Nullary Ari = 0 + iota
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
	Multary

	Arity = Variadic | Nullary | Unary | Binary | Ternary |
		Quaternary | Quinary | Senary | Septenary |
		Octonary | Novenary | Denary | Multary
)

func (e Ari) Identity() Item { return e }
func (e Ari) Type() Type     { return Arity }
func (e Ari) Name() Key      { return Key(e.String()) }
func (e Ari) Id() Id         { return Id(e) }
func (e Ari) Cons(args ...Item) (Item, Cons) {
	if len(args) > 0 {
		if len(args) > 1 {
			var items = make([]Item, 0, len(args))
			for _, arg := range args {
				if ari, ok := arg.(Ari); ok {
					items = append(items, ari)
				}
				return EnumVal(items), e.Cons
			}
		}
		if ari, ok := args[0].(Ari); ok {
			return ari, e.Cons
		}
	}
	return e, e.Cons
}
func (e Ari) Compare(arg Ari) Rank {
	if e != arg {
		if e < arg {
			return Lesser
		}
		if e > arg {
			return Greater
		}
	}
	return Equal
}

///}}}
// FIXITY CONSTANTS{{{
//go:generate stringer -type Fixity
type Fixity uint8

const (
	Infix Fixity = 1<<iota - 1
	Prefix
	Postfix

	Fixities = Infix | Prefix | Postfix
)

// fixity of expression
func (e Fixity) Identity() Item { return e }
func (e Fixity) Type() Type     { return Enum }
func (e Fixity) Uint() uint     { return uint(e) }
func (e Fixity) Len() int       { return bits.Len(uint(e)) }
func (e Fixity) Id() Id         { return Id(bits.Len(uint(e))) }
func (e Fixity) Compare(arg Fixity) Rank {
	if e != arg {
		if e < arg {
			return Lesser
		}
		if e > arg {
			return Greater
		}
	}
	return Equal
}

func (e NnaFnc) Identity() Item    { return e }
func (e NnaFnc) Type() Type        { return Function }
func (e NnaFnc) Arity() Ari        { return Nullary }
func (e NnaFnc) Call(...Item) Item { return e() }
func (e NnaFnc) String() string    { return e.Arity().String() }

func (e UnaFnc) Identity() Item { return e }
func (e UnaFnc) Type() Type     { return Function }
func (e UnaFnc) Arity() Ari     { return Unary }
func (e UnaFnc) String() string { return e.Arity().String() }
func (e UnaFnc) Call(args ...Item) Item {
	if len(args) == 1 {
		return e(args[0])
	}
	return None
}

func (e BinFnc) Identity() Item { return e }
func (e BinFnc) Type() Type     { return Function }
func (e BinFnc) Arity() Ari     { return Binary }
func (e BinFnc) String() string { return e.Arity().String() }
func (e BinFnc) Call(args ...Item) Item {
	if len(args) == 2 {
		return e(args[0], args[1])
	}
	return None
}

func (e NarFnc) Identity() Item { return e }
func (e NarFnc) Type() Type     { return Function }
func (e NarFnc) Arity() Ari     { return Multary }
func (e NarFnc) String() string { return e.Arity().String() }
func (e NarFnc) Call(args ...Item) Item {
	if len(args) > 0 {
		return e(args...)
	}
	return None
}

func Define(fnc Fnc, ret Type, pat ...Type) Def {
	return func(args ...Item) (Item, Pattern) {
		if len(args) > 0 {
		}
		return fnc, Pattern(append(pat, ret))
	}
}
func (d Def) Fnc() Fnc               { f, _ := d(); return f.(Fnc) }
func (d Def) Pattern() Pattern       { _, p := d(); return p }
func (d Def) Identity() Item         { return d }
func (d Def) Len() int               { return len(d.Pattern()) }
func (d Def) Arity() Ari             { return Ari(d.Len() - 1) }
func (d Def) Type() Type             { return d.Pattern()[d.Len()-1] }
func (d Def) Call(args ...Item) Item { return d.Fnc().Call(args...) }
