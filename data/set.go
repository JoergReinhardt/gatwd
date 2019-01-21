package data

func NewPair(l, r Data) Paired { return PairVal{l, r} }

// implements Paired flagged Pair
func (p PairVal) String() string {
	return p.Left().String() + ": " + p.Right().String()
}
func (p PairVal) Left() Data         { return p.l }
func (p PairVal) Right() Data        { return p.r }
func (p PairVal) Both() (Data, Data) { return p.l, p.r }

// implements Mapped flagged Set

func NewSet(acc ...Paired) Mapped {
	var m = make(map[StrVal]Data)
	for _, pair := range acc {
		m[pair.Left().(StrVal)] = pair.Right()
	}
	return SetVal(m)
}

func (s SetVal) Flag() BitFlag { return Set.Flag() }
func (s SetVal) Keys() []Data {
	var keys = []Data{}
	for k, _ := range s {
		keys = append(keys, k)
	}
	return keys
}
func (s SetVal) Data() []Data {
	var dat = []Data{}
	for _, d := range s {
		dat = append(dat, d)
	}
	return dat
}
func (s SetVal) Accs() []Paired {
	var pairs = []Paired{}
	for k, d := range s {
		pairs = append(pairs, PairVal(PairVal{k, d}))
	}
	return pairs
}
func (s SetVal) Get(acc Data) Data             { return s[acc.(StrVal)] }
func (s SetVal) Set(acc Data, dat Data) Mapped { s[acc.(StrVal)] = acc.(StrVal); return s }
