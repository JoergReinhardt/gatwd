package functions

import d "github.com/joergreinhardt/gatwd/data"

type (
	//// NATIVE VALUE CONSTRUCTORS
	DataConst   func() d.Native
	DataSlice   func(...Native) d.DataSlice
	DataGoSlice func(...Native) d.Sliceable
	DataSet     func(...Native) d.Mapped
	DataPair    func(...Native) d.PairVal

	//// NATIVE EXPRESSION CONSTRUCTOR
	DataExpr func(...d.Native) Expression
)

//// NATIVE EXPRESSION CONSTRUCTOR
///
// returns an expression with native return type implementing the callable
// interface
func New(inf ...interface{}) Native {
	return NewData(d.New(inf...))
}

func NewData(args ...d.Native) Native {

	var nats = make([]d.Native, 0, len(args))
	for _, arg := range args {
		nats = append(nats, arg)
	}
	var nat = d.NewData(nats...)
	var match = nat.TypeNat().Match

	switch {
	case match(d.Slice):
		return DataSlice(func(args ...Native) d.DataSlice {
			return nat.(d.DataSlice)
		})
	case match(d.Unboxed):
		return DataGoSlice(func(args ...Native) d.Sliceable {
			return nat.(d.Sliceable)
		})
	case match(d.Pair):
		return DataPair(func(args ...Native) d.PairVal {
			return nat.(d.PairVal)
		})
	case match(d.Map):
		return DataSet(func(args ...Native) d.Mapped {
			return nat.(d.Mapped)
		})
	}
	return DataConst(func() d.Native {
		return nat
	})
}

// ATOMIC NATIVE VALUE CONSTRUCTOR
func (n DataConst) Call(...Expression) Expression { return n }
func (n DataConst) Eval() d.Native                { return n() }
func (n DataConst) TypeFnc() TyFnc                { return Data }
func (n DataConst) TypeNat() d.TyNat              { return n().TypeNat() }
func (n DataConst) FlagType() d.Uint8Val          { return Flag_Function.U() }
func (n DataConst) String() string                { return n().String() }
func (n DataConst) TypeName() string              { return n().TypeName() }
func (n DataConst) Type() Typed {
	return Define(n().TypeNat().TypeName(), NewData(n().TypeNat()))
}

// NATIVE SLICE VALUE CONSTRUCTOR
func (n DataSlice) Call(args ...Expression) Expression { return n }
func (n DataSlice) Len() int                           { return n().Len() }
func (n DataSlice) TypeFnc() TyFnc                     { return Data }
func (n DataSlice) Eval() d.Native                     { return n() }
func (n DataSlice) Head() d.Native                     { return n().Head() }
func (n DataSlice) Tail() d.Sequential                 { return n().Tail() }
func (n DataSlice) Shift() (d.Native, d.DataSlice)     { return n().Shift() }
func (n DataSlice) SliceNat() []d.Native               { return n().Slice() }
func (n DataSlice) Get(key d.Native) d.Native          { return n().Get(key) }
func (n DataSlice) GetInt(idx int) d.Native            { return n().GetInt(idx) }
func (n DataSlice) Range(s, e int) d.Sliceable         { return n().Range(s, e) }
func (n DataSlice) Empty() bool                        { return n().Empty() }
func (n DataSlice) Copy() d.Native                     { return n().Copy() }
func (n DataSlice) TypeNat() d.TyNat                   { return n().TypeNat() }
func (n DataSlice) ElemType() d.TyNat                  { return n().ElemType() }
func (n DataSlice) String() string                     { return n().String() }
func (n DataSlice) TypeName() string                   { return n().TypeName() }
func (n DataSlice) FlagType() d.Uint8Val               { return Flag_Function.U() }
func (n DataSlice) Slice() []d.Native                  { return n().Slice() }
func (n DataSlice) Type() Typed {
	return Define(n().TypeName(), NewData(n.TypeNat()))
}
func (n DataSlice) SliceExpr() []Expression {
	var slice = make([]Expression, 0, n.Len())
	for _, nat := range n.Slice() {
		slice = append(slice, NewData(nat))
	}
	return slice
}

func (n DataGoSlice) Call(args ...Expression) Expression { return n }
func (n DataGoSlice) TypeFnc() TyFnc                     { return Data }
func (n DataGoSlice) Eval() d.Native                     { return n() }
func (n DataGoSlice) Len() int                           { return n().Len() }
func (n DataGoSlice) Get(key d.Native) d.Native          { return n().Get(key) }
func (n DataGoSlice) GetInt(idx int) d.Native            { return n().GetInt(idx) }
func (n DataGoSlice) Range(s, e int) d.Sliceable         { return n().Range(s, e) }
func (n DataGoSlice) Copy() d.Native                     { return n().Copy() }
func (n DataGoSlice) Empty() bool                        { return n().Empty() }
func (n DataGoSlice) Slice() []d.Native                  { return n().Slice() }
func (n DataGoSlice) TypeNat() d.TyNat                   { return n().TypeNat() }
func (n DataGoSlice) ElemType() d.TyNat                  { return n().ElemType() }
func (n DataGoSlice) TypeName() string                   { return n().TypeName() }
func (n DataGoSlice) FlagType() d.Uint8Val               { return Flag_Function.U() }
func (n DataGoSlice) String() string                     { return n().String() }
func (n DataGoSlice) Type() Typed                        { return Define(n.Eval().TypeName(), NewData(n.TypeNat())) }
func (n DataGoSlice) SliceExpr() []Expression {
	var slice = make([]Expression, 0, n.Len())
	for _, nat := range n.Slice() {
		slice = append(slice, NewData(nat))
	}
	return slice
}

// NATIVE PAIR VALUE CONSTRUCTOR
func (n DataPair) Call(args ...Expression) Expression { return n }
func (n DataPair) TypeFnc() TyFnc                     { return Data }
func (n DataPair) TypeNat() d.TyNat                   { return n().TypeNat() }
func (n DataPair) Eval() d.Native                     { return n() }
func (n DataPair) Left() d.Native                     { return n().Left() }
func (n DataPair) Right() d.Native                    { return n().Right() }
func (n DataPair) Both() (l, r d.Native)              { return n().Both() }
func (n DataPair) LeftType() d.TyNat                  { return n().LeftType() }
func (n DataPair) RightType() d.TyNat                 { return n().RightType() }
func (n DataPair) SubType() d.Typed                   { return n().TypeNat() }
func (n DataPair) TypeName() string                   { return n().TypeName() }
func (n DataPair) FlagType() d.Uint8Val               { return Flag_Function.U() }
func (n DataPair) String() string                     { return n().String() }
func (n DataPair) Type() Typed {
	return Define(n().TypeName(), NewData(n().TypeNat()))
}
func (n DataPair) Pair() Paired {
	return NewPair(
		NewData(n().Left()),
		NewData(n().Right()))
}
func (n DataPair) LeftExpr() Expression  { return NewData(n().Left()) }
func (n DataPair) RightExpr() Expression { return NewData(n().Right()) }
func (n DataPair) BothExpr() (l, r Expression) {
	return NewData(n().Left()),
		NewData(n().Right())
}

// NATIVE SET VALUE CONSTRUCTOR

func (n DataSet) Call(args ...Expression) Expression   { return n }
func (n DataSet) Eval() d.Native                       { return n() }
func (n DataSet) TypeFnc() TyFnc                       { return Data }
func (n DataSet) TypeNat() d.TyNat                     { return n().TypeNat() }
func (n DataSet) Len() int                             { return n().Len() }
func (n DataSet) Slice() []d.Native                    { return n().Slice() }
func (n DataSet) GetNat(acc d.Native) (d.Native, bool) { return n().Get(acc) }
func (n DataSet) SetNat(acc, val d.Native) d.Mapped    { return n().Set(acc, val) }
func (n DataSet) Delete(acc d.Native) bool             { return n().Delete(acc) }
func (n DataSet) Get(acc d.Native) (d.Native, bool)    { return n().Get(acc) }
func (n DataSet) Set(acc, val d.Native) d.Mapped       { return n().Set(acc, val) }
func (n DataSet) Keys() []d.Native                     { return n().Keys() }
func (n DataSet) Data() []d.Native                     { return n().Data() }
func (n DataSet) Fields() []d.Paired                   { return n().Fields() }
func (n DataSet) KeyType() d.TyNat                     { return n().KeyType() }
func (n DataSet) ValType() d.TyNat                     { return n().ValType() }
func (n DataSet) SubType() d.Typed                     { return n().TypeNat() }
func (n DataSet) TypeName() string                     { return n().TypeName() }
func (n DataSet) FlagType() d.Uint8Val                 { return Flag_Function.U() }
func (n DataSet) String() string                       { return n().String() }
func (n DataSet) Type() Typed {
	return Define(n().TypeName(), NewData(n()))
}
func (n DataSet) KeysExpr() []Expression {
	var exprs = make([]Expression, 0, n.Len())
	for _, key := range n().Keys() {
		exprs = append(exprs, NewData(key))
	}
	return exprs
}
func (n DataSet) DataExpr() []Expression {
	var exprs = make([]Expression, 0, n.Len())
	for _, val := range n().Data() {
		exprs = append(exprs, NewData(val))
	}
	return exprs
}
func (n DataSet) SliceExpr() []Expression {
	var slice = make([]Expression, 0, n.Len())
	for _, nat := range n.Fields() {
		slice = append(slice, NewData(nat))
	}
	return slice
}
func (n DataSet) Pairs() []Paired {
	var pairs = []Paired{}
	for _, field := range n.Fields() {
		pairs = append(
			pairs, NewPair(
				NewData(field.Left()),
				NewData(field.Right())))
	}
	return pairs
}

func NewNative(expr Expression) DataExpr {
	return func(args ...d.Native) Expression {
		if len(args) > 0 {
			var exprs = make([]Expression, 0, len(args))
			for _, arg := range args {
				exprs = append(exprs, NewData(arg))
			}
			return expr.Call(exprs...)
		}
		return expr
	}
}

func (n DataExpr) Eval() d.Native                     { return n }
func (n DataExpr) TypeFnc() TyFnc                     { return Data }
func (n DataExpr) TypeNat() d.TyNat                   { return d.Function }
func (n DataExpr) Call(args ...Expression) Expression { return n().Call(args...) }
func (n DataExpr) FlagType() d.Uint8Val               { return Flag_Function.U() }
func (n DataExpr) String() string                     { return n().String() }
func (n DataExpr) TypeName() string                   { return n().TypeName() }
func (n DataExpr) Type() Typed {
	return Define(n.TypeNat().TypeName(), NewData(n.TypeNat()))
}
