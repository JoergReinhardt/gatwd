package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

// type system implementation
type (
	Map    func(...Callable) Callable
	Bind   func(f, g Callable) Callable
	Fold   func(Callable, Callable, ...Callable) Callable
	Filter func(Callable, ...Callable) bool
	Zip    func(l, r Callable) Paired

	MapPaired    func(...Callable) Paired
	BindPaired   func(f, g Paired) Callable
	FoldPaired   func(Paired, Paired, ...Callable) Paired
	FilterPaired func(Paired, ...Callable) bool
	Split        func(Paired) (Callable, Callable)

	// GENERALIZED CONUMEABLE & CONSUMEABLE PAIRS
	ConsumeVal  func(...Callable) (Callable, ConsumeVal)
	ConsPairVal func(...Callable) (Callable, ConsPairVal)
)

//// CURRY
func Curry(exprs ...Callable) Callable {
	if len(exprs) > 0 {
		if len(exprs) > 1 {
			return exprs[0].Call(Curry(exprs[1:]...))
		}
		return exprs[0].Call()
	}
	return NewNone()
}

//// CONSUMEABLE FUNCTOR
func NewConsumeable(cons Consumeable) ConsumeVal {
	return ConsumeVal(func(args ...Callable) (Callable, ConsumeVal) {
		var head Callable
		head, cons = cons.Consume()
		if head != nil {
			return head, NewConsumeable(cons)
		}
		return nil, NewConsumeable(cons)
	})
}

func (m ConsumeVal) TypeName() string { return "(" + m.Head().TypeName() + ")" }
func (m ConsumeVal) TypeFnc() TyFnc   { return Collection }
func (m ConsumeVal) SubType() TyFnc   { return m.Head().TypeFnc() }
func (m ConsumeVal) TypeNat() d.TyNat {
	return m.Head().TypeNat()
}
func (m ConsumeVal) Consume() (Callable, Consumeable) {
	var head Callable
	head, m = m()
	if head == nil {
		return nil, m
	}
	return head, m
}
func (m ConsumeVal) Head() Callable {
	var head, _ = m()
	return head
}
func (m ConsumeVal) Tail() Consumeable {
	return m
}
func (m ConsumeVal) Call(args ...Callable) Callable {
	return m.Head().Call(args...)
}
func (m ConsumeVal) Eval(args ...d.Native) d.Native {
	return m.Head().Eval()
}
func (m ConsumeVal) String() string {
	return m.Head().String()
}

//// CONSUMEABLE PAIRS FUNCTOR
func NewConsumeablePairs(expr Paired) ConsPairVal {
	return func(args ...Callable) (Callable, ConsPairVal) {
		var pair Callable
		var arg = expr.Call(args...)
		switch {
		case arg.TypeFnc().Match(Pair):
			if val, ok := arg.(Paired); ok {
				pair = val
			}

		case arg.TypeFnc().Match(Collection):
			if col, ok := arg.(ConsumeVal); ok {
				var left, right Callable
				left, arg = col.Consume()
				right, arg = col.Consume()
				pair = NewPair(left, right)
			}

		default:
			pair = NewPair(arg.TypeFnc(), arg)
		}
		return pair,
			NewConsumeablePairs(expr)
	}
}

func (c ConsPairVal) Call(args ...Callable) Callable {
	var head, _ = c()
	if len(args) > 0 {
		if len(args) > 1 {
			return head.Call(args...)
		}
		return head.Call(args[0])
	}
	return head
}
func (c ConsPairVal) Eval(args ...d.Native) d.Native {
	var head, _ = c()
	return head.Eval()
}
func (c ConsPairVal) Ident() Callable {
	return c
}
func (c ConsPairVal) Consume() (Callable, Consumeable) {
	return c.Head(), c.Tail()
}
func (c ConsPairVal) Head() Callable {
	h, _ := c()
	return h
}
func (c ConsPairVal) Tail() Consumeable {
	_, t := c()
	return t
}
func (c ConsPairVal) TypeName() string { return "(" + c.Head().TypeName() + ")" }
func (c ConsPairVal) TypeFnc() TyFnc   { return Collection }
func (c ConsPairVal) SubType() TyFnc   { return c.Head().TypeFnc() }
func (c ConsPairVal) TypeNat() d.TyNat {
	return c.Head().TypeNat()
}
func (c ConsPairVal) String() string {
	return c.Head().String()
}

//// MAP
func MapL(list ListCol, mapf Map) ListCol {
	return ListCol(func(args ...Callable) (Callable, ListCol) {
		var head Callable
		head, list = list()
		if head == nil {
			return nil, list
		}
		if len(args) > 0 {
			return mapf(head.Call(args...)), MapL(list, mapf)
		}
		return mapf(head), MapL(list, mapf)
	})
}

func MapF(fnc Consumeable, fmap Map) ConsumeVal {
	return ConsumeVal(func(args ...Callable) (Callable, ConsumeVal) {
		var head, tail = fnc.Consume()
		if head == nil {
			return nil, MapF(tail, fmap)
		}
		if len(args) > 0 {
			return fmap(head.Call(args...)),
				MapF(tail, fmap)
		}
		return fmap(head),
			MapF(tail, fmap)
	})
}

func MapP(pairs ConsumeablePairs, pmap MapPaired) ConsPairVal {
	return ConsPairVal(func(args ...Callable) (Callable, ConsPairVal) {
		// decapitate list to get head and list continuation
		var pair Paired
		pair, pairs = pairs.ConsumePair()
		if pair == nil { // return empty head
			return nil, MapP(pairs, pmap)
		}
		// return result of applying arguments to fmap and the
		// list continuation
		if len(args) > 0 {
			return pmap(pair.Call(args...)),
				MapP(pairs, pmap)
		}
		return pmap(pair),
			MapP(pairs, pmap)
	})
}

//// BIND
func BindL(fl, gl ListCol, bind Bind) ConsumeVal {
	return ConsumeVal(func(args ...Callable) (Callable, ConsumeVal) {
		var f, g Callable
		f, fl = fl()
		g, gl = gl()
		if f != nil {
			if g != nil {
				return bind(f, g), BindL(fl, gl, bind)
			}
		}
		return nil, BindL(fl, gl, bind)
	})
}

func BindF(fa, ga Consumeable, bind Bind) ConsumeVal {
	return ConsumeVal(func(args ...Callable) (Callable, ConsumeVal) {
		var f, g Callable
		f, fa = fa.Consume()
		g, ga = ga.Consume()
		if f != nil {
			if g != nil {
				return bind(f, g), BindF(fa, ga, bind)
			}
		}
		return nil, BindF(fa, ga, bind)
	})
}

func BindP(fp, gp ConsumeablePairs, bind Bind) ConsPairVal {
	return ConsPairVal(func(args ...Callable) (Callable, ConsPairVal) {
		var f, g Paired
		f, fp = fp.ConsumePair()
		g, gp = gp.ConsumePair()
		if f != nil {
			if g != nil {
				return bind(f, g), BindP(fp, gp, bind)
			}
		}
		return nil, BindP(fp, gp, bind)
	})
}

//// FOLD
func FoldL(list ListCol, elem Callable, fold Fold) ListCol {
	return ListCol(func(args ...Callable) (Callable, ListCol) {
		var head, tail = list()
		if head == nil {
			return nil, list
		}
		if len(args) > 0 {
			elem = fold(elem, head, args...)
			return elem, FoldL(tail, elem, fold)
		}
		elem = fold(elem, head)
		return elem, FoldL(tail, elem, fold)
	})
}

func FoldF(cons Consumeable, elem Callable, fold Fold) ConsumeVal {
	return ConsumeVal(func(args ...Callable) (Callable, ConsumeVal) {
		var head Callable
		head, cons = cons.Consume()
		if head == nil {
			return nil, FoldF(cons, elem, fold)
		}
		if len(args) > 0 {
			elem = fold(elem, head, args...)
			return elem, FoldF(cons, elem, fold)
		}
		elem = fold(elem, head)
		return elem, FoldF(cons, elem, fold)
	})
}

func FoldP(pairs ConsumeablePairs, elem Callable, fold Fold) ConsPairVal {
	return ConsPairVal(func(args ...Callable) (Callable, ConsPairVal) {
		var pair Paired
		pair, pairs = pairs.ConsumePair()
		if pair == nil {
			return nil, FoldP(pairs, elem, fold)
		}
		if len(args) > 0 {
			elem = fold(elem, pair, args...)
			return elem, FoldP(pairs, elem, fold)
		}
		elem = fold(elem, pair)
		return elem, FoldP(pairs, elem, fold)
	})
}

//// FILTER FUNCTOR
func FilterL(list ListCol, filter Filter) ListCol {
	return ListCol(
		func(args ...Callable) (Callable, ListCol) {
			var head, tail = list()
			if head == nil {
				return nil, list
			}
			// filter either returns true & head is returned, or
			// FilterL will be called recursively
			if !filter(head, args...) {
				return FilterL(tail, filter)(args...)
			}
			return head, FilterL(tail, filter)
		})
}

func FilterF(cons Consumeable, filter Filter) ConsumeVal {
	return ConsumeVal(
		func(args ...Callable) (Callable, ConsumeVal) {
			var head, tail = cons.Consume()
			if head == nil {
				return nil, FilterF(cons, filter)
			}
			if !filter(head, args...) {
				return FilterF(tail, filter)(args...)
			}
			return head, FilterF(tail, filter)
		})
}

func FilterP(pairs ConsumeablePairs, filter Filter) ConsPairVal {
	return ConsPairVal(
		func(args ...Callable) (Callable, ConsPairVal) {
			var pair Paired
			pair, pairs = pairs.ConsumePair()
			if pair == nil {
				return nil, FilterP(pairs, filter)
			}
			if !filter(pair, args...) {
				return FilterP(pairs, filter)(args...)
			}
			return pair, FilterP(pairs, filter)
		})
}

//// ZIP
func ZipL(llist, rlist ListCol, zip Zip) ListCol {
	return ListCol(
		func(args ...Callable) (Callable, ListCol) {
			var lhead, ltail = llist()
			var rhead, rtail = rlist()
			if lhead == nil || rhead == nil {
				return nil, ZipL(llist, rlist, zip)
			}
			if len(args) > 0 {
				return zip(lhead, rhead).Call(args...), ZipL(ltail, rtail, zip)
			}
			return zip(lhead, rhead), ZipL(ltail, rtail, zip)
		})
}

func ZipF(lcons, rcons Consumeable, zip Zip) ConsumeVal {
	return ConsumeVal(
		func(args ...Callable) (Callable, ConsumeVal) {
			var lhead, ltail = lcons.Consume()
			var rhead, rtail = rcons.Consume()
			if lhead == nil || rhead == nil {
				return nil,
					ZipF(lcons, rcons, zip)
			}
			if len(args) > 0 {
				return zip(lhead, rhead).Call(args...),
					ZipF(ltail, rtail, zip)
			}
			return zip(lhead, rhead),
				ZipF(ltail, rtail, zip)
		})
}

//// SPLIT
func SplitP(pairs ConsumeablePairs, split Split) func() (
	Callable,
	Callable,
	Consumeable,
	Consumeable,
) {
	var l, r Callable
	var lv, rv = NewVector(), NewVector()
	return func() (Callable, Callable, Consumeable, Consumeable) {
		var pair Paired
		pair, pairs = pairs.ConsumePair()
		if l != nil {
			lv = lv.Append(l)
		}
		if r != nil {
			rv = rv.Append(r)
		}
		l, r = split(pair)
		return l, r, lv, rv
	}
}
