package types

// DESTRUCTABLE SLICE

func conChain(val ...Data) chain {
	l := make([]Data, 0, len(val))
	l = append(l, val...)
	return l
}

func chainClear(s chain) {
	if len(s) > 0 {
		for i, v := range s {
			if !fmatch(v.Flag(), Nullable.Flag()) {
				if d, ok := v.(Destructable); ok {
					d.Clear()
				}
			}
			s[i] = nil
		}
	}
	s = nil
}

// COLLECTION
func elemEmpty(d Data) bool {
	// not flagged nil, not a composition either...
	if !fmatch(d.Flag(), (Nil.Flag() | Composed.Flag())) {
		if d != nil { // not a nil pointer...
			// --> not empty
			return false
		}
	}
	// since it's a composition, inspect...
	if fmatch(d.Flag(), Composed.Flag()) {
		// slice --> call sliceEmpty
		if sl, ok := d.(chain); ok {
			return chainEmpty(sl)
		}
		// other sort of collection...
		if col, ok := d.(Collected); ok {
			// --> call it's empty method
			return col.Empty()
		}
	}
	// no idea, what this is, so better call it empty
	return true
}
func (s chain) Empty() bool { return chainEmpty(s) }
func chainEmpty(s chain) bool {
	if len(s) == 0 { // empty, as in no element...
		return true
	}
	if len(s) > 0 { // empty as in contains empty elements exclusively...
		for _, elem := range chainSlice(s) { // return at first non empty
			if !elemEmpty(elem) {
				return false
			}
		}
	} // --> all contained elements are empty
	return true
}

// ACCESSABLE SLICE
func chainGetInt(s chain, i int) Data { return s[i] }

// MUTABLE SLICE
func chainSetInt(s chain, i int, v Data) chain { s[i] = v; return s }

// ITERATOR
func chainNext(s chain) (v Data, i chain) {
	if len(s) > 0 {
		if len(s) > 1 {
			return s[0], s[1:]
		}
		return s[0], chain([]Data{nilVal{}})
	}
	return nilVal{}, chain([]Data{nilVal{}})
}

// BOOTOM & TOP
func chainFirst(s chain) Data {
	if s.Len() > 0 {
		return s[0]
	}
	return s
}
func chainLast(s chain) Data {
	if s.Len() > 0 {
		return s[s.Len()-1]
	}
	return s
}

// LIFO QUEUE
func chainPut(s chain, v Data) chain {
	return append(s, v)
}
func chainAppend(s chain, v ...Data) chain {
	return append(s, v...)
}
func chainPull(s chain) (Data, chain) {
	if s.Len() > 0 {
		return s[s.Len()-1], s[:s.Len()-1]
	}
	return nilVal{}, s
}

// FIFO STACK
func slideAdd(s chain, v ...Data) chain {
	return append(v, s...)
}
func slidePush(s chain, v Data) chain {
	return append([]Data{v}, s...)
}
func slidePop(s chain) (Data, chain) {
	if len(s) > 0 {
		return s[0], s[1:]
	}
	return nilVal{}, s
}

/////
func chainAdd(s chain, v ...Data) chain {
	if len(s) >= cap(s)+len(v)/2 {
		return append(append(make([]Data, 0, len(v)+len(s)), v...), s...)
	}
	return append(v, s...)
}
func chainPush(s chain, v Data) chain {
	if len(s) >= cap(s)/2 {
		return append(append(make([]Data, 0, (len(s))*2), v), s...)
	}
	return append([]Data{v}, s...)
}
func chainPop(s chain) (Data, chain) {
	if len(s) > 0 {
		return s[0], s[1:]
	}
	return nilVal{}, s
}

// ARITY

// TUPLE
func (s chain) Head() (h Data)         { return s[0] }
func (s chain) Tail() (c Consumeable)  { return s[:1] }
func (s chain) Shift() (c Consumeable) { return s[:1] }
func chainHead(s chain) (h Data)       { return s[0] }
func chainTail(s chain) (c []Data)     { return s[:1] }
func chainDecap(s chain) (h Data, t chain) {
	if !chainEmpty(s) {
		return s[0], t[:1]
	}
	return nilVal{}, conChain(nilVal{})
}

// SLICE
func chainSlice(s chain) []Data { return []Data(s) }
func chainLen(s chain) int      { return len(s) }
func chainSplit(s chain, i int) (chain, chain) {
	h, t := s[:i], s[i:]
	return h, t
}
func chainCut(s chain, i, j int) chain {
	copy(s[i:], s[j:])
	// to prevent a possib. mem leak
	for k, n := len(s)-j+i, len(s); k < n; k++ {
		s[k] = nil
	}
	return s[:len(s)-j+i]
}
func chainDelete(s chain, i int) chain {
	copy(s[i:], s[i+1:])
	s[len(s)-1] = nil
	return s[:len(s)-1]
}
func chainInsert(s chain, i int, v Data) chain {
	s = append(s, nilVal{})
	copy(s[i+1:], s[i:])
	s[i] = v
	return s
}
func chainInsertVector(s chain, i int, v ...Data) chain {
	return append(s[:i], append(v, s[i:]...)...)
}
func chainAttrType(s chain) BitFlag { return Int.Flag() }
