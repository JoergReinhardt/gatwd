package functions

import (
	"sort"
	"strings"
)

type (
	//// INTERNAL HELPER TYPES
	sorter struct {
		parms []Functor
		less  func(s []Functor, a, b int) bool
	}
	searcher struct {
		*sorter
		compare func(a, b Functor) int
		search  func([]Functor, Functor) func(int) bool
	}
	//// COLLECTION TYPES
	VecVal  func(...Functor) []Functor
	ListVal func(...Functor) (Functor, ListVal)
)

// SORTER
func newSorter(
	s []Functor,
	l func(s []Functor, a, b int) bool,
) *sorter {
	return &sorter{s, l}
}
func (s sorter) Slice() []Functor   { return s.parms }
func (s sorter) Len() int           { return len(s.parms) }
func (s sorter) Less(a, b int) bool { return s.less(s.parms, a, b) }
func (s *sorter) Swap(a, b int) {
	(*s).parms[b], (*s).parms[a] =
		(*s).parms[a], (*s).parms[b]
}
func (s *sorter) Sort() []Functor {
	sort.Sort(s)
	return s.parms
}

// SEARCHER
func newSearcher(
	s []Functor,
	compare func(a, b Functor) int,
) *searcher {
	return &searcher{
		sorter: newSorter(s, func(s []Functor, a, b int) bool {
			return compare(s[a], s[b]) < 0
		}),
		compare: compare,
		search: func(s []Functor, match Functor) func(int) bool {
			return func(idx int) bool {
				return compare(s[idx], match) >= 0
			}
		}}
}
func (s *searcher) Index(match Functor) int {

	if sort.IsSorted(s) {
		return sort.Search(len(s.parms),
			s.search(newSorter(
				s.parms, s.less,
			).parms, match))
	}

	return sort.Search(len(s.parms),
		s.search(newSorter(
			s.parms, s.less,
		).Sort(), match))
}

func (s *searcher) Search(match Functor) Functor {

	var idx = s.Index(match)

	if idx >= 0 && idx < len(s.parms) {
		if s.compare(s.parms[idx], match) == 0 {
			return s.parms[idx]
		}
	}

	return NewNone()
}

//////////////////////////////////////////////////////////////////////////////////////////
//// VECTORS (SLICES) OF VALUES
///
// sequential vector provides random access to sequential data. appends
// arguments in the order they where passed in, at the end of slice, when
// called.
func NewVecFromApp(grp Applicative) VecVal {
	if grp.Type().Match(Vector) {
		if vec, ok := grp.(VecVal); ok {
			return vec
		}
		return NewVector(grp.(RandomAcc).Slice()...)
	}
	var (
		vec        = []Functor{}
		head, tail = grp.Continue()
	)
	for head, tail = tail.Continue(); !tail.Empty(); {
		if !IsNone(head) {
			vec = append(vec, head)
		}
	}
	return NewVector(vec...)
}
func NewVector(elems ...Functor) VecVal {
	// return slice of elements, when not empty
	return func(args ...Functor) []Functor {
		if len(args) > 0 {
			return append(elems, args...)
		}
		return elems
	}
}
func (v VecVal) ConsVec(args ...Functor) VecVal {
	return NewVector(v(args...)...)
}
func (v VecVal) Cons(arg Functor) Applicative {
	if IsNone(arg) {
		return v
	}
	return v.ConsVec(arg)
}
func (v VecVal) Head() Functor {
	if v.Len() > 0 {
		return v()[0]
	}
	return NewNone()
}
func (v VecVal) Tail() Applicative {
	if v.Len() > 1 {
		return NewVector(v()[1:]...)
	}
	return NewVector()
}
func (v VecVal) Continue() (Functor, Applicative) {
	return v.Head(), v.Tail()
}
func (v VecVal) Concat(grp Sequential) Applicative {
	if grp.Empty() {
		return v
	}
	return NewList(v()...).Concat(grp)
}
func (v VecVal) Last() Functor {
	if v.Len() > 0 {
		return v()[v.Len()-1]
	}
	return NewNone()
}
func (v VecVal) First() Functor { return v.Head() }
func (v VecVal) Append(args ...Functor) Queued {
	return NewVector(append(v(), args...)...)
}
func (v VecVal) Push(arg Functor) Stacked {
	if !IsNone(arg) {
		return NewVector(append(v(), arg)...)
	}
	return v
}
func (v VecVal) Pop() (Functor, Stacked) {
	var (
		head = v.Last()
		tail Stacked
	)
	if v.Len() > 1 {
		tail = NewVector(v()[:v.Len()-1]...)
	} else {
		tail = NewVector()
	}
	return head, tail
}
func (v VecVal) Put(arg Functor) Queued {
	if !IsNone(arg) {
		return NewVector(append(v(), arg)...)
	}
	return v
}
func (v VecVal) Pull() (Functor, Queued) {
	if v.Len() > 1 {
		return v()[0], NewVector(v()[1:]...)
	}
	return v()[0], NewVector()
}
func (v VecVal) Len() int         { return len(v()) }
func (v VecVal) Null() VecVal     { return NewVector() }
func (v VecVal) Type() Decl       { return Declare(Vector, v.TypeElem()) }
func (v VecVal) TypeFnc() TyFnc   { return Vector }
func (v VecVal) TypeElem() Decl   { return v.Head().Type() }
func (v VecVal) Slice() []Functor { return v() }
func (v VecVal) Flatten() VecVal {
	var elems = make([]Functor, 0, v.Len())
	for _, elem := range v() {
		if IsVect(elem) {
			if vec, ok := elem.(VecVal); ok {
				elems = append(elems, vec.Flatten()()...)
			}
		}
		elems = append(elems, elem)
	}
	return NewVector(elems...)
}
func (v VecVal) Generator() GenVal {
	return func() (Functor, GenVal) {
		return v.Head(), v.Tail().(VecVal).Generator()
	}
}
func (v VecVal) Accumulator() AccVal {
	return func(args ...Functor) (Functor, AccVal) {
		if len(args) > 0 {
			v = NewVector(v(args...)...)
		}
		return v.Head(), v.Tail().(VecVal).Accumulator()
	}
}
func (v VecVal) Call(args ...Functor) Functor {
	var head, tail = v.Continue()
	if len(args) > 0 {
		return NewPair(head.Call(args...), tail)
	}
	return NewPair(head, tail)
}

func (v VecVal) Get(i int) (Functor, bool) {
	if i < v.Len() {
		return v()[i], true
	}
	return NewNone(), false
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
func (v VecVal) Clear() VecVal     { return NewVector(v()[:0]...) }
func (v VecVal) Sequence() ListVal { return NewList(v()...) }
func (v VecVal) Sort(
	less func(a, b Functor) bool,
) Applicative {
	var s = newSorter(
		v(),
		func(s []Functor, a, b int) bool {
			return less(s[a], s[b])
		},
	).Sort()
	return NewVector(s...)
}
func (v VecVal) Search(
	match Functor,
	compare func(a, b Functor) int,
) Functor {
	return newSearcher(v(), compare).Search(match)
}
func (v VecVal) SearchIdx(
	match Functor,
	compare func(a, b Functor) int,
) int {
	return newSearcher(v(), compare).Index(match)
}

///////////////////////////////////////////////////////////////////////////////
//// LINKED LIST TYPE
///
// linked list type implementing sequential
// wrap arbitrary applicative in a list
func NewListFromApp(grp Applicative) ListVal {
	return ListVal(func(args ...Functor) (Functor, ListVal) {
		if len(args) > 0 {
			if len(args) > 1 {
				var head, tail = grp.Concat(
					NewList(args...)).Continue()
				return head, NewListFromApp(tail)
			}
			var head, tail = NewListFromApp(
				grp.Cons(args[0])).Continue()
			return head, NewListFromApp(tail)
		}
		var head, tail = grp.Continue()
		return head, NewListFromApp(tail)
	})
}

// new list with, or without content
func NewList(elems ...Functor) ListVal {

	// return empty list able to be extended by cons, when no initial
	// elements are given/left
	if len(elems) == 0 {
		return func(args ...Functor) (Functor, ListVal) {
			if len(args) > 0 {
				if len(args) > 1 {
					return args[0], NewList(args[1:]...)
				}
				return args[0], nil
			}
			// return instance of none as head and a nil pointer as
			// tail, if neither elements nor arguments where passed
			return NewNone(), nil
		}
	}

	// at least one of the initial elements is left‥.
	return func(args ...Functor) (Functor, ListVal) {
		// if arguments are passed, prepend those and return first
		// argument as head‥.
		if len(args) > 0 {
			// ‥.put arguments up front of preceeding elements
			if len(args) > 1 {
				return args[0], NewList(
					append(args, elems...)...)
			}
			// use single argument as new head of sequence and
			// preceeding elements as tail
			return args[0], NewList(elems...)
		}

		// no arguments given, but more than one element left → return
		// first element as head, and remaining elements as tail of
		// sequence
		if len(elems) > 1 {
			return elems[0], NewList(elems[1:]...)
		}
		// return last element and empty sequence
		return elems[0], nil

	}
}

func (s ListVal) Cons(arg Functor) Applicative {
	if IsNone(arg) {
		return s
	}
	return ListVal(func(late ...Functor) (Functor, ListVal) {
		if len(late) > 0 {
			return s.Cons(arg).(ListVal)(late...)
		}
		return arg, s
	})
}

func (s ListVal) Concat(grp Sequential) Applicative {
	if !s.Empty() {
		return ListVal(func(args ...Functor) (Functor, ListVal) {
			if len(args) > 0 {
				var head, tail = s(args...)
				return head, tail.Concat(grp).(ListVal)
			}
			var head, tail = s.Continue()
			return head, tail.Concat(grp).(ListVal)
		})
	}
	return grp.(Applicative)
}

func (s ListVal) Head() Functor {
	var cur, _ = s()
	return cur
}
func (s ListVal) Tail() Applicative {
	var _, tail = s()
	if tail != nil {
		return tail
	}
	return NewList()
}
func (s ListVal) Continue() (Functor, Applicative) {
	return s.Head(), s.Tail()
}

func (v ListVal) Push(arg Functor) Stacked { return v.Cons(arg).(ListVal) }
func (v ListVal) Pop() (Functor, Stacked)  { return v() }
func (s ListVal) First() Functor           { return s.Head() }
func (s ListVal) Null() ListVal            { return NewList() }
func (s ListVal) TypeElem() Decl           { return s.Head().Type() }
func (s ListVal) TypeFnc() TyFnc           { return Group }
func (s ListVal) Type() Decl               { return Declare(Group, s.TypeElem()) }
func (s ListVal) Vector() VecVal           { return NewVector(s.Slice()...) }
func (s ListVal) Empty() bool {
	var _, tail = s()
	return tail == nil
}
func (s ListVal) Call(args ...Functor) Functor {
	var head, tail = s.Continue()
	if len(args) > 0 {
		return NewPair(head.Call(args...), tail)
	}
	return NewPair(head, tail)
}
func (s ListVal) Flatten() ListVal {
	return func(args ...Functor) (Functor, ListVal) {
		var (
			head Functor
			tail ListVal
		)
		if len(args) > 0 {
			head, tail = s(args...)
		} else {
			head, tail = s()
		}
		if IsList(head) {
			if list, ok := head.(ListVal); ok {
				return list.Flatten().Concat(
					tail.Flatten()).(ListVal)()
			}
		}
		return head, tail.Flatten()
	}
}
func (s ListVal) Slice() []Functor {
	var (
		slice      []Functor
		head, tail = s()
	)
	for !head.Type().Match(None) && !tail.Empty() {
		slice = append(slice, head)
		head, tail = tail()
	}
	slice = append(slice, head)
	return slice
}

func (s ListVal) String() string {
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
