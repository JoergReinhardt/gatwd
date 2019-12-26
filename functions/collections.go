package functions

import (
	"sort"
	"strings"
)

type (

	//// COLLECTIONS
	VecVal func(...Expression) []Expression
	SeqVal func(...Expression) (Expression, SeqVal)
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

func (v VecVal) Continue() (Expression, Continuation) { return v.Current(), v.Next() }
func (v VecVal) Current() Expression {
	if v.Len() > 0 {
		return v()[v.Len()-1]
	}
	return NewNone()
}
func (v VecVal) Next() Continuation {
	if v.Len() > 1 {
		return NewVector(v()[:v.Len()-1]...)
	}
	return NewVector()
}
func (v VecVal) Len() int                             { return len(v()) }
func (v VecVal) Null() VecVal                         { return NewVector() }
func (v VecVal) Type() TyComp                         { return Def(Vector, v.TypeElem()) }
func (v VecVal) TypeFnc() TyFnc                       { return Vector }
func (v VecVal) TypeElem() TyComp                     { return v.Current().Type() }
func (v VecVal) ConsVec(args ...Expression) VecVal    { return NewVector(v(args...)...) }
func (v VecVal) Cons(args ...Expression) Sequential   { return v.ConsVec(args...) }
func (v VecVal) Call(args ...Expression) Expression   { return v.ConsVec(args...) }
func (v VecVal) Append(args ...Expression) Sequential { return v.ConsVec(args...) }
func (v VecVal) AppendVec(vec VecVal) VecVal          { return v.Append(vec()...).(VecVal) }
func (v VecVal) Pull() (Expression, Sequential)       { return v.Current(), v.Next().(Sequential) }
func (v VecVal) Push(args ...Expression) Sequential   { return NewVector(append(args, v()...)...) }
func (v VecVal) Pop() (Expression, Sequential) {
	if v.Len() == 0 {
		return NewNone(), NewVector()
	}
	if v.Len() == 1 {
		return v()[0], NewVector()
	}
	return v()[0], NewVector(v()[1:]...)
}
func (v VecVal) Slice() []Expression { return v() }
func (v VecVal) First() Expression {
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

func (v VecVal) Get(i int) (Expression, bool) {
	if i < v.Len() {
		return v()[i], true
	}
	return NewNone(), false
}

func (v VecVal) Reverse() VecVal {
	return NewVector(reverse(v())...)
}
func (v VecVal) End() bool {
	if v.Len() == 0 {
		return true
	}
	return false
}
func (v VecVal) String() string {
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
	)
	return NewVector(s.Sort()...)
}
func (v VecVal) Search(
	less func(a, b Expression) bool,
	match func(Expression) bool,
) Expression {
	var s = newSearcher(
		newSorter(
			v(),
			func(slice []Expression, a, b int) bool {
				return less(slice[a], slice[b])
			},
		),
		func(i int, slice []Expression) bool {
			return match(slice[i])
		},
	)
	return (*s).Search()
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
	var sorted = *s
	sort.Sort(&sorted)
	return sorted.slice
}

type searcher struct {
	*sorter
	match func(int, []Expression) bool
}

func newSearcher(
	s *sorter,
	m func(int, []Expression) bool,
) *searcher {
	return &searcher{s, m}
}
func (s *searcher) Search() Expression {
	var so = *s
	(&so).Sort()
	if idx := sort.Search(
		s.Len(),
		func(i int) bool {
			return s.match(i, s.slice)
		}); idx >= 0 {
		return s.slice[idx]
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
func (s SeqVal) Continue() (Expression, Continuation) { return s() }
func (s SeqVal) Current() Expression {
	var cur, _ = s()
	return cur
}
func (s SeqVal) Next() Continuation {
	var _, tail = s()
	return tail
}
func (s SeqVal) End() bool {
	if head, tail := s(); tail == nil && head.Type().Match(None) {
		return true
	}
	return false
}

func (s SeqVal) Cons(args ...Expression) Sequential {
	if len(args) == 0 {
		return s
	}
	if len(args) == 1 {
		return SeqVal(func(late ...Expression) (Expression, SeqVal) {
			if len(late) > 0 {
				if len(late) > 1 {
					return late[0],
						s.Cons(append(late[1:], args[0])...).(SeqVal)
				}
				return late[0], s.Cons(args[0]).(SeqVal)
			}
			return args[0], s
		})
	}
	return SeqVal(func(late ...Expression) (Expression, SeqVal) {
		if len(late) > 0 {
			if len(late) > 1 {
				return late[0],
					s.Cons(append(late[1:], args...)...).(SeqVal)
			}
			return late[0], s.Cons(args...).(SeqVal)
		}
		return args[0], s.Cons(args[1:]...).(SeqVal)
	})

}
func (s SeqVal) ConSeqVal(prefix SeqVal) SeqVal {
	var head, tail = prefix()
	// if tail is empty‥.
	if tail.End() {
		// if head is none, return original s
		if head.Type().Match(None) {
			return s
		}
		// return a sequence starting with head yielded by prepended
		// seqval, followed by s as its tail
		return SeqVal(func(args ...Expression) (Expression, SeqVal) {
			if len(args) > 0 {
				if len(args) > 1 {
					head, tail = prefix(args...)
					return head, s.ConSeqVal(tail)
				}
			}
			return head, s
		})
	}
	// tail is not empty yet, return a sequence starting with yielded head
	// followed by remaining tail consed to s recursively
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		if len(args) > 0 {
			head, tail = prefix(args...)
			return head, s.ConSeqVal(tail)
		}
		return head, s.ConSeqVal(tail)
	})

}

func (s SeqVal) Append(appendix ...Expression) Sequential {

	// return imediately, if no elements where given to append
	if len(appendix) == 0 {
		return s
	}

	// yield head & tail
	var head, tail = s()

	// if tail is empty
	if tail.End() {
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
func (s SeqVal) AppendSeqVal(seq SeqVal) SeqVal {
	var head, tail = s()
	if tail.End() {
		if head.Type().Match(None) {
			return seq
		}
		return SeqVal(func(args ...Expression) (Expression, SeqVal) {
			if len(args) > 0 {
				head, tail = s(args...)
				return head, tail.AppendSeqVal(seq)
			}
			return head, s
		})
	}
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		if len(args) > 0 {
			head, tail = s(args...)
			return head, tail.AppendSeqVal(seq)
		}
		return head, tail.AppendSeqVal(seq)
	})
}

func (s SeqVal) Pop() (Expression, Sequential)      { return s() }
func (s SeqVal) Push(args ...Expression) Sequential { return s.Cons(args...) }
func (s SeqVal) Pull() (Expression, Sequential) {
	var (
		acc        = []Expression{}
		head, tail = s()
	)
	for !tail.End() {
		acc = append(acc, head)
		head, tail = tail()
	}
	return head, NewVector(acc...)
}

func (s SeqVal) Null() SeqVal     { return NewSequence() }
func (s SeqVal) TypeElem() TyComp { return s.Current().Type() }
func (s SeqVal) TypeFnc() TyFnc   { return Sequence }
func (s SeqVal) Type() TyComp     { return Def(Sequence, s.TypeElem()) }
func (s SeqVal) Call(args ...Expression) Expression {
	if len(args) > 0 {
		return NewPair(s.Cons(args...).Continue())
	}
	return NewPair(s.Continue())
}

func (s SeqVal) String() string {
	var (
		hstr, tstr string
		head, tail = s()
	)
	for !tail.End() {
		hstr = hstr + "( " + head.String() + " "
		tstr = tstr + ")"
		head, tail = tail()
	}
	hstr = hstr + "( " + head.String() + " "
	tstr = tstr + ")"
	return hstr + tstr
}
