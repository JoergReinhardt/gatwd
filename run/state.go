package run

import (
	d "github.com/JoergReinhardt/gatwd/data"
	f "github.com/JoergReinhardt/gatwd/functions"
)

// length is the length of the entire frame including header in segments, where
// each segment is uint64, so that length + 1 is the address of the info header
// of the next frame
type Length d.Uint8Val

func (l Length) TypePrime() d.TyNative       { return d.Flag }
func (l Length) TypeFnc() f.TyFnc            { return f.HigherOrder }
func (l Length) Flag() d.BitFlag             { return l.TypeFnc().Flag() }
func (l Length) Eval(a ...d.Native) d.Native { return l.Flag() }
func (l Length) Int() int                    { return int(l) }

// layout is a flag marking which of the arguments are pointer
type Layout d.Uint32Val

func (l Layout) TypePrime() d.TyNative       { return d.Flag }
func (l Layout) TypeFnc() f.TyFnc            { return f.HigherOrder }
func (l Layout) Flag() d.BitFlag             { return l.TypeFnc().Flag() }
func (l Layout) Eval(a ...d.Native) d.Native { return l.Flag() }
func (l Layout) Int() int                    { return int(l) }

type Ftype d.Uint8Val

func (t Ftype) Eval(...d.Native) d.Native { return t }
func (t Ftype) TypeFnc() f.TyFnc          { return f.Type }
func (t Ftype) TypeNat() d.TyNative       { return d.Flag }
func (t Ftype) Flag() d.BitFlag           { return d.BitFlag(uint64(d.Uint8Val(t).Uint())) }
func (t Ftype) Match(f Ftype) bool {
	if t&f != 0 {
		return true
	}
	return false
}

//go:generate stringer -type=Ftype
const (
	Select Ftype = 1
	Update Ftype = 1 << iota
	Continuation
	Return
)

// INFO TABLE
// how many arguments are expected (also see layout)
type Arity d.Uint8Val

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

func (a Arity) Eval(v ...d.Native) d.Native { return a }
func (a Arity) Int() int                    { return int(a) }
func (a Arity) Flag() d.BitFlag             { return d.BitFlag(a) }
func (a Arity) TypeNat() d.TyNative         { return d.Flag }
func (a Arity) TypeFnc() f.TyFnc            { return f.HigherOrder }
func (a Arity) Match(arg Arity) bool        { return a == arg }

// properys relevant for application
type Propertys d.Uint8Val

//go:generate stringer -type Propertys
const (
	Default Propertys = 0
	PostFix Propertys = 1
	InFix   Propertys = 1 + iota
	// ⌐: PreFix
	Atomic
	// ⌐: Composit
	Eager
	// ⌐: Lazy
	Right
	// ⌐: Left_Bound
	Mutable
	// ⌐: Imutable
	SideEffect
	// ⌐: Pure
	Data
	// ⌐: Function
)

func (p Propertys) TypePrime() d.TyNative       { return d.Flag }
func (p Propertys) TypeFnc() f.TyFnc            { return f.HigherOrder }
func (p Propertys) Flag() d.BitFlag             { return p.TypeFnc().Flag() }
func (p Propertys) Eval(a ...d.Native) d.Native { return p.Flag() }
func (p Propertys) Match(arg Propertys) bool {
	if p&arg != 0 {
		return true
	}
	return false
}

type Info struct {
	Length    // Uint8
	Arity     // uint8	      ‥.
	Propertys // Uint8	      ‥.
	Ftype     // uint8
	Layout    // uint32
}

func newInfo(
	length Length,
	arity Arity,
	props Propertys,
	ftype Ftype,
	layout Layout,
) Info {
	return Info{
		length,
		arity,
		props,
		ftype,
		layout,
	}
}
