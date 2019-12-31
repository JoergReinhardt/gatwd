package functions

import (
	"sort"
	"strings"
)

type (

	//// COLLECTIONS
	VecVal func(...Expression) []Expression
)

//////////////////////////////////////////////////////////////////////////////////////////
//// VECTORS (SLICES) OF VALUES
///
// sequential vector provides random access to sequential data. appends
// arguments in the order they where passed in, at the end of slice, when
// called.
func NewVecFormGroup(grp Group) VecVal {
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
		vec = append(vec, head)
	}
	return NewVector(vec...)
}
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
func (v VecVal) Continue() (Expression, Group) {
	return v.Head(), v.Tail()
}
func (v VecVal) Concat(grp Continuation) Group {
	if grp.Empty() {
		return v
	}
	return NewSequence(v()...).Concat(grp)
}
func (v VecVal) Suffix() Directional {
	if v.Len() > 1 {
		return NewVector(v()[1:]...)
	}
	return NewVector()
}
func (v VecVal) Prefix() BiDirectional {
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
func (v VecVal) Prepend(dir Group) Directional {
	var (
		slice      = []Expression{}
		head, tail = dir.Continue()
	)
	slice = append(slice, head)
	for head, tail = tail.Continue(); !tail.Empty(); {
		slice = append(slice, head)
	}
	return NewVector(append(slice, v()...)...)
}
func (v VecVal) PrependArgs(args ...Expression) Directional {
	return NewVector(append(args, v()...)...)
}
func (v VecVal) Append(apendix Group) Directional {
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
func (v VecVal) AppendVec(vec VecVal) VecVal {
	return NewVector(append(v(), vec()...)...)
}
func (v VecVal) AppendArgs(args ...Expression) Directional {
	return NewVector(append(v(), args...)...)
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
