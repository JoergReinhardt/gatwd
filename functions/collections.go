package functions

import (
	"sort"
	"strings"
)

type (

	//// COLLECTION TYPES
	VecVal  func(...Expression) []Expression
	ListVal func(...Expression) (Expression, ListVal)

	//// INTERNAL HELPER TYPES
	sorter struct {
		parms []Expression
		less  func(s []Expression, a, b int) bool
	}
	searcher struct {
		*sorter
		match   Expression
		compare func(a, b Expression) int
		search  func([]Expression) func(int) bool
	}
)

// SORTER
func newSorter(
	s []Expression,
	l func(s []Expression, a, b int) bool,
) *sorter {
	return &sorter{s, l}
}
func (s sorter) Slice() []Expression { return s.parms }
func (s sorter) Len() int            { return len(s.parms) }
func (s sorter) Less(a, b int) bool  { return s.less(s.parms, a, b) }
func (s *sorter) Swap(a, b int) {
	(*s).parms[b], (*s).parms[a] =
		(*s).parms[a], (*s).parms[b]
}
func (s *sorter) Sort() []Expression {
	sort.Sort(s)
	return s.parms
}

// SEARCHER
func newSearcher(
	s []Expression,
	match Expression,
	compare func(a, b Expression) int,
) *searcher {
	return &searcher{
		sorter: newSorter(s, func(s []Expression, a, b int) bool {
			return compare(s[a], s[b]) < 0
		}),
		match:   match,
		compare: compare,
		search: func(s []Expression) func(int) bool {
			return func(idx int) bool {
				return compare(s[idx], match) >= 0
			}
		}}
}
func (s *searcher) Index() int {

	if sort.IsSorted(s) {
		return sort.Search(len(s.parms),
			s.search(newSorter(
				s.parms, s.less,
			).parms))
	}

	return sort.Search(len(s.parms),
		s.search(newSorter(
			s.parms, s.less,
		).Sort()))
}

func (s *searcher) Search() Expression {

	var idx = s.Index()

	if idx >= 0 && idx < len(s.parms) {
		if s.compare(s.parms[idx], s.match) == 0 {
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
func NewVecFormGroup(grp Topological) VecVal {
	if grp.Type().Match(Vector) {
		if vec, ok := grp.(VecVal); ok {
			return vec
		}
		return NewVector(grp.(Vectorized).Slice()...)
	}
	var (
		vec        = []Expression{}
		head, tail = grp.Continue()
	)
	for head, tail = tail.Continue(); !tail.Empty(); {
		if !IsNone(head) {
			vec = append(vec, head)
		}
	}
	return NewVector(vec...)
}
func NewVector(elems ...Expression) VecVal {
	// return slice of elements, when not empty
	return func(args ...Expression) []Expression {
		if len(args) > 0 {
			return append(elems, args...)
		}
		return elems
	}
}
func (v VecVal) ConsVec(args ...Expression) VecVal {
	return NewVector(v(args...)...)
}
func (v VecVal) Cons(arg Expression) Topological {
	if IsNone(arg) {
		return v
	}
	return v.ConsVec(arg)
}
func (v VecVal) Head() Expression {
	if v.Len() > 0 {
		return v()[0]
	}
	return NewNone()
}
func (v VecVal) Tail() Topological {
	if v.Len() > 1 {
		return NewVector(v()[1:]...)
	}
	return NewVector()
}
func (v VecVal) Continue() (Expression, Topological) {
	return v.Head(), v.Tail()
}
func (v VecVal) Concat(grp Continuous) Topological {
	if grp.Empty() {
		return v
	}
	return NewList(v()...).Concat(grp)
}
func (v VecVal) Last() Expression {
	if v.Len() > 0 {
		return v()[v.Len()-1]
	}
	return NewNone()
}
func (v VecVal) First() Expression { return v.Head() }
func (v VecVal) Append(args ...Expression) Queued {
	return NewVector(append(v(), args...)...)
}
func (v VecVal) Push(arg Expression) Stacked {
	if !IsNone(arg) {
		return NewVector(append(v(), arg)...)
	}
	return v
}
func (v VecVal) Pop() (Expression, Stacked) {
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
func (v VecVal) Put(arg Expression) Queued {
	if !IsNone(arg) {
		return NewVector(append(v(), arg)...)
	}
	return v
}
func (v VecVal) Pull() (Expression, Queued) {
	if v.Len() > 1 {
		return v()[0], NewVector(v()[1:]...)
	}
	return v()[0], NewVector()
}
func (v VecVal) Len() int            { return len(v()) }
func (v VecVal) Null() VecVal        { return NewVector() }
func (v VecVal) Type() TyDef         { return Def(Vector, v.TypeElem()) }
func (v VecVal) TypeFnc() TyFnc      { return Vector }
func (v VecVal) TypeElem() TyDef     { return v.Head().Type() }
func (v VecVal) Slice() []Expression { return v() }
func (v VecVal) Flatten() VecVal {
	var elems = make([]Expression, 0, v.Len())
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
	return func() (Expression, GenVal) {
		return v.Head(), v.Tail().(VecVal).Generator()
	}
}
func (v VecVal) Accumulator() AccVal {
	return func(args ...Expression) (Expression, AccVal) {
		if len(args) > 0 {
			v = NewVector(v(args...)...)
		}
		return v.Head(), v.Tail().(VecVal).Accumulator()
	}
}
func (v VecVal) Call(args ...Expression) Expression {
	var head, tail = v.Continue()
	if len(args) > 0 {
		return NewPair(head.Call(args...), tail)
	}
	return NewPair(head, tail)
}

func (v VecVal) Get(i int) (Expression, bool) {
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
	less func(a, b Expression) bool,
) Topological {
	var s = newSorter(
		v(),
		func(s []Expression, a, b int) bool {
			return less(s[a], s[b])
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

///////////////////////////////////////////////////////////////////////////////
//// LINKED LIST TYPE
///
// linked list type implementing sequential
func NewListFromGroup(grp Topological) ListVal {
	return ListVal(func(args ...Expression) (Expression, ListVal) {
		if len(args) > 0 {
			if len(args) > 1 {
				var head, tail = grp.Concat(
					NewList(args...)).Continue()
				return head, NewListFromGroup(tail)
			}
			var head, tail = NewListFromGroup(
				grp.Cons(args[0])).Continue()
			return head, NewListFromGroup(tail)
		}
		var head, tail = grp.Continue()
		return head, NewListFromGroup(tail)
	})
}
func NewList(elems ...Expression) ListVal {
	// return empty list able to be extended by cons, when no initial
	// elements are given/left
	if len(elems) == 0 {
		return func(args ...Expression) (Expression, ListVal) {
			if len(args) > 0 {
				if len(args) > 1 {
					return args[0], NewList(args[1:]...)
				}
				return args[0], NewList()
			}
			// return instance of none as head and a nil pointer as
			// tail, if neither elements nor arguments where passed
			return NewNone(), nil
		}
	}

	// at least one of the initial elements is left‥.
	return func(args ...Expression) (Expression, ListVal) {
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
		return elems[0], NewList()

	}
}

func (s ListVal) Cons(arg Expression) Topological {
	if IsNone(arg) {
		return s
	}
	return ListVal(func(late ...Expression) (Expression, ListVal) {
		if len(late) > 0 {
			return s.Cons(arg).(ListVal)(late...)
		}
		return arg, s
	})
}

func (s ListVal) Concat(grp Continuous) Topological {
	if !s.Empty() {
		return ListVal(func(args ...Expression) (Expression, ListVal) {
			if len(args) > 0 {
				var head, tail = s(args...)
				return head, tail.Concat(grp).(ListVal)
			}
			var head, tail = s.Continue()
			return head, tail.Concat(grp).(ListVal)
		})
	}
	return grp.(Topological)
}

func (s ListVal) Head() Expression {
	var cur, _ = s()
	return cur
}
func (s ListVal) Tail() Topological {
	var _, tail = s()
	if tail != nil {
		return tail
	}
	return NewList()
}
func (s ListVal) Continue() (Expression, Topological) {
	return s.Head(), s.Tail()
}

func (s ListVal) Vector() VecVal              { return NewVector(s.Slice()...) }
func (v ListVal) Push(arg Expression) Stacked { return v.Cons(arg).(ListVal) }
func (v ListVal) Pop() (Expression, Stacked)  { return v() }
func (s ListVal) First() Expression           { return s.Head() }
func (s ListVal) Null() ListVal               { return NewList() }
func (s ListVal) TypeElem() TyDef             { return s.Head().Type() }
func (s ListVal) TypeFnc() TyFnc              { return Group }
func (s ListVal) Type() TyDef                 { return Def(Group, s.TypeElem()) }
func (s ListVal) Empty() bool {
	var _, tail = s()
	return tail == nil
}
func (s ListVal) Call(args ...Expression) Expression {
	var head, tail = s.Continue()
	if len(args) > 0 {
		return NewPair(head.Call(args...), tail)
	}
	return NewPair(head, tail)
}
func (s ListVal) Flatten() ListVal {
	return func(args ...Expression) (Expression, ListVal) {
		var (
			head Expression
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
func (s ListVal) Slice() []Expression {
	var (
		slice      []Expression
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
