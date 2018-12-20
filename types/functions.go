package types

// ACCESSABLE SLICE
func get(s slice, i int) Value        { return s[i] }
func Get(s slice, attr IndexAt) Value { return s[attr.Idx()] }
func set(s slice, i int, v Value)     { s[i] = v }

func Set(s slice, attr IndexAt, val Value) { s[attr.Idx()] = val }

// ITERATOR
func Next(s slice) (v Value, i Iterable) {
	if len(s) > 0 {
		v = s[0]
		if len(s) > 1 {
			i = &collection{s[:1]}
		}
	}
	return v, i
}

// BOOTOM & TOP
func First(s slice) Value {
	if s.Len() > 0 {
		return s[0]
	}
	return s
}
func Last(s slice) Value {
	if s.Len() > 0 {
		return s[s.Len()-1]
	}
	return s
}

// LIFO QUEUE
func Put(s slice, v Value) {
	if len(s) == cap(s) {
		s = append(append(make([]Value, 0, len(s)*2), s...), v)
	}
	s = append(s, v)
}
func Append(s slice, v ...Value) {
	if len(s) == cap(s) {
		s = append(append(make([]Value, 0, (len(s)+len(v))), s...), v...)
	}
	s = append(s, v...)
}
func Pull(s slice) (v Value) {
	if s.Len() > 0 {
		s, v = s[:s.Len()-1], s[s.Len()-1]
	}
	return v
}

// FIFO STACK
func Add(s slice, v ...Value) {
	if len(s) == cap(s)+len(v) {
		s = append(append(make([]Value, 0, len(v)+len(s)), v...), s...)
	}
	s = append(v, s...)
}
func Push(s slice, v Value) {
	if len(s) == cap(s) {
		s = append(append(make([]Value, 0, (len(s))*2), v), s...)
	}
	s = append([]Value{v}, s...)
}
func Pop(s slice) (v Value) {
	if s.Len() > 0 {
		v, s = s[0], s[1:]
	}
	return v
}

// ARITY

// TUPLE
func Head(s slice) (h Value)                { return s[0] }
func Tail(s slice) (c Value)                { return &collection{s[:1]} }
func HeadNary(s slice, arity int) (h Value) { return &collection{s[:arity]} }
func TailNary(s slice, arity int) (c Value) { return &collection{s[arity:]} }
func Decap(s slice) (h Value, t Tupular) {
	return h, t
}
func DecapNary(s slice, arity int) (h *collection, t *collection) {
	if s.Len()+1 > arity {
		return &collection{s[:arity]}, &collection{s[arity:]}
	}
	return h, t
}

// SLICE
func Len(s slice) int { return len(s) }
func Split(s slice, i int) (*collection, *collection) {
	h, t := s[:i], s[i:]
	return &collection{h}, &collection{t}
}
func Cut(s slice, i, j int) {
	copy(s[i:], s[j:])
	for k, n := len(s)-j+i, len(s); k < n; k++ {
		s[k] = nil // <- prevents possib. mem leak
	}
	s = s[:len(s)-j+i]
}
func Delete(s slice, i int) {
	copy(s[i:], s[i+1:])
	s[len(s)-1] = nil
	s = s[:len(s)-1]
}
func Insert(s slice, i int, v Value) {
	s = append(s, nilVal{})
	copy(s[i+1:], s[i:])
	s[i] = v
}
func InsertVariadic(s slice, i int, v ...Value) {
	s = append(s[:i], append(v, s[i:]...)...)
}
func AttrType(s slice) flag { return Int.Type() }
