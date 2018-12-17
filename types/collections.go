package types

import (
	"strconv"
)

func newSlice(val ...Value) *slice {
	l := make([]Value, 0, len(val))
	l = append(l, val...)
	return &slice{l}
}
func newTypedSlice(t Type, val ...Value) *typedSlice {
	for _, v := range val {
		t = t.Concat(v.Type().Flag())
	}
	return &typedSlice{t.Flag(), newSlice(val...)}
}

// internal slice instance, base of almost all collection implementations
type slice struct {
	s []Value
}

// VALUE
func (s slice) Type() (t Type)       { return Arry.Type() }
func (s slice) Value() (v Value)     { return v }
func (s slice) Ref() (r interface{}) { return r }
func (s slice) String() (str string) {
	for i, v := range s.s {
		str = str + "\t" + strconv.Itoa(i) + "\t" + v.String() + "\n"
	}
	return str
}
func (s slice) Copy() (v Value) {
	sl := []Value{}
	for _, val := range s.s {
		sl = append(sl, val.Copy())
	}
	v = &slice{sl}
	return v
}

// SLICE
func (s slice) Len() int       { return len(s.s) }
func (s slice) Slice() []Value { return s.s }
func (s slice) AttrType() Type { return Int.Type() }

// MUTABLE SLICE
func (s slice) get(i int) Value            { return s.s[i] }
func (s slice) Get(acc IdxAcc) Value       { return s.s[acc.Idx()] }
func (s *slice) set(i int, v Value)        { (*s).s[i] = v }
func (s *slice) Set(acc IdxAcc, val Value) { (*s).s[acc.Idx()] = val }

// ITERATOR
func (s slice) Next() (v Value, i Iterable) {
	if len(s.s) > 0 {
		v = s.s[0]
		if len(s.s) > 1 {
			i = &slice{s.s[:1]}
		}
	}
	return v, i
}

// BOOTOM & TOP
func (s slice) First() Value {
	if s.Len() > 0 {
		return s.s[0]
	}
	return s
}
func (s slice) Last() Value {
	if s.Len() > 0 {
		return s.s[s.Len()-1]
	}
	return s
}

// LIFO QUEUE
func (s *slice) Put(v Value) {
	if len(s.s) == cap(s.s) {
		(*s).s = append(append(make([]Value, 0, len(s.s)*2), s.s...), v)
	}
	(*s).s = append(s.s, v)
}
func (s *slice) Append(v ...Value) {
	if len(s.s) == cap(s.s) {
		(*s).s = append(append(make([]Value, 0, (len(s.s)+len(v))), s.s...), v...)
	}
	(*s).s = append(s.s, v...)
}
func (s *slice) Pull() (v Value) {
	if s.Len() > 0 {
		(*s).s, v = (*s).s[:s.Len()-1], (*s).s[s.Len()-1]
	}
	return v
}

// FIFO STACK
func (s *slice) Add(v ...Value) {
	if len(s.s) == cap(s.s)+len(v) {
		(*s).s = append(append(make([]Value, 0, len(v)+len(s.s)), v...), s.s...)
	}
	(*s).s = append(v, s.s...)
}
func (s *slice) Push(v Value) {
	if len(s.s) == cap(s.s) {
		(*s).s = append(append(make([]Value, 0, (len(s.s))*2), v), s.s...)
	}
	(*s).s = append([]Value{v}, s.s...)
}
func (s *slice) Pop() (v Value) {
	if s.Len() > 0 {
		v, (*s).s = s.s[0], s.s[1:]
	}
	return v
}

// TUPLE
func (s slice) Decap() (h Value, t Tupled) {
	if s.Len() > 0 {
		h, t = s.s[0], &slice{s.s[1:]}
	}
	return h, t
}
func (s slice) Head() (h Value)  { return s.s[0] }
func (s slice) Tail() (c Tupled) { return newTupled(s.s[:1]...) }

// ARITY
func (s slice) Arity() int    { return ary(s) }
func (s slice) Empty() bool   { return empty(s) }
func (s slice) Split() Tupled { return Tupled(s) }

//// typed slice embeds slice and only needs its own methods implemented
// internal typed slice instance, embeds the base slice and adds type flag to
// keep track of content types
type typedSlice struct {
	t Flag
	*slice
}

func (s typedSlice) UnaryTyped() bool    { return s.t.Match(Unary.Flag()) }
func (s *typedSlice) AttrType() Type     { return s.t.Type() }
func (s *typedSlice) Get(i int) Value    { return s.s[i] }
func (s *typedSlice) Set(i int, v Value) { (*s).s[i] = v }
func (s *typedSlice) Put(v Value) {
	(*s).t = s.t.Concat(v.Type().Flag())
	(*s).s = append(s.s, v)
}
func (s *typedSlice) Append(v ...Value) {
	for _, val := range v {
		(*s).t = s.t.Concat(val.Type().Flag())
		(*s).s = append(s.s, val)
	}
}
func (s *typedSlice) Push(v Value) {
	(*s).t = s.t.Concat(v.Type().Flag())
	(*s).s = append(s.s, v)
}
func (s *typedSlice) Add(v ...Value) {
	for _, val := range v {
		(*s).t = s.t.Concat(val.Type().Flag())
	}
	(*s).s = append(v, s.s...)
}
func (s typedSlice) MultiTyped() bool {
	if s.t.Flag().Count() > 1 {
		return true
	}
	return false
}

//// SLICE HELPERS ////
func ary(v Value) int {
	if v, ok := v.(Value); ok {
		if v.Type().Match(Unary.Flag()) {
			return 1
		}
		if l, ok := v.(Size); ok {
			if l.Len() > 0 {
				return l.Len()
			}
		}
	}
	return 0
}
func flat(v Value) bool {
	if v, ok := v.(Value); ok {
		if v.Type().Match(Unary.Flag()) {
			return true
		}
		if a, ok := v.(Array); ok {
			for _, f := range a.Slice() {
				if !flat(f) {
					return false
				}
			}
		}
	}
	return true
}
func empty(v Value) bool {
	if v, ok := v.(Value); ok {
		if v.Type().Match(Unary.Flag()) {
			if v.Value() != nil {
				return true
			}
			if v.Type().Match(Nil.Flag()) {
				return true
			}
			return false
		}
		if l, ok := v.(Size); ok {
			if l.Len() > 0 {
				if a, ok := v.(Array); ok {
					for _, v := range a.Slice() {
						if !empty(v) {
							return false
						}
					}
				}
			}
		}
	}
	return true
}
