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
	"unicode"
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
	// STRINGER ∷ printable
	Stringer interface {
		String() string
	}
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
		Cons(...Obj) Obj
	}
	// LABLE ∷ lable text
	Lab interface {
		Lable() string
	}
	// FLAG ∷ labled bit flag
	Flg interface {
		Kind() Kind
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
		Lookup(string) Sym
	}
	// CATEGORY ∷ set of objects of same kind of unit
	Cat interface {
		Tab        // member[0] = pi, cu ↑
		Unit() Obj // cu
		Zero() Cat
	}
	Type interface {
		Cat
		Type() (Flg, Cat)
	}
	// KIND
	// 'kind' is the category of types
	// Q: which kind of type are we talking about?
	// A: the kind of type we are currently dealing with.
	// Q: which unit does this type have?
	// A: the kind of unit, all types of this kind have.
	Kind interface {
		Type
		Kind() Kind
	}

	//Sum Obj // ∑ [Objₜ] = collection type of particular objects type
	//Pro Obj // ∏ Obj₁|Obj₂|…|Objₙ = enum type composed of n subtype flags

	Const func() Obj         // constant object
	Pair  func() (Obj, Obj)  // pair of objects
	Link  func() (Obj, Pair) // linked objects (list, tree, …)
	Vect  []Obj              // sum of objects

	UnaOp func(Obj) Obj      // unary operation
	BinOp func(a, b Obj) Obj // binary operation
	GenOp func(...Obj) Obj   // n-nary operation
)

func headArg(os []Obj) Obj {
	if len(os) > 0 {
		return os[0]
	}
	return None
}
func scndArg(os []Obj) Obj {
	if len(os) > 1 {
		return os[1]
	}
	return None
}
func tailArgs(os []Obj) Obj {
	if len(os) > 1 {
		return Vect(os[1:])
	}
	return None
}
func pairTail(os []Obj) Obj {
	if len(os) > 2 {
		return Vect(os[2:])
	}
	return None
}
func args(os []Obj) Obj {
	if len(os) > 0 {
		if len(os) > 1 {
			var v = make(Vect, 0, len(os))
			for _, o := range os {
				v = append(v, o)
			}
			return v
		}
	}
	return None
}

func (c Const) Ident() Obj         { return c }
func (c Const) Unit() Obj          { return c() } // flag, symbol, instance…
func (c Const) Id() int            { return c.Unit().Id() }
func (c Const) Cons(os ...Obj) Obj { return args(os) }
func consConst(o Obj) Const        { return func() Obj { return o } }

func (p Pair) Ident() Obj         { return p }
func (p Pair) Head() Obj          { l, _ := p(); return l }
func (p Pair) Tail() Obj          { _, r := p(); return r }
func (p Pair) Unit() Obj          { return p.Head() }
func (p Pair) Id() int            { return p.Head().Id() }
func (p Pair) Pid() int           { return p.Tail().Id() }
func (p Pair) Cons(os ...Obj) Obj { return consPair(headArg(os), scndArg(os)) }
func consPair(l, r Obj) Pair      { return func() (l, r Obj) { return l, r } }

func (l Link) Ident() Obj { return l }
func (l Link) Head() Obj  { h, _ := l(); return h }
func (l Link) Tail() Obj  { _, t := l(); return t }
func (l Link) Unit() Obj  { return l.Head() }
func (l Link) Id() int    { return l.Head().Id() }
func (l Link) Pid() int   { return l.Tail().Id() }
func (l Link) Cons(os ...Obj) Obj {
	return consLink(consPair(headArg(os), scndArg(os)), pairTail(os))
}
func consLink(p Pair, o Obj) Link { return func() (Obj, Pair) { return o, p } }

func (v Vect) Ident() Obj     { return v }
func (v Vect) Len() int       { return len(v) }
func (v Vect) Empty() bool    { return v.Len() == 0 }
func (v Vect) Single() bool   { return v.Len() == 1 }
func (v Vect) Double() bool   { return v.Len() == 2 }
func (v Vect) Multiple() bool { return v.Len() > 2 }
func (Vect) Finite() bool     { return true }
func (v Vect) Id() int {
	if v.Empty() {
		return 0
	}
	return v[0].Id()
}
func (v Vect) Head() Obj {
	if !v.Empty() {
		return consVect(v[0])
	}
	return None
}
func (v Vect) Tail() Obj {
	if !v.Empty() {
		return consVect(v[1:]...)
	}
	return None
}
func (v Vect) Unit() Obj { return v.Head() }
func (v Vect) Cons(os ...Obj) Obj {
	return v.Head()
}
func consVect(os ...Obj) Vect { return os }

/// SYMBOL & SYMBOL TABLE
// buildin and runtime defined operators, functions, named values, keywords…
// all have compile time symbols stored in a dynamic symbol table.  a symbols
// uid is its index position in its containing table.

// SORT LABLES & FLAGS
type lableSort []string

func (n lableSort) Sort() []string {
	var t = n
	sort.Strings(t)
	return t
}
func (n lableSort) Sorted() bool { return sort.StringsAreSorted(n) }

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

type (
	sym    func() (int, string)
	labTab []string
)

func (s sym) Id() int       { var uid, _ = s(); return uid }
func (s sym) Lable() string { var _, name = s(); return name }
func (s sym) Ident() Obj    { return s }
func (s sym) Sym() Sym      { return s }
func (s sym) Cons(os ...Obj) Obj {
	var (
		id  int
		lab = ""
	)
	if len(os) > 0 {
		id = os[0].Id()
		if len(os) > 1 {
			if sym, ok := os[1].(Sym); ok {
				lab = sym.Lable()
			}
			if str, ok := os[1].(Stringer); ok {
				lab = str.String()
			}
			if sym, ok := os[0].(Sym); ok {
				lab = sym.Lable()
			}
			if str, ok := os[0].(Stringer); ok {
				lab = str.String()
			}
		}
		return conSym(id, lab)
	}
	return None
}

func conSym(id int, lab string) sym { return func() (int, string) { return id, lab } }

func (t labTab) Tab() Tab      { return t }
func (t labTab) Lable() string { return t[0] }
func (labTab) Finite() bool    { return true }
func (t labTab) Len() int      { return len(t) }
func (t labTab) empty() sym    { return conSym(0, "") }
func (t labTab) Unit() Obj     { return t.empty() }

func (t labTab) Member() []Sym {
	var os = make([]Sym, 0, len(t))
	for i, n := range t {
		os = append(os, conSym(i, n))
	}
	return os
}
func (t labTab) Cons(uid Obj) Sym {
	if oi, ok := uid.(consString); ok {
		if i := oi.Id(); i < len(t) {
			return conSym(i, t[i])
		}
	}
	return t.empty()
}

func (t labTab) Lookup(name string) Sym {
	for uid, sym := range t {
		if sym == name {
			return conSym(uid, t[uid])
		}
	}
	return t.empty()
}

// SYMBOL TABLE CONSTRUCTOR
func consLableTab(names ...string) labTab {
	var t = make([]string, len(names))
	for _, n := range names {
		t = append(t, n)
	}
	return labTab(t)
}

/// FLAG SET
type (
	flg    func() (uint, Tab)
	flgTab map[string]uint
)

func (f flg) set() flgTab { _, s := f(); return s.(flgTab) }
func (f flg) uid() uint   { u, _ := f(); return u }

func (f flg) Ident() Obj { return f }
func (f flg) Zero() Cat  { return f }
func (f flg) Sym() Sym   { return f }
func (f flg) Unit() Obj  { return f.Tab().Lookup("") }
func (f flg) Id() int    { return rank(f.Flg()) }
func (f flg) Card() int  { return card(f.uid()) }

func (f flg) Flg() uint             { return f.uid() }
func (f flg) Tab() Tab              { return Tab(f.set()) }
func (f flg) Lable() string         { return f.Sym().Lable() }
func (f flg) Lookup(lab string) Sym { return f.Tab().Lookup(lab) }
func (f flg) Cons(os ...Obj) Obj {
	if len(os) > 0 {
		for id, o := range os {
			if id < f.Card() {
				return consFlagTab(consFlg(1<<uint(id), flgTab{}))
			}
		}
	}
	return consFlagTab(consFlg(0, flgTab{}))
}

// FLAG CONSTRUCTOR
func consFlg(u uint, s Tab) flg { return func() (uint, Tab) { return u, s } }

func (s flgTab) empty() flg {
	var m = make(map[string]uint)
	m[""] = 0
	return consFlg(0, flgTab(m))
}
func (s flgTab) Tab() Tab      { return s }
func (s flgTab) Ident() Obj    { return s }
func (s flgTab) Id() int       { return rank(s[""]) }
func (s flgTab) Unit() Obj     { return s.empty() }
func (s flgTab) Zero() Cat     { return s }
func (s flgTab) Lable() string { return s.empty().Lable() }

// dereference symbol by name
func (s flgTab) Lookup(name string) Sym {
	var (
		u  uint
		ok bool
	)
	// unknown name → ident as zero element
	if u, ok = s[name]; !ok {
		return s.empty()
	}
	// return uint mapped to name as flag
	return consFlg(u, s)
}

// dereference symbol by rank (id)
func (s flgTab) Cons(os ...Obj) Obj {
	var (
		ft = flgTab(make(map[string]uint))
		fs uint
	)
	for i, o := range os {
		if f, ok := o.(Flg); ok {
			fs = fs | f.Flg()
			ft[f.Lable()] = f.Flg()
		} else {
			f = tflg(1 << uint(i))
			fs = fs | f.Flg()
			ft[f.Lable()] = f.Flg()
		}
	}
	ft["T"] = fs
	return flgTab(ft)
}

/// FLAG SET CONSTRUCTOR
// constructs flag sets dynamicly at runtime
func consFlagTab(syms ...Sym) flgTab {

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

/// CATEGORY
type cat func() (Sym, Tab)

func (c cat) Kind() Kind { return None }
func (c cat) Ident() Obj { return c }
func (c cat) Sym() Sym   { s, _ := c(); return s }
func (c cat) Tab() Tab   { _, t := c(); return t }

// parental context
func (c cat) Id() int       { return c.Sym().Id() }
func (c cat) Lable() string { return c.Sym().Lable() }

// categoric context
func (c cat) ConsCat(t tflg) Sym      { return consClass(t) }
func (c cat) Lookup(lable string) Sym { return c.Tab().Lookup(lable) }

// unit & zero
func (c cat) Zero() Cat { return None }
func (c cat) Unit() Obj { return consClass(None) }
func (c cat) Cons(os ...Obj) Obj {
}

// unit of each kind is the empty set of types of that kind
func consClass(t Flg) Cat {
	switch t.Kind() {
	case Truth:
		return cat(func() (Sym, Tab) { return Truth, consTruth(false) })
	case Uint:
		return cat(func() (Sym, Tab) { return Uint, consUint(0) })
	case Int:
		return cat(func() (Sym, Tab) { return Int, consInt(0) })
	case Flt:
		return cat(func() (Sym, Tab) { return Flt, consFloat(0.0) })
	}
	// define category during runtme, by user defined flag
	var (
		// create empty flag table, assign argument lable &
		// ident to element zero, to set parent category id
		tab      = flgTab{t.Lable(): t.Flg()}
		flag flg = func() (uint, Tab) { return 0, tab }
	)
	// return category with argument symbol & empty member table
	return flag
}

// unique id is id trace to root recursively
func root(c Cat, ids ...int) []int {
	if id := c.Id(); id != 0 { // for categorys not world
		return root( // call fuid recursively on current categorys ident
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
	Neg  consInt = -1 // negative
	Zero consInt = 0  // ambivalent
	Unit consInt = 1  // positive
)

//// CONSTRUCTORS
// a constant based flags uid is its rank amongst constants of the same type
func (f tflg) Kind() Kind       { return T }
func (f tflg) Ident() Obj       { return f }
func (f tflg) Flg() uint        { return uint(f) }
func (f tflg) Id() int          { return rank(f.Flg()) }
func (f tflg) Lable() string    { return f.String() }
func (f tflg) Type() (Flg, Cat) { return f, Kind(T) }
func (f tflg) Unit() Obj        { return Unit }
func (f tflg) Zero() Cat        { return None }
func (f tflg) Cons(id ...Obj) Obj {
	return tflg(1 << uint(id.Id()))
}
func (f tflg) Lookup(lab string) Sym {

	var fs = splitKind(f)

	for i, f := range fs {
		if f.Lable() == lab {
			return tflg(1 << uint(i))
		}
	}
	return None
}

///
// TRUTH
type consTruth bool

func (consTruth) Kind() Kind   { return Truth }
func (t consTruth) Ident() Obj { return t }
func (t consTruth) Bool() bool { return bool(t) }
func (t consTruth) Unit() Obj  { return consTruth(False) }
func (t consTruth) Zero() Cat  { return None }
func (t consTruth) Id() int {
	if t.Bool() {
		return 1
	}
	return 0
}
func (t consTruth) Lable() string {
	if t {
		return "True"
	}
	return "False"
}
func (t consTruth) Cons(id int) Sym {
	if id > 0 {
		return consTruth(true)
	}
	return consTruth(false)
}
func (t consTruth) Lookup(lab string) Sym {
	var (
		trues  = []string{"true", "True", "has", "is"}
		falses = []string{"false", "False", "has not", "is not"}
	)
	for i := 0; i < len(trues); i++ {
		if strings.Contains(lab, trues[i]) {
			return consTruth(true)
		}
		if strings.Contains(lab, falses[i]) {
			return consTruth(false)
		}
	}
	return None
}
func (t consTruth) Int() int       { return t.Id() }
func (t consTruth) Uint() uint     { return uint(t.Int()) }
func (t consTruth) Float() float64 { return float64(t.Int()) }

// NATURAL
type consUint uint

func (consUint) Kind() Kind        { return Uint }
func (n consUint) Ident() Obj      { return n }
func (n consUint) Id() int         { return n.Int() }
func (n consUint) Unit() Obj       { return consUint(Unit) }
func (n consUint) Zero() Cat       { return None }
func (n consUint) Cons(id int) Sym { return consUint(id) }
func (n consUint) Lable() string   { return strconv.Itoa(n.Id()) }
func (n consUint) Uint() uint      { return uint(n) }
func (n consUint) Int() int        { return int(n) }
func (n consUint) Float() float64  { return float64(n) }
func (n consUint) Bool() bool {
	if n > 0 {
		return true
	}
	return false
}
func (n consUint) Lookup(lab string) Sym {
	nn, err := strconv.Atoi(lab)
	if err != nil {
		return None
	}
	return consUint(nn)
}

// INTEGER
type consInt int

func (consInt) Kind() Kind        { return Int }
func (i consInt) Ident() Obj      { return i }
func (i consInt) Id() int         { return i.Int() }
func (i consInt) Unit() Obj       { return consInt(Unit) }
func (i consInt) Zero() Cat       { return None }
func (i consInt) Cons(id int) Sym { return consInt(id) }
func (i consInt) Lable() string   { return strconv.Itoa(i.Int()) }
func (i consInt) Uint() uint      { return uint(i) }
func (i consInt) Int() int        { return int(i) }
func (i consInt) Float() float64  { return float64(i) }
func (i consInt) Bool() bool {
	if i > 0 {
		return true
	}
	return false
}
func (i consInt) Lookup(lab string) Sym {
	ii, err := strconv.Atoi(lab)
	if err != nil {
		return None
	}
	return consInt(ii)
}

type consIntPos int

func (consIntPos) Kind() Kind { return Int }
func (i consIntPos) Ident() Obj {
	if i >= 0 {
		return i
	}
	return None
}
func (i consIntPos) Id() int {
	if i > 0 {
		return i.Int()
	}
	return 0
}
func (i consIntPos) Unit() Obj { return consIntPos(Unit) }
func (i consIntPos) Zero() Cat { return None }
func (i consIntPos) Lable() string {
	return strconv.Itoa(i.Ident().Id())
}
func (i consIntPos) Bool() bool {
	if i >= 0 {
		return true
	}
	return false
}
func (i consIntPos) Uint() uint {
	if i >= 0 {
		return uint(i)
	}
	return 1
}
func (i consIntPos) Int() int {
	if i >= 0 {
		return int(i)
	}
	return 1
}
func (i consIntPos) Float() float64 {
	if i >= 0 {
		return float64(i)
	}
	return 1.0
}
func (i consIntPos) Get(id int) Sym {
	if i >= 0 {
		return consIntPos(id)
	}
	return None
}
func (i consIntPos) Lookup(lab string) Sym {
	ii, err := strconv.Atoi(lab)
	if err != nil {
		return None
	}
	if ii >= 0 {
		return consInt(ii)
	}
	return None
}

type consIntNeg int

func (consIntNeg) Kind() Kind { return Int }
func (i consIntNeg) Ident() Obj {
	if i <= 0 {
		return i
	}
	return None
}
func (i consIntNeg) Id() int {
	if i > 0 {
		return -i.Int()
	}
	return 0
}
func (i consIntNeg) Unit() Obj { return consIntNeg(Neg) }
func (i consIntNeg) Zero() Cat { return None }
func (i consIntNeg) Lable() string {
	return strconv.Itoa(i.Ident().Id())
}
func (i consIntNeg) Bool() bool {
	if i <= 0 {
		return true
	}
	return false
}
func (i consIntNeg) Int() int {
	if i <= 0 {
		return int(i)
	}
	return -1
}
func (i consIntNeg) Float() float64 {
	if i <= 0 {
		return float64(i)
	}
	return -1.0
}
func (i consIntNeg) Get(id int) Sym {
	if i < 0 {
		return consIntNeg(id)
	}
	return None
}
func (i consIntNeg) Lookup(lab string) Sym {
	ii, err := strconv.Atoi(lab)
	if err != nil {
		return None
	}
	if ii <= 0 {
		return consIntNeg(ii)
	}
	return None
}

type consFloat float64

func (f consFloat) Kind() Kind      { return Flt }
func (f consFloat) Ident() Obj      { return f }
func (f consFloat) Int() int        { return int(f) }
func (f consFloat) Id() int         { return f.Int() }
func (f consFloat) Unit() Obj       { return consFloat(Unit) }
func (f consFloat) Zero() Cat       { return None }
func (f consFloat) Cons(id int) Sym { return consFloat(id) }
func (f consFloat) Lable() string   { return strconv.FormatFloat(float64(f), 'E', -1, 64) }
func (f consFloat) Uint() uint      { return uint(f) }
func (f consFloat) Float() float64  { return float64(f) }
func (f consFloat) Bool() bool {
	if f > 0 {
		return true
	}
	return false
}
func (f consFloat) Lookup(lab string) Sym {
	ff, err := strconv.ParseFloat(lab, 64)
	if err != nil {
		return None
	}
	return consFloat(ff)
}

type consString string

func (s consString) Kind() Kind     { return String }
func (s consString) Ident() Obj     { return s }
func (s consString) String() string { return string(s) }
func (s consString) Id() int        { return rank(s.Kind()) }
func (s consString) Unit() Obj      { return consString("") }
func (s consString) Zero() Cat      { return None }
func (s consString) Cons(os ...Obj) Obj {
	var (
		ss  = args(os)
		sep = ""
	)
	if len(ss) > 2 {
		// first rune is regarded seperator, if it's neither digit,
		// nor letter
		if !(unicode.IsDigit(rune(ss[0][0])) ||
			unicode.IsLetter(rune(ss[0][0]))) {
			sep = string(ss[0][0])
			if len(ss) > 1 {
				ss = ss[1:]
			}
		}
	}
	return consString(strings.Join(ss, sep))
}
func (s consString) Lable() string { return s }

func main() {}
