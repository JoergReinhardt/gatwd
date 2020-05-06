package main

import (
	"math/bits"
)

//// CATEGORY OBJECT INTERFACE AND UNIT TYPES
///
// every thing is an object and needs to implement the object interface 'Obj'.
// that includes internal parts of the type system, like type markers and
// names.  the interface demands a 'Type() int' method to return a unique
// numeric identification.
//
// the 'Ident() Obj' method needs to be implemented to return the native
// instance of whatever type implements the interface, aka it-'self'.
//
//   - runtime defined types are accessed by slice index.
//
//   - category types need quick set membership identification
//
//   - some kinds are named, others need anonymity.
//
// hence three sorts of identity markers exist:
//
//   - numeric unique id shared by every kind of type.
//
//   - binary bit flag for sets of categorys, with quick membership operation
//
//   - string representation of instance data, or name of its type
//
//\\
//// TYPE FLAG IMPLEMENTATION
///
// flags mark all kinds of category and provide the zero value for, most other
// product types (tuple/struct) and the type of categorys as such.
//
// the parent objects id is expressed in in the value rank of the flag within
// its set of constants of the same type.
//
// the flags string representation may be identical with the type lable, or
// category name, but not mandatory.  the private 'sym() Sym' method returns an
// intermediate symbol instance for convienience, composing flag rank and
// string representation to conform to the symbol interface, to be used by type
// constructors it gets passed to.  all categorys, one sub type of which is
// 'types', are instances of flg.  all category and parametric types have
// 'None' as their zero value within their own category and whatever the flag
// value is, in the parent category. sub types of the type category are most
// tuple/record/struct fields (may be implemented by hash map instead) and all
// parametric types including all base cases of sub-types of recursive
// parametric types.
//
// some native type like bool, or int is to be expected at the
// definition swamp floor… (might be a tur<d|tel>).
//\\
//// TYPE MARKER IMPLEMENTATIONS
///  ident methods return the actual instances uid and the instance itself.
//
// in case of a type marker, its the marker itself and it needs to implement
// ident in order to be viable argument for operations of the type system.
// when enclosed with an actual instance…
//
// !!!TYPE MARKER UID & IDENT METHODS NEED TO BE SHADOWED BY ENCLOSING OBJECT!!!
//
//\
/// FLAG INTERFACE
// instance methods
func isSet(f Flg) bool           { return card(f) > 1 }
func hasFlag(set, flag Flg) bool { return set&flag != 0 }
func card(f Flg) int             { return bits.OnesCount(uint(f)) }
func rank(f Flg) int             { return bits.Len(uint(f)) }
func consSet(f Flg, fs ...Flg) Set {
	if len(fs) > 0 {
		var set uint
		for _, flag := range fs {
			if isSet(flag) {
				fs = append(fs, splitSet(flag)...)
			}
			set = set | uint(flag)
		}
		return Set(set)
	}
	return Set(f)
}
func splitSet(f Flg) []Flg {
	var set = make([]Flg, 0, rank(f))
	if rank(f) > 1 {
		var flag = f
		for f != Flg(None) {
			var flag = Flg(f)
			f = f & Flg(flag)
			set = append(set, Flg(flag))
		}
		set = append(set, Flg(flag))
	}
	return set
}

type (
	Flg uint
	Set Flg // ∏ Flg₁|Flgₙ|…|Flgₙ

	Sym func() (int, string)
	Tab []Sym // ∏ Sym₁|Sym₂|…|Symₙ

	Obj interface {
		Uid() int   // ← cardinality
		Ident() Obj // ← (∑|∏ → unit | Flg → identity)
	}
	Sum Obj // ∑ [Objₜ]
	Pro Obj // ∏ Obj₁|Obj₂|…|Objₙ

	Const func() Obj
	Pair  func() (Obj, Obj)
	Link  func() (Obj, Pair)
	Val   func() (Sym, Obj)

	UnaOp func(Obj) Obj
	BinOp func(a, b Obj) Obj
	GenOp func(...Obj) Obj
)

/// CATEGORY FLAG IMPLEMENTATION
type tflg uint32

func (f tflg) Uid() int      { return rank(Flg(f)) }
func (f tflg) Lable() string { return f.String() }
func (f tflg) Ident() Obj    { return f }

/// SYMBOL
func (s Sym) Uid() int      { var uid, _ = s(); return uid }
func (s Sym) Lable() string { var _, name = s(); return name }

//go:generate stringer -type tflg
const (
	None tflg = 1<<iota - 1
	Truth
	Uint
	Int
	Flt
	Img
	Time
	Span
	Byte
	Rune
	Bytes
	String
	Vector
	Tuple
	List
	Zero
	One
	Min
	Max
	Fnc // map one classes objects onto another class
	Op  // operate on (take|return) objects of one class
	Just
	Either
	Or

	// truth values
	False bool = false
	True  bool = true

	Num = Uint | Int | Flt | Img // numeric
	Tmp = Time | Span            // temporal
	Bin = Byte | Bytes           // binary
	Txt = Rune | String          // textual
	Col = Lim | Fnc | Op         // collection
	Itv = Zero | One             // interval
	Lim = Min | Max              // limit
	Par = Fnc | Op               // parameterized

	Maybe = Just | None
	Alter = Either | Or

	Bounds = Lim | Itv

	T = Maybe | Truth | Num | Tmp | Bin | Txt |
		Col | Lim | Par | Itv | Maybe | Alter

	//  unit, neutral & sign
	Neg int = -1 // negative
	Amb int = 0  // ambivalent
	Pos int = 1  // positive
)

func main() {}
