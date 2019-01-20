package data

type pair [2]Data

func NewPair(l, r Data) Paired { return pair([2]Data{l, r}) }

// implements Paired flagged Pair
func (p pair) Flag() BitFlag { return Pair.Flag() }
func (p pair) String() string {
	return p.Left().String() + ": " + p.Right().String()
}
func (p pair) Left() Data         { return p[0] }
func (p pair) Right() Data        { return p[1] }
func (p pair) Both() (Data, Data) { return p[0], p[1] }

// implements Mapped flagged Set
type set map[StrVal]Data

func NewSet(acc ...Paired) Mapped {
	var m = make(map[StrVal]Data)
	for _, pair := range acc {
		m[pair.Left().(StrVal)] = pair.Right()
	}
	return set(m)
}

func (s set) Flag() BitFlag { return Set.Flag() }
func (s set) Keys() []Data {
	var keys = []Data{}
	for k, _ := range s {
		keys = append(keys, k)
	}
	return keys
}
func (s set) Data() []Data {
	var dat = []Data{}
	for _, d := range s {
		dat = append(dat, d)
	}
	return dat
}
func (s set) Accs() []Paired {
	var pairs = []Paired{}
	for k, d := range s {
		pairs = append(pairs, pair([2]Data{k, d}))
	}
	return pairs
}
func (s set) Get(acc Data) Data             { return s[acc.(StrVal)] }
func (s set) Set(acc Data, dat Data) Mapped { s[acc.(StrVal)] = acc.(StrVal); return s }
