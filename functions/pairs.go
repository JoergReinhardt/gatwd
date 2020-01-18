package functions

import (
	"fmt"
	"strconv"
	"strings"

	d "github.com/joergreinhardt/gatwd/data"
)

type (
	//// VALUE PAIRS
	TypePair  func(...Functor) (d.Typed, Functor)
	RealPair  func(...Functor) (float64, Functor)
	IndexPair func(...Functor) (int, Functor)
	KeyPair   func(...Functor) (string, Functor)
	ValPair   func(...Functor) (Functor, Functor)
	NatPair   func(...Functor) (d.Native, Functor)

	//// COLLECTIONS OF VALUE PAIRS
	KeyIndex []KeyPair
	KeyMap   map[string]Functor
	RealMap  map[float64]Functor
	TypeMap  map[d.BitFlag]Functor
)

///////////////////////////////////////////////////////////////////////////////
//// PAIRS OF VALUES
///
// pairs can be created empty, key & value may be constructed later
func NewEmptyPair() ValPair {
	return func(args ...Functor) (a, b Functor) {
		if len(args) > 0 {
			if len(args) > 1 {
				if len(args) > 2 {
					return NewPair(args[0], args[1]),
						NewList(args[2:]...)

				}
				return args[0], args[1]
			}
			return args[0], NewNone()
		}
		return NewNone(), NewNone()
	}
}

// new pair from two callable instances
func NewPair(l, r Functor) ValPair {
	return func(args ...Functor) (Functor, Functor) {
		if len(args) > 0 {
			if len(args) > 1 {
				if len(args) > 2 {
					return NewPair(args[0], args[1]),
						NewList(args[2:]...)
				}
				return args[0], args[1]
			}
			return args[0], r
		}
		return l, r
	}
}

func (p ValPair) Cons(arg Functor) Applicative    { return NewPair(arg, p) }
func (p ValPair) Concat(c Sequential) Applicative { return NewPair(p, c) }
func (p ValPair) Continue() (Functor, Applicative) {
	var (
		head Functor
		tail Sequential
		k, v = p()
	)
	if k.TypeFnc().Match(Continua) {
		head, tail = k.(Sequential).Continue()
		if v.TypeFnc().Match(Continua) {
			return head, tail.Concat(v.(Sequential))
		}
	}
	if v.TypeFnc().Match(Continua) {
		return head, tail.Concat(v.(Sequential))
	}
	return k, NewPair(v, NewNone())
}
func (p ValPair) Head() Functor {
	var h, _ = p.Continue()
	return h
}
func (p ValPair) Tail() Applicative {
	var _, t = p.Continue()
	return t
}

func (p ValPair) Both() (Functor, Functor) { return p() }
func (p ValPair) Swap() (Functor, Functor) { l, r := p(); return r, l }
func (p ValPair) Left() Functor            { l, _ := p(); return l }
func (p ValPair) Right() Functor           { _, r := p(); return r }

func (p ValPair) Pair() Paired        { return p }
func (p ValPair) SwappedPair() Paired { return NewPair(p.Right(), p.Left()) }
func (p ValPair) Slice() []Functor    { return []Functor{p.Left(), p.Right()} }

func (p ValPair) Key() Functor   { return p.Left() }
func (p ValPair) Value() Functor { return p.Right() }
func (p ValPair) TypeFnc() TyFnc { return Pair }
func (p ValPair) TypeElem() Decl {
	if p.Right() != nil {
		return p.Right().Type()
	}
	return Declare(None, Pair, None)
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
func (p ValPair) Type() Decl {
	return Declare(Pair, Declare(p.TypeKey(), p.TypeValue()))
}
func (p ValPair) Empty() bool {
	if p.Left() == nil || (!p.Left().Type().Match(None) &&
		(p.Right() == nil || (!p.Right().Type().Match(None)))) {
		return true
	}
	return false
}
func (p ValPair) String() string {
	return "(" + p.Key().String() + ", " + p.Value().String() + ")"
}
func (p ValPair) Call(args ...Functor) Functor {
	if len(args) > 0 {
		return NewPair(p.Key(), p.Value().Call(args...))
	}
	return p
}

//// NATIVE VALUE KEY PAIR
///
//
func NewNatPair(key d.Native, val Functor) NatPair {
	return func(...Functor) (d.Native, Functor) { return key, val }
}

func (a NatPair) KeyNat() d.Native         { key, _ := a(); return key }
func (a NatPair) Value() Functor           { _, val := a(); return val }
func (a NatPair) Left() Functor            { return Box(a.KeyNat()) }
func (a NatPair) Right() Functor           { return a.Value() }
func (a NatPair) Both() (Functor, Functor) { return a.Left(), a.Right() }
func (a NatPair) Pair() Paired             { return NewPair(a.Both()) }
func (a NatPair) Pairs() []Paired          { return []Paired{NewPair(a.Both())} }
func (a NatPair) Key() Functor             { return a.Left() }
func (a NatPair) TypeValue() d.Typed       { return a.Value().Type() }
func (a NatPair) TypeKey() d.Typed         { return a.KeyNat().Type() }
func (a NatPair) TypeFnc() TyFnc           { return Data | Pair }
func (p NatPair) Type() Decl               { return Declare(Pair, Declare(Key, p.TypeValue())) }
func (p NatPair) Call(args ...Functor) Functor {
	if len(args) > 0 {
		return NewNatPair(p.KeyNat(), p.Value().Call(args...))
	}
	return p
}

// implement swappable
func (p NatPair) Swap() (Functor, Functor) {
	l, r := p()
	return Box(l), r
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

func (p NatPair) Cons(arg Functor) Functor { return NewPair(arg, p) }
func (p NatPair) Continue() (Functor, Applicative) {
	var (
		head Functor
		tail Sequential
		k, v = p()
	)
	if v.TypeFnc().Match(Continua) {
		return head, tail.Concat(v.(Sequential))
	}
	return Box(k), NewPair(v, NewNone())
}
func (p NatPair) Head() Functor {
	var h, _ = p.Continue()
	return h
}
func (p NatPair) Tail() Applicative {
	var _, t = p.Continue()
	return t
}

//// STRING KEY PAIR
///
// pair composed of a string key and a functional value
func NewKeyPair(key string, val Functor) KeyPair {
	return func(...Functor) (string, Functor) { return key, val }
}

func (a KeyPair) KeyStr() string           { key, _ := a(); return key }
func (a KeyPair) Value() Functor           { _, val := a(); return val }
func (a KeyPair) Left() Functor            { return Box(d.StrVal(a.KeyStr())) }
func (a KeyPair) Right() Functor           { return a.Value() }
func (a KeyPair) Both() (Functor, Functor) { return a.Left(), a.Right() }
func (a KeyPair) Pair() Paired             { return NewPair(a.Both()) }
func (a KeyPair) Pairs() []Paired          { return []Paired{NewPair(a.Both())} }
func (a KeyPair) Key() Functor             { return a.Left() }
func (a KeyPair) TypeValue() d.Typed       { return a.Value().Type() }
func (a KeyPair) TypeElem() d.Typed        { return a.Value().Type() }
func (a KeyPair) TypeKey() d.Typed         { return Key }
func (a KeyPair) TypeFnc() TyFnc           { return Key | Pair }
func (p KeyPair) Call(args ...Functor) Functor {
	if len(args) > 0 {
		return NewKeyPair(p.KeyStr(), p.Value().Call(args...))
	}
	return p
}
func (p KeyPair) Type() Decl {
	return Declare(Key|Pair, Declare(Key, p.TypeValue()))
}

// implement swappable
func (p KeyPair) Swap() (Functor, Functor) {
	l, r := p()
	return Box(d.StrVal(l)), r
}
func (p KeyPair) SwappedPair() Paired { return NewPair(p.Right(), p.Left()) }

func (a KeyPair) Empty() bool {
	if a.Key() != nil && a.Value() != nil && a.Value().TypeFnc() != None {
		return false
	}
	return true
}
func (a KeyPair) String() string {
	return "(" + a.KeyStr() + " : " + a.Value().String() + ")"
}
func (p KeyPair) Cons(arg Functor) Applicative       { return NewPair(arg, p) }
func (p KeyPair) Concat(cons Sequential) Applicative { return NewPair(p, cons) }
func (p KeyPair) Continue() (Functor, Applicative) {
	var (
		head Functor
		tail Sequential
		k, v = p()
	)
	if v.TypeFnc().Match(Continua) {
		return head, tail.Concat(v.(Sequential))
	}
	return Box(d.StrVal(k)), NewPair(v, NewNone())
}
func (p KeyPair) Head() Functor {
	var h, _ = p.Continue()
	return h
}
func (p KeyPair) Tail() Applicative {
	var _, t = p.Continue()
	return t
}

//// INDEX PAIR
///
// pair composed of an integer and a functional value
func NewIndexPair(idx int, val Functor) IndexPair {
	return func(...Functor) (int, Functor) { return idx, val }
}
func (a IndexPair) Value() Functor           { _, val := a(); return val }
func (a IndexPair) Index() int               { idx, _ := a(); return idx }
func (a IndexPair) KeyIdx() int              { return a.Index() }
func (a IndexPair) Left() Functor            { return Box(d.IntVal(a.Index())) }
func (a IndexPair) Right() Functor           { return a.Value() }
func (a IndexPair) Both() (Functor, Functor) { return a.Left(), a.Right() }
func (a IndexPair) Pair() Paired             { return a }
func (a IndexPair) Pairs() []Paired          { return []Paired{NewPair(a.Both())} }
func (a IndexPair) Key() Functor             { return a.Left() }
func (a IndexPair) TypeFnc() TyFnc           { return Index | Pair }
func (a IndexPair) TypeKey() d.Typed         { return Index }
func (a IndexPair) TypeValue() d.Typed       { return a.Value().Type() }
func (a IndexPair) Type() Decl               { return Declare(Pair, Declare(Index, a.TypeValue())) }
func (p IndexPair) TypeElem() Decl           { return p.Value().Type() }
func (p IndexPair) Call(args ...Functor) Functor {
	if len(args) > 0 {
		return NewIndexPair(p.Index(), p.Value().Call(args...))
	}
	return p
}

// implement swappable
func (p IndexPair) Swap() (Functor, Functor) {
	l, r := p()
	return Box(d.New(l)), r
}
func (p IndexPair) SwappedPair() Paired { return NewPair(p.Right(), p.Left()) }
func (a IndexPair) Empty() bool {
	if a.Index() >= 0 && a.Value() != nil && a.Value().TypeFnc() != None {
		return true
	}
	return false
}
func (a IndexPair) String() string {
	return "(" + a.Key().String() + " : " + a.Value().String() + ")"
}

func (p IndexPair) Cons(arg Functor) Applicative    { return NewPair(arg, p) }
func (p IndexPair) Concat(c Sequential) Applicative { return NewPair(p, c) }
func (p IndexPair) Continue() (Functor, Applicative) {
	var (
		head Functor
		tail Sequential
		k, v = p()
	)
	if v.TypeFnc().Match(Continua) {
		return head, tail.Concat(v.(Sequential))
	}
	return Box(d.IntVal(k)), NewPair(v, NewNone())
}
func (p IndexPair) Head() Functor {
	var h, _ = p.Continue()
	return h
}
func (p IndexPair) Tail() Applicative {
	var _, t = p.Continue()
	return t
}

//// FLOATING PAIR
///
// pair composed of an integer and a functional value
func NewRealPair(flt float64, val Functor) RealPair {
	return func(...Functor) (float64, Functor) { return flt, val }
}
func (a RealPair) Real() float64                { flt, _ := a(); return flt }
func (a RealPair) Value() Functor               { _, val := a(); return val }
func (a RealPair) Left() Functor                { return Box(d.IntVal(a.Real())) }
func (a RealPair) Right() Functor               { return a.Value() }
func (a RealPair) Both() (Functor, Functor)     { return a.Left(), a.Right() }
func (a RealPair) Pair() Paired                 { return a }
func (a RealPair) Pairs() []Paired              { return []Paired{NewPair(a.Both())} }
func (a RealPair) Key() Functor                 { return a.Left() }
func (a RealPair) Call(args ...Functor) Functor { return a.Value().Call(args...) }
func (a RealPair) TypeFnc() TyFnc               { return Real | Pair }
func (a RealPair) TypeKey() d.Typed             { return Real }
func (a RealPair) TypeValue() d.Typed           { return a.Value().Type() }
func (a RealPair) Type() Decl                   { return Declare(Pair, Declare(Real, a.TypeValue())) }

// implement swappable
func (p RealPair) Swap() (Functor, Functor) {
	l, r := p()
	return Box(d.New(l)), r
}
func (p RealPair) SwappedPair() Paired { return NewPair(p.Right(), p.Left()) }
func (a RealPair) Empty() bool {
	if a.Real() >= 0 && a.Value() != nil && a.Value().TypeFnc() != None {
		return true
	}
	return false
}
func (a RealPair) String() string {
	return "(" + a.Key().String() + " : " + a.Value().String() + ")"
}
func (p RealPair) TypeElem() Decl { return p.Value().Type() }

func (p RealPair) Cons(arg Functor) Applicative    { return NewPair(arg, p) }
func (p RealPair) Concat(c Sequential) Applicative { return NewPair(p, c) }
func (p RealPair) Continue() (Functor, Applicative) {
	var (
		head Functor
		tail Sequential
		k, v = p()
	)
	if v.TypeFnc().Match(Continua) {
		return head, tail.Concat(v.(Sequential))
	}
	return Box(d.FltVal(k)), NewPair(v, NewNone())
}
func (p RealPair) Head() Functor {
	var h, _ = p.Continue()
	return h
}
func (p RealPair) Tail() Applicative {
	var _, t = p.Continue()
	return t
}

//// TYPE PAIR
///
// pair composed of a type flag and a functional value
func NewTypePair(typ d.Typed, val Functor) TypePair {
	return func(...Functor) (d.Typed, Functor) { return typ, val }
}
func (a TypePair) Value() Functor { _, val := a(); return val }
func (a TypePair) KeyTyped() d.Typed {
	var t, _ = a.Key().(d.Typed)
	return t
}
func (a TypePair) KeyDef() Decl {
	if Kind_Decl.Match(a.KeyTyped().Kind()) {
		return a.Key().(Decl)
	}
	return Declare(a.KeyTyped())
}
func (a TypePair) Key() Functor                 { return a.KeyDef() }
func (a TypePair) Left() Functor                { return a.KeyDef() }
func (a TypePair) Right() Functor               { return a.Value() }
func (a TypePair) Both() (Functor, Functor)     { return a.Left(), a.Right() }
func (a TypePair) Pair() Paired                 { return a }
func (a TypePair) Pairs() []Paired              { return []Paired{NewPair(a.Both())} }
func (a TypePair) Call(args ...Functor) Functor { return a.Value().Call(args...) }
func (a TypePair) TypeFnc() TyFnc               { return Type | Pair }
func (a TypePair) TypeKey() d.Typed             { return Type }
func (a TypePair) TypeValue() d.Typed           { return a.Value().Type() }
func (a TypePair) Type() Decl                   { return Declare(Pair, Declare(Type, a.TypeValue())) }

// implement swappable
func (p TypePair) Swap() (Functor, Functor) {
	l, r := p()
	return Box(d.New(l)), r
}
func (p TypePair) SwappedPair() Paired { return NewPair(p.Right(), p.Left()) }
func (a TypePair) Empty() bool {
	if a.KeyDef().Match(None) && a.Right().Type().Match(None) {
		return true
	}
	return false
}
func (a TypePair) String() string {
	return "(" + a.Key().String() + " : " + a.Value().String() + ")"
}
func (p TypePair) TypeElem() Decl { return p.Value().Type() }

func (p TypePair) Cons(arg Functor) Applicative    { return NewPair(arg, p) }
func (p TypePair) Concat(c Sequential) Applicative { return NewPair(p, c) }
func (p TypePair) Continue() (Functor, Applicative) {
	var (
		head Functor
		tail Sequential
		k, v = p.KeyDef(), p.Value()
	)
	if v.TypeFnc().Match(Continua) {
		return head, tail.Concat(v.(Sequential))
	}
	return k, NewPair(v, NewNone())
}
func (p TypePair) Head() Functor {
	var h, _ = p.Continue()
	return h
}
func (p TypePair) Tail() Applicative {
	var _, t = p.Continue()
	return t
}

///////////////////////////////////////////////////////////////////////////////
//// KEY INDEX
///  key index keeps index position of key/value pairs stored in a hash map in
//   order.
func NewKeyIndex(pairs ...KeyPair) KeyIndex { return pairs }

func (i KeyIndex) Call(...Functor) Functor { return i }

func (i KeyIndex) Len() int       { return len(i) }
func (i KeyIndex) Vector() VecVal { return NewVector(i.Slice()...) }
func (i KeyIndex) TypeFnc() TyFnc { return Key | Pair | Vector }
func (i KeyIndex) Type() Decl {
	return Declare(Vector, Declare(Pair, Declare(String, T)))
}
func (i KeyIndex) GetByKey(key string) Functor {
	var s = newSearcher(
		i.Slice(),
		func(key, arg Functor) int {
			return strings.Compare(
				key.String(), arg.String())
		})
	return s.Search(Box(d.StrVal(key)))
}
func (i KeyIndex) GetByIdx(idx int) Functor {
	if idx < i.Len() {
		return i[idx]
	}
	return NewNone()
}
func (i KeyIndex) String() string {
	var str string
	for i, p := range i {
		str = str + strconv.Itoa(i) +
			"\t:\t" + p.KeyStr() + "\n"
	}
	return str
}
func (i KeyIndex) Slice() []Functor {
	var slice = make([]Functor, 0, i.Len())
	for _, p := range i {
		slice = append(slice, p)
	}
	return slice
}
func (i KeyIndex) Keys() []string {
	var strs = make([]string, 0, i.Len())
	for _, p := range i {
		strs = append(strs, p.KeyStr())
	}
	return strs
}
func (i KeyIndex) Values() []Functor {
	var vals = make([]Functor, 0, i.Len())
	for _, p := range i {
		vals = append(vals, p.Value())
	}
	return vals
}
func (i KeyIndex) InvertPairs() KeyMap {
	var m = map[string]Functor{}
	for n := 0; n < i.Len(); n++ {
		m[i[n].KeyStr()] = NewIndexPair(n, i[n])
	}
	return m
}
func (i KeyIndex) InvertVals() KeyMap {
	var m = map[string]Functor{}
	for _, v := range i {
		m[v.KeyStr()] = v
	}
	return m
}
func (i KeyIndex) InvertIdx() KeyMap {
	var m = map[string]Functor{}
	for n, v := range i {
		m[v.KeyStr()] = Box(d.IntVal(n))
		n -= 1
	}
	return m
}

///////////////////////////////////////////////////////////////////////////////
//// KEY MAP
///
func NewKeyMap(pairs ...KeyPair) KeyMap {
	var m = map[string]Functor{}
	for _, pair := range pairs {
		m[pair.KeyStr()] = pair.Value()
	}
	return m
}
func (k KeyMap) Type() Decl                   { return Declare(Key, HashMap) }
func (k KeyMap) TypeFnc() TyFnc               { return Key | HashMap }
func (k KeyMap) Call(args ...Functor) Functor { return k }
func (k KeyMap) String() string {
	var str = "{\n}"
	for k, v := range k {
		str = str + k + " ∷ " + v.String() + "\n"
	}
	str = str + "}"
	return str
}
func (k KeyMap) Get(key string) Functor {
	if val, ok := k[key]; ok {
		return val
	}
	return NewNone()
}
func (k KeyMap) GetPair(key string) KeyPair {
	if val, ok := k[key]; ok {
		return NewKeyPair(key, val)
	}
	return NewKeyPair("", NewNone())
}
func (k KeyMap) Pairs() []KeyPair {
	var pairs = make([]KeyPair, 0, len(k))
	for k, v := range k {
		pairs = append(pairs, NewKeyPair(k, v))
	}
	return pairs
}

//// REAL MAP
///
func NewRealMap(pairs ...RealPair) RealMap {
	var m = map[float64]Functor{}
	for _, pair := range pairs {
		m[pair.Real()] = pair.Value()
	}
	return m
}
func (k RealMap) Type() Decl                   { return Declare(Real, HashMap) }
func (k RealMap) TypeFnc() TyFnc               { return Real | HashMap }
func (k RealMap) Call(args ...Functor) Functor { return k }
func (k RealMap) String() string {
	var str = "{\n}"
	for k, v := range k {
		str = str + fmt.Sprintf("%f", k) + " ∷ " +
			v.String() + "\n"
	}
	str = str + "}"
	return str
}
func (k RealMap) Get(key float64) Functor {
	if val, ok := k[key]; ok {
		return val
	}
	return NewNone()
}
func (k RealMap) GetPair(key float64) RealPair {
	if val, ok := k[key]; ok {
		return NewRealPair(key, val)
	}
	return NewRealPair(0.0, NewNone())
}
func (k RealMap) Pairs() []RealPair {
	var pairs = make([]RealPair, 0, len(k))
	for k, v := range k {
		pairs = append(pairs, NewRealPair(k, v))
	}
	return pairs
}

//// KEY MAP
///
func NewTypeMap(pairs ...TypePair) TypeMap {
	var m = map[d.BitFlag]Functor{}
	for _, pair := range pairs {
		m[pair.KeyTyped().Flag()] = pair.Value()
	}
	return m
}
func (k TypeMap) Type() Decl                   { return Declare(Type, HashMap) }
func (k TypeMap) TypeFnc() TyFnc               { return Type | HashMap }
func (k TypeMap) Call(args ...Functor) Functor { return k }
func (k TypeMap) String() string {
	var str = "{\n}"
	for k, v := range k {
		str = str + d.Typed(k).TypeName() + " ∷ " + v.String() + "\n"
	}
	str = str + "}"
	return str
}
func (k TypeMap) Get(key d.Typed) Functor {
	if val, ok := k[key.Flag()]; ok {
		return val
	}
	return NewNone()
}
func (k TypeMap) GetPair(key d.Typed) TypePair {
	if val, ok := k[key.Flag()]; ok {
		return NewTypePair(key, val)
	}
	return NewTypePair(None, NewNone())
}
func (k TypeMap) Pairs() []TypePair {
	var pairs = make([]TypePair, 0, len(k))
	for k, v := range k {
		pairs = append(pairs, NewTypePair(k, v))
	}
	return pairs
}
