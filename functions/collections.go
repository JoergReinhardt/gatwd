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
	VecVal   func(...Expression) []Expression
	StackVal func(...Expression) []Expression
	MapVal   func(...Expression) map[string]Expression
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
func (p ValPair) TypeElem() TyComp {
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
func (p ValPair) Type() TyComp {
	if p.TypeKey().Match(None) && p.TypeValue().Match(None) {
		return Def(Pair, Def(None, None))
	}
	return Def(Pair, Def(p.TypeKey(), p.TypeValue()))
}
func (p ValPair) End() bool { return p.Empty() }
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
func (p ValPair) Step() Expression                     { return p.Left() }
func (p ValPair) Next() Continuation                   { return NewPair(p.Right(), NewNone()) }
func (p ValPair) Continue() (Expression, Continuation) { return p.Step(), p.Next() }

//// NATIVE VALUE KEY PAIR
///
//
func NewNatPair(key d.Native, val Expression) NatPair {
	return func(...Expression) (Expression, d.Native) { return val, key }
}

func (a NatPair) KeyNat() d.Native                   { _, key := a(); return key }
func (a NatPair) Value() Expression                  { val, _ := a(); return val }
func (a NatPair) Left() Expression                   { return a.Value() }
func (a NatPair) Right() Expression                  { return Box(a.KeyNat()) }
func (a NatPair) Both() (Expression, Expression)     { return a.Left(), a.Right() }
func (a NatPair) Pair() Paired                       { return NewPair(a.Both()) }
func (a NatPair) Pairs() []Paired                    { return []Paired{NewPair(a.Both())} }
func (a NatPair) Key() Expression                    { return a.Right() }
func (a NatPair) Call(args ...Expression) Expression { return a.Value().Call(args...) }
func (a NatPair) TypeValue() d.Typed                 { return a.Value().Type() }
func (a NatPair) TypeKey() d.Typed                   { return a.KeyNat().Type() }
func (a NatPair) TypeFnc() TyFnc                     { return Data | Pair }
func (p NatPair) Type() TyComp {
	if p.TypeKey().Match(None) && p.TypeValue().Match(None) {
		return Def(Pair, Def(Key, None))
	}
	return Def(Pair, Def(Key, p.TypeValue()))
}

// implement swappable
func (p NatPair) Swap() (Expression, Expression) {
	l, r := p()
	return Box(r), l
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
func (p NatPair) Step() Expression                     { return p.Left() }
func (p NatPair) Next() Continuation                   { return NewPair(p.Right(), NewNone()) }
func (p NatPair) Continue() (Expression, Continuation) { return p.Step(), p.Next() }

//// STRING KEY PAIR
///
// pair composed of a string key and a functional value
func NewKeyPair(key string, val Expression) KeyPair {
	return func(...Expression) (Expression, string) { return val, key }
}

func (a KeyPair) KeyStr() string                     { _, key := a(); return key }
func (a KeyPair) Value() Expression                  { val, _ := a(); return val }
func (a KeyPair) Left() Expression                   { return a.Value() }
func (a KeyPair) Right() Expression                  { return Box(d.StrVal(a.KeyStr())) }
func (a KeyPair) Both() (Expression, Expression)     { return a.Left(), a.Right() }
func (a KeyPair) Pair() Paired                       { return NewPair(a.Both()) }
func (a KeyPair) Pairs() []Paired                    { return []Paired{NewPair(a.Both())} }
func (a KeyPair) Key() Expression                    { return a.Right() }
func (a KeyPair) Call(args ...Expression) Expression { return a.Value().Call(args...) }
func (a KeyPair) TypeValue() d.Typed                 { return a.Value().Type() }
func (a KeyPair) TypeElem() d.Typed                  { return a.Value().Type() }
func (a KeyPair) TypeKey() d.Typed                   { return Key }
func (a KeyPair) TypeFnc() TyFnc                     { return Key | Pair }
func (p KeyPair) Type() TyComp {
	if p.TypeKey().Match(None) && p.TypeValue().Match(None) {
		return Def(Pair, Def(Key, None))
	}
	return Def(Key|Pair, Def(p.TypeKey(), p.TypeValue()))
}

// implement swappable
func (p KeyPair) Swap() (Expression, Expression) {
	l, r := p()
	return Box(d.StrVal(r)), l
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
func (p KeyPair) Step() Expression                     { return p.Left() }
func (p KeyPair) Next() Continuation                   { return NewPair(p.Right(), NewNone()) }
func (p KeyPair) Continue() (Expression, Continuation) { return p.Step(), p.Next() }

//// INDEX PAIR
///
// pair composed of an integer and a functional value
func NewIndexPair(idx int, val Expression) IndexPair {
	return func(...Expression) (Expression, int) { return val, idx }
}
func (a IndexPair) Index() int                         { _, idx := a(); return idx }
func (a IndexPair) Value() Expression                  { val, _ := a(); return val }
func (a IndexPair) Left() Expression                   { return a.Value() }
func (a IndexPair) Right() Expression                  { return Box(d.IntVal(a.Index())) }
func (a IndexPair) Both() (Expression, Expression)     { return a.Left(), a.Right() }
func (a IndexPair) Pair() Paired                       { return a }
func (a IndexPair) Pairs() []Paired                    { return []Paired{NewPair(a.Both())} }
func (a IndexPair) Key() Expression                    { return a.Right() }
func (a IndexPair) Call(args ...Expression) Expression { return a.Value().Call(args...) }
func (a IndexPair) TypeFnc() TyFnc                     { return Index | Pair }
func (a IndexPair) TypeKey() d.Typed                   { return Index }
func (a IndexPair) TypeValue() d.Typed                 { return a.Value().Type() }
func (a IndexPair) Type() TyComp {
	if a.TypeKey().Match(None) && a.TypeValue().Match(None) {
		return Def(Pair, Def(Index, None))
	}
	return Def(Pair, Def(Index, a.TypeValue()))
}

// implement swappable
func (p IndexPair) Swap() (Expression, Expression) {
	l, r := p()
	return Box(d.New(r)), l
}
func (p IndexPair) SwappedPair() Paired { return NewPair(p.Right(), p.Left()) }
func (a IndexPair) End() bool           { return a.End() }
func (a IndexPair) Empty() bool {
	if a.Index() >= 0 && a.Value() != nil && a.Value().TypeFnc() != None {
		return true
	}
	return false
}
func (a IndexPair) String() string {
	return "(" + a.Right().String() + " : " + a.Left().String() + ")"
}
func (p IndexPair) Step() Expression                     { return p.Left() }
func (p IndexPair) Next() Continuation                   { return NewPair(p.Right(), NewNone()) }
func (p IndexPair) Continue() (Expression, Continuation) { return p.Step(), p.Next() }

////////////////////////////////////////////////////////////////////////////////
//// SORTER
///
// sorter is a helper struct to sort vector elements inline
type sorter struct {
	Slice []Expression
	By
}

func newSorter(slice []Expression, by By) *sorter {
	return &sorter{slice, by}
}

func (t *sorter) Swap(i, j int)     { (*t).Slice[j], (*t).Slice[i] = (*t).Slice[i], (*t).Slice[j] }
func (t sorter) Less(i, j int) bool { return t.By(i, j) }
func (t sorter) Len() int           { return len(t.Slice) }

// sort interface. the'By' type implements 'sort.Less() int' and is the
// function type of a parameterized sort & search function.
type By func(a, b int) bool

func (by By) Sort(slice []Expression) []Expression {
	var sorter = newSorter(slice, by)
	sort.Sort(sorter)
	return sorter.Slice
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

// default list operation prepends at the beginning of the list
func (v VecVal) Cons(args ...Expression) Sequential   { return NewVector(append(args, v()...)...) }
func (v VecVal) Concat(args ...Expression) Sequential { return NewVector(v(args...)...) }
func (v VecVal) ConcatVec(args ...Expression) VecVal  { return NewVector(v(args...)...) }
func (v VecVal) Len() int                             { return len(v()) }
func (v VecVal) Step() Expression {
	if v.Len() > 0 {
		return v()[0]
	}
	return NewNone()
}
func (v VecVal) Next() Continuation {
	if v.Len() > 1 {
		return NewVector(v()[1:]...)
	}
	return NewVector()
}
func (v VecVal) Continue() (Expression, Continuation) { return v.Step(), v.Next() }
func (v VecVal) Null() VecVal                         { return NewVector() }
func (v VecVal) TypeFnc() TyFnc                       { return Vector }
func (v VecVal) TypeElem() TyComp                     { return v.Step().Type() }
func (v VecVal) Type() TyComp                         { return Def(Vector, v.TypeElem()) }
func (v VecVal) Slice() []Expression                  { return v() }
func (v VecVal) Call(args ...Expression) Expression {
	if len(args) > 0 {
		return NewPair(v.Step().Call(args...), v.Next())
	}
	return NewPair(v.Step(), v.Next())
}

func (v VecVal) End() bool {
	if v.Len() != 0 || !v.Step().Type().Match(None) {
		return false
	}
	return true
}
func (v VecVal) String() string {
	var strs = []string{}
	for _, str := range v() {
		strs = append(strs, str.String())
	}
	return "[" + strings.Join(strs, ", ") + "]"
}

func (v VecVal) First() Expression { return v.Step() }

func (v VecVal) Last() Expression {
	if v.Len() > 0 {
		return v()[v.Len()-1]
	}
	return nil
}

func (v VecVal) Reverse() VecVal {
	return NewVector(reverse(v())...)
}

func (v VecVal) Clear() VecVal { return NewVector() }

func (v VecVal) Sequence() SeqVal {
	return func(args ...Expression) (Expression, SeqVal) {
		var head, tail = v.Continue()
		if len(args) > 0 {
			return head, NewVector(tail.(VecVal)(args...)...).Sequence()
		}
		return head, tail.(VecVal).Sequence()
	}
}

func (v VecVal) Get(i int) (Expression, bool) {
	if i < v.Len() {
		return v()[i], true
	}
	return NewNone(), false
}
func (v VecVal) Sort(by By) VecVal {
	var s = newSorter(v(), by)
	sort.Sort(s)
	return NewVector(s.Slice...)
}

func (v VecVal) Search(
	by By,
	match func(arg Expression) bool,
) Expression {
	var s = newSorter(v(), by)
	sort.Sort(s)
	for _, elem := range s.Slice {
		if match(elem) {
			return elem
		}
	}
	return NewNone()
}

func (v VecVal) SearchAll(
	by func(i, j int) bool,
	match func(arg Expression) bool,
) VecVal {
	var (
		s   = newSorter(v(), by)
		vec = NewVector()
	)
	sort.Sort(s)
	for _, elem := range s.Slice {
		if match(elem) {
			vec = vec.ConcatVec(elem)
		}
	}
	return vec
}

///////////////////////////////////////////////////////////////////////////////
//// RECURSIVE LIST OF VALUES
///
// lazy implementation of recursively linked list. backed by slice. returns
// last element put in as head. prepends arguments when called to become new
// head of list, one at a time, thereby reversing argument order.
func NewStack(elems ...Expression) StackVal {
	if len(elems) == 0 {
		return StackVal(func(args ...Expression) []Expression {
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

// default list operation prepends at the beginning of the list
func (l StackVal) Cons(args ...Expression) Sequential      { return NewStack(l(args...)...) }
func (l StackVal) ConsStack(args ...Expression) StackVal   { return NewStack(l(args...)...) }
func (l StackVal) Concat(args ...Expression) Sequential    { return NewStack(append(l(), args...)...) }
func (l StackVal) ConcatStack(args ...Expression) StackVal { return NewStack(append(l(), args...)...) }
func (l StackVal) Len() int                                { return len(l()) }
func (l StackVal) Step() Expression {
	if l.Len() > 0 {
		return l()[l.Len()-1]
	}
	return NewNone()
}
func (l StackVal) Next() Continuation {
	if l.Len() > 1 {
		return NewStack(l()[:l.Len()-1]...)
	}
	return NewStack()
}
func (l StackVal) Continue() (Expression, Continuation) { return l.Step(), l.Next() }
func (l StackVal) Null() StackVal                       { return NewStack() }
func (l StackVal) TypeFnc() TyFnc                       { return List }
func (l StackVal) TypeElem() TyComp                     { return l.Step().Type() }
func (l StackVal) Type() TyComp                         { return Def(List, l.TypeElem()) }
func (l StackVal) Slice() []Expression                  { return l() }
func (l StackVal) Sequence() Sequential {
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		var head, tail = l.Continue()
		if len(args) > 0 {
			return head, NewStack(
				tail.(StackVal)(args...)...,
			).Sequence().(SeqVal)
		}
		return head, tail.(StackVal).Sequence().(SeqVal)
	})
}
func (l StackVal) Call(args ...Expression) Expression {
	if len(args) > 0 {
		return NewPair(l.Step().Call(args...), l.Next())
	}
	return NewPair(l.Step(), l.Next())
}

func (l StackVal) End() bool {
	if l.Len() != 0 || !l.Step().Type().Match(None) {
		return false
	}
	return true
}
func (l StackVal) String() string {
	var (
		args       = []string{}
		head, list = l.Continue()
	)
	for list != nil {
		args = append(args, head.String())
		head, list = list.Continue()
	}
	return "(" + strings.Join(args, ", ") + ")"
}

func (v StackVal) Get(i int) (Expression, bool) {
	if i < v.Len() {
		return v()[i], true
	}
	return NewNone(), false
}
func (v StackVal) Sort(by By) StackVal {
	var s = newSorter(v(), by)
	sort.Sort(s)
	return NewStack(s.Slice...)
}

func (v StackVal) Search(
	by By,
	match func(arg Expression) bool,
) Expression {
	var s = newSorter(v(), by)
	sort.Sort(s)
	for _, elem := range s.Slice {
		if match(elem) {
			return elem
		}
	}
	return NewNone()
}

func (v StackVal) SearchAll(
	by func(i, j int) bool,
	match func(arg Expression) bool,
) StackVal {
	var (
		s     = newSorter(v(), by)
		stack = NewStack()
	)
	sort.Sort(s)
	for _, elem := range s.Slice {
		if match(elem) {
			stack = stack.ConsStack(elem)
		}
	}
	return stack
}

///////////////////////////////////////////////////////////////////////////////
//// MAP VALUE
///
// sequential vector provides random access to sequential data. appends
// arguments in the order they where passed in, at the end of slice, when
// called
func NewMap(pairs ...KeyPair) MapVal {
	var (
		val = map[string]Expression{}
		cp  = func(m map[string]Expression) map[string]Expression {
			var cpval = map[string]Expression{}
			for k, v := range m {
				cpval[k] = v
			}
			return cpval
		}
	)
	if len(pairs) == 0 {
		for _, pair := range pairs {
			val[string(pair.KeyStr())] = pair.Value()
		}
		return MapVal(func(args ...Expression) map[string]Expression {
			if len(args) > 0 {
				var cpval = cp(val)
				for _, arg := range args {
					if arg.Type().Match(Pair | Key) {
						if pair, ok := arg.(KeyPair); ok {
							cpval[string(pair.KeyStr())] = pair.Value()
						}
					}
				}
				return cpval
			}
			return val
		})
	}
	return func(args ...Expression) map[string]Expression {
		if len(args) > 0 {
			var cpval = cp(val)
			for _, arg := range args {
				if arg.TypeFnc().Match(Pair | Key) {
					if pair, ok := arg.(KeyPair); ok {
						cpval[string(pair.KeyStr())] = pair.Value()
					}
				}
			}
		}
		return val
	}
}

// default operation depends on map state. For empty maps, elements will be
// added to the map, for maps containing elements allready, arguments are cast
// to string, looked up and the corresponding values are returned.
func (m MapVal) Call(args ...Expression) Expression {
	if len(args) > 0 {
		var result = make([]Expression, 0, len(args))
		for _, arg := range args {
			if found, ok := m()[arg.String()]; ok {
				result = append(result, found)
				continue
			}
		}
		return NewVector(result...)
	}
	return m
}
func (m MapVal) KeyVals() ([]string, []Expression) {
	var (
		keys = make([]string, 0, m.Len())
		vals = make([]Expression, 0, m.Len())
	)
	for k, v := range m() {
		keys = append(keys, k)
		vals = append(vals, v)
	}
	return keys, vals
}
func (m MapVal) Keys() []string {
	var k, _ = m.KeyVals()
	return k
}
func (m MapVal) Values() []Expression {
	var _, v = m.KeyVals()
	return v
}
func (m MapVal) KeyPairs() []KeyPair {
	var (
		keys, vals = m.KeyVals()
		pairs      = make([]KeyPair, 0, m.Len())
	)
	for n, key := range keys {
		pairs = append(pairs, NewKeyPair(key, vals[n]))
	}
	return pairs
}
func (m MapVal) Get(key string) Expression {
	if found, ok := m()[key]; ok {
		return found
	}
	return NewNone()
}
func (m MapVal) Type() TyComp   { return Def(HashMap, m.TypeElem()) }
func (m MapVal) TypeFnc() TyFnc { return HashMap }
func (m MapVal) TypeElem() TyComp {
	if m.Len() > 0 {
		return m()[m.Keys()[0]].Type()
	}
	return Def(None)
}
func (m MapVal) Len() int { return len(m()) }
func (m MapVal) String() string {
	var strs = make([]string, 0, m.Len())
	for k, v := range m() {
		strs = append(strs, k+" ∷ "+v.String())
	}
	return "{" + strings.Join(strs, " ") + "}"
}
