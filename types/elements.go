package types

// TUPLE CELL
//
// provides the pointer referenceable struct base for all tuples, linked lists,
// trees... both fields can either be empty, or contain a collection. for a
// unary, it's a flat value. for n-nary's it's a collection of n fields, that
// are neighter head nor tail and constitute for the 'n' in n-nary. tail may
// eithere be empty, contain a flat value, or a collection.
func newTupled(v ...Value) Tupled {
	var tup Tupled
	return tup
}

// provides a pointer of known type
type cell struct {
	arity int
	typ   Type
	cells Value
}

func fetchUnary(v Value) Value {
	if v != nil {
		if v, ok := v.(Value); ok {
			t := v.Type()
			if t.Flag().Match(Nil.Flag()) {
				return NilVal{}
			}
			if t.Flag().Match(Unary.Flag()) {
				return v.Value()
			}
		}
	}
	return NilVal{}
}
func emptyVal(v Value) bool {
	if v := fetchUnary(v); v != nil {
		if !v.Flag().Match(Unary.Flag()) {
			if !v.Flag().Match(Nil.Flag()) {
				return true
			}
			return false
		}
		switch {
		}
	}
	return true
}
func decapN(arity int, s *slice) (*slice, *slice) {
	return (*s).DecapNary(arity)
}
func newEmpty(v Value) *cell {
	return &cell{0, Nil.Type(), NilVal{}}
}
func newUnary(v Value) *cell {
	return &cell{1, v.Type(), newSlice(v)}
}
func newNary(head []Value, tail []Value) *cell {
	var c *cell
	return c
}
func newTuple(head Value, tail []Value) *cell {
	var c *cell
	return c
}
func newCell(arity int, typ Type, v ...Value) *cell {
	len := len(v)
	if len <= arity {
		return &cell{arity, typ, newSlice(v...)}
	}
	return &cell{
		arity,
		typ.Type(),
		newSlice( // <-- heads last elements needs to be postponed...
			append(v[:arity-1], // until the cell holding the tail...
				&cell{ // has been created, containing...
					1,                      // a unary of type tuple...
					typ.Type(),             // to hold the slice...
					newSlice(v[arity:]...), // and append the final elements...
				},
			)...),
	}
}
func (c cell) Arity() int {
	if val, ok := c.cells.(Value); ok {
		vt := val.Flag()
		switch {
		case vt.Match(Nil.Flag()):
			return 0
		case vt.Match(Unary.Flag()):
			return 1
		case vt.Match(Slice.Flag()):
			return len(val.(SliceVal))
		default:
			// we have something... consider it flat by definition
			return 1
		}
	}
	return -1
}
func (c cell) Unary() bool {
	if c.Arity() == 1 {
		return true
	}
	return false
}
func (c cell) Empty() bool {
	if c.cells == nil {
		return true
	}
	if val, ok := c.cells.(Value); ok {
		vt := val.Flag()
		switch {
		case vt.Match(Nil.Flag()):
			return true
		case vt.Match(Unary.Flag()):
			return false
		case vt.Match(Slice.Flag()):
			if len(val.(SliceVal)) > 0 {
				return false
			}
			return true
		default:
			// all other collections are considered flat by
			// definition, if intendet otherwise they should have
			// been passed in as a slice
			return false
		}
	}
	return true
}

func (c cell) Value() Value     { return c }
func (c cell) Ref() interface{} { return &c }
func (c cell) Flag() Flag       { return c.typ.Flag() }
func (c cell) Type() Type       { return c.typ.Type() }
func (c cell) Copy() Value {
	return newCell(
		c.arity,
		c.Type(),
		c.cells.Copy(),
	)
}
func (c cell) String() string {
	return c.Head().String() + "\t" + c.Tail().String()
}
func (c *cell) Decap() (head Value, tail Tupled) {
	return head, tail
}
func (c *cell) Head() (head Value) { return head }
func (c *cell) Tail() (tail Value) { return tail }

func empty(v Value) bool { return false }
func unary(v Value) bool { return false }
func arity(v Value) int  { return 0 }
