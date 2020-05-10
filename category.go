/*
 CATEGORY OBJECT INTERFACE AND UNIT TYPES

Clearup Category

Category is cluttered by dispatch of lable/id/flag accessors.
Category will only be accessed by index via 'Obj.Id() -> int'.
  ⇒ category constructor needs to derive Kind from int Id:
    ⇒ kind needs to be kind[0]
      ⇒ kind needs to return all []kind
        ← in order to access sibling kinds
      ⇒ kind needs to return its kind: kind[0].Kind()
        ← in order to walk tree for generation of uuid

  '''
  Category  ∷ C  →  Obj…

    id   C  = id    C →  int		   |  pos in parent category
    card C  = card  C →  int		   |  length
    name C  = name  C →  string		   |  name of category (flag lable, or symbol)
					   |
    cons    = C  →  C₀			   |  cons empty root category
    cons O… = O… →  (O.Id()… → Cₙ) → Cₙ₊ₘ  |  cons 'cat from cat & objects
					   |
    kind    = C  →  Cₙ…			   |  kind ∅ returns all kinds	    → C…
    kind Cₙ = Cₙ →  (Cₙ → C₀) → Cₚₐᵣₑₙₜ	   |  kind C returns kind <of> O(C) → Cₚ
    kind Oₙ = Oₙ →  (id Oₙ → Cₙ) → C₍ₕᵢₗ₎  |  kind O returns kind <of> O₍   → Cₒ
  '''

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
	// ID ∷ numeric id
	// LABLE ∷ lable text
	Lab interface {
		Id() int
		Lable() string
	}
	// FLAG ∷ labled bit flag
	Flg interface {
		Flg() uint
		Lab
	}
	// IDENTITY ∷ unique instance of object
	Ident interface {
		Ident() Obj
	}
	Relat interface {
		Kind(...Obj) Cat
		Cons(...Obj) Obj
	}
	// OBJECT ∷
	// ⇒ kind _  →  Cₚₐᵣₑₙₜ  <|> !_ARGS_IGNORED_!_CONSTANT_! (PARENT)
	// ⇒ O… <cons> O  →  O
	Obj interface { // ← bits = 1 !
		Ident // Ident() Obj
		Relat // K(…O) C, C(…O) O
		Lab   // Id int, L string
	}
	// CATEGORY ∷ C  →  [O]			   <|>  set of objects
	// cons	    = ∅  →  C₀			   <|>  cons all sub categorys
	// cons	C   = C  →  [C]			   <|>  cons parent category
	// cons O…  = O… →  (O.Id()… → Cₙ) → Cₙ₊ₘ  <|>  cons new cat' from cat & objects
	//					   <|>
	// kind    = C  →  Cₙ…			   <|>  kind ∅ returns all kinds      → C…
	// kind Cₙ = Cₙ →  (Cₙ → C₀) → Cₚₐᵣₑₙₜ	   <|>  kind C returns kind <of> O(C) → Cₚ
	// kind Oₙ = Oₙ →  (id Oₙ → Cₙ) → C₍ₕᵢₗ₎   <|>  kind O returns kind <of> O₍   → Cₒ
	// ⇒ (kind C →  Cₚₐᵣₑₙₜ| kind O →  C₍ₕᵢₗ₎)
	// ⇒ O… <cons> C  →  C'
	Cat interface { // ← bits > 1 !
		Ident // Ident() Obj
		Relat // K(…O) C, C(…O) O
		Flg   // Id int, L string, F uint
	}
	// LINKED ∷ objects linked by name
	Lnk interface {
		Head() Obj
		Tail() Obj
		Ident
	}
	//Sum Obj // ∑ [Objₜ] = collection type of particular objects type
	//Pro Obj // ∏ Obj₁|Obj₂|…|Objₙ = enum type composed of n subtype flags

	Const func() Ident          // constant object
	Pair  func() (Ident, Ident) // pair of objects
	Link  func() (Ident, Pair)  // linked objects (list, tree, …)
	Vect  []Ident               // sum of objects

	UnaOp func(Ident) Ident      // unary operation
	BinOp func(a, b Ident) Ident // binary operation
	GenOp func(...Ident) Ident   // n-nary operation
)
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
func (f tflg) Ident() Obj    { return f }
func (f tflg) Flg() uint     { return uint(f) }
func (f tflg) Id() int       { return rank(f.Flg()) }
func (f tflg) Lable() string { return f.String() }
func (f tflg) Unit() Obj     { return Unit }
func (f tflg) Cons(id ...Obj) Obj {
	return tflg(1 << uint(id.Id()))
}

///
// TRUTH
type consTruth bool

func (t consTruth) Ident() Obj { return t }
func (t consTruth) Bool() bool { return bool(t) }
func (t consTruth) Unit() Obj  { return consTruth(False) }
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
func (t consTruth) Cons(os ...Obj) Obj {
	if id > 0 {
		return consTruth(true)
	}
	return consTruth(false)
}
func (t consTruth) Lookup(lab string) Obj {
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

func (n consUint) Ident() Obj         { return n }
func (n consUint) Id() int            { return n.Int() }
func (n consUint) Unit() Obj          { return consUint(Unit) }
func (n consUint) Zero() Cat          { return None }
func (n consUint) Cons(os ...Obj) Obj { return consUint(id) }
func (n consUint) Lable() string      { return strconv.Itoa(n.Id()) }
func (n consUint) Uint() uint         { return uint(n) }
func (n consUint) Int() int           { return int(n) }
func (n consUint) Float() float64     { return float64(n) }
func (n consUint) Bool() bool {
	if n > 0 {
		return true
	}
	return false
}
func (n consUint) Lookup(lab string) Obj {
	nn, err := strconv.Atoi(lab)
	if err != nil {
		return None
	}
	return consUint(nn)
}

// INTEGER
type consInt int

func (i consInt) Ident() Obj      { return i }
func (i consInt) Id() int         { return i.Int() }
func (i consInt) Unit() Obj       { return consInt(Unit) }
func (i consInt) Zero() Cat       { return None }
func (i consInt) Cons(id int) Obj { return consInt(id) }
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
func (i consInt) Lookup(lab string) Obj {
	ii, err := strconv.Atoi(lab)
	if err != nil {
		return None
	}
	return consInt(ii)
}

type consIntPos int

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
func (i consIntPos) Get(id int) Obj {
	if i >= 0 {
		return consIntPos(id)
	}
	return None
}
func (i consIntPos) Lookup(lab string) Obj {
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
func (i consIntNeg) Get(id int) Obj {
	if i < 0 {
		return consIntNeg(id)
	}
	return None
}
func (i consIntNeg) Lookup(lab string) Obj {
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

func (f consFloat) Ident() Obj      { return f }
func (f consFloat) Int() int        { return int(f) }
func (f consFloat) Id() int         { return f.Int() }
func (f consFloat) Unit() Obj       { return consFloat(Unit) }
func (f consFloat) Zero() Cat       { return None }
func (f consFloat) Cons(id int) Obj { return consFloat(id) }
func (f consFloat) Lable() string   { return strconv.FormatFloat(float64(f), 'E', -1, 64) }
func (f consFloat) Uint() uint      { return uint(f) }
func (f consFloat) Float() float64  { return float64(f) }
func (f consFloat) Bool() bool {
	if f > 0 {
		return true
	}
	return false
}
func (f consFloat) Lookup(lab string) Obj {
	ff, err := strconv.ParseFloat(lab, 64)
	if err != nil {
		return None
	}
	return consFloat(ff)
}

type consString string

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
