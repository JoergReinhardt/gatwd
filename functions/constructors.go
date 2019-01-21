/*
DATA CONSTRUCTORS
*/
package functions

import (
	d "github.com/JoergReinhardt/godeep/data"
)

type (
	ConstFnc  func() Function
	UnaryFnc  func(Function) Function
	BinaryFnc func(a, b Function) Function
	NaryFnc   func(...Function) Function
	VecFnc    func() []Function
	TupleFnc  func() (Vectorized, []d.BitFlag)
	RecordFnc func() (Tupled, []Paired)
	ListFnc   func() (Function, Recursive)
)

// ONSTANT
// constant also conains immutable data, but it may be the result of a constant experssion
func newConstant(dat Function) ConstFnc        { return ConstFnc(func() Function { return dat }) }
func (c ConstFnc) Flag() d.BitFlag             { return d.Definition.Flag() }
func (c ConstFnc) Ident() Function             { return c }
func (c ConstFnc) Call(d ...Function) Function { return c() }

func (u UnaryFnc) Call(d ...Function) Function {
	if len(d) > 0 {
		return u(d[0])
	}
	return nil
}
func (b BinaryFnc) Call(d ...Function) Function {
	if len(d) > 1 {
		return b(d[0], d[1])
	}
	return nil
}
func (n NaryFnc) Call(d ...Function) Function {
	if len(d) > 0 {
		return n(d...)
	}
	return nil
}

// TUPLE

// VECTOR
// vector keeps a slice of data instances
func conVec(vec Vectorized, dd ...Function) Vectorized {
	return vectorConstructor(append(vec.Slice(), dd...)...)
}
func vectorConstructor(dd ...Function) Vectorized {
	return VecFnc(func() []Function { return dd })
}
func newVector(dd ...Function) Vectorized {
	return VecFnc(func() (vec []Function) {
		for _, d := range dd {
			vec = append(vec, NewFncData(d))
		}
		return vec
	})
}

// base implementation functions/sliceable interface
func (v VecFnc) Head() Function {
	if v.Len() > 0 {
		return v.Vector()[0]
	}
	return nil
}
func (v VecFnc) Len() int { return len(v()) }
func (v VecFnc) Empty() bool {
	if len(v()) > 0 {
		for _, dat := range v() {
			if !d.Nil.Flag().Match(dat.Flag()) {
				return false
			}
		}
	}
	return true
}
func (v VecFnc) Flag() d.BitFlag { return d.Vector.Flag() }
func (v VecFnc) Ident() Function { return v }
func (v VecFnc) Tail() []Function {
	if v.Len() > 1 {
		return v.Vector()[1:]
	}
	return nil
}
func (v VecFnc) DeCap() (Function, []Function) {
	return v.Head(), v.Tail()
}
func (v VecFnc) Vector() []Function { return v() }
func (v VecFnc) Slice() []Function  { return v() }
func (v VecFnc) Call(d ...Function) Function {
	if len(d) > 0 {
		conVec(v, d...)
	}
	return v
}

// LINKED LIST
// base implementation of linked lists
func conRecurse(rec Recursive, d ...Function) Recursive {
	if len(d) > 0 {
		if len(d) > 1 {
			return ListFnc(func() (Function, Recursive) {
				return d[0], conRecurse(rec, d...)
			})
		}
		return ListFnc(func() (Function, Recursive) {
			return d[0], rec
		})
	}
	return nil
}
func newRecursive(d ...Function) Recursive {
	if len(d) > 0 {
		if len(d) > 1 {
			return ListFnc(func() (Function, Recursive) { return d[0], newRecursive(d[1:]...) })
		}
		return ListFnc(func() (Function, Recursive) { return d[0], nil })
	}
	return nil
}
func (l ListFnc) Ident() Function { return l }
func (l ListFnc) Head() Function  { h, _ := l(); return h }
func (l ListFnc) Tail() Recursive { _, t := l(); return t }
func (l ListFnc) Call(d ...Function) Function {
	if len(d) > 0 {
		return conRecurse(l, d...)
	}
	return l
}
func (l ListFnc) DeCap() (Function, Recursive) { return l() }
func (l ListFnc) Flag() d.BitFlag              { return d.List.Flag() }
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

// TUPLE
func conTuple(tup Tupled, dat ...Function) Tupled {
	return newTuple(append(tup.Slice(), dat...)...)
}
func newTuple(dat ...Function) Tupled {
	var flags []d.BitFlag
	for _, data := range dat {
		flags = append(flags, data.Flag())
	}
	var vec = vectorConstructor(dat...)
	return TupleFnc(func() (Vectorized, []d.BitFlag) {
		return vec, flags
	})
}
func (t TupleFnc) Flags() []d.BitFlag            { _, f := t(); return f }
func (t TupleFnc) DeCap() (Function, []Function) { v, _ := t(); return v.DeCap() }
func (t TupleFnc) Slice() []Function             { v, _ := t(); return v.Slice() }
func (t TupleFnc) Head() Function                { v, _ := t(); return v.Head() }
func (t TupleFnc) Tail() []Function              { v, _ := t(); return v.Tail() }
func (t TupleFnc) Empty() bool                   { v, _ := t(); return v.Empty() }
func (t TupleFnc) Len() int                      { v, _ := t(); return v.Len() }
func (t TupleFnc) Flag() d.BitFlag               { return d.Tuple.Flag() }
func (t TupleFnc) Ident() Function               { return t }
func (t TupleFnc) Call(d ...Function) Function {
	if len(d) > 0 {
		return conTuple(t, d...)
	}
	return t
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
func newRecord(pairs ...Paired) Recorded {
	var sig = []Paired{}
	var dat = []Function{}
	for _, pair := range pairs {
		sig = append(sig, NewPair(pair.Left(), pair.Right().Flag()))
		dat = append(dat, pair)
	}
	var tup = newTuple(dat...)
	return RecordFnc(func() (Tupled, []Paired) {
		return tup, sig
	})
}
func (r RecordFnc) Ident() Function { return r }
func (r RecordFnc) Call(d ...Function) Function {
	if len(d) > 0 {
		var pairs = []Paired{}
		for _, pair := range d {
			pairs = append(pairs, pair.(Paired))
		}
		return conRecord(r, pairs...)
	}
	return r
}
func (r RecordFnc) ArgSig() []Paired              { _, pairs := r(); return pairs }
func (r RecordFnc) Tuple() Tupled                 { tup, _ := r(); return tup }
func (r RecordFnc) DeCap() (Function, []Function) { return r.Tuple().DeCap() }
func (r RecordFnc) Head() Function                { return r.Tuple().Head() }
func (r RecordFnc) Tail() []Function              { return r.Tuple().Tail() }
func (r RecordFnc) Slice() []Function             { return r.Tuple().Slice() }
func (r RecordFnc) Empty() bool                   { return r.Tuple().Empty() }
func (r RecordFnc) Len() int                      { return r.Tuple().Len() }
func (r RecordFnc) Flag() d.BitFlag               { return d.Record.Flag() }
