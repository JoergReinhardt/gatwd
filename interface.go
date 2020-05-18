package gatw

import (
	"math/bits"
	"strings"
)

type (
	//// ELEMENT OBJECT VALUE {{{
	Elem interface {
		Ident() Elem
	}
	Obj interface {
		Elem
		Id() int
	}
	Val interface {
		Elem
		Name() string
	}
	ElemFnc func() Elem
	ObjFnc  func() (int, Elem)
	ValFnc  func() (Name, Obj)

	//// IDENTITY FLAG NAME
	// numeric identity & lists there of
	Id  int
	Uid []int

	// symbol & list of symbols
	Name string
	Path []string

	ConPath func(Path, ...Name) Path

	// flag, flag set (bitwise OR set) & list of flags
	Flg  func() uint
	FSet func() uint
	FLst func() []uint

	ConFSet func(FLst, ...Flg) FSet
	ConFLst func(FLst, ...Flg) FLst

	//// FUNCTION ////////
	Define func(
		name string,
		expr Elem,
		args ...Elem,
	) Definition

	Definition func() (
		value Val,
		args []Elem,
	)

	//// VALUE ////
	Let func(string, Elem) Val

	////////////////////////
	/// LITERAL LEX PARSE /
	Parse func(string) Litrl
	Litrl func() (Id, string)

	//// CATEGORY FUNCTIONS VALUES LITERALS ELEMENTS & OBJECTS //////////
	Unit func(Elem) Val

	//// COMPOSED {{{
	/// SUM TYPES
	Sum    func() (Elem, Pair)
	ConSum func(Sum, ...Elem) Sum

	//// PRODUCT TYPES ////
	// TUPLE
	Tup    []Elem
	ConTup func([]Elem, ...Elem) Tup
	// }}}

	/////////////////
	// PAIR
	Pair func() (l, r Elem)

	// RECORD
	Record map[Name]Obj
	Field  func() (Name, Elem)
	Fields []Field

	// SEQUENCE
	Cons func([]Elem, ...Elem) Lst
	Conc func(Lst, ...Elem) Lst

	// list,sequence, reverse & flat
	Lst func() (Elem, Lst)
	Rev func() (Elem, Lst)
	Flt func() (Elem, Lst)

	Append func(Elem, ...Elem) Lst
	Concat func(Lst, ...Elem) Lst

	// elements mapped to Names
	Set    func() map[string]Elem
	ConSet func(Set, ...Elem) Set //}}}

	//// FUNCTIONS {{{
	// arity
	Unary  func(Elem) Elem
	Binary func(x, y Elem) Elem
	Nary   func(...Elem) Elem

	DefUnary  func(Lst, Elem) Unary
	DefBinary func(n Lst, x, y Elem) Binary
	DefNary   func(Lst, ...Elem) Nary
	//}}}

	//// MONADIC ////{{{
	/// atoren (alig, termin…)
	// gener|accumul-ator
	Generate func() (Elem, Generate)
	Accumult func(e Elem) (Elem, Accumult)

	ConsGenRtr func(func() (Elem, Generate)) Generate
	DefAccMltr func(func(Elem) (Elem, Accumult)) Accumult
	//}}}

	//// EMPTY | NO-OP {{{
	NTA struct{}
	NTO func() //}}}
)

/// IDENTITY METHODS {{{
func (o ObjFnc) Id() int         { i, _ := o(); return i }
func (i ObjFnc) Ident() Elem     { return i }
func (i ValFnc) Ident() Elem     { return i }
func (i ElemFnc) Ident() Elem    { return i }
func (i Sum) Ident() Elem        { return i }
func (i Pair) Ident() Elem       { return i }
func (i Lst) Ident() Elem        { return i }
func (i Set) Ident() Elem        { return i }
func (i Definition) Ident() Elem { return i }
func (i Define) Ident() Elem     { return i }

func (i Concat) Ident() Elem { return i }
func (i ConSet) Ident() Elem { return i }

// identy & uniqueness
func (i Id) Ident() Elem   { return i }
func (i Uid) Ident() Elem  { return i }
func (i Flg) Ident() Elem  { return i }
func (i FSet) Ident() Elem { return i }
func (i FLst) Ident() Elem { return i }
func (i Name) Ident() Elem { return i }
func (i Path) Ident() Elem { return i }

func (i ConFSet) Ident() Elem { return i }
func (i ConFLst) Ident() Elem { return i }

// composition and container
func (i Tup) Ident() Elem    { return i }
func (i ConTup) Ident() Elem { return i }

// function prototypes
func (i DefUnary) Ident() Elem   { return i }
func (i DefBinary) Ident() Elem  { return i }
func (i DefNary) Ident() Elem    { return i }
func (i ConsGenRtr) Ident() Elem { return i }
func (i DefAccMltr) Ident() Elem { return i }

func (i Unary) Ident() Elem    { return i }
func (i Binary) Ident() Elem   { return i }
func (i Nary) Ident() Elem     { return i }
func (i Generate) Ident() Elem { return i }
func (i Accumult) Ident() Elem { return i }

var (
	NE NTA
	NL Lst
)

/// initialize empty slice and list {{{
func init() {
	NE = NTA(struct{}{})
	NL = func() (Elem, Lst) { return NE, NL }
}

func (NTA) Ident() Elem    { return NE }
func (NTA) Symbol() string { return "¬" }
func (NTA) Name() string   { return "" }
func (NTA) Uid() []int     { return []int{} }
func (NTA) Id() int        { return 0 }
func (NTA) Len() int       { return 0 }
func (NTA) Empty() bool    { return true }

func flgUint(f Flg) uint { return f() }
func flgInt(f Flg) int   { return bits.Len(f()) }

func uintFlg(u uint) Flg { return func() uint { return u } }
func intFlg(i int) Flg   { return func() uint { return 1 << uint(i) } }
func idFlg(i int) Flg    { return func() uint { return 1 << uint(i) } }

func (s Flg) String() string  { return "undefined" }
func (s Flg) Has(f uint) bool { return s()&^f != 0 }
func (f Flg) Id() int         { return bits.OnesCount(f()) }
func (f Flg) Len() int        { return f.Len() }
func (f Flg) Head() int       { return bits.Len(f()) }
func (f Flg) Atom() bool      { return f.Head() == 1 }
func (f Flg) Path() []uint {
	var (
		u uint
		p = make([]uint, 0, f.Len())
	)
	for i := 0; i < f.Len(); i++ {
		if u = 1 << uint(i); f.Has(u) {
			p = append(p, u)
		}
	}
	return p
}

func (u Uid) Len() int   { return len(u) }
func (u Uid) Atom() bool { return len(u) == 1 }
func (u Uid) Head() Id   { return Id(u[len(u)-1]) }
func (u Uid) Path() []int {
	if len(u) > 1 {
		return u[:len(u)-1]
	}
	return []int{}
}

func conName(s string) Name     { return Name(s) }
func conPath(ss ...string) Path { return ss }
func (n Path) Len() int         { return len(n) }
func (n Path) Strings() []string {
	var strs = make([]string, 0, n.Len())
	for _, s := range n {
		strs = append(strs, s)
	}
	return strs
}
func (n Path) Atom() bool   { return n.Len() == 1 }
func (n Path) Name() string { return n.Strings()[n.Len()-1] }
func (n Path) Full() string { return strings.Join(n.Strings(), ".") }
func (n Path) Path() []Name {
	var s = []Name{}
	if n.Len() > 1 {
		for _, name := range n[:n.Len()-1] {
			s = append(s, Name(name))
		}
	}
	return s
}

// }}}
// }}}

func ConObj(i int, e Elem) Obj   { return ObjFnc(func() (int, Elem) { return i, e }) }
func ConVal(n string, o Obj) Val { return ValFnc(func() (Name, Obj) { return Name(n), o }) }

func (v ValFnc) Val() Name    { n, _ := v(); return n }
func (v ValFnc) Name() string { return string(v.Name()) }

func (d Definition) Value() Val   { val, _ := d(); return val }
func (d Definition) Args() []Elem { _, args := d(); return args }
func (d Definition) Name() string { return d.Value().Name() }

func Empty(l Lst) bool { return l.Empty() }

func (l Lst) Empty() bool { return EmptyL(l) }

func Head(l Lst) Elem { h, _ := l(); return h }
func Tail(l Lst) Lst  { _, s := l(); return s }
func Next(l Lst) Elem { _, s := l(); return Head(s) }
func EmptyL(l Lst) bool {
	if l.Ident() == NE {
		return true
	}
	return false
}

func ConPair(l, r Elem) Pair {
	return func() (a, b Elem) { return l, r }
}

func conTupl(e ...Elem) Tup { return e }
func (t Tup) Fst() Elem     { return Fst(t) }
func (t Tup) Scnd() Elem    { return Scnd(t) }
func (t Tup) Last() Elem    { return Last(t) }
func (t Tup) Prior() Elem   { return Last(t) }

func Fst(t Tup) Elem {
	var tup = t
	if len(tup) > 0 {
		return tup[0]
	}
	return conTupl()
}

func Scnd(t Tup) Elem {
	var tup = t
	if len(tup) > 1 {
		return tup[1]
	}
	return conTupl()
}

func Last(t Tup) Elem {
	var (
		tup = t
		l   = len(tup)
	)

	if l > 0 {
		return tup[l-1]
	}
	return conTupl()
}

func Prior(t Tup) Tup {
	var tup = t
	if len(tup) > 1 {
		return conTupl(tup[1:]...)
	}
	return conTupl()
}

func ConcatL(l, r Lst) Lst {
	var (
		head Elem
		tail Lst
	)
	// HEAD LEFT
	if head = Head(l); head != NE {
		if tail = Tail(l); tail != nil {
			// TAIL LEFT
			return func() (Elem, Lst) {
				return head, ConcatL(tail, r)
			}
		} else {
			// TAIL RIGHT
			return func() (Elem, Lst) {
				return head, r
			}
		}

	} else { // HEAD RIGHT
		if head = Head(r); head != nil {
			// TAIL RIGHT
			if tail = Tail(r); tail != nil {
				return func() (Elem, Lst) {
					return head, tail
				}
			}
			// NO TAIL
			return func() (Elem, Lst) {
				return head, NL
			}
		}
	}
	// NO HEAD
	return NL
}

func AppendL(l Lst, e ...Elem) Lst {
	if len(e) > 0 {
		if len(e) > 1 {
			return func() (Elem, Lst) {
				return e[0], AppendL(
					l, ConLst(e[1:]...))
			}
		}
		return func() (Elem, Lst) {
			return e[0], l
		}
	}
	return l
}

func ConLst(e ...Elem) Lst {
	if len(e) > 0 {
		if len(e) > 1 {
			return func() (Elem, Lst) {
				return e[0], ConLst(e[1:]...)
			}
		}
		return func() (Elem, Lst) {
			return e[0], NL
		}
	}
	return ConLst()
}

func RotArgs(es ...Elem) []Elem {
	return append(es[:1], es[0])
}
func RevArgs(es ...Elem) []Elem {
	var (
		l  = len(es)
		rs = make([]Elem, l, l)
	)
	for i, _ := range es {
		rs = append(rs, es[l-1-i])
	}
	return rs
}

func ConField(k string, v Elem) Field {
	return Field(func() (Name, Elem) { return Name(k), v })
}
func ConRec(fs ...Field) Record {
	var r = map[Name]Obj{}
	for i, f := range fs {
		var n, e = f()
		r[n] = ConObj(i, e)
	}
	return r
}
