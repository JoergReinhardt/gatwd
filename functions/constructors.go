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
	// TYPE IDENTITY & CONSTRUCTION
	TypeFnc func() (
		tid d.IntVal,
		name d.StrVal,
		signature d.StrVal,
		constructors []HigherOrderType,
	)
	SumTypeFnc  func() (HigherOrderType, []HigherOrderType)                // sum is a set of types (instance type & set)
	ProdTypeFnc func(t ...HigherOrderType) (ident, parent HigherOrderType) // product derives new type from existing type
	// TYPE INSTANCIATION (DATA CONSTRUCTORS)
	DataConstructor func() (Value, HigherOrderType)
	// HIGHER ORDER CLOSURES (BUILDTIN STATICLY DEFINED)
	PraedFnc  func(Value) Boolean  // result impl. Bool() bool
	OptionFnc func(Value) Optional // result impl.Maybe() bool, e.g. 'Just|None'
)

func NewTypeFnc(
	tid int,
	name string,
	signature string,
	cons ...HigherOrderType,
) TypeFnc {
	return TypeFnc(func() (
		id d.IntVal,
		name, sig d.StrVal,
		cons []HigherOrderType,
	) {
		return d.IntVal(tid), d.StrVal(name), d.StrVal(signature), cons
	})
}
func (t TypeFnc) Id() d.IntVal            { id, _, _, _ := t(); return id }
func (t TypeFnc) Name() d.StrVal          { _, name, _, _ := t(); return name }
func (t TypeFnc) Sig() d.StrVal           { _, _, sig, _ := t(); return sig }
func (t TypeFnc) Cons() []HigherOrderType { _, _, _, cons := t(); return cons }
func (t TypeFnc) TypePrime() d.TyPrime    { return d.Type }
func (t TypeFnc) TypeFnc() TyFnc          { return Type }
func (t TypeFnc) Ident() Value            { return t }
func (t TypeFnc) String() string {
	return strconv.Itoa(t.Id().Int()) + " " + t.Name().String()
}
func (t TypeFnc) Eval(p ...d.Primary) d.Primary {
	return d.NewPair(t.Id(), t.Name())
}

// DATA CONSTRUCTOR
//
// data constructors enclose over values together with the higher order type
// the instance is associated with.
func NewDataConstructor(
	value Value,
	hot HigherOrderType,
) DataConstructor {
	return DataConstructor(func() (Value, HigherOrderType) {
		return value, hot
	})
}
func (d DataConstructor) Ident() Value            { return d }
func (d DataConstructor) Value() Value            { val, _ := d(); return val }
func (d DataConstructor) Type() HigherOrderType   { _, ho := d(); return ho }
func (d DataConstructor) Id() d.IntVal            { return d.Type().Id() }
func (d DataConstructor) Name() d.StrVal          { return d.Type().Name() }
func (d DataConstructor) Sig() d.StrVal           { return d.Type().Name() }
func (d DataConstructor) Cons() []HigherOrderType { return d.Type().Cons() }
func (d DataConstructor) TypePrime() d.TyPrime    { return d.Value().TypePrime() }
func (d DataConstructor) TypeFnc() TyFnc          { return d.Value().TypeFnc() }
func (d DataConstructor) Eval(p ...d.Primary) d.Primary {
	return d.Value().Eval(p...)
}
func (d DataConstructor) String() string {
	return d.Value().String() + " ∷ " + d.Type().String()
}

// SUM TYPE
func NewSumType(sum func() (HigherOrderType, []HigherOrderType)) SumTypeFnc { return SumTypeFnc(sum) }
func (e SumTypeFnc) Id() d.IntVal                                           { return e.Type().Id() }
func (e SumTypeFnc) Name() d.StrVal                                         { return e.Type().Name() }
func (e SumTypeFnc) Sig() d.StrVal                                          { return e.Type().Sig() }
func (e SumTypeFnc) Cons() []HigherOrderType                                { return e.Type().Cons() }
func (e SumTypeFnc) Type() TypeFnc                                          { t, _ := e(); return t.(TypeFnc) }
func (e SumTypeFnc) Sum() []HigherOrderType                                 { _, sum := e(); return sum }
func (e SumTypeFnc) TypeFnc() TyFnc                                         { return Enum }
func (e SumTypeFnc) TypePrime() d.TyPrime                                   { return d.Enum }
func (e SumTypeFnc) Ident() Value                                           { return e }
func (e SumTypeFnc) Eval(dat ...d.Primary) d.Primary                        { return d.StrVal(e.String()) }
func (e SumTypeFnc) Call(v ...Value) Value {
	t, set := e()
	var vals = []Value{}
	for _, val := range set {
		vals = append(vals, val)
	}
	return NewPair(t, NewVector(vals...))
}
func (e SumTypeFnc) String() string {
	var mem, set = e()
	var str = mem.Name().String() + "∈"
	for i, enum := range set {
		str = str + enum.Name().String()
		if i < len(set)-1 {
			str = str + "|"
		}
	}
	return str
}

// product type constructor constructs a new product type, that derives new
// types from a base type. if a product get's called without arguments, it acts
// like a type identity function. a call with arguments, substitutes the base
// type as first parameter and appends the type function arguments to derive
// new types from, as additional parameters.
//
// PRODUCT TYPE
func NewProductType(
	tid int,
	name string,
	signature string,
	parent HigherOrderType,
	cons ...HigherOrderType,
) ProdTypeFnc {
	return ProdTypeFnc(
		func(...HigherOrderType) (typ, parent HigherOrderType) {
			return NewTypeFnc(tid, name, signature, cons...), parent
		})
}
func (e ProdTypeFnc) Id() d.IntVal                    { return e.Type().Id() }
func (e ProdTypeFnc) Name() d.StrVal                  { return e.Type().Name() }
func (e ProdTypeFnc) Sig() d.StrVal                   { return e.Type().Sig() }
func (e ProdTypeFnc) Cons() []HigherOrderType         { return e.Type().Cons() }
func (e ProdTypeFnc) Ident() Value                    { return e }
func (e ProdTypeFnc) TypeFnc() TyFnc                  { return Type }
func (e ProdTypeFnc) TypePrime() d.TyPrime            { return d.Type }
func (e ProdTypeFnc) Type() HigherOrderType           { t, _ := e(); return t }
func (e ProdTypeFnc) Parent() HigherOrderType         { _, p := e(); return p }
func (e ProdTypeFnc) Eval(dat ...d.Primary) d.Primary { return d.StrVal(e.String()) }
func (e ProdTypeFnc) Call(v ...Value) Value           { return e }
func (e ProdTypeFnc) String() string {
	var str string
	return str
}

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
func (c ConstFnc) TypePrime() d.TyPrime          { return c().TypePrime() }
func (c ConstFnc) TypeFnc() TyFnc                { return Function }
func (c ConstFnc) Ident() Value                  { return c }
func (c ConstFnc) Eval(p ...d.Primary) d.Primary { return c() }
func (c ConstFnc) Call(d ...Value) Value {
	return c().(ConstFnc)()
}

///// UNARY FUNCTION
func NewUnaryFnc(fnc func(f Value) Value) UnaryFnc {
	return UnaryFnc(func(f Value) Value { return fnc(f) })
}
func (u UnaryFnc) TypePrime() d.TyPrime          { return d.Function.TypePrime() }
func (u UnaryFnc) TypeFnc() TyFnc                { return Function }
func (u UnaryFnc) Ident() Value                  { return u }
func (u UnaryFnc) Eval(p ...d.Primary) d.Primary { return u }
func (u UnaryFnc) Call(d ...Value) Value {
	return u(d[0])
}

///// BINARY FUNCTION
func NewBinaryFnc(fnc func(a, b Value) Value) BinaryFnc {
	return BinaryFnc(func(a, b Value) Value { return fnc(a, b) })
}
func (b BinaryFnc) TypePrime() d.TyPrime          { return d.Function.TypePrime() }
func (b BinaryFnc) TypeFnc() TyFnc                { return Function }
func (b BinaryFnc) Ident() Value                  { return b }
func (b BinaryFnc) Eval(p ...d.Primary) d.Primary { return b }
func (b BinaryFnc) Call(d ...Value) Value         { return b(d[0], d[1]) }

///// NARY FUNCTION
func NewNaryFnc(fnc func(f ...Value) Value) NaryFnc {
	return NaryFnc(func(f ...Value) Value { return fnc(f...) })
}
func (n NaryFnc) TypePrime() d.TyPrime          { return d.Function.TypePrime() }
func (n NaryFnc) TypeFnc() TyFnc                { return Function }
func (n NaryFnc) Ident() Value                  { return n }
func (n NaryFnc) Eval(p ...d.Primary) d.Primary { return n }
func (n NaryFnc) Call(d ...Value) Value         { return n(d...) }

/////////////////////////////////////////////////////////
//// ASSOCIATIVE VECTOR (VECTOR OF PAIRS)
func conAssocVec(vec Accessable, pp ...Paired) Accessable {
	return conAssocVecFromPairs(append(vec.Pairs(), pp...)...)
}
func conAssocVecFromPairs(pp ...Paired) Accessable {
	return AssocVecFnc(func() []Paired { return pp })
}
func NewAssocVector(pp ...Paired) Accessable {
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
func (v AssocVecFnc) SetVal(key, value Value) Associative {
	if idx := v.Search(key); idx >= 0 {
		var pairs = v.Pairs()
		pairs[idx] = NewPair(key, value)
		return NewAssocVector(pairs...)
	}
	return NewAssocVector(append(v.Pairs(), NewPair(key, value))...).(AssocVecFnc)
}
func (v AssocVecFnc) Sort(flag d.TyPrime) {
	var ps = newPairSorter(v()...)
	ps.Sort(flag)
	v = NewAssocVector(ps...).(AssocVecFnc)
}

///////////////////////////////////////////////////
// VECTOR
// vector keeps a slice of data instances
func conVec(vec Vectorized, dd ...Value) Vectorized {
	return conVecFromValues(append(vec.Slice(), dd...)...)
}
func conVecFromValues(dd ...Value) Vectorized {
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

func (v AssocVecFnc) Call(d ...Value) Value {
	if len(d) > 0 {
		for _, val := range d {
			if pair, ok := val.(Paired); ok {
				v = v.Con(pair)
			}
		}
	}
	return v
}
func (v AssocVecFnc) Con(p ...Paired) AssocVecFnc {
	return v.Con(p...)
}
func (v AssocVecFnc) DeCap() (Paired, []Paired) {
	return v.Head(), v.Tail()
}
func (v AssocVecFnc) TypeFnc() TyFnc { return AssocVec }
func (v AssocVecFnc) TypePrime() d.TyPrime {
	if len(v()) > 0 {
		return d.Vector | v.Head().TypePrime()
	}
	return d.Vector | d.Nil.TypePrime()
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
			if !d.Nil.TypePrime().Flag().Match(dat.TypePrime().Flag()) {
				return false
			}
		}
	}
	return true
}
func (v VecFnc) TypeFnc() TyFnc { return Vector }
func (v VecFnc) TypePrime() d.TyPrime {
	if len(v()) > 0 {
		return d.Vector.TypePrime() | v.Head().TypePrime()
	}
	return d.Vector.TypePrime() | d.Nil.TypePrime()
}
func (v VecFnc) Ident() Value                  { return v }
func (v VecFnc) Eval(p ...d.Primary) d.Primary { return NewVector(v()...) }
func (v VecFnc) DeCap() (Value, []Value) {
	return v.Head(), v.Tail()
}
func (v VecFnc) Vector() []Value          { return v() }
func (v VecFnc) Slice() []Value           { return v() }
func (v VecFnc) Con(arg ...Value) []Value { return append(v(), arg...) }
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
func (v VecFnc) Sort(flag d.TyPrime) {
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
func (l ListFnc) TypeFnc() TyFnc                { return List }
func (l ListFnc) TypePrime() d.TyPrime          { return d.List.TypePrime() | l.Head().TypePrime() }
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
		flags = append(flags, data.TypePrime().Flag())
	}
	var vec = conVecFromValues(dat...)
	return TupleFnc(func() (Vectorized, []d.BitFlag) {
		return vec, flags
	})
}
func (t TupleFnc) Con(dat ...Value) Tupled { return conTuple(t, dat...) }
func (t TupleFnc) Flags() []d.BitFlag      { _, f := t(); return f }
func (t TupleFnc) DeCap() (Value, []Value) { v, _ := t(); return v.DeCap() }
func (t TupleFnc) Slice() []Value          { v, _ := t(); return v.Slice() }
func (t TupleFnc) Head() Value             { v, _ := t(); return v.Head() }
func (t TupleFnc) Tail() []Value           { v, _ := t(); return v.Tail() }
func (t TupleFnc) Empty() bool             { v, _ := t(); return v.Empty() }
func (t TupleFnc) Len() int                { v, _ := t(); return v.Len() }
func (t TupleFnc) TypePrime() d.TyPrime {
	var flag d.BitFlag
	for _, elem := range t.Slice() {
		flag = flag | elem.TypePrime().Flag()
	}
	return d.TyPrime(flag | d.Tuple.Flag())
}
func (t TupleFnc) TypeFnc() TyFnc                { return Tuple }
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
	return t
}
func (t TupleFnc) Search(praed Value) int { return newDataSorter(t.Slice()...).Search(praed) }
func (t TupleFnc) Sort(flag d.TyPrime) {
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
		sig = append(sig, NewPair(pair.Left(), New(pair.Right().TypePrime())))
		dat = append(dat, pair)
	}
	var tup = NewTuple(dat...)
	return RecordFnc(func() (Tupled, []Paired) {
		return tup, sig
	})
}
func (r RecordFnc) Ident() Value                  { return r }
func (r RecordFnc) Con(p ...Paired) Recorded      { return conRecord(r, p...) }
func (r RecordFnc) Eval(p ...d.Primary) d.Primary { return r.Tuple() }
func (r RecordFnc) ArgSig() []Paired              { _, pairs := r(); return pairs }
func (r RecordFnc) Tuple() Tupled                 { tup, _ := r(); return tup }
func (r RecordFnc) DeCap() (Value, []Value)       { return r.Tuple().DeCap() }
func (r RecordFnc) Head() Value                   { return r.Tuple().Head() }
func (r RecordFnc) Tail() []Value                 { return r.Tuple().Tail() }
func (r RecordFnc) Slice() []Value                { return r.Tuple().Slice() }
func (r RecordFnc) Empty() bool                   { return r.Tuple().Empty() }
func (r RecordFnc) Len() int                      { return r.Tuple().Len() }
func (r RecordFnc) TypePrime() d.TyPrime          { return d.Record }
func (r RecordFnc) TypeFnc() TyFnc                { return Record }
func (r RecordFnc) Search(praed Value) int {
	return newPairSorter(r.Pairs()...).Search(praed)
}
func (r RecordFnc) Sort(flag d.TyPrime) {
	var ps = newPairSorter(r.Pairs()...)
	ps.Sort(flag)
	r = NewRecord(ps...).(RecordFnc)
}
func (r RecordFnc) Get(i int) Value {
	if i < r.Len() {
		return r.Slice()[i]
	}
	return nil
}
func (r RecordFnc) Pairs() []Paired {
	var pairs = []Paired{}
	for _, val := range r.Slice() {
		pairs = append(pairs, val.(Paired))
	}
	return pairs
}
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
func (r RecordFnc) SetVal(key Value, value Value) Associative {
	if idx := r.Search(key); idx >= 0 {
		var pairs = r.Pairs()
		pairs[idx] = NewPair(key, value)
		return NewRecord(pairs...)
	}
	return NewRecord(append(r.Pairs(), NewPair(key, value))...)
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

// PRAEDICATE
func NewPraedicate(pred func(scrut Value) Boolean) PraedFnc { return PraedFnc(pred) }
func (p PraedFnc) TypeFnc() TyFnc                           { return Predicate }
func (p PraedFnc) TypePrime() d.TyPrime                     { return d.Bool }
func (p PraedFnc) Ident() Value                             { return p }
func (p PraedFnc) String() string {
	return "T → λpredicate  → Bool"
}
func (p PraedFnc) Eval(dat ...d.Primary) d.Primary {
	return d.BoolVal(p(NewPrimaryConstatnt(dat[0])).Bool())
}
func (p PraedFnc) Call(v ...Value) Value {
	if len(v) > 0 {
		return p(v[0]).(Value)
	}
	return NewNone()
}

// OPTION
func NewOption(option func(Value) Optional) OptionFnc { return OptionFnc(option) }
func (p OptionFnc) TypeFnc() TyFnc                    { return Option }
func (p OptionFnc) TypePrime() d.TyPrime              { return d.Bool }
func (p OptionFnc) Ident() Value                      { return p }
func (p OptionFnc) String() string {
	return "T → λoption  → Option"
}
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
