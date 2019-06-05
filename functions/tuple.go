package functions

import d "github.com/joergreinhardt/gatwd/data"

type (
	//// TUPLE
	TupleElem     func(...Callable) (int, Callable)
	TupleVal      func(...Callable) []TupleElem
	TupleType     func(...Callable) TupleVal
	TupleTypeCons func(...Callable) TupleType
)

//// TUPLE TYPE
///
//
func NewTupleType(defaults ...Callable) TupleType {
	var signature = []d.Paired{}
	for _, ini := range defaults {
		signature = append(
			signature,
			d.NewPair(
				ini.TypeNat(),
				ini.TypeFnc(),
			),
		)
	}
	return func(args ...Callable) TupleVal {
		var tuples = []TupleElem{}
		for pos, elem := range defaults {
			tuples = append(
				tuples,
				NewTupleElem(elem, pos))
		}
		tuples = createTuple(signature, tuples, args...)
		return func(vals ...Callable) []TupleElem {
			if len(vals) > 0 {
				return applyTuple(signature, tuples, vals...)
			}
			return tuples
		}
	}
}

func applyTuple(
	signature []d.Paired,
	tuples []TupleElem,
	args ...Callable,
) []TupleElem {
	if len(args) > 0 {
		for pos, arg := range args {
			if arg.TypeFnc().Match(Pair | Index) {
				if pair, ok := arg.(IndexPair); ok {
					var idx, val = pair.Index(), pair.Value()
					var result = val.Call(tuples[idx].Value())
					if len(signature) > idx {
						if result.TypeFnc().Match(
							signature[idx].Right().(TyFnc),
						) && result.TypeNat().Match(
							signature[idx].Left().(d.TyNat),
						) {
							tuples[idx] = NewTupleElem(result, idx)
							continue
						}
					}
				}
			}
			if arg.TypeFnc().Match(Tuple | Element) {
				if tup, ok := arg.(TupleElem); ok {
					var idx, val = tup()
					var result = val.Call(tuples[idx].Value())
					if len(signature) > idx {
						if val.TypeFnc().Match(
							signature[idx].Right().(TyFnc),
						) && val.TypeNat().Match(
							signature[idx].Left().(d.TyNat),
						) {
							tuples[idx] = NewTupleElem(result, idx)
							continue
						}
					}
				}
			}
			var result = arg.Call(tuples[pos].Value())
			if result.TypeFnc().Match(
				signature[pos].Right().(TyFnc),
			) && result.TypeNat().Match(
				signature[pos].Left().(d.TyNat),
			) {
				tuples[pos] = NewTupleElem(result, pos)
			}
		}
	}
	return tuples
}

//// CREATE TUPLE
func createTuple(
	signature []d.Paired,
	tuples []TupleElem,
	args ...Callable,
) []TupleElem {
	if len(args) > 0 {
		for pos, arg := range args {
			if arg.TypeFnc().Match(Pair | Index) {
				if pair, ok := arg.(IndexPair); ok {
					var val, idx = pair.Value(), pair.Index()
					if len(signature) > idx {
						if val.TypeFnc().Match(
							signature[idx].Right().(TyFnc),
						) && val.TypeNat().Match(
							signature[idx].Left().(d.TyNat),
						) {
							tuples[idx] = NewTupleElem(val, idx)
							continue
						}
					}
				}
			}
			if arg.TypeFnc().Match(Tuple | Element) {
				if tup, ok := arg.(TupleElem); ok {
					var idx, val = tup.Index(), tup.Value()
					if len(signature) > idx {
						if val.TypeFnc().Match(
							signature[idx].Right().(TyFnc),
						) && val.TypeNat().Match(
							signature[idx].Left().(d.TyNat),
						) {
							tuples[idx] = NewTupleElem(val, idx)
							continue
						}
					}
				}
			}
			if args[pos].TypeFnc().Match(
				signature[pos].Right().(TyFnc),
			) && args[pos].TypeNat().Match(
				signature[pos].Left().(d.TyNat),
			) {
				tuples[pos] = NewTupleElem(args[pos], pos)
			}
		}
	}
	return tuples
}

func (t TupleType) String() string                 { return "Type " + t().String() }
func (t TupleType) Call(args ...Callable) Callable { return t().Call(args...) }
func (t TupleType) Eval(args ...d.Native) d.Native { return t().Eval(args...) }
func (t TupleType) TypeFnc() TyFnc {
	return Constructor |
		Tuple |
		Type |
		t().TypeFnc()
}

func (t TupleType) TypeNat() d.TyNat {
	return d.Functor |
		t().TypeNat()
}

//// TUPLE VALUE
func (t TupleVal) Len() int { return len(t()) }
func (t TupleVal) String() string {
	var elems = t()
	var l = len(elems)
	var str = "Tuple: "
	for pos, elem := range elems {
		str = str + elem.String()
		if pos < l-1 {
			str = str + ", "
		}
	}
	return str
}
func (t TupleVal) TypeFnc() TyFnc {
	var elems = t()
	var fnc = Tuple
	for _, elem := range elems {
		fnc = fnc | elem.TypeFnc()
	}
	return fnc
}
func (t TupleVal) TypeNat() d.TyNat {
	var elems = t()
	var nat = d.Functor
	for _, elem := range elems {
		nat = nat | elem.TypeNat()
	}
	return nat
}
func (t TupleVal) Call(args ...Callable) Callable {
	var tlen = len(t())
	var elems = []Callable{}
	for pos, elem := range args {
		if pos < tlen {
			elems = append(elems, elem)
		}
	}
	var vec = NewVector()
	for _, elem := range t(elems...) {
		vec = ConsVector(vec, elem)
	}
	return vec
}
func (t TupleVal) Eval(args ...d.Native) d.Native {
	var vals = []Callable{}
	for _, val := range args {
		vals = append(vals, AtomVal(val.Eval))
	}
	var tups = t(vals...)
	var slice = d.NewSlice()
	for _, tup := range tups {
		slice.Append(tup)
	}
	return slice
}

//// TUPLE ELEMENT
func NewTupleElem(val Callable, idx int) TupleElem {
	return func(args ...Callable) (int, Callable) {
		if len(args) > 0 {
			return idx, val.Call(args...)
		}
		return idx, val
	}
}
func (e TupleElem) Value() Callable                { var _, val = e(); return val }
func (e TupleElem) Index() int                     { var idx, _ = e(); return idx }
func (e TupleElem) String() string                 { return e.Value().String() }
func (e TupleElem) TypeFnc() TyFnc                 { return Tuple | Element | e.Value().TypeFnc() }
func (e TupleElem) TypeNat() d.TyNat               { return d.Functor | e.Value().TypeNat() }
func (e TupleElem) Call(args ...Callable) Callable { return e.Value().Call(args...) }
func (e TupleElem) Eval(args ...d.Native) d.Native { return e.Value().Eval(args...) }
