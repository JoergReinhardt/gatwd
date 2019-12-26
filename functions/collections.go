package functions

import (
	"sort"
	"strings"

	d "github.com/joergreinhardt/gatwd/data"
)

type (
	//// VALUE PAIRS
	ValPair   func(...Expression) (Expression, Expression)
	NatPair   func(...Expression) (d.Native, Expression)
	KeyPair   func(...Expression) (string, Expression)
	IndexPair func(...Expression) (int, Expression)
	RealPair  func(...Expression) (float64, Expression)

	//// COLLECTIONS
	KeyMap   func(...Expression) map[string]Expression
	IndexMap func(...Expression) map[int]Expression
	RealMap  func(...Expression) map[float64]Expression
	VecVal   func(...Expression) []Expression
	SeqVal   func(...Expression) (Expression, SeqVal)
)

///////////////////////////////////////////////////////////////////////////////
//// PAIRS OF VALUES
///
// pairs can be created empty, key & value may be constructed later
func NewEmptyPair() ValPair {
	return func(args ...Expression) (a, b Expression) {
		if len(args) > 0 {
			if len(args) > 1 {
				return args[0], args[1]
			}
			return args[0], NewNone()
		}
		return NewNone(), NewNone()
	}
}

// new pair from two callable instances
func NewPair(l, r Expression) ValPair {
	return func(args ...Expression) (Expression, Expression) {
		if len(args) > 0 {
			if len(args) > 1 {
				return args[0], args[1]
			}
			return args[0], r
		}
		return l, r
	}
}
func (p ValPair) Pair() Paired                   { return p }
func (p ValPair) Both() (Expression, Expression) { return p() }
func (p ValPair) Swap() (Expression, Expression) { l, r := p(); return r, l }
func (p ValPair) Left() Expression               { l, _ := p(); return l }
func (p ValPair) Right() Expression              { _, r := p(); return r }
func (p ValPair) SwappedPair() Paired            { return NewPair(p.Right(), p.Left()) }
func (p ValPair) Slice() []Expression            { return []Expression{p.Left(), p.Right()} }
func (p ValPair) Key() Expression                { return p.Left() }
func (p ValPair) Value() Expression              { return p.Right() }
func (p ValPair) TypeFnc() TyFnc                 { return Pair }
func (p ValPair) TypeElem() TyComp {
	if p.Right() != nil {
		return p.Right().Type()
	}
	return Def(None, Pair, None)
}
func (p ValPair) TypeKey() d.Typed {
	if p.Left() != nil {
		return p.Left().Type()
	}
	return None
}
func (p ValPair) TypeValue() d.Typed {
	if p.Right() != nil {
		return p.Right().Type()
	}
	return None
}
func (p ValPair) Type() TyComp {
	return Def(Pair, Def(p.TypeKey(), p.TypeValue()))
}
func (p ValPair) End() bool { return p.Empty() }
func (p ValPair) Empty() bool {
	if p.Left() == nil || (!p.Left().Type().Match(None) &&
		(p.Right() == nil || (!p.Right().Type().Match(None)))) {
		return true
	}
	return false
}
func (p ValPair) String() string {
	return "(" + p.Left().String() + ", " + p.Right().String() + ")"
}
func (p ValPair) Call(args ...Expression) Expression {
	return NewPair(p.Key(), p.Value().Call(args...))
}
func (p ValPair) Current() Expression                  { return p.Left() }
func (p ValPair) Next() Continuation                   { return NewPair(p.Right(), NewNone()) }
func (p ValPair) Continue() (Expression, Continuation) { return p.Current(), p.Next() }

//// NATIVE VALUE KEY PAIR
///
//
func NewNatPair(key d.Native, val Expression) NatPair {
	return func(...Expression) (d.Native, Expression) { return key, val }
}

func (a NatPair) KeyNat() d.Native                   { key, _ := a(); return key }
func (a NatPair) Value() Expression                  { _, val := a(); return val }
func (a NatPair) Left() Expression                   { return Box(a.KeyNat()) }
func (a NatPair) Right() Expression                  { return a.Value() }
func (a NatPair) Both() (Expression, Expression)     { return a.Left(), a.Right() }
func (a NatPair) Pair() Paired                       { return NewPair(a.Both()) }
func (a NatPair) Pairs() []Paired                    { return []Paired{NewPair(a.Both())} }
func (a NatPair) Key() Expression                    { return a.Left() }
func (a NatPair) Call(args ...Expression) Expression { return a.Value().Call(args...) }
func (a NatPair) TypeValue() d.Typed                 { return a.Value().Type() }
func (a NatPair) TypeKey() d.Typed                   { return a.KeyNat().Type() }
func (a NatPair) TypeFnc() TyFnc                     { return Data | Pair }
func (p NatPair) Type() TyComp                       { return Def(Pair, Def(Key, p.TypeValue())) }

// implement swappable
func (p NatPair) Swap() (Expression, Expression) {
	l, r := p()
	return Box(l), r
}
func (p NatPair) SwappedPair() Paired { return NewPair(p.Right(), p.Left()) }

func (a NatPair) End() bool { return a.Empty() }
func (a NatPair) Empty() bool {
	if a.Key() != nil && a.Value() != nil && a.Value().TypeFnc() != None {
		return false
	}
	return true
}
func (a NatPair) String() string {
	return "(" + a.Right().String() + " : " + a.Left().String() + ")"
}
func (p NatPair) Current() Expression                  { return p.Left() }
func (p NatPair) Next() Continuation                   { return NewPair(p.Right(), NewNone()) }
func (p NatPair) Continue() (Expression, Continuation) { return p.Current(), p.Next() }

//// STRING KEY PAIR
///
// pair composed of a string key and a functional value
func NewKeyPair(key string, val Expression) KeyPair {
	return func(...Expression) (string, Expression) { return key, val }
}

func (a KeyPair) KeyStr() string                     { key, _ := a(); return key }
func (a KeyPair) Value() Expression                  { _, val := a(); return val }
func (a KeyPair) Left() Expression                   { return Box(d.StrVal(a.KeyStr())) }
func (a KeyPair) Right() Expression                  { return a.Value() }
func (a KeyPair) Both() (Expression, Expression)     { return a.Left(), a.Right() }
func (a KeyPair) Pair() Paired                       { return NewPair(a.Both()) }
func (a KeyPair) Pairs() []Paired                    { return []Paired{NewPair(a.Both())} }
func (a KeyPair) Key() Expression                    { return a.Left() }
func (a KeyPair) Call(args ...Expression) Expression { return a.Value().Call(args...) }
func (a KeyPair) TypeValue() d.Typed                 { return a.Value().Type() }
func (a KeyPair) TypeElem() d.Typed                  { return a.Value().Type() }
func (a KeyPair) TypeKey() d.Typed                   { return Key }
func (a KeyPair) TypeFnc() TyFnc                     { return Key | Pair }
func (p KeyPair) Type() TyComp {
	return Def(Key|Pair, Def(p.TypeKey(), p.TypeValue()))
}

// implement swappable
func (p KeyPair) Swap() (Expression, Expression) {
	l, r := p()
	return Box(d.StrVal(l)), r
}
func (p KeyPair) SwappedPair() Paired { return NewPair(p.Right(), p.Left()) }

func (a KeyPair) End() bool { return a.Empty() }
func (a KeyPair) Empty() bool {
	if a.Key() != nil && a.Value() != nil && a.Value().TypeFnc() != None {
		return false
	}
	return true
}
func (a KeyPair) String() string {
	return "(" + a.Right().String() + " : " + a.Left().String() + ")"
}
func (p KeyPair) Current() Expression                  { return p.Value() }
func (p KeyPair) Next() Continuation                   { return NewPair(p.Right(), NewNone()) }
func (p KeyPair) Continue() (Expression, Continuation) { return p.Current(), p.Next() }

//// INDEX PAIR
///
// pair composed of an integer and a functional value
func NewIndexPair(idx int, val Expression) IndexPair {
	return func(...Expression) (int, Expression) { return idx, val }
}
func (a IndexPair) Index() int                         { idx, _ := a(); return idx }
func (a IndexPair) Value() Expression                  { _, val := a(); return val }
func (a IndexPair) Left() Expression                   { return Box(d.IntVal(a.Index())) }
func (a IndexPair) Right() Expression                  { return a.Value() }
func (a IndexPair) Both() (Expression, Expression)     { return a.Left(), a.Right() }
func (a IndexPair) Pair() Paired                       { return a }
func (a IndexPair) Pairs() []Paired                    { return []Paired{NewPair(a.Both())} }
func (a IndexPair) Key() Expression                    { return a.Left() }
func (a IndexPair) Call(args ...Expression) Expression { return a.Value().Call(args...) }
func (a IndexPair) TypeFnc() TyFnc                     { return Index | Pair }
func (a IndexPair) TypeKey() d.Typed                   { return Index }
func (a IndexPair) TypeValue() d.Typed                 { return a.Value().Type() }
func (a IndexPair) Type() TyComp                       { return Def(Pair, Def(Index, a.TypeValue())) }

// implement swappable
func (p IndexPair) Swap() (Expression, Expression) {
	l, r := p()
	return Box(d.New(l)), r
}
func (p IndexPair) SwappedPair() Paired { return NewPair(p.Right(), p.Left()) }
func (a IndexPair) End() bool           { return a.Empty() }
func (a IndexPair) Empty() bool {
	if a.Index() >= 0 && a.Value() != nil && a.Value().TypeFnc() != None {
		return true
	}
	return false
}
func (a IndexPair) String() string {
	return "(" + a.Right().String() + " : " + a.Left().String() + ")"
}
func (p IndexPair) Current() Expression                  { return p.Left() }
func (p IndexPair) Next() Continuation                   { return NewPair(p.Right(), NewNone()) }
func (p IndexPair) Continue() (Expression, Continuation) { return p.Current(), p.Next() }

//// FLOATING PAIR
///
// pair composed of an integer and a functional value
func NewRealPair(flt float64, val Expression) RealPair {
	return func(...Expression) (float64, Expression) { return flt, val }
}
func (a RealPair) Real() float64                      { flt, _ := a(); return flt }
func (a RealPair) Value() Expression                  { _, val := a(); return val }
func (a RealPair) Left() Expression                   { return Box(d.IntVal(a.Real())) }
func (a RealPair) Right() Expression                  { return a.Value() }
func (a RealPair) Both() (Expression, Expression)     { return a.Left(), a.Right() }
func (a RealPair) Pair() Paired                       { return a }
func (a RealPair) Pairs() []Paired                    { return []Paired{NewPair(a.Both())} }
func (a RealPair) Key() Expression                    { return a.Left() }
func (a RealPair) Call(args ...Expression) Expression { return a.Value().Call(args...) }
func (a RealPair) TypeFnc() TyFnc                     { return Real | Pair }
func (a RealPair) TypeKey() d.Typed                   { return Real }
func (a RealPair) TypeValue() d.Typed                 { return a.Value().Type() }
func (a RealPair) Type() TyComp                       { return Def(Pair, Def(Real, a.TypeValue())) }

// implement swappable
func (p RealPair) Swap() (Expression, Expression) {
	l, r := p()
	return Box(d.New(l)), r
}
func (p RealPair) SwappedPair() Paired { return NewPair(p.Right(), p.Left()) }
func (a RealPair) End() bool           { return a.Empty() }
func (a RealPair) Empty() bool {
	if a.Real() >= 0 && a.Value() != nil && a.Value().TypeFnc() != None {
		return true
	}
	return false
}
func (a RealPair) String() string {
	return "(" + a.Right().String() + " : " + a.Left().String() + ")"
}
func (p RealPair) Current() Expression                  { return p.Left() }
func (p RealPair) Next() Continuation                   { return NewPair(p.Right(), NewNone()) }
func (p RealPair) Continue() (Expression, Continuation) { return p.Current(), p.Next() }

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
