package functions

import (
	"sort"
	"strings"

	d "github.com/joergreinhardt/gatwd/data"
)

type (
	//// VALUE PAIRS
	ValPair   func(...Expression) (Expression, Expression)
	NatPair   func(...Expression) (Expression, d.Native)
	KeyPair   func(...Expression) (Expression, string)
	TypePair  func(...Expression) (Expression, Typed)
	IndexPair func(...Expression) (Expression, int)

	//// COLLECTIONS
	VecVal  func(...Expression) []Expression
	ListVal func(...Expression) (Expression, ListVal)
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
func (p ValPair) Left() Expression               { l, _ := p(); return l }
func (p ValPair) Right() Expression              { _, r := p(); return r }
func (p ValPair) Swap() (Expression, Expression) { l, r := p(); return r, l }
func (p ValPair) SwappedPair() Paired            { return NewPair(p.Right(), p.Left()) }
func (p ValPair) Slice() []Expression            { return []Expression{p.Left(), p.Right()} }
func (p ValPair) Key() Expression                { return p.Right() }
func (p ValPair) Value() Expression              { return p.Left() }
func (p ValPair) TypeFnc() TyFnc                 { return Pair }
func (p ValPair) TypeElem() TyPattern {
	if p.Right() != nil {
		return p.Left().Type()
	}
	return Def(None, Pair, None)
}
func (p ValPair) TypeKey() d.Typed {
	if p.Right() != nil {
		return p.Right().Type()
	}
	return None
}
func (p ValPair) TypeValue() d.Typed {
	if p.Left() != nil {
		return p.Left().Type()
	}
	return None
}
func (p ValPair) Type() TyPattern {
	if p.TypeKey().Match(None) && p.TypeValue().Match(None) {
		return Def(Pair, Def(None, None))
	}
	return Def(Pair, Def(p.TypeKey(), p.TypeValue()))
}

func (p ValPair) Empty() bool {
	if p.Left() == nil || (!p.Left().TypeFnc().Flag().Match(None) &&
		(p.Right() == nil || (!p.Right().TypeFnc().Flag().Match(None)))) {
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

///////////////////////////////////////////////////////////////////////////////
//// ASSOCIATIVE PAIRS
///
//// NATIVE VALUE KEY PAIR
///
//
func NewNatPair(key d.Native, val Expression) NatPair {
	return func(...Expression) (Expression, d.Native) { return val, key }
}

func (a NatPair) KeyNat() d.Native                   { _, key := a(); return key }
func (a NatPair) Value() Expression                  { val, _ := a(); return val }
func (a NatPair) Left() Expression                   { return a.Value() }
func (a NatPair) Right() Expression                  { return DecData(a.KeyNat()) }
func (a NatPair) Both() (Expression, Expression)     { return a.Left(), a.Right() }
func (a NatPair) Pair() Paired                       { return NewPair(a.Both()) }
func (a NatPair) Pairs() []Paired                    { return []Paired{NewPair(a.Both())} }
func (a NatPair) Key() Expression                    { return a.Right() }
func (a NatPair) Call(args ...Expression) Expression { return a.Value().Call(args...) }
func (a NatPair) TypeValue() d.Typed                 { return a.Value().Type() }
func (a NatPair) TypeKey() d.Typed                   { return a.KeyNat().Type() }
func (a NatPair) TypeFnc() TyFnc                     { return Data | Pair }
func (p NatPair) Type() TyPattern {
	if p.TypeKey().Match(None) && p.TypeValue().Match(None) {
		return Def(Pair, Def(Key, None))
	}
	return Def(Pair, Def(Key, p.TypeValue()))
}

// implement swappable
func (p NatPair) Swap() (Expression, Expression) {
	l, r := p()
	return DecData(r), l
}
func (p NatPair) SwappedPair() Paired { return NewPair(p.Right(), p.Left()) }

func (a NatPair) Empty() bool {
	if a.Key() != nil && a.Value() != nil && a.Value().TypeFnc() != None {
		return false
	}
	return true
}
func (a NatPair) String() string {
	return "(" + a.Right().String() + " : " + a.Left().String() + ")"
}

//// STRING KEY PAIR
///
// pair composed of a string key and a functional value
func NewKeyPair(key string, val Expression) KeyPair {
	return func(...Expression) (Expression, string) { return val, key }
}

func (a KeyPair) KeyStr() string                     { _, key := a(); return key }
func (a KeyPair) Value() Expression                  { val, _ := a(); return val }
func (a KeyPair) Left() Expression                   { return a.Value() }
func (a KeyPair) Right() Expression                  { return DecData(d.StrVal(a.KeyStr())) }
func (a KeyPair) Both() (Expression, Expression)     { return a.Left(), a.Right() }
func (a KeyPair) Pair() Paired                       { return NewPair(a.Both()) }
func (a KeyPair) Pairs() []Paired                    { return []Paired{NewPair(a.Both())} }
func (a KeyPair) Key() Expression                    { return a.Right() }
func (a KeyPair) Call(args ...Expression) Expression { return a.Value().Call(args...) }
func (a KeyPair) TypeValue() d.Typed                 { return a.Value().Type() }
func (a KeyPair) TypeKey() d.Typed                   { return Key }
func (a KeyPair) TypeFnc() TyFnc                     { return Key | Pair }
func (p KeyPair) Type() TyPattern {
	if p.TypeKey().Match(None) && p.TypeValue().Match(None) {
		return Def(Pair, Def(Key, None))
	}
	return Def(Key|Pair, Def(p.TypeKey(), p.TypeValue()))
}

// implement swappable
func (p KeyPair) Swap() (Expression, Expression) {
	l, r := p()
	return DecData(d.StrVal(r)), l
}
func (p KeyPair) SwappedPair() Paired { return NewPair(p.Right(), p.Left()) }

func (a KeyPair) Empty() bool {
	if a.Key() != nil && a.Value() != nil && a.Value().TypeFnc() != None {
		return false
	}
	return true
}
func (a KeyPair) String() string {
	return "(" + a.Right().String() + " : " + a.Left().String() + ")"
}

///////////////////////////////////////////////////////////////////////////////
//// INDEX PAIR
///
// pair composed of an integer and a functional value
func NewIndexPair(idx int, val Expression) IndexPair {
	return func(...Expression) (Expression, int) { return val, idx }
}
func (a IndexPair) Index() int                         { _, idx := a(); return idx }
func (a IndexPair) Value() Expression                  { val, _ := a(); return val }
func (a IndexPair) Left() Expression                   { return a.Value() }
func (a IndexPair) Right() Expression                  { return DecData(d.IntVal(a.Index())) }
func (a IndexPair) Both() (Expression, Expression)     { return a.Left(), a.Right() }
func (a IndexPair) Pair() Paired                       { return a }
func (a IndexPair) Pairs() []Paired                    { return []Paired{NewPair(a.Both())} }
func (a IndexPair) Key() Expression                    { return a.Right() }
func (a IndexPair) Call(args ...Expression) Expression { return a.Value().Call(args...) }
func (a IndexPair) TypeFnc() TyFnc                     { return Index | Pair }
func (a IndexPair) TypeKey() d.Typed                   { return Index }
func (a IndexPair) TypeValue() d.Typed                 { return a.Value().Type() }
func (a IndexPair) Type() TyPattern {
	if a.TypeKey().Match(None) && a.TypeValue().Match(None) {
		return Def(Pair, Def(Index, None))
	}
	return Def(Pair, Def(Index, a.TypeValue()))
}

// implement swappable
func (p IndexPair) Swap() (Expression, Expression) {
	l, r := p()
	return DecData(d.New(r)), l
}
func (p IndexPair) SwappedPair() Paired { return NewPair(p.Right(), p.Left()) }
func (a IndexPair) Empty() bool {
	if a.Index() >= 0 && a.Value() != nil && a.Value().TypeFnc() != None {
		return true
	}
	return false
}
func (a IndexPair) String() string {
	return "(" + a.Right().String() + " : " + a.Left().String() + ")"
}

//////////////////////////////////////////////////////////////////////////////////////////
//// VECTORS (SLICES) OF VALUES
///
// helper function to reverse argument sets
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

// sequential vector provides random access to sequential data. appends
// arguments in the order they where passed in, at the end of slice, when
// called
func NewVector(elems ...Expression) VecVal {
	if len(elems) == 0 {
		return VecVal(func(args ...Expression) []Expression {
			if len(args) > 0 {
				return NewVector(args...)()
			}
			return []Expression{}
		})
	}
	var match = func(args []Expression) bool {
		for _, arg := range args {
			if !elems[0].Type().Match(arg.Type()) {
				return false
			}
		}
		return true
	}
	return func(args ...Expression) []Expression {
		if len(args) > 0 {
			if match(args) {
				return append(elems, args...)
			}
		}
		return elems
	}
}

// default operation
func (v VecVal) Prepend(args ...Expression) Sequential {
	return NewVector(append(reverse(args), v()...)...)
}
func (v VecVal) PrependVec(args ...Expression) VecVal {
	return NewVector(append(reverse(args), v()...)...)
}
func (v VecVal) Append(args ...Expression) Sequential {
	return v.Cons(args...)
}
func (v VecVal) AppendVec(args ...Expression) VecVal {
	return v.ConsVec(args...)
}

// prepends arguments at head of list in reversed order, to emulate arguments
// added once at a time recursively.
func (v VecVal) Cons(args ...Expression) Sequential {
	return NewVector(append(v(), args...)...)
}

func (v VecVal) ConsVec(args ...Expression) VecVal {
	return NewVector(append(v(), args...)...)
}

// appends arguments to the vector, or returns unaltered vector, when no
// arguments are passed.
func (v VecVal) Call(args ...Expression) Expression {
	var (
		head Expression
		tail Consumeable
	)
	if len(args) > 0 {
		head, tail = NewVector(v(args...)...).Consume()
		return NewPair(head, tail)
	}
	head, tail = v.Consume()
	return NewPair(head, tail)
}
func (v VecVal) Slice() []Expression { return v() }
func (v VecVal) Len() int            { return len(v()) }
func (v VecVal) TypeFnc() TyFnc      { return Vector }
func (v VecVal) Type() TyPattern {
	if v.Len() > 0 {
		return Def(Vector, v.Head().Type())
	}
	return Def(Vector, None)
}
func (v VecVal) TypeElem() TyPattern {
	if v.Len() > 0 {
		return v.Head().Type()
	}
	return Def(None)
}

func (v VecVal) Head() Expression {
	if v.Len() > 0 {
		return v()[0]
	}
	return NewNone()
}

func (v VecVal) Tail() Consumeable {
	if v.Len() > 1 {
		return NewVector(v()[1:]...)
	}
	return nil
}

func (v VecVal) TailVec() VecVal {
	if v.Len() > 1 {
		return NewVector(v.Tail().(VecVal)()...)
	}
	return nil
}

func (v VecVal) Consume() (Expression, Consumeable) {
	return v.Head(), v.Tail()
}

func (v VecVal) ConsumeVec() (Expression, VecVal) {
	return v.Head(), v.TailVec()
}

func (v VecVal) First() Expression { return v.Head() }

func (v VecVal) Last() Expression {
	if v.Len() > 0 {
		return v()[v.Len()-1]
	}
	return nil
}

func (v VecVal) Reverse() VecVal {
	return NewVector(reverse(v())...)
}

func (v VecVal) Empty() bool {
	if len(v()) > 0 {
		return false
	}
	return true
}

func (v VecVal) Clear() VecVal { return NewVector() }

func (v VecVal) Sequential() SeqVal {
	return func(args ...Expression) (Expression, SeqVal) {
		var head, tail = v.ConsumeVec()
		if len(args) > 0 {
			return head, NewVector(tail(args...)...).Sequential()
		}
		return head, tail.Sequential()
	}
}

func (v VecVal) Get(i int) (Expression, bool) {
	if i < v.Len() {
		return v()[i], true
	}
	return NewNone(), false
}

func (v VecVal) Set(i int, val Expression) (Vectorized, bool) {
	if i < v.Len() {
		var slice = v()
		slice[i] = val
		return VecVal(
			func(elems ...Expression) []Expression {
				return slice
			}), true

	}
	return v, false
}
func (v VecVal) Sort(less func(a, b Expression) bool) VecVal {
	var s = vecSort(func() ([]Expression, func(i, j Expression) bool) {
		return v(), less
	})
	sort.Sort(s)
	var vec, _ = s()
	return NewVector(vec...)
}
func (v VecVal) Search(
	less func(a, b Expression) bool,
	match func(arg Expression) bool,
) Expression {
	var s = vecSort(func() ([]Expression, func(i, j Expression) bool) {
		return v(), less
	})
	sort.Sort(s)
	var vec, _ = s()
	for _, elem := range vec {
		if match(elem) {
			return elem
		}
	}
	return NewNone()
}
func (v VecVal) SearchAll(
	less func(a, b Expression) bool,
	match func(arg Expression) bool,
) VecVal {
	var s = vecSort(func() ([]Expression, func(i, j Expression) bool) {
		return v(), less
	})
	sort.Sort(s)
	var vec, _ = s()
	var res = []Expression{}
	for _, elem := range vec {
		if match(elem) {
			res = append(res, elem)
		}
	}
	return NewVector(res...)
}

// vector sorter with parametrizes less method
type vecSort func() ([]Expression, func(i, j Expression) bool)

func (v vecSort) Len() int {
	var s, _ = v()
	return len(s)
}
func (v vecSort) Less(i, j int) bool {
	var s, l = v()
	return l(s[i], s[j])
}
func (v vecSort) Swap(i, j int) {
	var s, l = v()
	s[i], s[j] = s[j], s[i]
	v = func() ([]Expression, func(Expression, Expression) bool) {
		return s, l
	}
}
func (v VecVal) String() string {
	var strs = []string{}
	for _, str := range v() {
		strs = append(strs, str.String())
	}
	return "[" + strings.Join(strs, ", ") + "]"
}

//// RECURSIVE LIST OF VALUES
///
// lazy implementation of recursively linked list. backed by slice. returns
// last element put in as head. prepends arguments when called to become new
// head of list, one at a time, thereby reversing argument order.
func NewList(elems ...Expression) ListVal {
	if len(elems) == 0 {
		return ListVal(func(args ...Expression) (Expression, ListVal) {
			if len(args) > 0 {
				return NewList(args...)()
			}
			return NewNone(), nil
		})
	}
	var match = func(args []Expression) bool {
		for _, arg := range args {
			if !elems[0].Type().Match(arg.Type()) {
				return false
			}
		}
		return true
	}
	return func(args ...Expression) (Expression, ListVal) {
		if len(args) > 0 {
			if match(args) {
				elems = append(elems, args...)
			}
		}
		var l = len(elems)
		if l > 0 {
			var head = elems[l-1]
			if l > 1 {
				return head, NewList(elems[:l-1]...)
			}
			return head, NewList()
		}
		return NewNone(), nil
	}
}

// default operation
func (l ListVal) Cons(elems ...Expression) Sequential {
	if len(elems) == 0 {
		return l
	}
	return ListVal(func(args ...Expression) (Expression, ListVal) {
		if len(args) > 0 {
			return l(append(elems, args...)...)
		}
		return l(elems...)
	})
}

// appends elements at the end of list in the order they where passed.
func (l ListVal) Prepend(elems ...Expression) Sequential { return l.Cons(elems...) }
func (l ListVal) Append(elems ...Expression) Sequential {
	return ListVal(func(args ...Expression) (Expression, ListVal) {
		if len(args) > 0 {
			var head, tail = l(args...)
			if tail.Empty() {
				return head, NewList(reverse(elems)...)
			}
			return head, tail.Append(elems...).(ListVal)
		}
		var head, tail = l()
		if tail.Empty() {
			return head, NewList(reverse(elems)...)
		}
		return head, tail.Append(elems...).(ListVal)
	})
}
func (l ListVal) Head() Expression                   { h, _ := l(); return h }
func (l ListVal) Tail() Consumeable                  { _, t := l(); return t }
func (l ListVal) TailList() ListVal                  { _, t := l(); return t }
func (l ListVal) Consume() (Expression, Consumeable) { return l() }
func (l ListVal) ConsumeList() (Expression, ListVal) {
	return l.Head(), l.TailList()
}
func (l ListVal) TypeFnc() TyFnc { return List }
func (l ListVal) Null() ListVal  { return NewList() }
func (l ListVal) TypeElem() TyPattern {
	if l.Len() > 0 {
		return l.Head().Type()
	}
	return Def(List, None)
}

func (l ListVal) Type() TyPattern {
	if l.Len() > 0 {
		return Def(List, l.Head().Type())
	}
	return Def(List, None)
}

func (l ListVal) Slice() []Expression {
	var (
		vec        = []Expression{}
		head, tail = l()
	)
	for tail != nil {
		vec = append(vec, head)
		head, tail = tail()
	}
	return vec
}

func (l ListVal) Call(args ...Expression) Expression {
	var (
		head Expression
		tail Sequential
	)
	if len(args) > 0 {
		head, tail = l(args...)
		return NewPair(head, tail)
	}
	head, tail = l()
	return NewPair(head, tail)
}

func (l ListVal) Empty() bool {
	if l.Tail() != nil {
		return false
	}
	return true
}

func (l ListVal) Len() int {
	var (
		length  int
		_, tail = l()
	)
	if tail != nil {
		length += 1 + tail.Len()
	}
	return length
}
func (l ListVal) String() string {
	var (
		args       = []string{}
		head, list = l()
	)
	for list != nil {
		args = append(args, head.String())
		head, list = list()
	}
	return "(" + strings.Join(args, ", ") + ")"
}
