/*
DATA CONSTRUCTORS

  implementations of 'precedence types', ake functional base-/ and collection types
*/
package functions

import (
	"strconv"

	d "github.com/JoergReinhardt/gatwd/data"
)

type (
	// FUNCTION CLOSURES
	ConstFnc  func() Value
	UnaryFnc  func(Value) Value
	BinaryFnc func(a, b Value) Value
	NaryFnc   func(...Value) Value
	// DATA COLLENCTION CLOSURES
	AssocVecFnc func() []Paired
	VecFnc      func() []Value
	ListFnc     func() (Value, Recursive)
	RecordFnc   func() (Tupled, []Paired)
	TupleFnc    func() (Vectorized, []d.BitFlag)
	// HIGHER ORDER TYPE DEFINITION
	TypeFnc func() (tid d.IntVal, name d.StrVal)
	EnumFnc func() (TypeFnc, []TypeFnc) // type of instance, set of alternatives
	// HIGHER ORDER CLOSURES (STATICLY DEFINED)
	PraedFnc  func(Value) Boolean  // result impl. Bool() bool
	OptionFnc func(Value) Optional // result impl.Maybe() bool, e.g. 'Just|None'
)

func NewType(tid int, name string) HigherOrderType {
	return TypeFnc(func() (d.IntVal, d.StrVal) { return d.IntVal(tid), d.StrVal(name) })
}
func (t TypeFnc) Name() d.StrVal          { _, name := t(); return name }
func (t TypeFnc) Id() d.IntVal            { id, _ := t(); return id }
func (t TypeFnc) TypePrim() d.TyPrimitive { return d.Type }
func (t TypeFnc) TypeHO() TyHigherOrder   { return Type }
func (t TypeFnc) Ident() Value            { return t }
func (t TypeFnc) String() string {
	return strconv.Itoa(t.Id().Int()) + " " + t.Name().String()
}
func (t TypeFnc) Eval(p ...d.Primary) d.Primary {
	return d.NewPair(t.Id(), t.Name())
}
func (t TypeFnc) Call(d ...Value) Value { return t }

// CONSTANT
//
// constant also conains immutable data that may be an instance of a type of
// the data package, or result of a function call guarantueed to allways return
// the same value.
func NewConstant(fnc func(...Value) Value) ConstFnc {
	return ConstFnc(func() Value { return fnc() })
}
func NewPrimaryConstatnt(prime d.Primary) ConstFnc {
	return func() Value { return NewFromData(prime) }
}
func (c ConstFnc) TypePrim() d.TyPrimitive       { return c().TypePrim() }
func (c ConstFnc) TypeHO() TyHigherOrder         { return Function }
func (c ConstFnc) Ident() Value                  { return c }
func (c ConstFnc) Eval(p ...d.Primary) d.Primary { return c() }
func (c ConstFnc) Call(d ...Value) Value {
	return c().(ConstFnc)()
}

///// UNARY FUNCTION
func NewUnaryFnc(fnc func(f Value) Value) UnaryFnc {
	return UnaryFnc(func(f Value) Value { return fnc(f) })
}
func (u UnaryFnc) TypePrim() d.TyPrimitive       { return d.Function.TypePrim() }
func (u UnaryFnc) TypeHO() TyHigherOrder         { return Function }
func (u UnaryFnc) Ident() Value                  { return u }
func (u UnaryFnc) Eval(p ...d.Primary) d.Primary { return u }
func (u UnaryFnc) Call(d ...Value) Value {
	return u(d[0])
}

///// BINARY FUNCTION
func NewBinaryFnc(fnc func(a, b Value) Value) BinaryFnc {
	return BinaryFnc(func(a, b Value) Value { return fnc(a, b) })
}
func (b BinaryFnc) TypePrim() d.TyPrimitive       { return d.Function.TypePrim() }
func (b BinaryFnc) TypeHO() TyHigherOrder         { return Function }
func (b BinaryFnc) Ident() Value                  { return b }
func (b BinaryFnc) Eval(p ...d.Primary) d.Primary { return b }
func (b BinaryFnc) Call(d ...Value) Value         { return b(d[0], d[1]) }

///// NARY FUNCTION
func NewNaryFnc(fnc func(f ...Value) Value) NaryFnc {
	return NaryFnc(func(f ...Value) Value { return fnc(f...) })
}
func (n NaryFnc) TypePrim() d.TyPrimitive       { return d.Function.TypePrim() }
func (n NaryFnc) TypeHO() TyHigherOrder         { return Function }
func (n NaryFnc) Ident() Value                  { return n }
func (n NaryFnc) Eval(p ...d.Primary) d.Primary { return n }
func (n NaryFnc) Call(d ...Value) Value         { return n(d...) }

/////////////////////////////////////////////////////////
//// ASSOCIATIVE VECTOR (VECTOR OF PAIRS)
func conAccVec(vec Accessable, pp ...Paired) Accessable {
	return accVecConstructor(append(vec.Pairs(), pp...)...)
}
func accVecConstructor(pp ...Paired) Accessable {
	return AssocVecFnc(func() []Paired { return pp })
}
func NewAccVector(pp ...Paired) Accessable {
	return AssocVecFnc(func() (pairs []Paired) {
		for _, pair := range pp {
			pairs = append(pairs, pair)
		}
		return pairs
	})
}
func (v AssocVecFnc) Slice() []Value {
	var fncs = []Value{}
	for _, pair := range v() {
		fncs = append(fncs, NewPair(pair.Left(), pair.Right()))
	}
	return fncs
}
func (v AssocVecFnc) Pairs() []Paired { return v() }
func (v AssocVecFnc) Head() Paired {
	if v.Len() > 0 {
		return v.Pairs()[0]
	}
	return nil
}
func (v AssocVecFnc) Tail() []Paired {
	if v.Len() > 1 {
		return v.Pairs()[1:]
	}
	return nil
}
func (v AssocVecFnc) Eval(p ...d.Primary) d.Primary {
	var slice = d.DataSlice{}
	for _, pair := range v() {
		d.SliceAppend(slice, d.NewPair(pair.Left(), pair.Right()))
	}
	return slice
}
func (v AssocVecFnc) Len() int                   { return len(v()) }
func (v AssocVecFnc) Empty() bool                { return ElemEmpty(v.Head()) && (len(v.Tail()) == 0) }
func (v AssocVecFnc) GetVal(praed Value) Paired  { return newPairSorter(v()...).Get(praed) }
func (v AssocVecFnc) Range(praed Value) []Paired { return newPairSorter(v()...).Range(praed) }
func (v AssocVecFnc) Search(praed Value) int     { return newPairSorter(v()...).Search(praed) }
func (v AssocVecFnc) Sort(flag d.TyPrimitive) {
	var ps = newPairSorter(v()...)
	ps.Sort(flag)
	v = NewAccVector(ps...).(AssocVecFnc)
}

///////////////////////////////////////////////////
// VECTOR
// vector keeps a slice of data instances
func conVec(vec Vectorized, dd ...Value) Vectorized {
	return vectorConstructor(append(vec.Slice(), dd...)...)
}
func vectorConstructor(dd ...Value) Vectorized {
	return VecFnc(func() []Value { return dd })
}
func NewVector(dd ...Value) Vectorized {
	return VecFnc(func() (vec []Value) {
		for _, dat := range dd {
			vec = append(vec, New(dat))
		}
		return vec
	})
}

func (v AssocVecFnc) DeCap() (Paired, []Paired) {
	return v.Head(), v.Tail()
}
func (v AssocVecFnc) TypeHO() TyHigherOrder { return AssocVec }
func (v AssocVecFnc) TypePrim() d.TyPrimitive {
	if len(v()) > 0 {
		return d.Vector | v.Head().TypePrim()
	}
	return d.Vector | d.Nil.TypePrim()
}

// base implementation functions/sliceable interface
func (v VecFnc) Head() Value {
	if v.Len() > 0 {
		return v.Vector()[0]
	}
	return nil
}
func (v VecFnc) Tail() []Value {
	if v.Len() > 1 {
		return v.Vector()[1:]
	}
	return nil
}
func (v VecFnc) Len() int { return len(v()) }
func (v VecFnc) Empty() bool {
	if len(v()) > 0 {
		for _, dat := range v() {
			if !d.Nil.TypePrim().Flag().Match(dat.TypePrim().Flag()) {
				return false
			}
		}
	}
	return true
}
func (v VecFnc) TypeHO() TyHigherOrder { return Vector }
func (v VecFnc) TypePrim() d.TyPrimitive {
	if len(v()) > 0 {
		return d.Vector.TypePrim() | v.Head().TypePrim()
	}
	return d.Vector.TypePrim() | d.Nil.TypePrim()
}
func (v VecFnc) Ident() Value                  { return v }
func (v VecFnc) Eval(p ...d.Primary) d.Primary { return NewVector(v()...) }
func (v VecFnc) DeCap() (Value, []Value) {
	return v.Head(), v.Tail()
}
func (v VecFnc) Vector() []Value { return v() }
func (v VecFnc) Slice() []Value  { return v() }
func (v VecFnc) Call(d ...Value) Value {
	if len(d) > 0 {
		conVec(v, d...)
	}
	return v
}
func (v VecFnc) Set(i int, val Value) Vectorized {
	if i < v.Len() {
		var slice = v()
		slice[i] = val
		return VecFnc(func() []Value { return slice })

	}
	return nil
}
func (v VecFnc) Get(i int) Value {
	if i < v.Len() {
		return v()[i]
	}
	return nil
}
func (v VecFnc) Search(praed Value) int { return newDataSorter(v()...).Search(praed) }
func (v VecFnc) Sort(flag d.TyPrimitive) {
	var ps = newDataSorter(v()...)
	ps.Sort(flag)
	v = NewVector(ps...).(VecFnc)
}

// RECURSIVE LIST
// base implementation of linked lists
func NewRecursiveList(dd ...Value) Recursive {
	if len(dd) > 0 {
		if len(dd) > 1 {
			return ListFnc(func() (Value, Recursive) {
				return dd[0], NewRecursiveList(dd[1:]...)
			})
		}
		return ListFnc(func() (Value, Recursive) { return dd[0], nil })
	}
	return nil
}
func (l ListFnc) Ident() Value                  { return l }
func (l ListFnc) Eval(p ...d.Primary) d.Primary { return NewPair(l.Head(), l.Tail()) }
func (l ListFnc) Head() Value                   { h, _ := l(); return h }
func (l ListFnc) Tail() Recursive               { _, t := l(); return t }
func (l ListFnc) TypeHO() TyHigherOrder         { return List }
func (l ListFnc) TypePrim() d.TyPrimitive       { return d.List.TypePrim() | l.Head().TypePrim() }
func (l ListFnc) Empty() bool {
	var h, _ = l()
	if h != nil {
		return false
	}
	return true
}
func (l ListFnc) Len() int {
	var _, t = l()
	if t != nil {
		return 1 + t.Len()
	}
	return 1
}
func (l ListFnc) Call(d ...Value) Value {
	if len(d) > 0 {
		var head, tail = l()
		return NewPair(head, tail)
	}
	return l
}
func (l ListFnc) DeCap() (Value, Recursive) {
	var head, rec = l()
	l = ListFnc(func() (Value, Recursive) { return l() })
	return head, rec
}
func (l ListFnc) Con(val Value) ListFnc {
	return ListFnc(func() (Value, Recursive) { return val, l })
}

// TUPLE
func conTuple(tup Tupled, dat ...Value) Tupled {
	return NewTuple(append(tup.Slice(), dat...)...)
}
func NewTuple(dat ...Value) Tupled {
	var flags []d.BitFlag
	for _, data := range dat {
		flags = append(flags, data.TypePrim().Flag())
	}
	var vec = vectorConstructor(dat...)
	return TupleFnc(func() (Vectorized, []d.BitFlag) {
		return vec, flags
	})
}
func (t TupleFnc) Flags() []d.BitFlag      { _, f := t(); return f }
func (t TupleFnc) DeCap() (Value, []Value) { v, _ := t(); return v.DeCap() }
func (t TupleFnc) Slice() []Value          { v, _ := t(); return v.Slice() }
func (t TupleFnc) Head() Value             { v, _ := t(); return v.Head() }
func (t TupleFnc) Tail() []Value           { v, _ := t(); return v.Tail() }
func (t TupleFnc) Empty() bool             { v, _ := t(); return v.Empty() }
func (t TupleFnc) Len() int                { v, _ := t(); return v.Len() }
func (t TupleFnc) TypePrim() d.TyPrimitive {
	var flag d.BitFlag
	for _, elem := range t.Slice() {
		flag = flag | elem.TypePrim().Flag()
	}
	return d.TyPrimitive(flag | d.Tuple.Flag())
}
func (t TupleFnc) TypeHO() TyHigherOrder         { return Tuple }
func (t TupleFnc) Eval(p ...d.Primary) d.Primary { return NewVector(t.Slice()...) }
func (t TupleFnc) Ident() Value                  { return t }
func (t TupleFnc) Call(d ...Value) Value {
	if len(d) > 0 {
		return conTuple(t, d...)
	}
	return t
}
func (t TupleFnc) Get(i int) Value {
	if i < t.Len() {
		return t.Slice()[i]
	}
	return nil
}
func (t TupleFnc) Set(i int, val Value) Vectorized {
	if i < t.Len() {
		var slice = t.Slice()
		slice[i] = val
		return NewTuple(slice...)
	}
	return nil
}
func (t TupleFnc) Search(praed Value) int { return newDataSorter(t.Slice()...).Search(praed) }
func (t TupleFnc) Sort(flag d.TyPrimitive) {
	var ps = newDataSorter(t.Slice()...)
	ps.Sort(flag)
	t = NewTuple(ps...).(TupleFnc)
}

// RECORD
func conRecord(rec Recorded, pairs ...Paired) Recorded {
	tup, ps := rec.(RecordFnc)()
	if len(pairs) > 0 {
		return RecordFnc(func() (Tupled, []Paired) {
			return tup, pairs
		})
	}
	return RecordFnc(func() (Tupled, []Paired) {
		return tup, ps
	})
}
func NewRecord(pairs ...Paired) Recorded {
	var sig = []Paired{}
	var dat = []Value{}
	for _, pair := range pairs {
		sig = append(sig, NewPair(pair.Left(), New(pair.Right().TypePrim())))
		dat = append(dat, pair)
	}
	var tup = NewTuple(dat...)
	return RecordFnc(func() (Tupled, []Paired) {
		return tup, sig
	})
}
func (r RecordFnc) Ident() Value                  { return r }
func (r RecordFnc) Eval(p ...d.Primary) d.Primary { return r.Tuple() }
func (r RecordFnc) Call(d ...Value) Value {
	if len(d) > 0 {
		var pairs = []Paired{}
		for _, pair := range d {
			pairs = append(pairs, pair.(Paired))
		}
		return conRecord(r, pairs...)
	}
	return r
}
func (r RecordFnc) ArgSig() []Paired        { _, pairs := r(); return pairs }
func (r RecordFnc) Tuple() Tupled           { tup, _ := r(); return tup }
func (r RecordFnc) DeCap() (Value, []Value) { return r.Tuple().DeCap() }
func (r RecordFnc) Head() Value             { return r.Tuple().Head() }
func (r RecordFnc) Tail() []Value           { return r.Tuple().Tail() }
func (r RecordFnc) Slice() []Value          { return r.Tuple().Slice() }
func (r RecordFnc) Empty() bool             { return r.Tuple().Empty() }
func (r RecordFnc) Len() int                { return r.Tuple().Len() }
func (r RecordFnc) Pairs() []Paired {
	var pairs = []Paired{}
	for _, val := range r.Slice() {
		pairs = append(pairs, val.(Paired))
	}
	return pairs
}
func (r RecordFnc) GetVal(p Value) Paired {
	_, pairs := r()
	ps := newPairSorter(pairs...)
	ps.Sort(d.Symbolic)
	idx := ps.Search(p)
	if idx != -1 {
		return ps[idx]
	}
	return nil
}
func (r RecordFnc) TypePrim() d.TyPrimitive { return d.Record }
func (r RecordFnc) TypeHO() TyHigherOrder   { return Record }

func (r RecordFnc) Get(i int) Value {
	if i < r.Len() {
		return r.Slice()[i]
	}
	return nil
}
func (r RecordFnc) Set(i int, pair Value) Vectorized {
	var slice = r.Slice()
	if i < len(slice) {
		var pairs = []Paired{}
		for _, val := range slice {
			if pair, ok := val.(Paired); ok {
				pairs = append(pairs, pair)
			}
		}
		return NewRecord(pairs...)
	}
	return nil
}
func (r RecordFnc) Search(praed Value) int { return newPairSorter(r.Pairs()...).Search(praed) }
func (r RecordFnc) Sort(flag d.TyPrimitive) {
	var ps = newPairSorter(r.Pairs()...)
	ps.Sort(flag)
	r = NewRecord(ps...).(RecordFnc)
}

// PRAEDICATE
func NewPraedicate(test func(scrut Value) Boolean) PraedFnc { return PraedFnc(test) }
func (p PraedFnc) TypeHO() TyHigherOrder                    { return Predicate }
func (p PraedFnc) TypePrim() d.TyPrimitive                  { return d.Bool }
func (p PraedFnc) String() string {
	return "T → λpredicate  → Bool"
}
func (p PraedFnc) Ident() Value { return p }
func (p PraedFnc) Eval(dat ...d.Primary) d.Primary {
	return d.BoolVal(p(NewPrimaryConstatnt(dat[0])).Bool())
}
func (p PraedFnc) Call(v ...Value) Value {
	if len(v) > 0 {
		return p(v[0])
	}
	return NewNone()
}

// OPTION
func NewOption(option func(Value) Optional) OptionFnc { return OptionFnc(option) }
func (p OptionFnc) TypeHO() TyHigherOrder             { return Option }
func (p OptionFnc) TypePrim() d.TyPrimitive           { return d.Bool }
func (p OptionFnc) String() string {
	return "T → λoption  → Option"
}
func (p OptionFnc) Ident() Value { return p }
func (p OptionFnc) Eval(dat ...d.Primary) d.Primary {
	if len(dat) > 0 {
		return p(NewPrimaryConstatnt(dat[0])).Eval(dat...)
	}
	return p(NewNone()).Eval(dat...)
}
func (p OptionFnc) Call(v ...Value) Value {
	if len(v) > 0 {
		return p(v[0])
	}
	return NewNone()
}

// ENUM
func NewEnumType(enum func() (TypeFnc, []TypeFnc)) EnumFnc { return EnumFnc(enum) }
func (e EnumFnc) TypeHO() TyHigherOrder                    { return Enum }
func (e EnumFnc) TypePrim() d.TyPrimitive                  { return d.Enum }
func (e EnumFnc) String() string {
	var mem, set = e()
	var str = mem.Name().String() + " ∈ "
	for i, enum := range set {
		str = str + enum.Name().String()
		if i < len(set)-1 {
			str = str + "|"
		}
	}
	return str
}
func (e EnumFnc) Ident() Value                    { return e }
func (e EnumFnc) Eval(dat ...d.Primary) d.Primary { return d.StrVal(e.String()) }
func (e EnumFnc) Call(v ...Value) Value {
	return NewNone()
}
