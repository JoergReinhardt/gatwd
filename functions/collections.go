package functions

import (
	"sort"
	"strings"
)

type (

	//// COLLECTIONS
	VecVal func(...Expression) []Expression
	SeqVal func(...Expression) (Expression, Group)
)

//////////////////////////////////////////////////////////////////////////////////////////
//// VECTORS (SLICES) OF VALUES
///
// sequential vector provides random access to sequential data. appends
// arguments in the order they where passed in, at the end of slice, when
// called.
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
func (v VecVal) Tail() Group {
	if v.Len() > 1 {
		return NewVector(v()[1:]...)
	}
	return NewVector()
}
func (v VecVal) Suffix() Group {
	if v.Len() > 1 {
		return NewVector(v()[1:]...)
	}
	return NewVector()
}
func (v VecVal) Prefix() Group {
	if v.Len() > 1 {
		return NewVector(v()[:v.Len()-1]...)
	}
	return NewVector()
}
func (v VecVal) Last() Expression {
	if v.Len() > 0 {
		return v()[v.Len()-1]
	}
	return NewNone()
}
func (v VecVal) First() Expression                 { return v.Head() }
func (v VecVal) Continue() (Expression, Group)     { return v.Head(), v.Tail() }
func (v VecVal) Pop() (Expression, Stack)          { return v.Head(), v.Tail().(Stack) }
func (v VecVal) Push(args ...Expression) Stack     { return NewVector(append(v(), args...)...) }
func (v VecVal) Pull() (Expression, Queue)         { return v.First(), v.Suffix().(Queue) }
func (v VecVal) Put(args ...Expression) Queue      { return NewVector(append(v(), args...)...) }
func (v VecVal) Slice() []Expression               { return v() }
func (v VecVal) Len() int                          { return len(v()) }
func (v VecVal) Null() VecVal                      { return NewVector() }
func (v VecVal) Type() TyComp                      { return Def(Vector, v.TypeElem()) }
func (v VecVal) TypeFnc() TyFnc                    { return Vector }
func (v VecVal) TypeElem() TyComp                  { return v.Head().Type() }
func (v VecVal) ConsVec(args ...Expression) VecVal { return NewVector(v(args...)...) }
func (v VecVal) Cons(appendix ...Expression) Group { return v.ConsVec(appendix...) }
func (v VecVal) ConsGroup(appendix Group) Group {
	if v.Len() == 0 {
		return NewSequence().ConsGroup(appendix)
	}
	return SeqVal(func(args ...Expression) (Expression, Group) {
		var (
			head Expression
			tail Continuation
		)
		if len(args) > 0 {
			head, tail = NewSequence(
				append(v(), args...)...,
			).Continue()
		}
		return head, tail.(SeqVal).ConsGroup(appendix).(SeqVal)
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
func (v VecVal) Append(apendix Group) VecVal {
	if apendix.Empty() {
		return v
	}
	var (
		slice      = []Expression{}
		head, tail = apendix.Continue()
	)
	for head, tail = tail.Continue(); !tail.Empty(); {
		slice = append(slice, head)
	}
	slice = append(v(), slice...)
	return VecVal(func(args ...Expression) []Expression {
		if len(args) > 0 {
			return append(slice, args...)
		}
		return slice
	})
}
func (v VecVal) AppendVec(vec VecVal) VecVal { return v.AppendArgs(vec()...) }
func (v VecVal) AppendArgs(args ...Expression) VecVal {
	return NewVector(append(v(), args...)...)
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
) Group {
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
	return SeqVal(func(args ...Expression) (Expression, Group) {
		var head, tail = con.Continue()
		if len(args) > 0 {
			head = head.Call(args...)
		}
		return head, NewSeqFromCon(tail)
	})
}
func NewSeqFromSeq(seq Group) SeqVal {
	return SeqVal(func(args ...Expression) (Expression, Group) {
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
		return func(args ...Expression) (Expression, Group) {
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
	return func(args ...Expression) (Expression, Group) {

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
func (s SeqVal) Head() Expression {
	var cur, _ = s()
	return cur
}
func (s SeqVal) Tail() Group {
	var _, tail = s()
	return tail
}
func (s SeqVal) Continue() (Expression, Group) {
	return s.Head(), s.Tail()
}
func (s SeqVal) Cons(suffix ...Expression) Group {
	if len(suffix) == 0 {
		return s
	}
	if len(suffix) == 1 {
		return SeqVal(func(late ...Expression) (Expression, Group) {
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
	return SeqVal(func(late ...Expression) (Expression, Group) {
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

func (s SeqVal) ConsGroup(suffix Group) Group {
	var head, tail = suffix.Continue()
	// if tail is empty‥.
	if tail.Empty() {
		// if head is none, return original s
		if head.Type().Match(None) {
			return s
		}
		// return a sequence starting with head yielded by prepended
		// seqval, followed by s as its tail
		return SeqVal(func(args ...Expression) (Expression, Group) {
			if len(args) > 0 {
				if len(args) > 1 {
					head, tail = suffix.Call(args...,
					).(Continuation).Continue()
					return head, s.ConsGroup(tail).(SeqVal)
				}
			}
			return head, s
		})
	}
	// tail is not empty yet, return a sequence starting with yielded head
	// followed by remaining tail consed to s recursively
	return SeqVal(func(args ...Expression) (Expression, Group) {
		if len(args) > 0 {
			head, tail = suffix.Call(args...,
			).(Continuation).Continue()
			return head, s.ConsGroup(tail).(SeqVal)
		}
		return head, s.ConsGroup(tail).(SeqVal)
	})

}
func (s SeqVal) ConsSeqVal(seq SeqVal) Group { return s.ConsGroup(seq).(Group) }

func (s SeqVal) Pop() (Expression, Stack)      { return s.Head(), s.Tail().(SeqVal) }
func (s SeqVal) Push(args ...Expression) Stack { return s.Cons(args...).(Stack) }
func (s SeqVal) First() Expression             { return s.Head() }
func (s SeqVal) Suffix() Expression            { return s.Tail() }

func (s SeqVal) AppendSeq(appendix SeqVal) SeqVal {
	var head, tail = s()
	if tail.Empty() {
		if head.Type().Match(None) {
			return appendix
		}
		return SeqVal(func(args ...Expression) (Expression, Group) {
			if len(args) > 0 {
				head, tail = s(args...)
				return head, tail.(SeqVal).AppendSeq(appendix)
			}
			return head, appendix
		})
	}
	return SeqVal(func(args ...Expression) (Expression, Group) {
		if len(args) > 0 {
			head, tail = s(args...)
			return head, tail.(SeqVal).AppendSeq(appendix)
		}
		return head, tail.(SeqVal).AppendSeq(appendix)
	})
}
func (s SeqVal) AppendArgs(args ...Expression) Group {
	return s.Append(NewSequence(args...))
}
func (s SeqVal) Append(appendix Group) Group {
	return s.AppendSeq(SeqVal(func(args ...Expression) (Expression, Group) {
		if len(args) > 0 {
			appendix = appendix.Cons(args...)
		}
		var head, tail = appendix.Continue()
		return head, NewSeqFromSeq(tail.(Group))
	}))
}

func (s SeqVal) Empty() bool {
	if head, tail := s(); tail == nil && head.Type().Match(None) {
		return true
	}
	return false
}

// sort takes a compare function to return a negative value, when a is lesser
// then b, zero when they are equal and a positive value, when a is greater
// then b and generates a sequence of sorted elements implementing quick sort
// by lazy list comprehension (split elements lesser & greater pivot).
func (s SeqVal) Sort(compare func(a, b Expression) int) Group {
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
	return Pass(list, lesser).(SeqVal).Sort(compare).
		Cons(pivot).ConsGroup(
		Pass(list, greater).(SeqVal).Sort(compare)).(Group)
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
		head, tail = tail.(SeqVal)()
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
		head, tail = tail.(SeqVal)()
	}
	hstr = hstr + "(" + head.String()
	tstr = tstr + ")"
	return hstr + tstr
}
