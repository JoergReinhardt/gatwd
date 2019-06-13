package functions

import d "github.com/joergreinhardt/gatwd/data"

type (
	//// TUPLE
	TupleCons func(...Callable) TupleVal
	TupleVal  func(...Callable) Callable
)

func NewTupleCons(args ...Callable) Callable {
	var value Callable
	return value
}

//// TUPLE TYPE
///
//
func (t TupleCons) CompType() TyComp {
	var typ TyComp
	return typ
}
func (t TupleCons) TypeNat() d.TyNat { return t().TypeNat() }
func (t TupleCons) TypeFnc() TyFnc   { return Tuple | t().TypeFnc() }

//func (t TupleCons) TypeName() string { return t().TypeName() }
func (t TupleCons) String() string { return t().String() }

func (t TupleCons) Eval() d.Native                 { return t().Eval() }
func (t TupleCons) Call(args ...Callable) Callable { return t(args...) }

//// TUPLE VALUE
//func (t TupleVal) Len() int { return len(t()) }
func (t TupleVal) String() string {
	var str string
	return str
}
func (t TupleVal) TypeFnc() TyFnc   { return Tuple }
func (t TupleVal) TypeNat() d.TyNat { return d.Functor }
func (t TupleVal) Eval() d.Native   { return d.NewNil() }
func (t TupleVal) Call(args ...Callable) Callable {
	var val Callable
	return val
}
