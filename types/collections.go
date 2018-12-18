package types

import (
	"strconv"
)

func newSlice(val ...Value) *slice {
	l := make([]Value, 0, len(val))
	l = append(l, val...)
	return &slice{l}
}
func newTypedSlice(t Type, val ...Value) *flatTypedSlice {
	for _, v := range val {
		t = t.Flag().Concat(v.Flag())
	}
	return &flatTypedSlice{t.Flag(), newSlice(val...)}
}
func guardArity(arity int, v ...Value) []Value {
	return v[:arity]
}
func nestSlice(arity int, v []Value) [][]Value {
	var acc [][]Value
	for len(v) > arity {
		v, acc = v[arity:], append(acc, v[0:arity:arity])
	}
	return append(acc, v)
}

// internal slice instance, base of almost all collection implementations
type slice struct {
	s []Value
}

// VALUE
func (s slice) Type() Type           { return Ordered.Type() }
func (s slice) Flag() Flag           { return Ordered.Flag() }
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

// ACCESSABLE SLICE
func (s slice) get(i int) Value                  { return s.s[i] }
func (s slice) Get(attr OrdinalAttr) Value       { return s.s[attr.Idx()] }
func (s *slice) set(i int, v Value)              { (*s).s[i] = v }
func (s *slice) Set(attr OrdinalAttr, val Value) { (*s).s[attr.Idx()] = val }

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

// ARITY

// TUPLE
func (s *slice) Head() (h Value)              { return (*s).s[0] }
func (s *slice) Tail() (c Value)              { return &slice{s.s[:1]} }
func (s *slice) HeadNary(arity int) (h Value) { return &slice{(*s).s[:arity]} }
func (s *slice) TailNary(arity int) (c Value) { return &slice{s.s[arity:]} }
func (s slice) Decap() (h Value, t Tupled) {
	return h, t
}
func (s *slice) DecapNary(arity int) (h *slice, t *slice) {
	if (*s).Len()+1 > arity {
		return &slice{(*s).s[:arity]}, &slice{(*s).s[arity:]}
	}
	return h, t
}

// SLICE
func (s slice) Len() int       { return len(s.s) }
func (s slice) Slice() []Value { return s.s }
func (s slice) Split(i int) (*slice, *slice) {
	h, t := s.s[:i], s.s[i:]
	return &slice{h}, &slice{t}
}
func (s *slice) Cut(i, j int) {
	copy((*s).s[i:], s.s[j:])
	for k, n := len(s.s)-j+i, len(s.s); k < n; k++ {
		(*s).s[k] = nil // <- prevents possib. mem leak
	}
	(*s).s = s.s[:len(s.s)-j+i]
}
func (s *slice) Delete(i int) {
	copy((*s).s[i:], s.s[i+1:])
	(*s).s[len(s.s)-1] = nil
	(*s).s = s.s[:len(s.s)-1]
}
func (s *slice) Insert(i int, v Value) {
	(*s).s = append((*s).s, NilVal{})
	copy(s.s[i+1:], s.s[i:])
	(*s).s[i] = v
}
func (s *slice) InsertVariadic(i int, v ...Value) {
	(*s).s = append((*s).s[:i], append(v, s.s[i:]...)...)
}
func (s slice) AttrType() Flag { return Int.Flag() }

//// typed slice embeds slice and only needs its own methods implemented
// internal typed slice instance, embeds the base slice and adds type flag to
// keep track of content types
type flatTypedSlice struct {
	t Flag
	*slice
}

func (s flatTypedSlice) UnaryTyped() bool    { return s.t.Match(Unary) }
func (s *flatTypedSlice) AttrType() Flag     { return s.t.Flag() }
func (s *flatTypedSlice) Get(i int) Value    { return s.s[i] }
func (s *flatTypedSlice) Set(i int, v Value) { (*s).s[i] = v }
func (s *flatTypedSlice) Put(v Value) {
	(*s).t = s.t.Concat(v.Flag())
	(*s).s = append(s.s, v)
}
func (s *flatTypedSlice) Append(v ...Value) {
	(*s).slice.Append(v...)
	for _, val := range v {
		(*s).t = s.t.Concat(val.Type())
	}
}
func (s *flatTypedSlice) Push(v Value) {
	(*s).t = s.t.Concat(v.Type())
	(*s).slice.Push(v)
}
func (s *flatTypedSlice) Add(v ...Value) {
	for _, val := range v {
		(*s).t = s.t.Concat(val.Type())
	}
	(*s).slice.Add(v...)
}
func (s flatTypedSlice) MultiTyped() bool {
	if s.t.Flag().Count() > 1 {
		return true
	}
	return false
}
