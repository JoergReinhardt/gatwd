package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

// type system implementation
type (
	/// CURRY FUNCTION
	Curry func(Callable, ...Callable) Callable

	/// APPLY FUNCTION
	Apply func(NaryExpr, ...Callable) Callable

	/// MAP FUNCTION
	Map       func(...Callable) Callable
	MapPaired func(...Paired) Paired

	/// FOLD FUNCTION
	Fold func(Callable, Callable, ...Callable) Callable

	/// FILTER FUNCTION
	Filter       func(Callable, ...Callable) bool
	FilterPaired func(Paired, ...Paired) bool

	/// ZIP FUNCTION
	Zip func(l, r Callable) Paired

	/// SPLIT FUNCTION
	Split func(Callable, ...Callable) Paired

	/// BIND
	// bind operator (>>=) binds the return value of one monad to be the
	// argument of another
	Bind func(f, g Callable) MonadicVal

	/// FUNCTORS
	// all functors can be applyed to map & fold by implementing the
	// consumeable interface. that renders all consumeables to be functors
	FunctorVal func(...Callable) (Callable, FunctorVal)

	/// APPLICAPLE
	// applicables enclose a functor value and an apply function that is
	// called and passed the functor and optional arguments to, whenever
	// the applicable value is evalueated, or called
	ApplicableVal func(...Callable) (Callable, ApplicableVal)

	/// MONADIC
	// monadic values provide mappings between two or more functor types by
	// taking functors as arguments and returning a functor value as result
	// and a new instance of the monadic value type to compute the next
	// result from.
	MonadicVal func(...Callable) (Callable, MonadicVal)
)

//// CURRY
func ConsCurry(f, g NaryExpr, args ...Callable) Callable {
	if len(args) > 0 {
		return f(g(args...))
	}
	return f(g())
}
func RecCurry(args ...Callable) Callable {
	if len(args) > 0 {
		var f = args[0].Call
		if len(args) > 1 {
			var g = args[1].Call
			if len(args) > 2 {
				return f(g(
					RecCurry(
						args[2:]...,
					),
				))
			}
			return f(g())
		}
		return f()
	}
	return NewNone()
}

// FUNCTOR
// new functor encloses a flat callable expression to implement consumeable so
// that it can be mapped over to return new results depending solely on the
// passed arguments for each consequtive call. the wrapping is ommited, should
// the passed expression implement the consumeable interface already and the
// expression will be type asserted and returned instead.
func NewFunctor(expr Callable) FunctorVal {
	if expr.TypeFnc().Match(Consumeables) {
		return func(args ...Callable) (Callable, FunctorVal) {
			return expr.Call(args...), expr.(FunctorVal)
		}
	}
	return func(args ...Callable) (Callable, FunctorVal) {
		if len(args) > 0 {
			if len(args) > 1 {
				return expr.Call(args...), NewFunctor(expr)
			}
			return expr.Call(args[0]), NewFunctor(expr)
		}
		return expr, NewFunctor(expr)
	}
}

func (c FunctorVal) Call(args ...Callable) Callable {
	var head, _ = c()
	if len(args) > 0 {
		if len(args) > 1 {
			return head.Call(args...)
		}
		return head.Call(args[0])
	}
	return head
}
func (c FunctorVal) Eval(args ...d.Native) d.Native {
	var head, _ = c()
	if len(args) > 0 {
		if len(args) > 1 {
			return head.Eval(args...)
		}
		return head.Eval(args[0])
	}
	return head.Eval()
}
func (c FunctorVal) Ident() Callable                  { return c }
func (c FunctorVal) Consume() (Callable, Consumeable) { return c.Head(), c.Tail() }
func (c FunctorVal) Head() Callable                   { h, _ := c(); return h }
func (c FunctorVal) Tail() Consumeable                { _, t := c(); return t }
func (c FunctorVal) TypeFnc() TyFnc                   { return Functor | c.Head().TypeFnc() }
func (c FunctorVal) TypeNat() d.TyNat                 { return c.Head().TypeNat() }
func (c FunctorVal) String() string                   { return c.Head().String() }

func MapC(cons Consumeable, fmap Map) FunctorVal {
	return FunctorVal(func(args ...Callable) (Callable, FunctorVal) {
		// decapitate list to get head and list continuation
		var head, tail = cons.Consume()
		if head == nil { // return empty head
			return nil, NewFunctor(cons)
		}
		// return result of applying arguments to fmap and the
		// list continuation
		if len(args) > 0 {
			return fmap(head).Call(args...),
				MapC(tail, fmap)
		}
		return fmap(head),
			MapC(tail, fmap)
	})
}

func MapL(list ListVal, mapf Map) ListVal {
	return ListVal(func(args ...Callable) (Callable, ListVal) {
		// decapitate list to get head and list continuation
		var head, tail = list()
		if head == nil { // return empty head
			return nil, list
		}
		// return result of applying arguments to fmap and the
		// list continuation
		if len(args) > 0 {
			return mapf(head).Call(args...), MapL(tail, mapf)
		}
		return mapf(head), MapL(tail, mapf)
	})
}

func MapF(fnc FunctorVal, fmap Map) FunctorVal {
	return FunctorVal(func(args ...Callable) (Callable, FunctorVal) {
		// decapitate list to get head and list continuation
		var head, fnc = fnc()
		if head == nil { // return empty head
			return nil, fnc
		}
		// return result of applying arguments to fmap and the
		// list continuation
		if len(args) > 0 {
			return fmap(head).Call(args...),
				MapF(fnc, fmap)
		}
		return fmap(head),
			MapF(fnc, fmap)
	})
}

func FoldL(list ListVal, elem Callable, fold Fold) ListVal {
	return ListVal(func(args ...Callable) (Callable, ListVal) {
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

func FoldF(cons Consumeable, elem Callable, fold Fold) FunctorVal {
	return FunctorVal(func(args ...Callable) (Callable, FunctorVal) {
		var head Callable
		head, cons = cons.Consume()
		if head == nil {
			return nil, NewFunctor(cons)
		}
		if len(args) > 0 {
			elem = fold(elem, head, args...)
			return elem, FoldF(cons, elem, fold)
		}
		elem = fold(elem, head)
		return elem, FoldF(cons, elem, fold)
	})
}

// FILTER FUNCTOR LATE BINDING
func FilterL(list ListVal, filter Filter) ListVal {
	return ListVal(
		func(args ...Callable) (Callable, ListVal) {
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

func FilterF(cons Consumeable, filter Filter) FunctorVal {
	return FunctorVal(
		func(args ...Callable) (Callable, FunctorVal) {
			var head, tail = cons.Consume()
			if head == nil {
				return nil, NewFunctor(cons)
			}
			if !filter(head, args...) {
				return FilterF(tail, filter)(args...)
			}
			return head, FilterF(tail, filter)
		})
}

func ZipL(llist, rlist ListVal, zip Zip) ListVal {
	return ListVal(
		func(args ...Callable) (Callable, ListVal) {
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

func ZipF(lcons, rcons Consumeable, zip Zip) FunctorVal {
	return FunctorVal(
		func(args ...Callable) (Callable, FunctorVal) {
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

// APPLY FUNCTION

// APPLICATIVE
// appliccable encloses over a consumeable-/ and an apply expression. whenn
// called, expression and optional arguments are passed to the apply function
// and the yielded result will be returned
func NewApply(arg Callable) Apply {
	return Apply(func(expr NaryExpr, args ...Callable) Callable {
		if len(args) > 0 {
			return arg.Call(args...)
		}
		return arg
	})
}

func NewApplicable(cons Consumeable, apply Apply) ApplicableVal {
	return func(args ...Callable) (Callable, ApplicableVal) {
		var head, tail = cons.Consume()
		if head == nil {
			return nil, NewApplicable(tail, apply)
		}
		if len(args) > 0 {
			if len(args) > 1 {
				return apply(head.Call, args...),
					NewApplicable(tail, apply)
			}
			return apply(head.Call, args[0]),
				NewApplicable(tail, apply)
		}
		return apply(head.Call),
			NewApplicable(tail, apply)

	}
}

func (c ApplicableVal) Call(args ...Callable) Callable {
	var head Callable
	if len(args) > 0 {
		if len(args) > 1 {
			head, _ = c(args...)
			return head
		}
		head, _ = c(args[0])
		return head
	}
	head, _ = c()
	return head
}
func (c ApplicableVal) Eval(args ...d.Native) d.Native {
	var head, _ = c()
	if len(args) > 0 {
		if len(args) > 1 {
			return head.Eval(args...)
		}
		return head.Eval(args[0])
	}
	return head.Eval()
}
func (c ApplicableVal) Ident() Callable                  { return c }
func (c ApplicableVal) Consume() (Callable, Consumeable) { return c() }
func (c ApplicableVal) Head() Callable                   { h, _ := c(); return h }
func (c ApplicableVal) Tail() Consumeable                { _, t := c(); return t }
func (c ApplicableVal) TypeFnc() TyFnc                   { return Applicable | c.Head().TypeFnc() }
func (c ApplicableVal) TypeNat() d.TyNat                 { return c.Head().TypeNat() }
func (c ApplicableVal) String() string                   { return c.Head().String() }

// MONADIC
func BindC(col Consumeable, bind Bind) MonadicVal {
	var head Callable
	return MonadicVal(
		func(args ...Callable) (Callable, MonadicVal) {
			head, col = col.Consume()
			if head == nil {
				return nil, NewMonad(col, bind)
			}
			if len(args) > 0 {
				return bind(head, col)(args...)
			}
			return bind(head, col)()
		})
}

func BindL(list ListVal, bind Bind) MonadicVal {
	return MonadicVal(
		func(args ...Callable) (Callable, MonadicVal) {
			var head Callable
			head, list = list()
			if head == nil {
				return nil, NewMonad(list, bind)
			}
			if len(args) > 0 {
				return bind(head, list)(args...)
			}
			return bind(head, list)()
		})
}

func BindF(fnc FunctorVal, bind Bind) MonadicVal {
	return MonadicVal(
		func(args ...Callable) (Callable, MonadicVal) {
			var head Callable
			head, fnc = fnc()
			if head == nil {
				return nil, NewMonad(fnc, bind)
			}
			if len(args) > 0 {
				return bind(head, fnc)(args...)
			}
			return bind(head, fnc)()
		})
}

func BindA(fna ApplicableVal, bind Bind) MonadicVal {
	return MonadicVal(
		func(args ...Callable) (Callable, MonadicVal) {
			var head Callable
			head, fna = fna()
			if head == nil {
				return nil, NewMonad(fna, bind)
			}
			if len(args) > 0 {
				return bind(head, fna)(args...)
			}
			return bind(head, fna)()
		})
}

func BindM(fnm MonadicVal, bind Bind) MonadicVal {
	return MonadicVal(
		func(args ...Callable) (Callable, MonadicVal) {
			var head Callable
			head, fnm = fnm()
			if head == nil {
				return nil, NewMonad(fnm, bind)
			}
			if len(args) > 0 {
				return BindM(fnm, bind)(args...)
			}
			return BindM(fnm, bind)()
		})
}

func NewMonad(cons Consumeable, bind Bind) MonadicVal {
	var head Callable
	return MonadicVal(
		func(args ...Callable) (Callable, MonadicVal) {
			if len(args) > 0 {
				for _, arg := range args {
					var match = arg.TypeFnc().Match
					// bind argument if bindable
					if match(Consumeables) {
						switch {
						case match(Collections):
							if col, ok := arg.(Consumeable); ok {
								return BindC(col, bind)()
							}
						case match(List):
							if list, ok := arg.(ListVal); ok {
								return BindL(list, bind)()
							}
						case match(Functor):
							if fnc, ok := arg.(FunctorVal); ok {
								return BindF(fnc, bind)()
							}
						case match(Applicable):
							if fna, ok := arg.(ApplicableVal); ok {
								return BindA(fna, bind)()
							}
						case match(Monad) || match(IO):
							if fnm, ok := arg.(MonadicVal); ok {
								return BindM(fnm, bind)()
							}
						}
					}
				}
				// consume head, to yield expression‥.
				head, cons = cons.Consume()
				if head == nil {
					return nil, NewMonad(cons, bind)
				}
				// apply non bindable arguments to head
				return head.Call(args...), NewMonad(cons, bind)
			}
			// call head without arguments
			return head.Call(), NewMonad(cons, bind)
		})
}

func (c MonadicVal) Call(args ...Callable) Callable {
	var head, _ = c(args...)
	return head
}
func (c MonadicVal) Eval(args ...d.Native) d.Native {
	return c.Call(NatToFnc(args...)...)
}
func (c MonadicVal) Ident() Callable                  { return c }
func (c MonadicVal) Consume() (Callable, Consumeable) { return c() }
func (c MonadicVal) Head() Callable                   { h, _ := c(); return h }
func (c MonadicVal) Tail() Consumeable                { _, t := c(); return t }
func (c MonadicVal) String() string                   { return c.Head().String() }
func (c MonadicVal) TypeNat() d.TyNat                 { return c.Head().TypeNat() }
func (c MonadicVal) TypeFnc() TyFnc                   { return Monad | c.Head().TypeFnc() }
