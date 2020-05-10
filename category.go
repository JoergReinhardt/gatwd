/*
 CATEGORY OBJECT INTERFACE AND UNIT TYPES

 every thing is an object and needs to implement the object interface 'Obj'.
 that includes internal parts of the type system, like type markers and
 names.  the interface demands a 'Type() int' method to return a unique
 numeric identification.

 the 'Ident() Obj' method needs to be implemented to return the native
 instance of whatever type implements the interface, aka it-'self'.

   - runtime defined types are accessed by slice index.

   - category types need quick set membership identification

   - some kinds are named, others need anonymity.

 hence three sorts of identity markers exist:

   - numeric unique id shared by every kind of type.

   - binary bit flag for sets of categorys, with quick membership operation

   - string representation of instance data, or name of its type


 TYPE FLAG IMPLEMENTATION

 flags mark all kinds of category and provide the zero value for, most other
 product types (tuple/struct) and the type of categorys as such.

 the parent objects id is expressed in in the value rank of the flag within
 its set of constants of the same type.

 the flags string representation may be identical with the type lable, or
 category name, but not mandatory.  the private 'sym() Sym' method returns an
 intermediate symbol instance for convienience, composing flag rank and
 string representation to conform to the symbol interface, to be used by type
 constructors it gets passed to.  all categorys, one sub type of which is
 'types', are instances of flg.  all category and parametric types have
 'None' as their zero value within their own category and whatever the flag
 value is, in the parent category. sub types of the type category are most
 tuple/record/struct fields (may be implemented by hash map instead) and all
 parametric types including all base cases of sub-types of recursive
 parametric types.

 some native type like bool, or int is to be expected at the
 definition swamp floor… (might be a tur<d|tel>).

 TYPE MARKER IMPLEMENTATIONS
 ident methods return the actual instances uid and the instance itself.

 in case of a type marker, its the marker itself and it needs to implement
 ident in order to be viable argument for operations of the type system.
 when enclosed with an actual instance…

 !!!TYPE MARKER UID & IDENT METHODS NEED TO BE SHADOWED BY ENCLOSING OBJECT!!!
*/
package main

import (
	"math/bits"
	"sort"
	"strconv"
	"strings"
)

/// FLAG INTERFACE
// interface methods
func isSet(f uint) bool           { return card(f) > 1 }
func hasFlag(set, flag uint) bool { return set&flag != 0 }
func card(f uint) int             { return bits.OnesCount(uint(f)) }
func rank(f uint) int             { return bits.Len(uint(f)) }
func splitKind(f tflg) []Flg {
	var (
		u   = uint(f)
		l   = rank(u)
		set = make([]Flg, 0, l)
	)
	if l > 1 {
		for f != 0 {
			var u = f
			f = f & u
			set = append(set, tflg(u))
		}
		set = append(set, tflg(u))
	}
	return set
}
func splitSet(f uint) []uint {
	var set = make([]uint, 0, rank(f))
	if rank(f) > 1 {
		var flag = f
		for f != 0 {
			var flag = f
			f = f & flag
			set = append(set, flag)
		}
		set = append(set, flag)
	}
	return set
}

type (
	// UID ∷ numeric id
	Id interface {
		Id() int
	}
	// IDENTITY ∷ unique instance of object
	Ident interface {
		Ident() Obj
	}
	// OBJECT ∷  unique identifyable instance
	Obj interface {
		Id
		Ident
	}
	// LABLE ∷ lable text
	Lab interface {
		Lable() string
	}
	// FLAG ∷ labled bit flag
	Flg interface {
		Flg() uint
		Lab
	}
	// SYMBOL ∷ object labled with name, bit flag, or other lable
	Sym interface {
		Lab // Flg | Lab
		Obj
	}
	// LINKED ∷ objects linked by name
	Lnk interface {
		Head() Obj
		Tail() Obj
	}
	// TABLE ∷ object uid to name map
	Tab interface {
		Get(int) Sym
		Lookup(string) Sym
	}
	// CATEGORY ∷ set of objects of same kind of unit
	Cat interface {
		Sym
		Tab        // member[0] = pi, cu ↑
		Unit() Obj // cu
		Zero() Cat
	}
	Type interface {
		Cat
		Type() (Flg, Cat)
	}

	//Sum Obj // ∑ [Objₜ] = collection type of particular objects type
	//Pro Obj // ∏ Obj₁|Obj₂|…|Objₙ = enum type composed of n subtype flags

	Const func() Obj         // constant object
	Pair  func() (Obj, Obj)  // pair of objects
	Link  func() (Obj, Pair) // linked objects (list, tree, …)
	Vect  func() []Obj       // sum of objects

	UnaOp func(Obj) Obj      // unary operation
	BinOp func(a, b Obj) Obj // binary operation
	GenOp func(...Obj) Obj   // n-nary operation
)

func (c Const) Ident() Obj  { return c }
func (c Const) Unit() Obj   { return c() } // flag, symbol, instance…
func (c Const) Id() int     { return c.Unit().Id() }
func consConst(o Obj) Const { return func() Obj { return o } }

func (p Pair) Ident() Obj    { return p }
func (p Pair) Head() Obj     { l, _ := p(); return l }
func (p Pair) Tail() Obj     { _, r := p(); return r }
func (p Pair) Unit() Obj     { return p.Head() }
func (p Pair) Id() int       { return p.Head().Id() }
func (p Pair) Pid() int      { return p.Tail().Id() }
func (p Pair) Finite() bool  { return true }
func consPair(l, r Obj) Pair { return func() (l, r Obj) { return l, r } }

func (l Link) Ident() Obj         { return l }
func (l Link) Head() Obj          { h, _ := l(); return h }
func (l Link) Tail() Obj          { _, t := l(); return t }
func (l Link) Unit() Obj          { return l.Head() }
func (l Link) Id() int            { return l.Head().Id() }
func (l Link) Pid() int           { return l.Tail().Id() }
func consLink(p Pair, o Obj) Link { return func() (Obj, Pair) { return o, p } }

func (v Vect) Ident() Obj     { return v }
func (v Vect) Len() int       { return len(v()) }
func (v Vect) Empty() bool    { return v.Len() == 0 }
func (v Vect) Single() bool   { return v.Len() == 1 }
func (v Vect) Double() bool   { return v.Len() == 2 }
func (v Vect) Multiple() bool { return v.Len() > 2 }
func (Vect) Finite() bool     { return true }
func (v Vect) Id() int {
	if v.Empty() {
		return 0
	}
	return v()[0].Id()
}
func (v Vect) Head() Obj {
	if !v.Empty() {
		return consVect(v()[0])
	}
	return None
}
func (v Vect) Tail() Obj {
	if !v.Empty() {
		return consVect(v()[1:]...)
	}
	return None
}
func (v Vect) Unit() Obj      { return v.Head() }
func consVect(os ...Obj) Vect { return func() []Obj { return os } }

/// SYMBOL & SYMBOL TABLE
// buildin and runtime defined operators, functions, named values, keywords…
// all have compile time symbols stored in a dynamic symbol table.  a symbols
// uid is its index position in its containing table.
type (
	sym  func() (int, string)
	stab []string
)

func (s sym) Id() int       { var uid, _ = s(); return uid }
func (s sym) Lable() string { var _, name = s(); return name }
func (s sym) Ident() Obj    { return s }
func (s sym) Sym() Sym      { return s }

func conSym(idx int, sym string) sym { return func() (int, string) { return idx, sym } }

func (t stab) Tab() Tab      { return t }
func (t stab) Lable() string { return t[0] }
func (stab) Finite() bool    { return true }
func (t stab) Len() int      { return len(t) }
func (t stab) zero() sym     { return conSym(0, "") }
func (t stab) Unit() Obj     { return t.zero() }

func (t stab) Member() []Sym {
	var os = make([]Sym, 0, len(t))
	for i, n := range t {
		os = append(os, conSym(i, n))
	}
	return os
}
func (t stab) Get(uid int) Sym {
	if uid < len(t) {
		return conSym(uid, t[uid])
	}
	return t.zero()
}

func (t stab) Lookup(name string) Sym {
	for uid, sym := range t {
		if sym == name {
			return conSym(uid, t[uid])
		}
	}
	return t.zero()
}

// SYMBOL TABLE CONSTRUCTOR
func consTab(names ...string) stab {
	var t = make([]string, len(names))
	for _, n := range names {
		t = append(t, n)
	}
	return stab(t)
}

/// FLAG SET
type (
	flg  func() (uint, Tab)
	fset map[string]uint
)

func (f flg) Ident() Obj { return f }

func (f flg) set() fset { _, s := f(); return s.(fset) }
func (f flg) uid() uint { u, _ := f(); return u }

func (f flg) Tab() Tab  { return Tab(f.set()) }
func (f flg) Flg() uint { return f.uid() }

func (f flg) Id() int       { return rank(f.Flg()) }
func (f flg) Sym() Sym      { return f.set().Get(rank(f.uid())) }
func (f flg) Lable() string { return f.Sym().Lable() }

// FLAG CONSTRUCTOR
func consFlg(u uint, s Tab) flg { return func() (uint, Tab) { return u, s } }

func (fset) Finite() bool { return true }
func (s fset) Len() int   { return len(s) }
func (s fset) Tab() Tab   { return s }
func (s fset) zero() flg {
	var m = make(map[string]uint)
	return consFlg(0, fset(m))
}
func (s fset) Unit() Obj     { return s.zero() }
func (s fset) Lable() string { return s.zero().Lable() }

// dereference symbol by name
func (s fset) Lookup(name string) Sym {
	var (
		u  uint
		ok bool
	)
	// unknown name → ident as zero element
	if u, ok = s[name]; !ok {
		return s.zero()
	}
	// return uint mapped to name as flag
	return consFlg(u, s)
}

// dereference symbol by rank (id)
func (s fset) Get(i int) Sym {
	var f = uint(1) << uint(i)
	for _, u := range s {
		if u == f {
			return consFlg(u, s)
		}
	}
	return s.zero()
}

/// FLAG SET CONSTRUCTOR
// constructs flag sets dynamicly at runtime
func consFlagSet(syms ...Sym) fset {

	var (
		l = len(syms)
		m = make(map[string]uint, l)
	)

	// init empty flag set
	if l == 0 {
		m[""] = 0
	}

	// map lable to flag
	for i := 0; i < l; i++ {
		// map names to flags derived by rank
		m[syms[i].Lable()] = 1 << uint(i)
	}

	// return map
	return m
}

// sort slice of flags
type nameSort []string

func (n nameSort) Sort() []string {
	var t = n
	sort.Strings(t)
	return t
}
func (n nameSort) Sorted() bool { return sort.StringsAreSorted(n) }

type flagSort []Flg

func (f flagSort) Len() int           { return len(f) }
func (f flagSort) Less(i, j int) bool { return f[i].Flg() < f[j].Flg() }
func (f flagSort) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f flagSort) Sorted() bool       { return sort.IsSorted(f) }
func (f flagSort) Sort() []Flg {
	if !f.Sorted() {
		var s = f
		sort.Sort(s)
		return s
	}
	return f
}

/// CATEGORY
type cat func() (Sym, Tab)

func (c cat) Sym() Sym { s, _ := c(); return s }
func (c cat) Tab() Tab { _, t := c(); return t }

// parental context
func (c cat) Ident() Obj    { return c }
func (c cat) Unit() Obj     { return c.Sym().Ident() }
func (c cat) Lable() string { return c.Sym().Lable() }
func (c cat) Zero() Cat     { return c.Get(0).(Cat) }
func (c cat) Id() int       { return c.Sym().Id() }

// categoric context
func (c cat) Get(id int) Sym          { return c.Tab().Get(id) }
func (c cat) GetObj(o Obj) Sym        { return c.Tab().Get(o.Id()) }
func (c cat) Lookup(lable string) Sym { return c.Tab().Lookup(lable) }
func (c cat) Derives() []int          { return fuid(c) }

// unique id is id trace to root recursively
func fuid(c Cat, ids ...int) []int {
	if id := c.Id(); id != 0 { // for categorys not world
		return fuid( // call fuid recursively on current categorys ident
			c.Zero(),           // c[0] = ident of cat c
			append(ids, id)..., // pass on all ids
		)
	}
	// append world and return ancestry vector
	return append(ids, 0)
}

// tflg is the base set of flags needed to express categorys and types there
// of.
type tflg uint32

// a constant based flags uid is its rank amongst constants of the same type
func (f tflg) Ident() Obj       { return f }
func (f tflg) Flg() uint        { return uint(f) }
func (f tflg) Id() int          { return rank(f.Flg()) }
func (f tflg) Lable() string    { return f.String() }
func (f tflg) Type() (Flg, Cat) { return f, Kind(T) }
func (f tflg) Unit() Obj        { return Unit }
func (f tflg) Zero() Cat        { return Zero }
func (f tflg) Get(id int) Sym   { return tflg(1 << uint(id)) }
func (f tflg) Lookup(lab string) Sym {

	var fs = splitKind(f)

	for i, f := range fs {
		if f.Lable() == lab {
			return tflg(1 << uint(i))
		}
	}
	return None
}

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
	Fnc // map one classes objects onto another class
	Op  // operate on (take|return) objects of one class
	Just
	Either
	Or

	// TRUTH
	False bool = false
	True  bool = true

	Num = Uint | Int | Flt | Img // numeric
	Tmp = Time | Span            // temporal
	Bin = Byte | Bytes           // binary
	Txt = Rune | String          // textual
	Col = Vector | Tuple | List  // collection
	Par = Fnc | Op               // parameterized

	Maybe = Just | None
	Alter = Either | Or

	T = Truth | Num | Tmp | Bin | Txt | Col |
		Par | Maybe | Alter

	//  unit, neutral & sign
	Neg  tInt = -1 // negative
	Zero tInt = 0  // ambivalent
	Unit tInt = 1  // positive
)

//// CONSTRUCTORS
///
// TRUTH
type tTruth bool

func (t tTruth) Ident() Obj { return t }
func (t tTruth) Bool() bool { return bool(t) }
func (t tTruth) Unit() Obj  { return tTruth(false) }
func (t tTruth) Zero() Cat  { return tTruth(true) }
func (t tTruth) Id() int {
	if t.Bool() {
		return 1
	}
	return 0
}
func (t tTruth) Lable() string {
	if t {
		return "True"
	}
	return "False"
}
func (t tTruth) Get(id int) Sym {
	if id > 0 {
		return tTruth(true)
	}
	return tTruth(false)
}
func (t tTruth) Lookup(lab string) Sym {
	var (
		trues  = []string{"true", "True", "has", "is"}
		falses = []string{"false", "False", "has not", "is not"}
	)
	for i := 0; i < len(trues); i++ {
		if strings.Contains(lab, trues[i]) {
			return tTruth(true)
		}
		if strings.Contains(lab, falses[i]) {
			return tTruth(false)
		}
	}
	return tTruth(false)
}

// NATURAL
type tNat int

func (n tNat) Ident() Obj     { return n }
func (n tNat) Nat() uint      { return uint(n) }
func (n tNat) Int() int       { return int(n) }
func (n tNat) Id() int        { return int(n.Nat()) }
func (n tNat) Unit() Obj      { return tNat(1) }
func (n tNat) Zero() Cat      { return tNat(0) }
func (n tNat) Get(id int) Sym { return tNat(id) }
func (n tNat) Lable() string  { return strconv.Itoa(n.Id()) }
func (n tNat) Lookup(lab string) Sym {
	nn, err := strconv.Atoi(lab)
	if err != nil {
		return None
	}
	return tNat(nn)
}

// INTEGER
type tInt int

func (i tInt) Ident() Obj     { return i }
func (i tInt) Int() int       { return int(i) }
func (i tInt) Id() int        { return i.Int() }
func (i tInt) Unit() Obj      { return tInt(1) }
func (i tInt) Zero() Cat      { return tInt(0) }
func (i tInt) Get(id int) Sym { return tInt(id) }
func (i tInt) Lable() string  { return strconv.Itoa(i.Int()) }
func (i tInt) Lookup(lab string) Sym {
	ii, err := strconv.Atoi(lab)
	if err != nil {
		return None
	}
	return tInt(ii)
}

// KIND
// 'kind' is the category of types
// Q: which kind of type are we talking about?
// A: the kind of type we are currently dealing with.
// Q: which unit does this type have?
// A: the kind of unit, all types of this kind have.
type Kind tflg

func (Kind) Id() int    { return 0 }
func (Kind) Ident() Obj { return T }
func (Kind) Zero() Cat  { return Cat(None) }
func (Kind) Unit() Obj  { return Unit }

func (Kind) split() []Flg   { return splitKind(T) }
func (Kind) Get(id int) Sym { return tflg(1 << uint(id)) }
func (t Kind) Lookup(lab string) Sym {

	var fs = flagSort(t.split()).Sort()

	for i, f := range fs {
		if f.Lable() == lab {
			return t.Get(i)
		}
	}
	return None
}

func (t Kind) Lable() string {
	var (
		fs = flagSort(t.split()).Sort()
		l  = len(fs)
		ls = make([]string, 0, l)
	)
	for _, f := range fs {
		ls = append(ls, f.Lable())
	}
	return strings.Join(ls, " | ")
}

func main() {}
