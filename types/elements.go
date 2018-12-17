package types

import (
	"strconv"
)

////// USER DEFINED TYPES ///////////////
/// higher order constructors to dynamicly generate, allocate and validat
// typesafe collections according to signatures defined by the user during
// runtime,
type ConsumeableCollectConstructor func(...Value) Consumeable

func newConsumeable(ElemCollectConstructor, ...Attribute) func(...Value) Consumeable {
	return func(v ...Value) Consumeable {
		var a Consumeable
		return a
	}
}

type ElemCollectConstructor func(...Value) Element

func newElemCollectConstructor(t ...Type) ElemCollectConstructor {
	return func(v ...Value) Element {
		var e Element
		return e
	}
}

type NTupleConstructor func(...Value) Tupled

func NewNTupleConstructor(n int) NTupleConstructor {
	return func(v ...Value) Tupled {
		var nt Tupled
		return nt
	}
}

////// BASE STRUCTURAL ELEMENTS /////////

/// element index / field key accessor (part of a values identity, should it
// happen to be part of an engulfing data structure.
type Access struct {
	pair
}

func (a Access) AccType() Type         { return a.Head().Type() }
func (a Access) Acc() Value            { return a.Tail() }
func newAcc(t Flag, val Value) *Access { return &Access{*newPair(t, val)} }

type ValAcc struct {
	*Access
	val Value
}

func (a ValAcc) Value() Value { return a.val }

func newValAcc(accType ValType, accVal interface{}, val interface{}) ValAcc {
	return ValAcc{newAcc(Make(accType).(Flag), Make(accVal)), Make(val)}
}

type IdxAcc struct {
	*Access
	val Value
}

func (a IdxAcc) Value() Value { return a.val }
func (a IdxAcc) Idx() int     { return int(a.Acc().(IntVal)) }

func newIdxAcc(idx int, val interface{}) IdxAcc {
	return IdxAcc{newAcc(Make(Int).(Flag), Make(idx).(IntVal)), Make(val)}
}

type StrAcc struct {
	*Access       // type of key AND actual key!!!
	val     Value // contained value
}

func (a StrAcc) Key() string  { return string(a.Acc().(StringVal)) }
func (a StrAcc) Value() Value { return a.val }

func newStrAcc(key string, val interface{}) StrAcc {
	return StrAcc{newAcc(Make(String).(Flag), Make(key)), Make(val)}
}

const TAIL_SLICE_MAX = 15

func newTupled(v ...Value) Tupled {
	if c, ok := newTuple(v...).(Tupled); ok {
		return c
	}
	return nil
}
func newTuple(v ...Value) (c Tupled) {
	if len(v) > 0 {
		if len(v) > 1 {
			if len(v) > 2 {
				if len(v) > 3 {
					if len(v) > TAIL_SLICE_MAX {
						return &cell{newleave(v[0]), newSlice(v[1:]...)}
					}
					return &cell{newleave(v[0]), newTailSlice(v[1:]...)}
				}
				return &cell{newleave(v[0]), newPair(v[1], v[2])}
			}
			return newPair(v[0], v[1])
		}
		return newleave(v[0])
	}
	return nil
}

type cell struct {
	val  Value
	tail Tupled
}

func (c cell) Type() Type       { return Tuple.Type() }
func (c cell) Value() Value     { return c }
func (c cell) Ref() interface{} { return &c }
func (c cell) Head() Value {
	if c.val != nil {
		if c.val.(Tupled).Empty() {
			return c.val
		}
	}
	return &leave{}
}
func (c cell) Tail() Tupled {
	if c.val != nil {
		if c.val.(Tupled).Empty() {
			return c.tail
		}
	}
	return &leave{}
}
func (c *cell) Decap() (v Value, rc Tupled) {
	if c.val != nil && !c.val.(Tupled).Empty() {
		v = (*c).val
		if c.tail != nil && !c.val.(Tupled).Empty() {
			if !c.tail.Flat() {
				(*c).val, (*c).tail = (*c).tail.Decap()
				return v, c
			}
			if c.tail.Flat() {
				return c.tail.Head(), &leave{c.tail.(Tupled).Value().(Value)}
			}
			return c.tail.Head(), c.tail.Tail()
		}
		return &leave{c.val}, &leave{}
	}
	return &leave{}, &leave{}
}
func (c cell) Empty() bool {
	if c.val != nil {
		if !c.val.Type().Match(Nil.Flag()) {
			if c.tail != nil {
				if !c.tail.Empty() {
					return true
				}
			}
		}
	}
	return false
}
func (c cell) Unary() bool { return false }
func (c cell) Arity() (i int) {
	if !c.Empty() {
		if !c.Head() {
			i = i + 1
			if !c.Tail() {
				if !c.Tail().Unary() {
					i = i + c.Arity()
				}
				i = i + 1
			}
		}
	}
	return i
}
func (c cell) Flat() bool {
	if c.tail != nil {
		if c.val.Type().Match(Flat.Flag()) {
			if c.tail.Type().Match(Flat.Flag()) {
				return true
			}
		}
	}
	return false
}
func (c cell) String() string {
	return c.val.String() + "\t" + c.tail.String()
}
func (c cell) Copy() Value { return &cell{c.val.Copy(), c.tail.Copy().(Tupled)} }

//// TAIL SLICE /////
type tailSlice struct {
	s []Value
}

func (t *tailSlice) Decap() (Value, Tupled) {
	var v Value
	if len(t.s) > 0 {
		if len(t.s) > 1 {
			v, (*t).s = t.s[0], t.s[1:]
			return v, t
		}
		return t.s[0], &leave{}
	}
	return &leave{}, &leave{}
}
func (c tailSlice) Type() Type       { return Tuple.Type() }
func (c tailSlice) Value() Value     { return c }
func (c tailSlice) Ref() interface{} { return &c }
func (c tailSlice) Head() Value {
	if len(c.s) > 0 {
		return c.s[0]
	}
	return nil
}
func (c tailSlice) Tail() Tupled {
	if len(c.s) > 0 {
		return c.s[1].(Tupled)
	}
	return nil
}
func (c tailSlice) String() string {
	var str string
	for i, v := range c.s {
		str = str + "\t" + strconv.Itoa(i) + "\t" + v.String() + "\n"
	}
	return str
}
func (c tailSlice) Copy() Value { return &tailSlice{append([]Value{}, c.s...)} }
func (c tailSlice) Unary() bool { return false }
func (c tailSlice) Empty() bool { return empty(c) }
func (c tailSlice) Arity() int {
	if len(c.s) > 1 {
		for _, v := range c.s {
			if !v.(Tupled).Flat() {
				return false
			}
		}
	}
	return true
}
func newTailSlice(v ...Value) *tailSlice { return &tailSlice{v} }

////// PAIR OF ELEMENTS //////
type pair struct {
	*leave
	tail *leave
}

func (c pair) Flat() bool {
	if c.leave.Flat() && c.tail.Flat() {
		return true
	}

}
func (c pair) Empty() bool { return empty(c) }
func (c pair) Arity() int  { return ary(c) }
func (c pair) Unary() bool {
	if !c.tail.Empty() {
		if !c.leave.Empty() {
			if (c.leave.Unary() || c.tail.Unary()) &&
				!(c.leave.Unary() && c.tail.Unary()) {
				return true
			}
		}
	}
	return false
}
func (c pair) Type() Type       { return Tuple.Type() }
func (c pair) Value() Value     { return c }
func (c pair) Ref() interface{} { return &c }
func (c *pair) Decap() (Value, Tupled) {
	if !c.Empty() {
		if c.tail.Empty() {
			if c.leave.Empty() {
				return &leave{}, &leave{}
			}
			return c.leave, &leave{}
		}
	}
	return c.leave, c.tail
}
func (c pair) Head() Value  { return c.leave }
func (c pair) Tail() Tupled { return c.leave }
func (c pair) String() string {
	return c.leave.String() + "\t" + c.tail.String()
}
func (c pair) Copy() Value { return &pair{c.val.Copy().(*leave), c.tail.Copy().(*leave)} }

func newPair(a, b Value) *pair { return &pair{newleave(a), newleave(b)} }

////// leave ELEMENT //////
type leave struct {
	val Value
}

func (c leave) Arity() int { return 1 }
func (c leave) Empty() bool {
	if c.val != nil {
		if c.Value() != nil {
			if !c.val.Type().Match(Nil.Flag()) {
				return false
			}
		}
	}
	return true
}
func (c leave) Flat() bool {
	if !c.Empty() { // empty allready checks bounds
		if !c.val.(Tupled).Flat() {
			return true
		}
	}
	return true
}
func (c leave) Type() Type {
	if c.Empty() {
		return Nil.Type()
	}
	return Tuple.Type()
}
func (c leave) Value() Value     { return c }
func (c leave) Ref() interface{} { return &c }
func (c leave) Unary() bool      { return true }
func (c leave) Decap() (Value, Tupled) {
	if c.val != nil {
		if c.val.(Tupled).Empty() {
			return &leave{}, &leave{}
		}
	}
	return c.Head(), &leave{}
}
func (c leave) Tail() Tupled {
	if !c.Empty() {
		if !c.val.(Tupled).Empty() {
			return c.val.(Tupled)
		}
	}
	return &leave{}
}
func (c leave) Head() Value {
	if !c.Empty() {
		if !c.val.(Tupled).Empty() {
			return c.val
		}
	}
	return &leave{}
}
func (c leave) String() string {
	return c.val.String()
}
func (c leave) Copy() Value   { return &leave{c.val.Copy()} }
func newleave(v Value) *leave { return &leave{v} }
