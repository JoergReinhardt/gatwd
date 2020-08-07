package gatw

import (
	"math/bits"
	"strings"
	"unicode"
)

// CONSTANTS{{{

// CATEGORY CONSTANTS{{{
// CLASS CONSTANTS{{{

//go:generate stringer -type C
//go:generate stringer -type Rank
//go:generate stringer -type Option
const (
	// TYPE CLASSES
	None C = 1<<iota - 1
	// SUM TYPES
	Unit
	Pair
	Enum
	Sequence
	// PRODUCT TYPES
	Tuple
	Record
	// SHAPES OF TYPES
	Atom
	Function
	Composition
	// TYPE CLASSES
	Continuation // returns a current item and next continuation
	Definition   // argument pattern and return type of function, or expression
	Parametric   // yielded output depends on argument types
	Polymorph    // yields type from a set of return types
	Variant      // tuple|either…or|maybe…none
	Predicate    // yields trut value
	Optional     // yields a value, or not
	Equality     // lesser|greater|equal
	BitFlag      // labeled by flag
	Keyword      // record field, value, or function name
	Symbol       // labeled by keyword
	Number       // labeled by indexed
	Position     // position of element in an expression
	Ordered      // elements of this type can be ordered
	Element      // positioned (index) element of enum
	Member       // typed element of tuple (Type: Item)
	Field        // named field of record
	Partial      // partialy applied function
	Data         // constructors for constant types
	Kind         // a kind of type

	Maybe       = None | Unit // 'if': Unit OR None
	Shapes      = Atom | Function | Composition
	Sums        = Enum | Sequence       // sum types
	Products    = Pair | Tuple | Record // product types
	Collections = Sums | Products
	Classes     = Continuation | Definition | Parametric | Polymorph |
		Variant | Predicate | Optional | Equality | BitFlag |
		Keyword | Symbol | Number | Position | Ordered |
		Element | Member | Field | Partial | Data
	Types = Maybe | Collections | Classes | Shapes

	// OPTIONAL / ALTERNATIVE (IF‥ELSE|EITHER…OR)
	Either Option = 0
	Or     Option = 1

	// PRAEDICATE
	False Truth = false
	True  Truth = true

	// EQUALITY &| ORDER
	Lesser  Rank = -1
	Equal   Rank = 0
	Greater Rank = 1
) //}}}

//func (e C) String() string { return "" }
// {{{
type ( /// TYPE INTERFACE
	C      uint
	Rank   int8
	Option int8
	Truth  bool

	Id  int // rank of member in set of types
	Flg uint
	Key string

	Uid     []Id
	Pattern func() ([]Type, Comp)

	// ITEM INTERFACE
	Item interface {
		Type() Type
		Identity() Item
	}

	// FLAG INTERFACE
	Flag interface {
		Uint() uint
		String() string
	}
	// TYPE INTERFACE
	Type interface {
		Id() Id
		Name() Key
		Type() Type
		Identity() Item
	}
	// SIGNATURE (FUNCTION) TYPE
	Signature interface {
		Type
		Signature() Pattern
	}

	Cons func(...Item) (Item, Cons)

	// BOXED VALUE
	Value func() (Item, Type)

	// EXPRESSSION AND EVALUATION
	PairVal func(...Item) (l, r Item) // ← binary tree, when > 2 elements
	EnumVal []Item
	TplVal  PairVal // (Pattern, Enum)
	RecVal  map[Key]Item
	SeqVal  func(...Item) (Item, SeqVal) // ← backed by linked list

	Elem     PairVal
	TupElem  PairVal
	RecField PairVal

	Eval func(...Item) Item

	// FUNCTION INTERFACE
	Fnc interface {
		Item
		Call(...Item) Item
	}

	// FUNCTION DEFINITION
	Def func(...Item) (Item, Pattern)

	// FUNCTION PROTOTYPES
	NnaFnc func() Item
	UnaFnc func(Item) Item
	BinFnc func(x, y Item) Item
	NarFnc func(...Item) Item

	// PREDICATE PROTOTYPES
	PredFnc func(...Item) Truth
) // }}}

// ID LABLE
func (i Id) Identity() Item { return i }
func (i Id) Id() Id         { return i }
func (i Id) Int() int       { return int(i) }
func (i Id) Type() Type     { return Kind | Number }
func (i Id) Cons(args ...Item) (Item, Cons) {
	if len(args) > 0 {
		if len(args) > 1 {
			var ids = make([]Id, 0, len(args))
			for _, arg := range args {
				ids = append(ids, arg.Type().Id())
			}
			return Uid(ids), i.Cons
		}
		return Id(args[0].Type().Id()), i.Cons
	}
	return i, i.Cons
}
func (i Id) Compare(arg Id) Rank {
	if arg != i {
		if arg < i {
			return Lesser
		}
		if arg > i {
			return Greater
		}
	}
	return Equal
}

// NAME LABLE
func (n Key) Identity() Item { return n }
func (n Key) Type() Type     { return Kind | Symbol }
func (n Key) Name() Key      { return Key(string(n)) }
func (n Key) Runes() []rune  { return []rune(n) }
func (n Key) String() string { return string(n) }
func (n Key) Cons(args ...Item) (Item, Cons) {
	if len(args) > 0 {
		var names = make([]string, 0, len(args))
		for _, arg := range args {
			names = append(names,
				string(arg.Type().Name()))
		}
		return Key(strings.Join(names, ".")), n.Cons
	}
	return n, n.Cons
}
func (n Key) Capital() bool {
	if len(n.Runes()) > 0 {
		return unicode.IsUpper(n.Runes()[0])
	}
	return false
}
func (n Key) Compare(arg Key) Rank {
	return Rank(strings.Compare(n.String(), arg.String()))
}
func (n Key) Composed() bool {
	return strings.Contains(string(n.Name()), ".")
}
func (n Key) Polymorph() bool {
	return strings.Contains(string(n.Name()), "|")
}

// the Flag interface is supposed to be implemented by all user defined types
// marked by a uint bit-flag constants and therefore capable of fast set
// operations.  its easiely implemented by using 'go generate stringer' on your
// custom uint constant type and implementing an additional 'Uint() uint'
// method, in order to convert any instance of your flag type to uint for
// bitwise operations.  to implement a proper user defined type, all methods of
// the 'Type' & 'Item' interfaces have to be implemented as well.  there is a
// bunch of free functions, all with names chosen by convention to start with
// 'Flag', which can convieniently be wrapped as methods of any constant type
// that implements the flag interface.  see 'C' category flag type constant for
// an example.
func FlagUint(f uint) uint      { return f }
func FlagId(f uint) Id          { return Id(bits.Len(f)) }
func FlagMaxBitSet(f uint) int  { return bits.Len(f) }
func FlagNumBitsSet(f uint) int { return bits.OnesCount(f) }
func FlagIsBitSet(f uint) bool  { return FlagNumBitsSet(f) > 1 }
func FlagSplitSet(f uint) []uint {
	var flags = make([]uint, 0, bits.OnesCount(uint(f)))
	if FlagIsBitSet(f) {
		// generate flag per bit that is set
		for i := 0; i < bits.OnesCount(uint(f)); i++ {
			if bit := bits.RotateLeft(uint(f), i); bit == 1 {
				flags = append(flags, 1<<uint(i))
			}
		}
	}
	return flags
}

// the 'C' category constant is the root of the category tree.  it can be
// extended by the user at runtime.  all things definable must be assigned to a
// base category.  'C' implements the type interface.
//
// C.Cons(...Item) Item
//
// polymorph method that returns the tree of all defined categorys and the unit
// type of 'empty category', when called empty.  performs a lookup to return a
// type, or function definition/constructor, when passed a keyword, flag, or
// id.  creates a new category, member of an existing category, or sub-category
// there of, when passed a pair with a uid to address the category of which to
// create a member of and a type, or data definition/constructor.
//func (t C) String() string { return "" }
func (t C) Identity() Item { return t }
func (t C) Type() Type     { return Types }
func (t C) Uint() uint     { return uint(t) }
func (t C) Name() Key      { return Key(t.String()) }
func (t C) Id() Id         { return FlagId(t.Uint()) }
func (t C) Max() int       { return FlagMaxBitSet(t.Uint()) }
func (t C) Len() int       { return FlagNumBitsSet(t.Uint()) }
func (t C) IsBitSet() bool { return FlagIsBitSet(t.Uint()) }
func (t C) SplitSet() Pattern {
	var (
		us  = FlagSplitSet(t.Uint())
		pat = make([]Type, 0, len(us))
	)
	for u := range us {
		pat = append(pat, C(u))
	}
	return pat
}

func (t C) Cons(args ...Item) (Item, Cons) {
	if len(args) > 0 {
		if len(args) > 1 {
			var flag C
			for _, arg := range args {
				if c, ok := arg.(C); ok {
					flag = flag | c
				}
			}
			return flag, t.Cons
		}
		if c, ok := args[0].(C); ok {
			return c, t.Cons
		}
		return None, t.Cons
	}
	return t, t.Cons
}

//}}}

// PREDICATE FUNCTION
func ConTruth(b bool) Truth { return Truth(b) }

func (t Truth) Identity() Item { return t }
func (t Truth) Type() Type     { return t }
func (t Truth) Name() Key      { return Key("Truth." + t.String()) }
func (t Truth) Compare(arg Truth) Rank {
	if t != arg {
		if arg == true {
			return Lesser
		}
		if arg == false {
			return Greater
		}
	}
	return Equal
}
func (t Truth) Id() Id {
	if bool(t) {
		return Id(1)
	}
	return Id(0)
}
func (t Truth) String() string {
	if bool(t) {
		return "True"
	}
	return "False"
}
func (t Truth) Cons(args ...Item) (Item, Cons) {
	if len(args) > 0 {
		if len(args) > 1 {
			var items = make([]Item, 0, len(args))
			for _, arg := range args {
				if truth, ok := arg.(Truth); ok {
					items = append(items, truth)
				}
			}
			return EnumVal(items), t.Cons
		}
		if truth, ok := args[0].(Truth); ok {
			return truth, t.Cons
		}
		return None, t.Cons
	}
	return t, t.Cons
}

func ConPredicate(pred func(...Item) Truth) PredFnc {
	return PredFnc(pred)
}

func (p PredFnc) Identity() Item { return p }
func (p PredFnc) Type() Type     { return Predicate }
func (p PredFnc) String() string { return string(p.Type().Name()) }

// TYPED, BOXED ITEM TYPE AS BASE FOR JUST, EITER, OR…
func Box(i Item, t Type) Value {
	return Value(func() (Item, Type) { return i, t })
}

func (b Value) Type() Type     { _, t := b(); return t }
func (b Value) Identity() Item { i, _ := b(); return i }
