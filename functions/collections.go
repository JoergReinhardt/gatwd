package functions

import (
	"sort"
	"strings"
)

type (

	//// COLLECTIONS
	VecVal func(...Expression) []Expression
	SeqVal func(...Expression) (Expression, SeqVal)

	//// GENERATOR | ACCUMULATOR
	GenVal func() (Expression, GenVal)
	AccVal func(...Expression) (Expression, AccVal)
)

//////////////////////////////////////////////////////////////////////////////////////////
//// VECTORS (SLICES) OF VALUES
///
// sequential vector provides random access to sequential data. appends
// arguments in the order they where passed in, at the end of slice, when
// called
func NewVector(elems ...Expression) VecVal {
	// returns empty slice of expressions when no elements are given
	if len(elems) == 0 {
		return VecVal(func(args ...Expression) []Expression {
			if len(args) > 0 {
				return NewVector(args...)()
			}
			return []Expression{}
		})
	}
	// return slice of elements, when not empty
	return func(args ...Expression) []Expression {
		if len(args) > 0 {
			return append(elems, args...)
		}
		return elems
	}
}

func (v VecVal) Head() Expression {
	if v.Len() > 0 {
		return v()[0]
	}
	return NewNone()
}
func (v VecVal) Last() Expression {
	if v.Len() > 0 {
		return v()[v.Len()-1]
	}
	return NewNone()
}
func (v VecVal) First() Expression { return v.Head() }
func (v VecVal) Tail() Continuation {
	if v.Len() > 1 {
		return NewVector(v()[1:]...)
	}
	return NewVector()
}
func (v VecVal) Prefix() VecVal {
	if v.Len() > 1 {
		return NewVector(v()[:v.Len()-1]...)
	}
	return NewVector()
}
func (v VecVal) Suffix() VecVal { return v.Tail().(VecVal) }
func (v VecVal) Continue() (Expression, Continuation) {
	return v.First(), v.Suffix()
}
func (v VecVal) Push(args ...Expression) Stack          { return NewVector(append(v(), args...)...) }
func (v VecVal) Put(args ...Expression) Stack           { return NewVector(append(args, v()...)...) }
func (v VecVal) Pop() (Expression, Stack)               { return v.Last(), v.Prefix() }
func (v VecVal) Pull() (Expression, Queue)              { return v.Head(), v.Suffix() }
func (v VecVal) Slice() []Expression                    { return v() }
func (v VecVal) Len() int                               { return len(v()) }
func (v VecVal) Null() VecVal                           { return NewVector() }
func (v VecVal) Type() TyComp                           { return Def(Vector, v.TypeElem()) }
func (v VecVal) TypeFnc() TyFnc                         { return Vector }
func (v VecVal) TypeElem() TyComp                       { return v.Head().Type() }
func (v VecVal) ConsVec(args ...Expression) VecVal      { return NewVector(v(args...)...) }
func (v VecVal) Cons(appendix ...Expression) Sequential { return v.ConsVec(appendix...) }
func (v VecVal) ConsContinue(appendix Continuation) Sequential {
	if v.Len() == 0 {
		return NewSequence().ConsContinue(appendix)
	}
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		var (
			head Expression
			tail Continuation
		)
		if len(args) > 0 {
			head, tail = NewSequence(
				append(v(), args...)...,
			).Continue()
		}
		return head, tail.(SeqVal).ConsContinue(appendix).(SeqVal)
	})
}
func (v VecVal) Call(args ...Expression) Expression {
	if len(args) > 0 {
		var head, tail = NewVector(v(args...)...).Continue()
		return NewPair(head, tail)
	}
	var head, tail = NewVector(v()...).Continue()
	return NewPair(head, tail)
}
func (v VecVal) AppendVec(appendix VecVal) VecVal    { return v.Append(appendix()...).(VecVal) }
func (v VecVal) Append(appendix ...Expression) Queue { return v.ConsVec(appendix...) }

func (v VecVal) Get(i int) (Expression, bool) {
	if i < v.Len() {
		return v()[i], true
	}
	return NewNone(), false
}

func (v VecVal) Reverse() VecVal {
	return NewVector(reverse(v())...)
}
func (v VecVal) Empty() bool {
	if v.Len() == 0 {
		return true
	}
	return false
}
func (v VecVal) String() string {
	if v.Empty() {
		return "[]"
	}
	var strs = []string{}
	for _, str := range v() {
		strs = append(strs, str.String())
	}
	return "[" + strings.Join(strs, ", ") + "]"
}
func (v VecVal) Clear() VecVal    { return NewVector(v()[:0]...) }
func (v VecVal) Sequence() SeqVal { return NewSequence(v()...) }
func (v VecVal) Sort(
	less func(a, b Expression) bool,
) Sequential {
	var s = newSorter(
		v(),
		func(slice []Expression, a, b int) bool {
			return less(slice[a], slice[b])
		},
	).Sort()
	return NewVector(s...)
}
func (v VecVal) Search(
	match Expression,
	compare func(a, b Expression) int,
) Expression {
	return newSearcher(v(), match, compare).Search()
}

type sorter struct {
	slice []Expression
	less  func(
		slice []Expression,
		a, b int,
	) bool
}

/// vector helper functions
func newSorter(
	s []Expression,
	l func(slice []Expression, a, b int) bool,
) *sorter {
	return &sorter{s, l}
}
func (s sorter) Slice() []Expression { return s.slice }
func (s sorter) Len() int            { return len(s.slice) }
func (s sorter) Less(a, b int) bool  { return s.less(s.slice, a, b) }
func (s *sorter) Swap(a, b int) {
	(*s).slice[b], (*s).slice[a] = (*s).slice[a], (*s).slice[b]
}
func (s *sorter) Sort() []Expression {
	sort.Sort(s)
	return s.slice
}

type searcher struct {
	slice   []Expression
	match   Expression
	compare func(a, b Expression) int
	lesser  func(slice []Expression, a, b int) bool
	search  func([]Expression) func(int) bool
}

func newSearcher(
	slice []Expression,
	match Expression,
	compare func(a, b Expression) int,
) *searcher {
	return &searcher{
		slice:   slice,
		match:   match,
		compare: compare,
		lesser: func(slice []Expression, a, b int) bool {
			return compare(slice[a], slice[b]) < 0
		},
		search: func(slice []Expression) func(int) bool {
			return func(idx int) bool {
				return compare(slice[idx], match) >= 0
			}
		},
	}
}
func (s *searcher) Search() Expression {
	var idx = sort.Search(len(s.slice),
		s.search(newSorter(
			s.slice, s.lesser,
		).Sort()))
	if idx >= 0 && idx < len(s.slice) {
		if s.compare(s.slice[idx], s.match) == 0 {
			return s.slice[idx]
		}
	}
	return NewNone()
}
func reverse(args []Expression) (rev []Expression) {
	if len(args) > 1 {
		var l = len(args)
		rev = make([]Expression, l, l)
		for i, arg := range args {
			rev[l-1-i] = arg
		}
		return rev
	}
	return args
}

///////////////////////////////////////////////////////////////////////////////
//// SEQUENCE TYPE
///
// generic sequential type
func NewSeqFromCon(con Continuation) SeqVal {
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		var head, tail = con.Continue()
		if len(args) > 0 {
			head = head.Call(args...)
		}
		return head, NewSeqFromCon(tail)
	})
}
func NewSeqFromSeq(seq Sequential) SeqVal {
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		if len(args) > 0 {
			seq = seq.Cons(args...)
		}
		var head, tail = seq.Continue()
		return head, NewSeqFromCon(tail)
	})
}
func NewSequence(elems ...Expression) SeqVal {

	// return empty list able to be extended by cons, when no initial
	// elements are given/left
	if len(elems) == 0 {
		return func(args ...Expression) (Expression, SeqVal) {
			if len(args) > 0 {
				if len(args) > 1 {
					return args[0], NewSequence(args[1:]...)
				}
				return args[0], NewSequence()
			}
			// return instance of none as head and a nil pointer as
			// tail, if neither elements nor arguments where passed
			return NewNone(), nil
		}
	}

	// at least one of the initial elements is left‥.
	return func(args ...Expression) (Expression, SeqVal) {

		// if arguments are passed, prepend those and return first
		// argument as head‥.
		if len(args) > 0 {
			// ‥.put arguments up front of preceeding elements
			if len(args) > 1 {
				return args[0],
					NewSequence(
						append(
							args,
							elems...,
						)...)
			}
			// use single argument as new head of sequence and
			// preceeding elements as tail
			return args[0],
				NewSequence(elems...)
		}

		// no arguments given, but more than one element left → return
		// first element as head, and remaining elements as tail of
		// sequence
		if len(elems) > 1 {
			return elems[0],
				NewSequence(elems[1:]...)
		}
		// return last element and empty sequence
		return elems[0], NewSequence()

	}
}
func (s SeqVal) Continue() (Expression, Continuation) {
	return s.Head(), s.Tail()
}
func (s SeqVal) Head() Expression {
	var cur, _ = s()
	return cur
}
func (s SeqVal) Tail() Continuation {
	var _, tail = s()
	return tail
}
func (s SeqVal) Empty() bool {
	if head, tail := s(); tail == nil && head.Type().Match(None) {
		return true
	}
	return false
}

func (s SeqVal) Cons(suffix ...Expression) Sequential {
	if len(suffix) == 0 {
		return s
	}
	if len(suffix) == 1 {
		return SeqVal(func(late ...Expression) (Expression, SeqVal) {
			if len(late) > 0 {
				if len(late) > 1 {
					return late[0],
						s.Cons(append(late[1:], suffix[0])...).(SeqVal)
				}
				return late[0], s.Cons(suffix[0]).(SeqVal)
			}
			return suffix[0], s
		})
	}
	return SeqVal(func(late ...Expression) (Expression, SeqVal) {
		if len(late) > 0 {
			if len(late) > 1 {
				return late[0],
					s.Cons(append(late[1:], suffix...)...).(SeqVal)
			}
			return late[0], s.Cons(suffix...).(SeqVal)
		}
		return suffix[0], s.Cons(suffix[1:]...).(SeqVal)
	})

}

func (s SeqVal) ConsContinue(suffix Continuation) Sequential {
	var head, tail = suffix.Continue()
	// if tail is empty‥.
	if tail.Empty() {
		// if head is none, return original s
		if head.Type().Match(None) {
			return s
		}
		// return a sequence starting with head yielded by prepended
		// seqval, followed by s as its tail
		return SeqVal(func(args ...Expression) (Expression, SeqVal) {
			if len(args) > 0 {
				if len(args) > 1 {
					head, tail = suffix.Call(args...,
					).(Continuation).Continue()
					return head, s.ConsContinue(tail).(SeqVal)
				}
			}
			return head, s
		})
	}
	// tail is not empty yet, return a sequence starting with yielded head
	// followed by remaining tail consed to s recursively
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		if len(args) > 0 {
			head, tail = suffix.Call(args...,
			).(Continuation).Continue()
			return head, s.ConsContinue(tail).(SeqVal)
		}
		return head, s.ConsContinue(tail).(SeqVal)
	})

}
func (s SeqVal) ConsSeq(seq SeqVal) Sequential { return s.ConsContinue(seq) }

func (s SeqVal) Append(appendix ...Expression) Queue {

	// return imediately, if no elements where given to append
	if len(appendix) == 0 {
		return s
	}

	// yield head & tail
	var head, tail = s()

	// if tail is empty
	if tail.Empty() {
		// if head matches none‥.
		if head.Type().Match(None) {
			// ‥.assign first element of appendix as head‥.
			head = appendix[0]
			// ‥.reassign remaining appendix, when more elements remain
			if len(appendix) > 1 {
				appendix = appendix[1:]
			} else { // ‥.or clear appendix
				appendix = appendix[:0]
			}
		}
		// return a sequence, that returns the appendix as its tail,
		// which might be empty, if its first element was assigned as
		// head, and there are no further appending alements
		return SeqVal(func(args ...Expression) (Expression, SeqVal) {
			if len(args) > 0 { // passed arguments are prepended
				if len(args) > 1 {
					return args[0],
						NewSequence(append(args[1:], appendix...)...)
				}
				return args[0], NewSequence(appendix...)
			}
			return head, NewSequence(appendix...)
		})
	}
	// length of appendix might be zero, since first element of appendix
	// might have been taken as head, so check‥.
	if len(appendix) > 0 {
		// return a sequence that uses the head that was yielded by
		// sequence, or the first element of the appendix and append
		// the appending elements to it's tail
		return SeqVal(func(args ...Expression) (Expression, SeqVal) {
			return head, tail.Append(appendix...).(SeqVal)
		})
	}
	// first element of appending elements was assigned to head, no further
	// appending elements are left, tail is most likely empty
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		return head, tail
	})

}
func (s SeqVal) AppendSeqVal(appendix SeqVal) SeqVal {
	var head, tail = s()
	if tail.Empty() {
		if head.Type().Match(None) {
			return appendix
		}
		return SeqVal(func(args ...Expression) (Expression, SeqVal) {
			if len(args) > 0 {
				head, tail = s(args...)
				return head, tail.AppendSeqVal(appendix)
			}
			return head, s
		})
	}
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		if len(args) > 0 {
			head, tail = s(args...)
			return head, tail.AppendSeqVal(appendix)
		}
		return head, tail.AppendSeqVal(appendix)
	})
}
func (s SeqVal) AppendSeq(appendix Sequential) Sequential {
	return s.AppendSeqVal(SeqVal(func(args ...Expression) (Expression, SeqVal) {
		if len(args) > 0 {
			appendix = appendix.Cons(args...)
		}
		var head, tail = appendix.Continue()
		return head, NewSeqFromSeq(tail.(Sequential))
	}))
}

func (s SeqVal) Pop() (Expression, Stack)      { return s() }
func (s SeqVal) Push(args ...Expression) Stack { return s.Cons(args...).(Stack) }
func (s SeqVal) Pull() (Expression, Queue) {
	var (
		acc        = []Expression{}
		head, tail = s()
	)
	for !tail.Empty() {
		acc = append(acc, head)
		head, tail = tail()
	}
	return head, NewVector(acc...)
}

// sort takes a compare function to return a negative value, when a is lesser
// then b, zero when they are equal and a positive value, when a is greater
// then b and generates a sequence of sorted elements implementing quick sort
// by lazy list comprehension (split elements lesser & greater pivot).
func (s SeqVal) Sort(compare func(a, b Expression) int) Sequential {
	// if list is empty, or has a single element, just return it
	if s.Empty() || s.Tail().Empty() {
		return s
	}
	var ( // use head as pivot & tail to filter lesser & greater from
		pivot, list = s()
		lesser      = NewTest(
			func(arg Expression) bool {
				if compare(pivot, arg) < 0 {
					return true
				}
				return false
			})
		greater = NewTest(
			func(arg Expression) bool {
				if compare(pivot, arg) >= 0 {
					return true
				}
				return false
			})
	)
	// lazy list comprehension quick sort
	return Pass(list, lesser).Sort(compare).
		Cons(pivot).ConsContinue(
		Pass(list, greater).Sort(compare))
}

func (s SeqVal) Null() SeqVal     { return NewSequence() }
func (s SeqVal) TypeElem() TyComp { return s.Head().Type() }
func (s SeqVal) TypeFnc() TyFnc   { return Sequence }
func (s SeqVal) Type() TyComp     { return Def(Sequence, s.TypeElem()) }
func (s SeqVal) Call(args ...Expression) Expression {
	if len(args) > 0 {
		var head, tail = s(args...)
		return NewPair(head, tail)
	}
	var head, tail = s()
	return NewPair(head, tail)
}
func (s SeqVal) Slice() []Expression {
	var (
		slice      []Expression
		head, tail = s()
	)
	for !head.Type().Match(None) && !tail.Empty() {
		slice = append(slice, head)
		head, tail = tail()
	}
	return slice
}
func (s SeqVal) Vector() VecVal { return NewVector(s.Slice()...) }

func (s SeqVal) String() string {
	if s.Empty() {
		return "()"
	}
	var (
		hstr, tstr string
		head, tail = s()
	)
	for !tail.Empty() {
		hstr = hstr + "(" + head.String() + " "
		tstr = tstr + ")"
		head, tail = tail()
	}
	hstr = hstr + "(" + head.String()
	tstr = tstr + ")"
	return hstr + tstr
}

//// GENERATOR
///
// expects an expression that returns an unboxed value, when called empty and
// some notion of 'next' value, relative to its arguments, if arguments where
// passed.
func NewGenerator(init, generate Expression) GenVal {
	return func() (Expression, GenVal) {
		var next = generate.Call(init)
		return init, NewGenerator(next, generate)
	}
}
func (g GenVal) Expr() Expression {
	var expr, _ = g()
	return expr
}
func (g GenVal) Generator() GenVal {
	var _, gen = g()
	return gen
}
func (g GenVal) Call(args ...Expression) Expression {
	if len(args) > 0 {
		return NewPair(g.Expr().Call(args...), g.Generator())
	}
	return NewPair(g.Expr(), g.Generator())
}
func (g GenVal) TypeFnc() TyFnc   { return Generator }
func (g GenVal) Type() TyComp     { return Def(Generator, g.Head().Type()) }
func (g GenVal) TypeElem() TyComp { return g.Head().Type() }
func (g GenVal) String() string   { return g.Head().String() }
func (g GenVal) Empty() bool {
	if g.Head().Type().Match(None) {
		return true
	}
	return false
}
func (g GenVal) Continue() (Expression, Continuation) { return g() }
func (g GenVal) Head() Expression                     { return g.Expr() }
func (g GenVal) Tail() Continuation                   { return g.Generator() }

//// ACCUMULATOR
///
// accumulator expects an expression as input, that returns itself unboxed,
// when called empty and returns a new accumulator accumulating its value and
// arguments to create a new accumulator, if arguments where passed.
func NewAccumulator(acc, fnc Expression) AccVal {
	return AccVal(func(args ...Expression) (Expression, AccVal) {
		if len(args) > 0 {
			acc = fnc.Call(append([]Expression{acc}, args...)...)
			return acc, NewAccumulator(acc, fnc)
		}
		return acc, NewAccumulator(acc, fnc)
	})
}

func (g AccVal) Result() Expression {
	var res, _ = g()
	return res
}
func (g AccVal) Accumulator() AccVal {
	var _, acc = g()
	return acc
}
func (g AccVal) Call(args ...Expression) Expression {
	if len(args) > 0 {
		var res, acc = g(args...)
		return NewPair(res, acc)
	}
	return g.Result()
}
func (g AccVal) TypeFnc() TyFnc { return Accumulator }
func (g AccVal) Type() TyComp {
	return Def(
		Accumulator,
		g.Head().Type().TypeRet(),
		g.Head().Type().TypeArgs(),
	)
}
func (g AccVal) String() string { return g.Head().String() }

func (a AccVal) Empty() bool {
	if a.Head().Type().Match(None) {
		return true
	}
	return false
}
func (g AccVal) Head() Expression                     { return g.Result() }
func (g AccVal) TypeElem() TyComp                     { return g.Head().Type() }
func (g AccVal) Tail() Continuation                   { return g.Accumulator() }
func (g AccVal) Continue() (Expression, Continuation) { return g() }
