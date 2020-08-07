package gatw

import (
	"strconv"
	"strings"
)

// COMPOSED TYPES {{{
func ConUid(args ...Id) Uid { return args }
func ConIds(args ...int) (u Uid) {
	u = make([]Id, 0, len(args))
	for _, i := range args {
		u = append(u, Id(i))
	}
	return u
}

func (u Uid) Identity() Item { return u }
func (u Uid) Type() Type     { return Number }
func (u Uid) Compare(args Uid) Rank {
	for n, arg := range args {
		if n < len(u) {
			if u[n] < arg {
				return Lesser
			}
			if u[n] > arg {
				return Greater
			}
		}
	}
	return Equal
}
func (u Uid) Len() int { return len(u) }
func (u Uid) Shift() (Id, Uid) {
	if u.Len() > 0 {
		if u.Len() > 1 {
			return u[0], u[1:]
		}
		return u[0], u[:0]
	}
	return 0, u[:0]
}
func (u Uid) Head() Id  { h, _ := u.Shift(); return h }
func (u Uid) Tail() Uid { _, t := u.Shift(); return t }
func (u Uid) Cons(args ...Item) (Item, Cons) {
	if len(args) > 0 {
		var ids = make([]Id, 0, len(args))
		for _, arg := range args {
			ids = append(ids, arg.Type().Id())
		}
	}
	return u, u.Cons
}
func (u Uid) String() string {
	var strs = make([]string, 0, u.Len())
	for _, i := range u {
		strs = append(strs, strconv.Itoa(i.Int()))
	}
	return strings.Join(strs, ".")
}

// TYPE SIGNATURE PATTERN
type Comp int8

//go:generate stringer -type Comp
const (
	Sum     Comp = 0
	Product Comp = 1
)

func ConPattern(c Comp, t ...Type) Pattern {
	return Pattern(func() ([]Type, Comp) { return t, c })
}

func (t Pattern) Pattern() []Type { p, _ := t(); return p }
func (t Pattern) Composal() Comp  { _, c := t(); return c }
func (t Pattern) Identity() Item  { return t }
func (t Pattern) Type() Type      { return t } // tuple type
func (t Pattern) Len() int        { return len(t.Pattern()) }
func (t Pattern) Id() Id          { return Tuple.Id() }
func (t Pattern) Name() Key       { return Key(t.String()) }
func (t Pattern) Head() Type      { var head, _ = t.Shift(); return head }
func (t Pattern) Tail() Pattern   { var _, tail = t.Shift(); return tail }
func (t Pattern) Shift() (Type, Pattern) {
	var l = t.Len()
	if l > 0 {
		if l > 1 {
			return t.Pattern()[0],
				ConPattern(t.Composal(),
					t.Pattern()[1:]...)
		}
		return t.Pattern()[0], ConPattern(t.Composal(),
			t.Pattern()[:0]...)
	}
	return None, t
}
func (t Pattern) Get(i int) Type {
	if i < t.Len() {
		return t.Pattern()[i]
	}
	return None
}
func (t Pattern) String() string {
	var str = make([]string, 0, t.Len())
	for _, typ := range t.Pattern() {
		str = append(str, string(typ.Name()))
	}
	return "[" + strings.Join(str, "|") + "]"
}

// }}}
// PAIR {{{
func ConPair(l, r Item) PairVal {
	return func(args ...Item) (a, b Item) {
		if len(args) > 0 {
			if len(args) > 1 {
				if len(args) > 2 {
					return ConPair(
						ConPair(args[0], args[1]),
						ConPair(l, r))(args[2:]...)
				}
				if len(args) == 2 {
					return ConPair(args[0], args[1]),
						ConPair(l, r)
				}
			}
			return args[0], ConPair(l, r)
		}
		return l, r
	}
}
func (e PairVal) Identity() Item { return e }
func (e PairVal) Type() Type {
	if e.Empty() {
		return None
	}
	return Enum
}

func (p PairVal) Empty() bool {
	var h, t = p()
	return h.Type().Id() == 0 &&
		t.Type().Id() == 0
}
func Left(p PairVal) Item  { var l, _ = p(); return l }
func Right(p PairVal) Item { var _, r = p(); return r }

func (p PairVal) Left() Item          { return Left(p) }
func (p PairVal) Right() Item         { return Right(p) }
func (p PairVal) Fst() Item           { return p.Left() }
func (p PairVal) Snd() Item           { return p.Right() }
func (p PairVal) Head() Item          { return p.Left() }
func (p PairVal) Tail() Item          { return p.Right() }
func (p PairVal) Both() (Item, Item)  { return p.Left(), p.Right() }
func (p PairVal) Sequence() SeqVal    { return ConSequence(p.Left(), p.Right()) }
func (p PairVal) Expression() EnumVal { return ConEnum(p.Left(), p.Right()) }
func (p PairVal) Cons(args ...Item) Item {
	if p.Empty() {
		if len(args) > 0 {
			if len(args) > 1 {
				if len(args) > 2 {
				}
			}
		}
	}
	if p.Right().Type() == None {
		if len(args) > 0 {
			if len(args) > 1 {
				if len(args) > 2 {
				}
			}
		}
	}
	if len(args) > 0 {
		if len(args) > 1 {
			if len(args) > 2 {
				return ConPair(p.Cons(
					args[0], args[1],
				),
					p.Cons(args[1:]...),
				)
			}
			return ConPair(p, ConPair(
				args[0], args[1]))
		}
		return ConPair(p, ConPair(
			args[0], None))
	}
	return p
}

//}}}

// ENUMERATION{{{
// FREE ENUMERATION FUNCTIONS
func First(e EnumVal) Item {
	if len(e) > 0 {
		return e[0]
	}
	return ConEnum()
}
func Second(e EnumVal) Item {
	if len(e) > 1 {
		return e[1]
	}
	return ConEnum()
}
func Last(e EnumVal) Item {
	var l = len(e)
	if l > 0 {
		return e[l-1]
	}
	return ConEnum()
}
func ExprTail(e EnumVal) EnumVal {
	if len(e) > 0 {
		if len(e) > 1 {
			return e[1:]
		}
		return e[0:]
	}
	return e[:0]
}
func Prior(e EnumVal) EnumVal {
	if len(e) > 1 {
		return ConEnum(e[1:]...)
	}
	return ConEnum()
}
func Next(e EnumVal) Item {
	if len(e) >= 1 {
		return e[1]
	}
	return ConEnum()
}
func Get(e EnumVal, i int) Item {
	if len(e) < i {
		return e[i]
	}
	return ConEnum()
}
func Rotate(es ...Item) EnumVal {
	return append(es[:1], es[0])
}
func Shift(es ...Item) (Item, EnumVal) {
	if len(es) > 1 {
		if len(es) > 1 {
			return es[0], es[:1]
		}
		return es[0], ConEnum()
	}
	return ConEnum(), ConEnum()
}
func Reverse(es ...Item) EnumVal {
	if len(es) > 0 {
		if len(es) > 1 {
			return ConEnum(es[0],
				Reverse(es[1:]...))
		}
		return ConEnum(es[0])
	}
	return []Item{}
}

// ENUM CONSTRUCTOR
func ConEnum(e ...Item) EnumVal { return e }

func (e EnumVal) Identity() Item            { return e }
func (e EnumVal) Shift() (Item, EnumVal)    { return Shift(e) }
func (e EnumVal) Empty() bool               { return len(e) == 0 }
func (e EnumVal) Id() Id                    { return e.Type().Id() }
func (e EnumVal) Len() int                  { return len(e) }
func (e EnumVal) Fst() Item                 { return First(e) }
func (e EnumVal) Snd() Item                 { return Second(e) }
func (e EnumVal) Last() Item                { return Last(e) }
func (e EnumVal) Next() Item                { return Next(e) }
func (e EnumVal) Head() Item                { return First(e) }
func (e EnumVal) Tail() EnumVal             { return ExprTail(e) }
func (e EnumVal) Prior() Item               { return Prior(e) }
func (e EnumVal) Reverse() Item             { return Reverse(e) }
func (e EnumVal) Cons(args ...Item) EnumVal { return append(e, args...) }
func (e EnumVal) Sequence() SeqVal          { return ConSequence(e...) }
func (e EnumVal) Pair() PairVal             { return ConPair(e.Head(), e.Tail()) }
func (e EnumVal) Get(i int) Item            { return Get(e, i) }
func (e EnumVal) GetElem(i int) Elem {
	return Elem(ConPair(Id(i), e.Get(i)))
}
func (e EnumVal) Elements() []Elem {
	var elems = make([]Elem, 0, e.Len())
	for i, v := range e {
		elems = append(elems, ConElem(i, v))
	}
	return elems
}
func (e EnumVal) Type() Type {
	if e.Empty() {
		return None
	}
	return Enum
}

// ENUMERATED ELEMENT
func ConElem(i int, e Item) Elem { return Elem(ConPair(Id(i), e)) }

func (e Elem) Identity() Item  { return e }
func (e Elem) Type() Type      { return Element }
func (e Elem) Id() Id          { return PairVal(e).Left().(Id) }
func (e Elem) Value() Item     { return PairVal(e).Right() }
func (e Elem) ValueType() Type { return PairVal(e).Right().Type() }

//}}}

// SEQUENCE {{{
func ConSequence(elems ...Item) SeqVal {
	return SeqVal(func(args ...Item) (Item, SeqVal) {
		if len(args) > 0 {
			elems = append(args, elems...)
		}
		if len(elems) > 0 {
			if len(elems) > 1 {
				return elems[0],
					ConSequence(elems[0:]...)
			}
			return elems[0], ConSequence()
		}
		return None, ConSequence()
	})
}
func (s SeqVal) Head() Item            { return Head(s) }
func (s SeqVal) Tail() SeqVal          { return Tail(s) }
func (s SeqVal) Fst() Item             { return s.Head() }
func (s SeqVal) Snd() Item             { return s.Tail().Head() }
func (s SeqVal) Shift() (Item, SeqVal) { return Head(s), Tail(s) }
func (s SeqVal) Left() Item            { return Head(s) }
func (s SeqVal) Right() Item           { return Tail(s) }
func (s SeqVal) Cons(args ...Item) Item {
	return SeqVal(func(args ...Item) (Item, SeqVal) {
		if len(args) > 0 {
			return s(args...)
		}
		return s()
	})
}
func (s SeqVal) Empty() bool {
	var head, tail = s()
	return head.Type() == None && tail.Empty()
}
func (s SeqVal) Type() Type {
	if s.Empty() {
		return None
	}
	return Enum
}
func (s SeqVal) Identity() Item { return s }
func (s SeqVal) Expression() EnumVal {
	var (
		head, tail = s()
		items      = []Item{head}
	)
	for head, tail = tail(); !tail.Empty(); {
		items = append(items, head)
	}
	return ConEnum(items...)
}
func (s SeqVal) Pair() PairVal {
	head, tail := s()
	return ConPair(head, tail)
}

// FREE SEQUENCE FUNCTIONS
func Head(s SeqVal) Item               { head, _ := s(); return head }
func Tail(s SeqVal) SeqVal             { _, tail := s(); return tail }
func Continue(s SeqVal) (Item, SeqVal) { item, tail := s(); return item, tail }

// }}}
// TUPLE {{{
func ConTuple(args ...Item) TplVal {
	var (
		pattern = make([]Type, 0, len(args))
		items   = make([]Item, 0, len(args))
	)
	for _, arg := range args {
		pattern = append(pattern, arg.Type())
		if arg.Type().Id() != 0 {
			items = append(items, arg)
		} else {
			items = append(items, None)
		}
	}
	return TplVal(ConPair(
		ConPattern(Sum, pattern...), ConEnum(args...)))
}
func (t TplVal) Identity() Item      { return t }
func (t TplVal) Expression() EnumVal { expr, _ := t(); return expr.(EnumVal) }
func (t TplVal) Signature() Pattern  { _, pat := t(); return pat.(Pattern) }
func (t TplVal) Len() int            { return t.Signature().Len() }
func (t TplVal) Type() Type          { return t.Signature() }
func (t TplVal) Get(i int) Item {
	if t.Len() > i {
		return t.Expression()[i]
	}
	return None
}
func (t TplVal) GetMember(i int) TupElem {
	if t.Len() > i {
		var item = t.Expression()[i]
		return ConTupElem(Id(i), item)
	}
	return ConTupElem(Id(-1), None)
}
func (t TplVal) Member() []TupElem {
	var member = make([]TupElem, 0, t.Len())
	for i, item := range t.Expression() {
		member = append(member, ConTupElem(Id(i), item))
	}
	return member
}
func (t TplVal) Name() Key {
	var names = make([]string, 0, t.Len())
	for _, t := range t.Signature() {
		names = append(names, string(t.Name()))
	}
	return Key(strings.Join(names, " "))
}

// TUPLE ELEMENT
func ConTupElem(i Id, e Item) TupElem { return TupElem(ConPair(i, e)) }

func (m TupElem) Identity() Item  { return m }
func (m TupElem) Type() Type      { return Member }
func (m TupElem) Id() Id          { return PairVal(m).Left().(Id) }
func (m TupElem) Value() Item     { return PairVal(m).Right() }
func (m TupElem) ValueType() Type { return PairVal(m).Right().Type() }

//}}}

// RECORD {{{
func ConRecord(ps ...PairVal) RecVal {
	var set = map[Key]Item{}
	return set
}

func (s RecVal) Identity() Item { return s }
func (s RecVal) Len() int       { return len(s) }
func (s RecVal) Type() Type     { return Record }
func (s RecVal) Empty() bool    { return len(s) == 0 }
func (s RecVal) Get(k string) Item {
	if i, ok := s[Key(k)]; ok {
		return i
	}
	return C(None)
}
func (s RecVal) GetMember(k string) RecField {
	return ConField(Key(k), s.Get(k))
}
func (s RecVal) Head() PairVal { return s.Pair().Left().(PairVal) }
func (s RecVal) Tail() RecVal  { return s.Pair().Right().(RecVal) }
func (s RecVal) Member() []RecField {
	var recs = make([]RecField, 0, s.Len())
	for k, v := range s {
		recs = append(recs, ConField(Key(k), v))
	}
	return recs
}
func (s RecVal) Fields() (f []PairVal) {
	f = make([]PairVal, 0, s.Len())
	for k, v := range s {
		f = append(f, ConPair(Key(k), v))
	}
	return f
}
func (s RecVal) Pair() PairVal {
	if s.Len() > 0 {
		var elems = s.Expression()
		if s.Len() > 1 {
			return ConPair(elems[0], EnumVal(elems[1:]))
		}
		return ConPair(elems[0], None)
	}
	return ConPair(None, None)
}
func (s RecVal) Expression() EnumVal {
	var items = make([]Item, 0, s.Len())
	for k, v := range s {
		items = append(items, ConPair(k, v))
	}
	return ConEnum(items...)
}
func (s RecVal) Sequence() SeqVal { return ConSequence(s.Expression()...) }
func (e RecVal) String() string {
	var str = make([]string, 0, e.Len())
	for k, _ := range e {
		str = append(str, string(k))
	}
	return strings.Join(str, "|")
}
func (e RecVal) Cons(args ...Item) Item {
	if len(args) > 0 {
		var (
			i   = 0
			arg = args[i]
		)
		if arg.Type() == Pair {
			for _, arg = range args {
				var p = arg.(PairVal)
				if p.Left().Type() == Keyword {
					e[p.Left().(Key)] = p.Right()
					return e
				}
			}
		}
	}
	return e
}

// record field
func ConField(k Key, i Item) RecField { return RecField(ConPair(k, i)) }

func (r RecField) Identity() Item  { return r }
func (r RecField) Type() Type      { return Field }
func (r RecField) Key() Key        { return PairVal(r).Left().(Key) }
func (r RecField) Value() Item     { return PairVal(r).Right() }
func (r RecField) ValueType() Type { return r.Value().Type() }

// }}}
