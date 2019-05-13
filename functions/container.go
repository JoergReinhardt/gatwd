/*
  FUNCTIONAL CONTAINERS

  containers implement enumeration of functional types, aka lists, vectors sets, pairs, tuples‥.
*/
package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	//// PREDICATE | CASE | CASE-SWITCH
	PredictExpr    func(...Callable) bool
	CaseExpr       func(...Callable) (Callable, bool)
	CaseSwitchExpr func(...Callable) (Callable, bool, Consumeable)

	//// MAYBE | JUST | NONE
	NoneVal       func()
	JustVal       func(...Callable) Callable
	MaybeVal      func(...Callable) Callable
	MaybeType     func(...Callable) MaybeVal
	MaybeTypeCons func(...Callable) MaybeType

	//// TUPLE
	TupleElem     func(...Callable) (Callable, int)
	TupleVal      func(...Callable) []TupleElem
	TupleType     func(...Callable) TupleVal
	TupleTypeCons func(...Callable) TupleType

	//// RECORD
	RecordField    func(...Callable) (string, Callable)
	RecordVal      func(...Callable) []RecordField
	RecordType     func(...Callable) RecordVal
	RecordTypeCons func(...Callable) RecordType

	//// STATIC EXPRESSIONS
	ConstantExpr func() Callable
	UnaryExpr    func(Callable) Callable
	BinaryExpr   func(a, b Callable) Callable
	NaryExpr     func(...Callable) Callable

	//// DATA VALUE
	DataVal func(args ...d.Native) d.Native
)

//// RECORD
///
//
func NewRecordType(iniargs ...Callable) RecordType {
	var signature = createSignature(iniargs...)
	var inifields = applyRecord(
		signature,
		inifields,
		iniargs...,
	)
	return func(args ...Callable) RecordVal {
		var fields = inifields
		return func(args ...Callable) []RecordField {
			if len(args) > 0 {
				fields = applyRecord(
					signature,
					fields,
					args...,
				)
			}
			return fields
		}
	}
}

func createSignature(
	args ...Callable,
) []KeyPair {
	var signature = make([]KeyPair, 0, len(args))
	for pos, arg := range args {
		// signature from record field argument
		if arg.TypeFnc().Match(Record | Element) {
			if field, ok := arg.(RecordField); ok {
				signature = append(signature, NewKeyPair(
					field.KeyStr(),
					NewPair(
						NewFromData(
							field.Value().TypeNat(),
						),
						NewFromData(
							field.Value().TypeFnc(),
						),
					)))
				continue
			}
		}
		// signature from pair argument
		if arg.TypeFnc().Match(Pair) {
			if pair, ok := arg.(Paired); ok {
				if pair.Left().TypeNat().Match(d.String) {
					if key, ok := pair.Left().Eval().(d.StrVal); ok {
						signature = append(signature, NewKeyPair(
							key.String(),
							NewPair(
								NewFromData(
									pair.Right().TypeNat(),
								),
								NewFromData(
									pair.Right().TypeFnc(),
								),
							)))
						continue
					}
				}
			}
		}
		// signature from alternating key/value arguments
		if arg.TypeNat().Match(d.String) {
			if key, ok := arg.Eval().(d.StrVal); ok {
				if len(args) > pos+1 {
					pos += 1
					arg = args[pos]
					signature = append(signature, NewKeyPair(
						key.String(),
						NewPair(
							NewFromData(
								arg.TypeNat(),
							),
							NewFromData(
								arg.TypeFnc(),
							),
						)))
					continue
				}
			}
			pos -= 1
		}
	}
	return signature
}

func applyRecord(
	signature []KeyPair,
	records []RecordField,
	args ...Callable,
) []RecordField {
	for sigpos, sig := range signature {
		for argpos, arg := range args {
			// apply record field
			if arg.TypeFnc().Match(Record | Element) {
				if recfield, ok := arg.(RecordField); ok {
					if sig.Value().TypeNat().Match(
						recfield.Value().TypeNat(),
					) && sig.Value().TypeFnc().Match(
						recfield.Value().TypeNat(),
					) && sig.KeyStr() == recfield.KeyStr() {
						records[sigpos] = recfield
						break
					}
				}
			}
			// apply pair
			if arg.TypeFnc().Match(Pair) {
				if pair, ok := arg.(Paired); ok {
					if pair.Left().TypeNat().Match(d.String) {
						if key, ok := pair.Left().Eval().(d.StrVal); ok {
							if sig.Value().TypeNat().Match(
								pair.Right().TypeNat(),
							) && sig.Value().TypeFnc().Match(
								pair.Right().TypeNat(),
							) && sig.KeyStr() == key.String() {
								records[sigpos] = NewRecordField(
									key.String(),
									pair.Right(),
								)
								break
							}
						}
					}
				}
			}
			// apply alternating key/value arguments
			if arg.TypeNat().Match(d.String) {
				if key, ok := arg.Eval().(d.StrVal); ok {
					if len(records) > argpos+1 {
						argpos += 1
						if sig.Value().TypeNat().Match(
							args[argpos].TypeNat(),
						) && sig.Value().TypeFnc().Match(
							args[argpos].TypeNat(),
						) && sig.KeyStr() == key.String() {
							records[sigpos] = NewRecordField(
								key.String(),
								args[argpos],
							)
							break
						}
					}
				}
				argpos -= 1
			}
		}
	}
	return records
}

//// RECORD TYPE
func (t RecordType) Ident() Callable                { return t }
func (t RecordType) String() string                 { return t().String() }
func (t RecordType) TypeFnc() TyFnc                 { return Constructor | Record | t().TypeFnc() }
func (t RecordType) TypeNat() d.TyNat               { return d.Functor | t().TypeNat() }
func (v RecordType) Call(args ...Callable) Callable { return v(args...) }
func (v RecordType) Eval(args ...d.Native) d.Native { return v(NatToFnc(args...)...) }

//// RECORD VALUE
func (v RecordVal) Ident() Callable { return v }
func (v RecordVal) GetKey(key string) (RecordField, bool) {
	for _, field := range v() {
		if field.KeyStr() == key {
			return field, true
		}
	}
	return EmptyRecordField(), false
}
func (v RecordVal) GetIdx(idx int) (RecordField, bool) {
	if idx < len(v()) {
		return v()[idx], true
	}
	return EmptyRecordField(), false
}
func (v RecordVal) SetKey(key string, val Callable) (RecordVal, bool) {
	if _, ok := v.GetKey(key); ok {
		_ = v(NewRecordField(key, val))
		return v, true
	}
	return v, false
}
func (v RecordVal) SetIdx(idx int, val Callable) (RecordVal, bool) {
	if field, ok := v.GetIdx(idx); ok {
		_ = v(NewRecordField(field.KeyStr(), val))
		return v, true
	}
	return v, false
}
func (v RecordVal) Consume() (Callable, Consumeable) {
	var fields = v()
	if len(fields) > 0 {
		if len(fields) > 1 {
			var args = make([]Callable, 0, len(fields)-1)
			for _, field := range fields {
				args = append(args, field)
			}
			return fields[0],
				NewVector(args...)
		}
		return fields[0], v
	}
	return EmptyRecordField(), v
}
func (v RecordVal) Head() Callable {
	if len(v()) > 0 {
		return v()[0]
	}
	return EmptyRecordField()
}
func (v RecordVal) Tail() Consumeable {
	if len(v()) > 1 {
		var args = []Callable{}
		for _, field := range v()[1:] {
			args = append(args, field)
		}
		return NewVector(args...)
	}
	return NewNone()
}
func (v RecordVal) Call(args ...Callable) Callable {
	_ = v(args...)
	return v
}
func (v RecordVal) Eval(args ...d.Native) d.Native {
	var vals = []Callable{}
	for _, arg := range args {
		vals = append(vals, DataVal(arg.Eval))
	}
	_ = v(vals...)
	return v
}
func (v RecordVal) TypeNat() d.TyNat {
	var typ = d.Functor
	for _, field := range v() {
		typ = typ | field.TypeNat()
	}
	return typ
}
func (v RecordVal) TypeFnc() TyFnc {
	var typ = Record
	for _, field := range v() {
		typ = typ | field.TypeFnc()
	}
	return typ
}
func (v RecordVal) String() string {
	var l = len(v())
	var str = "("
	for pos, field := range v() {
		str = str + field.String()
		if pos < l-1 {
			str = str + ", "
		}
	}
	return str + ")"
}

//// RECORD FIELD
func EmptyRecordField() RecordField {
	return func(...Callable) (string, Callable) { return "None", NewNone() }
}
func NewRecordField(key string, val Callable) RecordField {
	return func(args ...Callable) (string, Callable) { return key, val }
}
func (a RecordField) String() string {
	return a.Key().String() + " :: " + a.Value().String()
}
func (a RecordField) Call(args ...Callable) Callable {
	return a.Right().Call(args...)
}
func (a RecordField) Eval(args ...d.Native) d.Native {
	return a.Right().Eval(args...)
}
func (a RecordField) Both() (Callable, Callable) {
	var key, val = a()
	return NewFromData(d.StrVal(key)), val
}
func (a RecordField) Left() Callable {
	key, _ := a()
	return NewFromData(d.StrVal(key))
}
func (a RecordField) Right() Callable {
	_, val := a()
	return val
}
func (a RecordField) Empty() bool {
	if a.Left() == nil || (a.Right() == nil ||
		(!a.Right().TypeFnc().Flag().Match(None) ||
			!a.Right().TypeNat().Flag().Match(d.Nil))) {
		return true
	}
	return false
}
func (a RecordField) TypeNat() d.TyNat {
	return d.Functor |
		d.Pair |
		d.String |
		a.Value().TypeNat()
}
func (a RecordField) TypeFnc() TyFnc  { return Record | Element | a.Value().TypeFnc() }
func (a RecordField) Key() Callable   { return a.Left() }
func (a RecordField) Value() Callable { return a.Right() }
func (a RecordField) Pair() Paired    { return NewPair(a.Both()) }
func (a RecordField) Pairs() []Paired { return []Paired{NewPair(a.Both())} }
func (a RecordField) KeyStr() string  { return a.Left().Eval().String() }
func (a RecordField) Ident() Callable { return a }

//// TUPLE TYPE
///
//

func TupleTypeConstructor(elems ...Callable) TupleType {
	var signature = []d.Paired{}
	for _, ini := range elems {
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
		for pos, elem := range elems {
			tuples = append(
				tuples,
				NewTupleElem(elem, pos))
		}
		for pos, arg := range args {
			if pos < len(signature) && pos < len(signature) {
				if arg.TypeFnc().Match(signature[pos].Right().(TyFnc)) &&
					arg.TypeNat().Match(signature[pos].Left().(d.TyNat)) {
					tuples[pos] = NewTupleElem(arg, pos)
				}
			}
		}
		return func(vals ...Callable) []TupleElem {
			if len(vals) > 0 {
				return ApplyTuple(tuples, vals...)()
			}
			return tuples
		}
	}
}

//// APPLY TUPLE
func ApplyTuple(elems []TupleElem, args ...Callable) TupleVal {
	if len(args) > 0 {
		for pos, arg := range args {
			if arg.TypeFnc().Match(Pair | Index) {
				if pair, ok := arg.(IndexPair); ok {
					var idx, val = pair()
					if len(elems) > idx {
						if val.TypeFnc().Match(
							elems[idx].Value().TypeFnc(),
						) && val.TypeNat().Match(
							elems[idx].Value().TypeNat(),
						) {
							elems[idx] = NewTupleElem(val, idx)
							continue
						}
					}
				}
			}
			if arg.TypeFnc().Match(Tuple | Element) {
				if tup, ok := arg.(TupleElem); ok {
					var val, idx = tup()
					if len(elems) > idx {
						if val.TypeFnc().Match(
							elems[idx].Value().TypeFnc(),
						) && val.TypeNat().Match(
							elems[idx].Value().TypeNat(),
						) {
							elems[idx] = NewTupleElem(val, idx)
							continue
						}
					}
				}
			}
			if arg.TypeFnc().Match(
				elems[pos].Value().TypeFnc(),
			) && arg.TypeNat().Match(
				elems[pos].Value().TypeNat(),
			) {
				elems[pos] = NewTupleElem(arg, pos)
			}
		}
	}
	return func(args ...Callable) []TupleElem { return elems }
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
	return ApplyTuple(t(), elems...)
}
func (t TupleVal) Eval(args ...d.Native) d.Native {
	var vals = []Callable{}
	for _, val := range args {
		vals = append(vals, DataVal(val.Eval))
	}
	return ApplyTuple(t(), vals...)
}

//// TUPLE ELEMENT
func NewTupleElem(val Callable, idx int) TupleElem {
	return func(args ...Callable) (Callable, int) {
		if len(args) > 0 {
			return val.Call(args...), idx
		}
		return val, idx
	}
}
func (e TupleElem) Value() Callable                { var val, _ = e(); return val }
func (e TupleElem) Index() int                     { var _, pos = e(); return pos }
func (e TupleElem) String() string                 { return e.Value().String() }
func (e TupleElem) TypeFnc() TyFnc                 { return Tuple | Element | e.Value().TypeFnc() }
func (e TupleElem) TypeNat() d.TyNat               { return d.Functor | e.Value().TypeNat() }
func (e TupleElem) Call(args ...Callable) Callable { return e.Value().Call(args...) }
func (e TupleElem) Eval(args ...d.Native) d.Native { return e.Value().Eval(args...) }

//// HELPER FUNCTIONS TO HANDLE ARGUMENTS
///
// since every callable also needs to implement the eval interface and data as
// such allways boils down to native values, conversion between callable-/ &
// native arguments is frequently needed. arguments may also need to be
// reversed when intendet to be passed to certain recursive expressions, or
// returned by those
//
/// REVERSE ARGUMENTS
func RevArgs(args ...Callable) []Callable {
	var rev = []Callable{}
	for i := len(args) - 1; i > 0; i-- {
		rev = append(rev, args[i])
	}
	return rev
}

/// CONVERT NATIVE TO FUNCTIONAL
func NatToFnc(args ...d.Native) []Callable {
	var result = []Callable{}
	for _, arg := range args {
		result = append(result, NewFromData(arg))
	}
	return result
}

/// CONVERT FUNCTIONAL TO NATIVE
func FncToNat(args ...Callable) []d.Native {
	var result = []d.Native{}
	for _, arg := range args {
		result = append(result, arg.Eval())
	}
	return result
}

/// GROUP ARGUMENTS PAIRWISE
//
// assumes the arguments to either implement paired, or be alternating pairs of
// key & value. in case the number of passed arguments that are not pairs is
// uneven, last field will be filled up with a value of type none
func ArgsToPaired(args ...Callable) []Paired {
	var pairs = []Paired{}
	var alen = len(args)
	for i, arg := range args {
		if arg.TypeFnc().Match(Pair) {
			pairs = append(pairs, arg.(Paired))
		}
		if i < alen-2 {
			i = i + 1
			pairs = append(pairs, NewPair(arg, args[i]))
		}
		pairs = append(pairs, NewPair(arg, NewNone()))
	}
	return pairs
}

//// CASE EXPRESSION
///
// case evaluates first argument by applying it to the predicate and either
// returns the argument, if predicate yields true, a none instance and false if
// it's not. if more than one argument is given, additional arguments will be
// evaluated recursively.
func NewCaseExpr(expr Callable, pred PredictExpr) CaseExpr {
	return func(args ...Callable) (Callable, bool) {
		var arg Callable
		if len(args) > 0 {
			arg = args[0]
			if pred(arg) {
				return arg, true
			}
			if len(args) > 1 {
				args = args[1:]
				return NewCaseExpr(expr, pred)(args...)
			}
		}
		return NewNone(), false
	}
}
func (c CaseExpr) Ident() Callable  { return c }
func (c CaseExpr) String() string   { return "Case" }
func (c CaseExpr) TypeFnc() TyFnc   { return Case }
func (c CaseExpr) TypeNat() d.TyNat { return d.Functor }
func (c CaseExpr) Call(args ...Callable) Callable {
	var val Callable
	var ok bool
	if len(args) > 0 {
		val, ok = c(args[0])
		if len(args) > 1 {
			val = val.Call(args[1:]...)
		}
	}
	if ok {
		return val
	}
	return NewNone()
}

func (c CaseExpr) Eval(args ...d.Native) d.Native {
	var val d.Native
	var ok bool
	if len(args) > 0 {
		val, ok = c(NewFromData(args[0]))
		if len(args) > 1 {
			val = val.Eval(args[1:]...)
		}
	}
	if ok {
		return val.Eval()
	}
	return d.NilVal{}
}

//// CASE SWITCH
// takes first argument to apply to case. if first argument is the only passed
// argument, it will be reused and applyed to all cases until one matches, or
// cases are depleted.
func NewCaseSwitch(cases ...CaseExpr) CaseSwitchExpr {

	// vectorive cases to be consumeable
	var cas CaseExpr
	var vec VecVal
	if len(cases) > 0 {
		cas = cases[0]
		if len(cases) > 1 {
			cases = cases[1:]
		}
		var args []Callable
		for _, arg := range cases {
			args = append(args, arg)
		}
		vec = NewVector(args...)
	}

	// case switch encloses and consumes passed cases & applys them
	// recursively to the passed argument(s) to return the resulting value,
	// or argument, depending on the boolean indicator and a consumeable
	// containing the remaining cases.
	return func(args ...Callable) (Callable, bool, Consumeable) {
		if len(args) > 0 {
			var val, ok = cas(args...)
			if ok {
				return val, ok, vec
			}
		}
		return NewNone(), false, NewList()
	}
}
func (s CaseSwitchExpr) String() string   { return "CaseSwitch" }
func (s CaseSwitchExpr) TypeFnc() TyFnc   { return CaseSwitch }
func (s CaseSwitchExpr) TypeNat() d.TyNat { return d.Functor }
func (s CaseSwitchExpr) Call(args ...Callable) Callable {
	var val, ok, _ = s(args...)
	if ok {
		return val
	}
	return NewNone()
}
func (s CaseSwitchExpr) Eval(args ...d.Native) d.Native {
	var val, ok, _ = s(NatToFnc(args...)...)
	if ok {
		return val.Eval()
	}
	return d.NilVal{}
}

//// NONE VALUE
func NewNone() NoneVal                             { return func() {} }
func (n NoneVal) Ident() Callable                  { return n }
func (n NoneVal) Len() int                         { return 0 }
func (n NoneVal) String() string                   { return "⊥" }
func (n NoneVal) Eval(...d.Native) d.Native        { return nil }
func (n NoneVal) Value() Callable                  { return nil }
func (n NoneVal) Call(...Callable) Callable        { return nil }
func (n NoneVal) Empty() bool                      { return true }
func (n NoneVal) TypeFnc() TyFnc                   { return None }
func (n NoneVal) TypeNat() d.TyNat                 { return d.Nil }
func (n NoneVal) TypeName() string                 { return n.String() }
func (n NoneVal) Head() Callable                   { return NewNone() }
func (n NoneVal) Tail() Consumeable                { return NewNone() }
func (n NoneVal) Consume() (Callable, Consumeable) { return NewNone(), NewNone() }

//// PREDICATE
func NewPredicate(pred func(...Callable) bool) PredictExpr { return pred }
func (p PredictExpr) String() string                       { return "Predicate" }
func (p PredictExpr) TypeNat() d.TyNat                     { return d.Functor }
func (p PredictExpr) TypeFnc() TyFnc                       { return Predicate }
func (p PredictExpr) Eval(args ...d.Native) d.Native {
	if len(args) > 0 {
		return p.Call(NewFromData(args[0]))
	}
	return d.NilVal{}
}
func (p PredictExpr) Call(args ...Callable) Callable {
	if len(args) > 0 {
		return p.Call(args[0])
	}
	return NewNone()
}

//// MAYVE | JUST | NONE
///
// new maybe-type constructor returns a constructor of values of a distinct
// maybe type, as defined by the predicate passed to it and thereby effectively
// declares a new higher order type at runtime.
//
// apart from the predicate, a type signature can be passed, to be returned by
// the defined maybe data constructor, when called without arguments, to be
// returned by typeFnc, typeNat, string‥.
///
//// MAYBE TYPE CONSTRUCTOR
func NewMaybeTypeConstructor(pred PredictExpr) MaybeType {
	var constructor MaybeType
	constructor = func(args ...Callable) MaybeVal {
		if len(args) > 0 {
			if len(args) > 1 {
				var maybes = NewVector()
				for _, arg := range args {
					maybes = ConsVector(
						maybes,
						constructor(arg))
				}
				return NewMaybeValue(maybes)
			}
			var arg = args[0]
			if pred(arg) {
				return NewMaybeValue(MaybeVal(
					func(args ...Callable) Callable {
						if len(args) > 0 {
							return arg.Call(args...)
						}
						return arg
					}))
			}
		}
		return NewMaybeValue(NewNone())
	}
	return constructor
}

func (c MaybeTypeCons) String() string   { return "Maybe·Type·Constructor" }
func (c MaybeTypeCons) TypeFnc() TyFnc   { return Constructor | Maybe }
func (c MaybeTypeCons) TypeNat() d.TyNat { return d.Functor }
func (c MaybeTypeCons) Call(args ...Callable) Callable {
	if len(args) > 0 {
		if args[0].TypeFnc().Match(Predicate) {
			return c(args[0])
		}
	}
	return NewNone()
}

//// MAYBE TYPE
func (m MaybeType) String() string   { return "Maybe·Type" }
func (m MaybeType) TypeFnc() TyFnc   { return Maybe }
func (m MaybeType) TypeNat() d.TyNat { return d.Functor }
func (m MaybeType) Call(args ...Callable) Callable {
	if len(args) > 0 {
		if len(args) > 1 {
			var vals = []Callable{}
			for _, arg := range args {
				vals = append(
					vals,
					m(arg),
				)
			}
		}
		var arg = args[0]
		return m(arg)
	}
	return NewNone()
}
func (m MaybeType) Eval(args ...d.Native) d.Native {
	if len(args) > 0 {
		if len(args) > 1 {
			var vals = []Callable{}
			for _, arg := range args {
				vals = append(
					vals,
					m(DataVal(arg.Eval)),
				)
			}
		}
		var arg = args[0]
		return m(DataVal(arg.Eval))
	}
	return NewNone()
}

//// MAYBE VALUE
func (m MaybeVal) String() string                 { return m().String() }
func (m MaybeVal) TypeFnc() TyFnc                 { return m().TypeFnc() }
func (m MaybeVal) TypeNat() d.TyNat               { return m().TypeNat() }
func (m MaybeVal) Call(args ...Callable) Callable { return m().Call(args...) }
func (m MaybeVal) Eval(args ...d.Native) d.Native { return m().Eval(args...) }
func NewMaybeValue(iniargs ...Callable) MaybeVal {
	return func(args ...Callable) Callable {
		if len(iniargs) > 0 {
			if len(iniargs) > 1 {
				return NewJust(CurryN(iniargs...))
			}
			return NewJust(iniargs[0])
		}
		return NewNone()
	}
}

//// JUST VALUE
func NewJust(val Callable) JustVal {
	var just JustVal
	just = JustVal(
		func(args ...Callable) Callable {
			if len(args) > 0 {
				return val.Call(args...)
			}
			return val
		})
	return just
}
func (n JustVal) Ident() Callable   { return n }
func (n JustVal) Value() Callable   { return n() }
func (n JustVal) Head() Callable    { return n() }
func (n JustVal) Tail() Consumeable { return n }
func (n JustVal) Consume() (Callable, Consumeable) {
	return n(), NewNone()
}
func (n JustVal) String() string {
	if n().TypeFnc().Match(None) || n().TypeNat().Match(d.Nil) {
		return "None " + n().String()
	}
	return "Just·" + n().TypeNat().String() + " " + n().String()
}
func (n JustVal) Call(args ...Callable) Callable {
	return n().Call(args...)
}
func (n JustVal) Eval(args ...d.Native) d.Native {
	return n().Eval(args...)
}
func (n JustVal) Empty() bool {
	if n() != nil {
		if n().TypeFnc().Match(None) ||
			n().TypeNat().Match(d.Nil) {
			return false
		}
	}
	return true
}
func (n JustVal) TypeFnc() TyFnc {
	if n().TypeFnc().Match(None) || n().TypeNat().Match(d.Nil) {
		return n().TypeFnc()
	}
	return Just | n().TypeFnc()
}
func (n JustVal) TypeNat() d.TyNat {
	return n().TypeNat()
}
func (n JustVal) TypeName() string {
	if n().TypeFnc().Match(None) || n().TypeNat().Match(d.Nil) {
		return "None"
	}
	return "JustVal·" + n().TypeFnc().String()
}

//// STATIC FUNCTION EXPRESSIONS OF PREDETERMINED ARITY
///
// used to guard expression arity, or whenever a type is needed to have a non
// variadic argument signature.
//
/// CONSTANT EXPRESSION
func NewConstant(
	expr Callable,
) ConstantExpr {
	return func() Callable { return expr }
}
func (c ConstantExpr) TypeFnc() TyFnc            { return Functor }
func (c ConstantExpr) TypeNat() d.TyNat          { return c().TypeNat() }
func (c ConstantExpr) Eval(...d.Native) d.Native { return c().Eval() }
func (c ConstantExpr) Call(...Callable) Callable { return c() }
func (c ConstantExpr) Ident() Callable           { return c() }

/// UNARY EXPRESSION
func NewUnaryExpr(
	expr Callable,
) UnaryExpr {
	return func(arg Callable) Callable { return expr.Call(arg) }
}
func (u UnaryExpr) Ident() Callable               { return u }
func (u UnaryExpr) TypeFnc() TyFnc                { return Functor }
func (u UnaryExpr) TypeNat() d.TyNat              { return d.Functor.TypeNat() }
func (u UnaryExpr) Call(arg ...Callable) Callable { return u(arg[0]) }
func (u UnaryExpr) Eval(arg ...d.Native) d.Native { return u(NewFromData(arg...)) }

/// BINARY EXPRESSION
func NewBinaryExpr(
	expr Callable,
) BinaryExpr {
	return func(a, b Callable) Callable { return expr.Call(a, b) }
}

func (b BinaryExpr) Ident() Callable                { return b }
func (b BinaryExpr) TypeFnc() TyFnc                 { return Functor }
func (b BinaryExpr) TypeNat() d.TyNat               { return d.Functor.TypeNat() }
func (b BinaryExpr) Call(args ...Callable) Callable { return b(args[0], args[1]) }
func (b BinaryExpr) Eval(args ...d.Native) d.Native {
	return b(NewFromData(args[0]), NewFromData(args[1]))
}

/// NARY EXPRESSION
func NewNaryExpr(
	expr Callable,
) NaryExpr {
	return func(args ...Callable) Callable {
		return expr.Call(args...)
	}
}
func (n NaryExpr) Ident() Callable             { return n }
func (n NaryExpr) TypeFnc() TyFnc              { return Functor }
func (n NaryExpr) TypeNat() d.TyNat            { return d.Functor.TypeNat() }
func (n NaryExpr) Call(d ...Callable) Callable { return n(d...) }
func (n NaryExpr) Eval(args ...d.Native) d.Native {
	var params = []Callable{}
	for _, arg := range args {
		params = append(params, NewFromData(arg))
	}
	return n(params...)
}

//// DATA VALUE
///
// data value implements the callable interface but returns an instance of
// data/Value. the eval method of every native can be passed as argument
// instead of the value itself, as in 'DataVal(native.Eval)', to delay, or even
// possibly ommit evaluation of the underlying data value for cases where
// lazynes is paramount.
func New(inf ...interface{}) Callable { return NewFromData(d.New(inf...)) }

func NewDataVal(iniargs ...d.Native) DataVal {
	return DataVal(func(args ...d.Native) d.Native {
		var val d.Native
		if len(iniargs) > 0 {
			if len(iniargs) > 1 {
				val = d.NewSlice(iniargs...)
				if len(args) > 0 {
					val.Eval(args...)
				}
			}
			val = iniargs[0]
			if len(args) > 0 {
				val.Eval(args...)
			}
		}
		if len(args) > 0 {
			return d.NilVal{}.Eval(args...)
		}
		return d.NilVal{}
	})
}

func NewFromData(data ...d.Native) DataVal {
	var eval func(...d.Native) d.Native
	for _, val := range data {
		eval = val.Eval
	}
	return func(args ...d.Native) d.Native { return eval(args...) }
}

func (n DataVal) Eval(args ...d.Native) d.Native {
	if len(args) > 0 {
		if len(args) > 1 {
			return n().Eval(args...)
		}
		return n().Eval(args[0])
	}
	return n().Eval()
}

func (n DataVal) Call(vals ...Callable) Callable {
	if len(vals) > 0 {
		if len(vals) > 1 {
			return NewFromData(n(FncToNat(vals...)...))
		}
		return NewFromData(n.Eval(vals[0].Eval()))
	}
	return NewFromData(n.Eval())
}

func (n DataVal) TypeFnc() TyFnc   { return Data }
func (n DataVal) TypeNat() d.TyNat { return n().TypeNat() }
func (n DataVal) String() string   { return n().String() }
