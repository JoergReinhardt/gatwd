package types

func newSlice(val ...Value) *collection {
	l := make([]Value, 0, len(val))
	l = append(l, val...)
	return &collection{l}
}
func newTypedSlice(t Typed, val ...Value) *flatTypedSlice {
	for _, v := range val {
		t = t.Type().concat(v.Type())
	}
	return &flatTypedSlice{t.Type(), newSlice(val...)}
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

// internal collection instance, base of almost all collection implementations

// VALUE

// ACCESSABLE SLICE
func (s collection) get(i int) Value              { return s.s[i] }
func (s collection) Get(attr IndexAt) Value       { return s.s[attr.Idx()] }
func (s *collection) set(i int, v Value)          { (*s).s[i] = v }
func (s *collection) Set(attr IndexAt, val Value) { (*s).s[attr.Idx()] = val }

// ITERATOR
func (s collection) Next() (v Value, i Iterable) {
	if len(s.s) > 0 {
		v = s.s[0]
		if len(s.s) > 1 {
			i = &collection{s.s[:1]}
		}
	}
	return v, i
}

// BOOTOM & TOP
func (s collection) First() Value {
	if s.Len() > 0 {
		return s.s[0]
	}
	return s
}
func (s collection) Last() Value {
	if s.Len() > 0 {
		return s.s[s.Len()-1]
	}
	return s
}

// LIFO QUEUE
func (s *collection) Put(v Value) {
	if len(s.s) == cap(s.s) {
		(*s).s = append(append(make([]Value, 0, len(s.s)*2), s.s...), v)
	}
	(*s).s = append(s.s, v)
}
func (s *collection) Append(v ...Value) {
	if len(s.s) == cap(s.s) {
		(*s).s = append(append(make([]Value, 0, (len(s.s)+len(v))), s.s...), v...)
	}
	(*s).s = append(s.s, v...)
}
func (s *collection) Pull() (v Value) {
	if s.Len() > 0 {
		(*s).s, v = (*s).s[:s.Len()-1], (*s).s[s.Len()-1]
	}
	return v
}

// FIFO STACK
func (s *collection) Add(v ...Value) {
	if len(s.s) == cap(s.s)+len(v) {
		(*s).s = append(append(make([]Value, 0, len(v)+len(s.s)), v...), s.s...)
	}
	(*s).s = append(v, s.s...)
}
func (s *collection) Push(v Value) {
	if len(s.s) == cap(s.s) {
		(*s).s = append(append(make([]Value, 0, (len(s.s))*2), v), s.s...)
	}
	(*s).s = append([]Value{v}, s.s...)
}
func (s *collection) Pop() (v Value) {
	if s.Len() > 0 {
		v, (*s).s = s.s[0], s.s[1:]
	}
	return v
}

// ARITY

// TUPLE
func (s *collection) Head() (h Value)              { return (*s).s[0] }
func (s *collection) Tail() (c Value)              { return &collection{s.s[:1]} }
func (s *collection) HeadNary(arity int) (h Value) { return &collection{(*s).s[:arity]} }
func (s *collection) TailNary(arity int) (c Value) { return &collection{s.s[arity:]} }
func (s collection) Decap() (h Value, t Tupular) {
	return h, t
}
func (s *collection) DecapNary(arity int) (h *collection, t *collection) {
	if (*s).Len()+1 > arity {
		return &collection{(*s).s[:arity]}, &collection{(*s).s[arity:]}
	}
	return h, t
}

// SLICE
func (s collection) Len() int       { return len(s.s) }
func (s collection) Slice() []Value { return s.s }
func (s collection) Split(i int) (*collection, *collection) {
	h, t := s.s[:i], s.s[i:]
	return &collection{h}, &collection{t}
}
func (s *collection) Cut(i, j int) {
	copy((*s).s[i:], s.s[j:])
	for k, n := len(s.s)-j+i, len(s.s); k < n; k++ {
		(*s).s[k] = nil // <- prevents possib. mem leak
	}
	(*s).s = s.s[:len(s.s)-j+i]
}
func (s *collection) Delete(i int) {
	copy((*s).s[i:], s.s[i+1:])
	(*s).s[len(s.s)-1] = nil
	(*s).s = s.s[:len(s.s)-1]
}
func (s *collection) Insert(i int, v Value) {
	(*s).s = append((*s).s, nilVal{})
	copy(s.s[i+1:], s.s[i:])
	(*s).s[i] = v
}
func (s *collection) InsertVariadic(i int, v ...Value) {
	(*s).s = append((*s).s[:i], append(v, s.s[i:]...)...)
}
func (s collection) AttrType() flag { return Int.Type() }

//// typed slice embeds slice and only needs its own methods implemented
// internal typed slice instance, embeds the base slice and adds type flag to
// keep track of content types
type flatTypedSlice struct {
	t flag
	*collection
}

func (s flatTypedSlice) UnaryTyped() bool    { return s.t.match(Unarys) }
func (s *flatTypedSlice) AttrType() flag     { return s.t.Type() }
func (s *flatTypedSlice) Get(i int) Value    { return s.s[i] }
func (s *flatTypedSlice) Set(i int, v Value) { (*s).s[i] = v }
func (s *flatTypedSlice) Put(v Value) {
	(*s).t = s.t.concat(v.Type())
	(*s).s = append(s.s, v)
}
func (s *flatTypedSlice) Append(v ...Value) {
	(*s).collection.Append(v...)
	for _, val := range v {
		(*s).t = s.t.concat(val.Type())
	}
}
func (s *flatTypedSlice) Push(v Value) {
	(*s).t = s.t.concat(v.Type())
	(*s).collection.Push(v)
}
func (s *flatTypedSlice) Add(v ...Value) {
	for _, val := range v {
		(*s).t = s.t.concat(val.Type())
	}
	(*s).collection.Add(v...)
}
func (s flatTypedSlice) MultiTyped() bool {
	if s.t.Type().count() > 1 {
		return true
	}
	return false
}
