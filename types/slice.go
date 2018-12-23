package types

// DESTRUCTABLE SLICE

func Clear(s slice) {
	if len(s) > 0 {
		for i, v := range s {
			if !fmatch(v.Type(), Nullable) {
				if d, ok := v.(Destructable); ok {
					d.Clear()
				}
			}
			s[i] = nil
		}
	}
	s = nil
}

// ACCESSABLE SLICE
func get(s slice, i int) Evaluable        { return s[i] }
func Get(s slice, attr IndexAt) Evaluable { return s[attr.Idx()] }

// MUTABLE SLICE
func set(s slice, i int, v Evaluable) slice          { s[i] = v; return s }
func Set(s slice, attr IndexAt, val Evaluable) slice { s[attr.Idx()] = val; return s }

// ITERATOR
func Next(s slice) (v Evaluable, i slice) {
	if len(s) > 0 {
		if len(s) > 1 {
			return s[0], s[1:]
		}
		return s[0], slice([]Evaluable{nilVal{}})
	}
	return nilVal{}, slice([]Evaluable{nilVal{}})
}

// BOOTOM & TOP
func First(s slice) Evaluable {
	if s.Len() > 0 {
		return s[0]
	}
	return s
}
func Last(s slice) Evaluable {
	if s.Len() > 0 {
		return s[s.Len()-1]
	}
	return s
}

// LIFO QUEUE
func Put(s slice, v Evaluable) slice {
	if len(s) == cap(s) {
		return append(append(make([]Evaluable, 0, len(s)*2), s...), v)
	}
	return append(s, v)
}
func Append(s slice, v ...Evaluable) slice {
	if len(s) == cap(s) {
		return append(append(make([]Evaluable, 0, (len(s)+len(v))), s...), v...)
	}
	return append(s, v...)
}
func Pull(s slice) (Evaluable, slice) {
	if s.Len() > 0 {
		return s[s.Len()-1], s[:s.Len()-1]
	}
	return nilVal{}, s
}

// FIFO STACK
func Add(s slice, v ...Evaluable) slice {
	if len(s) == cap(s)+len(v) {
		return append(append(make([]Evaluable, 0, len(v)+len(s)), v...), s...)
	}
	return append(v, s...)
}
func Push(s slice, v Evaluable) slice {
	if len(s) == cap(s) {
		return append(append(make([]Evaluable, 0, (len(s))*2), v), s...)
	}
	return append([]Evaluable{v}, s...)
}
func Pop(s slice) (Evaluable, slice) {
	if s.Len() > 0 {
		return s[0], s[1:]
	}
	return nilVal{}, s
}

// ARITY

// TUPLE
func Head(s slice) (h Evaluable) { return s[0] }
func Tail(s slice) (c Evaluable) { return s[:1] }
func Decap(s slice) (h Evaluable, t Tupular) {
	return h, t
}

// N-TUPLE
func HeadNary(s slice, arity int) (h Evaluable) { return s[:arity] }
func TailNary(s slice, arity int) (c Evaluable) { return s[arity:] }
func DecapNary(s slice, arity int) (h Evaluable, t slice) {
	if s.Len()+1 > arity {
		return s[:arity], s[arity:]
	}
	return h, t
}

// SLICE
func Slice(s slice) []Evaluable { return []Evaluable(s) }
func Len(s slice) int           { return len(s) }
func Split(s slice, i int) (slice, slice) {
	h, t := s[:i], s[i:]
	return h, t
}
func Cut(s slice, i, j int) slice {
	copy(s[i:], s[j:])
	// to prevent a possib. mem leak
	for k, n := len(s)-j+i, len(s); k < n; k++ {
		s[k] = nil
	}
	return s[:len(s)-j+i]
}
func Delete(s slice, i int) slice {
	copy(s[i:], s[i+1:])
	s[len(s)-1] = nil
	return s[:len(s)-1]
}
func Insert(s slice, i int, v Evaluable) slice {
	s = append(s, nilVal{})
	copy(s[i+1:], s[i:])
	s[i] = v
	return s
}
func InsertVariadic(s slice, i int, v ...Evaluable) slice {
	return append(s[:i], append(v, s[i:]...)...)
}
func AttrType(s slice) flag { return Int.Type() }
