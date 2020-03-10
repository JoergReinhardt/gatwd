/*
  type aliases from data package are wrapped by data expressions to implement
  the callable interface.

  there are several types of data expressions:

  - data constant wraps all static data types
  - data slice wraps slices of native instances
  - data go-slice are slices of instances of native go types
  - data pair is the package data implementation of a value pair
  - data sets have one implementation per key type for a variety of keytypes
  - data expression is a generic function with a signature expecting one or
    many instances of package data native instances as its arguments and one
    data/native instance as return value
*/
package functions

import d "github.com/joergreinhardt/gatwd/data"

type (
	//// NATIVE VALUE CONSTRUCTORS
	///
	// atom has to be indirected by wrapping it in a function call, since
	// d.Native is an interface type and therefore cant be aliased and/or
	// extendet by further methods.
	Atom     func() d.Native
	DatFunc  func(...d.Native) d.Native
	DatSlice func() d.DataSlice
	GoSlice  func() d.Sliceable
	DatPair  func() d.PairVal
	DatMap   func() d.Mapped
)

//// DATA CONSTRUCTOR
///
// Nat(inf ...interface{}) Native
// takes any instance of a native go type and returns a data/native instance
// based on its value.
func Nat(inf ...interface{}) d.Native { return d.New(inf...) }

// Dat(inf ...interface{}) Native
// takes any instance of any of the native go types and returns it as instance
// of data/native boxed in a functional wrapper.
func Dat(inf ...interface{}) Native { return Box(d.New(inf...)) }

// Box(args ...d.Native) Native
// takes an instance of data/native interface and boxes it in a functional
// wrapper with propper type information and referencial tranparency
func Box(args ...d.Native) Native {
	// allocate data and assign match function to a value
	var (
		nat   = d.NewData(args...)
		match = nat.Type().Match
	)

	switch { // switch based on type match
	case match(d.Function):
		if fn, ok := nat.(d.Expression); ok {
			return DatFunc(func(args ...d.Native) d.Native {
				if len(args) > 0 {
					return fn(args...)
				}
				return fn()
			})
		}
	case match(d.Slice):
		if slice, ok := nat.(d.DataSlice); ok {
			return DatSlice(func() d.DataSlice {
				return slice
			})
		}
	case match(d.Unboxed):
		if unboxed, ok := nat.(d.Sliceable); ok {
			return GoSlice(func() d.Sliceable {
				return unboxed
			})
		}
	case match(d.Pair):
		if pair, ok := nat.(d.PairVal); ok {
			return DatPair(func() d.PairVal {
				return pair
			})
		}
	case match(d.Map):
		if hmap, ok := nat.(d.Mapped); ok {
			return DatMap(func() d.Mapped {
				return hmap
			})
		}
	}
	// if instance is neither of type function, nor a collection,
	// instanciate a native atomic constant.
	return Atom(func() d.Native { return nat })
}

// helper to generate type identifying pattern from native types
func patternFromNative(nat d.Native) (typed d.Typed) {
	switch {
	case nat.Type().Match(d.Pair):
		var p = nat.(d.PairVal)
		typed = Declare(p.TypeKey(), p.TypeValue())
	case nat.Type().Match(d.Unboxed):
		var u = nat.(d.Sliceable)
		typed = Declare(u.Type(), u.TypeElem())
	case nat.Type().Match(d.Slice):
		var s = nat.(d.Sliceable)
		typed = Declare(s.Type(), s.TypeElem())
	case nat.Type().Match(d.Map):
		var m = nat.(d.Mapped)
		typed = Declare(m.Type(), m.TypeKey(), m.TypeValue())
	default:
		typed = nat.Type()
	}
	return typed
}

//// NATIVE FUNCTION EXPRESSION
///
// function argument-/ and return types are not known by the framework, to make
// native expressions type safe, they need to be wrapped in a function
// definition
func (n DatFunc) TypeFnc() TyFnc                 { return Data }
func (n DatFunc) TypeNat() d.TyNat               { return n().Type() }
func (n DatFunc) String() string                 { return n().String() }
func (n DatFunc) Eval(args ...d.Native) d.Native { return n(args...) }
func (n DatFunc) Type() Decl {
	return Declare(Declare(Data, Value), patternFromNative(n()))
}
func (n DatFunc) Call(args ...Functor) Functor {
	if len(args) > 0 {
		var nats = make([]d.Native, 0, len(args))
		for _, arg := range args {
			if arg.TypeFnc().Match(Data) {
				if data, ok := arg.(Evaluable); ok {
					var eval = data.Eval()
					nats = append(nats, eval)
				}
			}
		}
		return Box(n(nats...))
	}
	return Box(n())
}

//// NATIVE ATOMIC CONSTANT
///
// functional indirection in order to allow for new set of methods indipendent
// from interface return type as defined in data package.
func (n Atom) Call(...Functor) Functor   { return n }
func (n Atom) Eval(...d.Native) d.Native { return n() }
func (n Atom) TypeFnc() TyFnc            { return Data }
func (n Atom) TypeNat() d.TyNat          { return n().Type() }
func (n Atom) String() string            { return n().String() }
func (n Atom) Type() Decl {
	return Declare(Declare(
		Data, Atomic,
	),
		patternFromNative(n()),
	)
}

// NATIVE SLICE VALUE
func (n DatSlice) Call(args ...Functor) Functor   { return n }
func (n DatSlice) TypeFnc() TyFnc                 { return Data }
func (n DatSlice) TypeNat() d.TyNat               { return n().Type() }
func (n DatSlice) Len() int                       { return n().Len() }
func (n DatSlice) Head() d.Native                 { return n().Head() }
func (n DatSlice) Tail() d.Sequential             { return n().Tail() }
func (n DatSlice) Shift() (d.Native, d.DataSlice) { return n().Shift() }
func (n DatSlice) SliceNat() []d.Native           { return n().Slice() }
func (n DatSlice) Get(key d.Native) d.Native      { return n().Get(key) }
func (n DatSlice) GetInt(idx int) d.Native        { return n().GetInt(idx) }
func (n DatSlice) Range(s, e int) d.Sliceable     { return n().Range(s, e) }
func (n DatSlice) Empty() bool                    { return n().Empty() }
func (n DatSlice) Copy() d.Native                 { return n().Copy() }
func (n DatSlice) ElemType() d.Typed              { return n().TypeElem() }
func (n DatSlice) String() string                 { return n().String() }
func (n DatSlice) Slice() []d.Native              { return n().Slice() }
func (n DatSlice) Type() Decl {
	return Declare(Declare(Data, Vector), patternFromNative(n()))
}
func (n DatSlice) Eval(args ...d.Native) d.Native {
	return d.SliceAppend(n(), args...)
}
func (n DatSlice) SliceExpr() []Functor {
	var slice = []Functor{}
	for _, nat := range n.Slice() {
		slice = append(slice, Box(nat))
	}
	return slice
}

// SLICES OF NATIVE VALUES
func (n GoSlice) Call(args ...Functor) Functor { return n }
func (n GoSlice) Eval(args ...d.Native) d.Native {
	return d.NewSlice(append(n.Slice(), args...)...)
}
func (n GoSlice) TypeFnc() TyFnc             { return Data }
func (n GoSlice) TypeNat() d.TyNat           { return n().Type() }
func (n GoSlice) Len() int                   { return n().Len() }
func (n GoSlice) Get(key d.Native) d.Native  { return n().Get(key) }
func (n GoSlice) GetInt(idx int) d.Native    { return n().GetInt(idx) }
func (n GoSlice) Range(s, e int) d.Sliceable { return n().Range(s, e) }
func (n GoSlice) Copy() d.Native             { return n().Copy() }
func (n GoSlice) Empty() bool                { return n().Empty() }
func (n GoSlice) Slice() []d.Native          { return n().Slice() }
func (n GoSlice) ElemType() d.Typed          { return n().TypeElem() }
func (n GoSlice) String() string             { return n().String() }
func (n GoSlice) Type() Decl {
	return Declare(Declare(Data, Vector), patternFromNative(n()))
}
func (n GoSlice) SliceExpr() []Functor {
	var slice = make([]Functor, 0, n().Len())
	for _, nat := range n.Slice() {
		slice = append(slice, Box(nat))
	}
	return slice
}

// NATIVE PAIR VALUE CONSTRUCTOR
func (n DatPair) Call(args ...Functor) Functor { return n }
func (n DatPair) Eval(...d.Native) d.Native    { return n() }
func (n DatPair) TypeFnc() TyFnc               { return Data }
func (n DatPair) TypeNat() d.TyNat             { return n().Type() }
func (n DatPair) Left() d.Native               { return n().Left() }
func (n DatPair) Right() d.Native              { return n().Right() }
func (n DatPair) Both() (l, r d.Native)        { return n().Both() }
func (n DatPair) LeftType() d.TyNat            { return n().TypeKey() }
func (n DatPair) RightType() d.TyNat           { return n().TypeValue() }
func (n DatPair) SubType() d.Typed             { return n().Type() }
func (n DatPair) String() string               { return n().String() }
func (n DatPair) LeftExpr() Functor            { return Box(n().Left()) }
func (n DatPair) RightExpr() Functor           { return Box(n().Right()) }
func (n DatPair) Empty() bool {
	if n.Left().Type().Match(d.Nil) &&
		n.Right().Type().Match(d.Nil) {
		return true
	}
	return false
}
func (n DatPair) BothExpr() (l, r Functor) {
	return Box(n().Left()),
		Box(n().Right())
}
func (n DatPair) Pair() Paired {
	return NewPair(
		Box(n().Left()),
		Box(n().Right()))
}
func (n DatPair) Type() Decl {
	return Declare(Declare(
		Data,
		Pair,
	), Declare(
		patternFromNative(n().Left()),
		patternFromNative(n().Right())))
}

// NATIVE MAP OF VALUES
func (n DatMap) Call(args ...Functor) Functor         { return n }
func (n DatMap) Eval(...d.Native) d.Native            { return n() }
func (n DatMap) TypeFnc() TyFnc                       { return Data }
func (n DatMap) TypeNat() d.TyNat                     { return n().Type() }
func (n DatMap) Len() int                             { return n().Len() }
func (n DatMap) Slice() []d.Native                    { return n().Slice() }
func (n DatMap) GetNat(acc d.Native) (d.Native, bool) { return n().Get(acc) }
func (n DatMap) SetNat(acc, val d.Native) d.Mapped    { return n().Set(acc, val) }
func (n DatMap) Delete(acc d.Native) bool             { return n().Delete(acc) }
func (n DatMap) Get(acc d.Native) (d.Native, bool)    { return n().Get(acc) }
func (n DatMap) Set(acc, val d.Native) d.Mapped       { return n().Set(acc, val) }
func (n DatMap) Keys() []d.Native                     { return n().Keys() }
func (n DatMap) Data() []d.Native                     { return n().Data() }
func (n DatMap) Fields() []d.Paired                   { return n().Fields() }
func (n DatMap) KeyType() d.Typed                     { return n().TypeKey() }
func (n DatMap) ValType() d.Typed                     { return n().TypeValue() }
func (n DatMap) SubType() d.Typed                     { return n().Type() }
func (n DatMap) String() string                       { return n().String() }
func (n DatMap) KeysExpr() []Functor {
	var exprs = make([]Functor, 0, n.Len())
	for _, key := range n().Keys() {
		exprs = append(exprs, Box(key))
	}
	return exprs
}
func (n DatMap) DataExpr() []Functor {
	var exprs = make([]Functor, 0, n.Len())
	for _, val := range n().Data() {
		exprs = append(exprs, Box(val))
	}
	return exprs
}
func (n DatMap) SliceExpr() []Functor {
	var slice = make([]Functor, 0, n.Len())
	for _, nat := range n.Fields() {
		slice = append(slice, Box(nat))
	}
	return slice
}
func (n DatMap) Pairs() []Paired {
	var pairs = []Paired{}
	for _, field := range n.Fields() {
		pairs = append(
			pairs, NewPair(
				Box(field.Left()),
				Box(field.Right())))
	}
	return pairs
}
func (n DatMap) Type() Decl {
	if n().Len() > 0 {
		return Declare(
			Declare(
				Data,
				Pair,
			), Declare(
				patternFromNative(n().First().Left()),
				patternFromNative(n().First().Right())))
	}
	return Declare(Declare(Data, Pair), Declare(None, None))
}
