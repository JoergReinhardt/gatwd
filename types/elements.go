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
	typ   Flag
	cells Value
}

func unaryEmpty(v Value) bool {
	if v != nil {
		if val, ok := v.(Value); ok {
			if v.Flag().Match(Unary) {
				if !v.Flag().Match(Nil) {
					return false
				}
			}
		}
	}
	return true
}
func isUnary(v Value) bool {
	if !unaryEmpty(v) {
		if v.Flag().Match(Unary) {
			return true
		}
	}
	return false
}
func fetchUnary(v Value) Value {
	if isUnary(v) {
		return v
	}
}
func tupleEmpty(v Value) bool {
	if unaryEmpty(v) {
		return true
	}
	v := v.Flag().Match(Tuple)
	switch {
	case unaryEmpty(v.Head()):
		return true
	case unaryEmpty(v.Tail()):
		return true
	case collectEmpty(v.Head()):
		return true
	case collectEmpty(v.Tail()):
		return true
	}
	return false
}
func isBinary(v Value) bool {
	if v.Flag().Match(Tuple) {
		if tup, ok := v.(Tupled); ok {
			if tup.Arity() == 2 {
				return true
			}
		}
	}
	return false
}
func isNary(v Value) bool {
	if v.Flag().Match(Tuple) {
		if tup, ok := v.(Tupled); ok {
			if tup.Arity() > 2 {
				return true
			}
		}
	}
	return false
}
func binaryEmpty(v Value) bool {
	if v == nil {
		if val, ok := v.(Value); ok {
			if !v.Flag().Match(Nil) {
				if isBinary(v) {
					head, tail := v.(Tupled).Head(), v.(Tupled).Tail()
					if !tupleEmpty(head) && !tupleEmpty(tail) {
						return false
					}
				}
			}
		}
	}
	return true
}
func fetchUnary(v Value) Value {
	if v != nil {
		if v, ok := v.(Value); ok {
			t := v.Flag()
			if t.Match(Nil) {
				return NilVal{}
			}
			if t.Flag().Match(Unary) {
				return v.Value()
			}
		}
	}
	return NilVal{}
}
func fetchHead(v Value) Value {
}
func decapN(arity int, s *slice) (*slice, *slice) {
	return (*s).DecapNary(arity)
}
func newEmpty(v Value) *cell {
	return &cell{0, Flag(Nil), NilVal{}}
}
func newUnary(v Value) *cell {
	return &cell{1, v.Flag(), newSlice(v)}
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
		return &cell{arity, typ.Flag(), newSlice(v...)}
	}
	return &cell{
		arity,
		typ.Flag(),
		newSlice( // <-- heads last elements needs to be postponed...
			append(v[:arity-1], // until the cell holding the tail...
				&cell{ // has been created, containing...
					1,                      // a unary of type tuple...
					typ.Flag(),             // to hold the slice...
					newSlice(v[arity:]...), // and append the final elements...
				},
			)...),
	}
}
func (c cell) Arity() int {
	if val, ok := c.cells.(Value); ok {
		vt := val.Flag()
		switch {
		case vt.Match(Nil):
			return 0
		case vt.Match(Unary):
			return 1
		case vt.Match(Slice):
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
func empty(v Value) bool {
	if v != nil {
		if v, ok := v.(Value); ok {
			switch {
			case !v.Flag().Match(Nil):
				return true
			case !v.Flag().Match(Unary):
				return false
			case !v.Flag().Match(Chained):
			case !v.Flag().Match(Linked):
			case !v.Flag().Match(Consumed):
			case !v.Flag().Match(Ordered):
			case !v.Flag().Match(Mapped):
				return false
			}
		}
	}
	return true
}
func (c cell) Empty() bool {
	if c.cells == nil {
		return true
	}
	if val, ok := c.cells.(Value); ok {
		vt := val.Flag()
		switch {
		case vt.Match(Nil):
			return true
		case vt.Match(Unary):
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
