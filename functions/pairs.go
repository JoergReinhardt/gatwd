package functions

import (
	"fmt"

	d "github.com/joergreinhardt/gatwd/data"
)

type (
	//// VALUE PAIRS
	ValPair   func(...Expression) (Expression, Expression)
	NatPair   func(...Expression) (d.Native, Expression)
	KeyPair   func(...Expression) (string, Expression)
	IndexPair func(...Expression) (int, Expression)
	RealPair  func(...Expression) (float64, Expression)

	//// COLLECTIONS OF VALUE PAIRS
	KeyMap  map[string]Expression
	RealMap map[float64]Expression
)

///////////////////////////////////////////////////////////////////////////////
//// PAIRS OF VALUES
///
// pairs can be created empty, key & value may be constructed later
func NewEmptyPair() ValPair {
	return func(args ...Expression) (a, b Expression) {
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
func NewPair(l, r Expression) ValPair {
	return func(args ...Expression) (Expression, Expression) {
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

func (p ValPair) Cons(arg Expression) Grouped {
	if p.Empty() {
		return NewPair(arg, NewNone())
	}
	if p.Right().Type().Match(None) {
		return NewPair(p.Left(), arg)
	}
	if p.Left().Type().Match(None) {
		return NewPair(arg, p.Right())
	}
	return NewPair(p, arg)
	return p
}
func (p ValPair) Continue() (Expression, Grouped) {
	var l, r = p()
	if !l.Type().Match(None) {
		if !r.Type().Match(None) {
			if r.Type().Match(Continua) {
				return l, r.(Grouped)
			}
		}
	}
	return NewNone(), NewEmptyPair()
}
func (p ValPair) Head() Expression           { return p.Key() }
func (p ValPair) Tail() Grouped              { return NewList(p.Value()) }
func (p ValPair) Concat(c Continued) Grouped { return NewList(p, c) }

func (p ValPair) Pair() Paired                   { return p }
func (p ValPair) Both() (Expression, Expression) { return p() }
func (p ValPair) Swap() (Expression, Expression) { l, r := p(); return r, l }
func (p ValPair) Left() Expression               { l, _ := p(); return l }
func (p ValPair) Right() Expression              { _, r := p(); return r }
func (p ValPair) SwappedPair() Paired            { return NewPair(p.Right(), p.Left()) }
func (p ValPair) Slice() []Expression {
	return []Expression{p.Left(), p.Right()}
}
func (p ValPair) Key() Expression   { return p.Left() }
func (p ValPair) Value() Expression { return p.Right() }
func (p ValPair) TypeFnc() TyFnc    { return Pair }
func (p ValPair) TypeElem() TyDef {
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
func (p ValPair) Type() TyDef {
	return Def(Pair, Def(p.TypeKey(), p.TypeValue()))
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
func (p ValPair) Call(args ...Expression) Expression {
	if len(args) > 0 {
		return NewPair(p.Key(), p.Value().Call(args...))
	}
	return p
}

//// NATIVE VALUE KEY PAIR
///
//
func NewNatPair(key d.Native, val Expression) NatPair {
	return func(...Expression) (d.Native, Expression) { return key, val }
}

func (a NatPair) KeyNat() d.Native               { key, _ := a(); return key }
func (a NatPair) Value() Expression              { _, val := a(); return val }
func (a NatPair) Left() Expression               { return Box(a.KeyNat()) }
func (a NatPair) Right() Expression              { return a.Value() }
func (a NatPair) Both() (Expression, Expression) { return a.Left(), a.Right() }
func (a NatPair) Pair() Paired                   { return NewPair(a.Both()) }
func (a NatPair) Pairs() []Paired                { return []Paired{NewPair(a.Both())} }
func (a NatPair) Key() Expression                { return a.Left() }
func (a NatPair) TypeValue() d.Typed             { return a.Value().Type() }
func (a NatPair) TypeKey() d.Typed               { return a.KeyNat().Type() }
func (a NatPair) TypeFnc() TyFnc                 { return Data | Pair }
func (p NatPair) Type() TyDef                    { return Def(Pair, Def(Key, p.TypeValue())) }
func (p NatPair) Call(args ...Expression) Expression {
	if len(args) > 0 {
		return NewNatPair(p.KeyNat(), p.Value().Call(args...))
	}
	return p
}

// implement swappable
func (p NatPair) Swap() (Expression, Expression) {
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

func (p NatPair) Cons(args ...Expression) Expression {
	return NewVector(Map(TakeN(NewVector(args...), 2),
		func(arg Expression) Expression {
			var (
				left       d.Native
				right      Expression
				head, tail = arg.(Grouped).Continue()
			)
			if head.Type().Match(Data) {
				left = head.(NatEval).Eval()
			} else {
				left = Nat(head.String())
			}
			right, tail = tail.Continue()
			return NewNatPair(left, right)
		}))
}
func (p NatPair) Continue() (Expression, Grouped) {
	return p.Key(), NewList(p.Value())
}
func (p NatPair) Head() Expression { return p.Key() }
func (p NatPair) Tail() Grouped    { return NewList(p.Value()) }

//// STRING KEY PAIR
///
// pair composed of a string key and a functional value
func NewKeyPair(key string, val Expression) KeyPair {
	return func(...Expression) (string, Expression) { return key, val }
}

func (a KeyPair) KeyStr() string                 { key, _ := a(); return key }
func (a KeyPair) Value() Expression              { _, val := a(); return val }
func (a KeyPair) Left() Expression               { return Box(d.StrVal(a.KeyStr())) }
func (a KeyPair) Right() Expression              { return a.Value() }
func (a KeyPair) Both() (Expression, Expression) { return a.Left(), a.Right() }
func (a KeyPair) Pair() Paired                   { return NewPair(a.Both()) }
func (a KeyPair) Pairs() []Paired                { return []Paired{NewPair(a.Both())} }
func (a KeyPair) Key() Expression                { return a.Left() }
func (a KeyPair) TypeValue() d.Typed             { return a.Value().Type() }
func (a KeyPair) TypeElem() d.Typed              { return a.Value().Type() }
func (a KeyPair) TypeKey() d.Typed               { return Key }
func (a KeyPair) TypeFnc() TyFnc                 { return Key | Pair }
func (p KeyPair) Call(args ...Expression) Expression {
	if len(args) > 0 {
		return NewKeyPair(p.KeyStr(), p.Value().Call(args...))
	}
	return p
}
func (p KeyPair) Type() TyDef {
	return Def(Key|Pair, Def(Key, p.TypeValue()))
}

// implement swappable
func (p KeyPair) Swap() (Expression, Expression) {
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
func (p KeyPair) Cons(args ...Expression) Expression {
	return NewVector(Map(TakeN(NewVector(args...), 2),
		func(arg Expression) Expression {
			var left, tail = arg.(Grouped).Continue()
			return NewKeyPair(left.String(), tail)
		}))
}
func (p KeyPair) Continue() (Expression, Grouped) {
	return p.Key(), NewList(p.Value())
}
func (p KeyPair) Head() Expression { return p.Key() }
func (p KeyPair) Tail() Grouped    { return NewList(p.Value()) }

//// INDEX PAIR
///
// pair composed of an integer and a functional value
func NewIndexPair(idx int, val Expression) IndexPair {
	return func(...Expression) (int, Expression) { return idx, val }
}
func (a IndexPair) Value() Expression              { _, val := a(); return val }
func (a IndexPair) Index() int                     { idx, _ := a(); return idx }
func (a IndexPair) KeyIdx() int                    { return a.Index() }
func (a IndexPair) Left() Expression               { return Box(d.IntVal(a.Index())) }
func (a IndexPair) Right() Expression              { return a.Value() }
func (a IndexPair) Both() (Expression, Expression) { return a.Left(), a.Right() }
func (a IndexPair) Pair() Paired                   { return a }
func (a IndexPair) Pairs() []Paired                { return []Paired{NewPair(a.Both())} }
func (a IndexPair) Key() Expression                { return a.Left() }
func (a IndexPair) TypeFnc() TyFnc                 { return Index | Pair }
func (a IndexPair) TypeKey() d.Typed               { return Index }
func (a IndexPair) TypeValue() d.Typed             { return a.Value().Type() }
func (a IndexPair) Type() TyDef                    { return Def(Pair, Def(Index, a.TypeValue())) }
func (p IndexPair) Call(args ...Expression) Expression {
	if len(args) > 0 {
		return NewIndexPair(p.Index(), p.Value().Call(args...))
	}
	return p
}

// implement swappable
func (p IndexPair) Swap() (Expression, Expression) {
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

func (p IndexPair) Cons(args ...Expression) Expression {
	return NewVector(Map(TakeN(NewVector(args...), 2),
		func(arg Expression) Expression {
			var (
				idx        = 0
				head, tail = arg.(Grouped).Continue()
			)
			if head.Type().Match(Data) {
				var nat = head.(NatEval).Eval()
				if nat.Type().Match(d.Numbers) {
					idx = int(nat.(d.Numeral).Int())
				}
			}
			return NewIndexPair(idx, tail)
		}))
}
func (p IndexPair) Continue() (Expression, Grouped) {
	return p.Key(), NewList(p.Value())
}
func (p IndexPair) Head() Expression { return p.Key() }
func (p IndexPair) Tail() Grouped    { return NewList(p.Value()) }

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
func (a RealPair) Type() TyDef                        { return Def(Pair, Def(Real, a.TypeValue())) }

// implement swappable
func (p RealPair) Swap() (Expression, Expression) {
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
func (p RealPair) Cons(args ...Expression) Expression {
	return NewVector(Map(TakeN(NewVector(args...), 2),
		func(arg Expression) Expression {
			var (
				idx        = 0.0
				head, tail = arg.(Grouped).Continue()
			)
			if head.Type().Match(Data) {
				var nat = head.(NatEval).Eval()
				if nat.Type().Match(d.Numbers) {
					idx = float64(nat.(d.Numeral).Float())
				}
			}
			return NewRealPair(idx, tail)
		}))
}
func (p RealPair) Continue() (Expression, Grouped) {
	return p.Key(), NewList(p.Value())
}
func (p RealPair) Head() Expression { return p.Key() }
func (p RealPair) Tail() Grouped    { return NewList(p.Value()) }

///////////////////////////////////////////////////////////////////////////////
//// KEY MAP
///
func NewKeyMap(pairs ...KeyPair) KeyMap {
	var m = map[string]Expression{}
	for _, pair := range pairs {
		m[pair.KeyStr()] = pair.Value()
	}
	return m
}
func (k KeyMap) Type() TyDef                        { return Def(Key, HashMap) }
func (k KeyMap) TypeFnc() TyFnc                     { return Key | HashMap }
func (k KeyMap) Call(args ...Expression) Expression { return k }
func (k KeyMap) String() string {
	var str = "{\n}"
	for k, v := range k {
		str = str + k + " ∷ " + v.String() + "\n"
	}
	str = str + "}"
	return str
}
func (k KeyMap) Get(key string) Expression {
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
	var m = map[float64]Expression{}
	for _, pair := range pairs {
		m[pair.Real()] = pair.Value()
	}
	return m
}
func (k RealMap) Type() TyDef                        { return Def(Real, HashMap) }
func (k RealMap) TypeFnc() TyFnc                     { return Real | HashMap }
func (k RealMap) Call(args ...Expression) Expression { return k }
func (k RealMap) String() string {
	var str = "{\n}"
	for k, v := range k {
		str = str + fmt.Sprintf("%f", k) + " ∷ " +
			v.String() + "\n"
	}
	str = str + "}"
	return str
}
func (k RealMap) Get(key float64) Expression {
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
