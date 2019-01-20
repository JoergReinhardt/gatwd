package functions

import (
	"math/big"
	"sort"
	"strings"

	d "github.com/JoergReinhardt/godeep/data"
)

func compareSymbolic(a, b Symbolic) int {
	return strings.Compare(a.String(), b.String())
}
func compareRational(a, b d.RatioVal) int {
	return ((*big.Rat)(&a)).Cmp((*big.Rat)(&b))
}
func compareUnsigned(a, b Unsigned) int {
	if a.Uint() < b.Uint() {
		return -1
	}
	if a.Uint() > b.Uint() {
		return 1
	}
	return 0
}
func compareInteger(a, b Integer) int {
	if a.Int() < b.Int() {
		return -1
	}
	if a.Int() > b.Int() {
		return 1
	}
	return 0
}
func compareIrrational(a, b Irrational) int {
	if a.Float() < b.Float() {
		return -1
	}
	if a.Float() > b.Float() {
		return 1
	}
	return 0
}
func compareFlag(a, b d.BitFlag) int {
	if a.Flag() < b.Flag() {
		return -1
	}
	if a.Flag() > b.Flag() {
		return 1
	}
	return 0
}
func compInt2BooIncl(i int) bool { return i >= 0 }
func compInt2BooExcl(i int) bool { return i > 0 }

type dataSorter []Data

func (d dataSorter) Empty() bool {
	if len(d) > 0 {
		for _, dat := range d {
			if !elemEmpty(dat) {
				return false

			}
		}
	}
	return true
}
func (d dataSorter) Len() int      { return len(d) }
func (d dataSorter) Swap(i, j int) { d[i], d[j] = d[j], d[i] }
func (ds dataSorter) Sort(argType d.Type) {
	less := newDataLess(argType, ds)
	sort.Slice(ds, less)
}

func (ds dataSorter) Search(praed Data) int {
	var idx = sort.Search(len(ds), newDataFind(ds, praed))
	if idx < len(ds) {
		switch {
		case praed.Flag().Match(d.Symbolic.Flag()):
			if ds[idx].(Symbolic).String() == praed.(Symbolic).String() {
				return idx
			}
		case praed.Flag().Match(d.Flag.Flag()):
			if ds[idx].Flag().Match(praed.Flag()) {
				return idx
			}
		case praed.Flag().Match(d.Integer.Flag()):
			if ds[idx].(Integer).Int() == praed.(Integer).Int() {
				return idx
			}
		case praed.Flag().Match(d.Irrational.Flag()):
			if ds[idx].(Irrational).Float() == praed.(Irrational).Float() {
				return idx
			}
		case praed.Flag().Match(d.Unsigned.Flag()):
			if ds[idx].(Unsigned).Uint() == praed.(Unsigned).Uint() {
				return idx
			}
		}
	}
	return -1
}

func newDataLess(argType d.Type, ds dataSorter) func(i, j int) bool {
	var f = argType.Flag()
	switch {
	case f.Match(d.Symbolic.Flag()):
		return func(i, j int) bool {
			return compInt2BooExcl(
				compareSymbolic( // Symbolic
					ds[j], ds[i]))
		}
	case f.Match(d.Flag.Flag()):
		return func(i, j int) bool {
			return ds[j].Flag() == ds[i].Flag()
		}
	case f.Match(d.Unsigned.Flag()):
		return func(i, j int) bool {
			return ds[j].(Unsigned) == ds[i].(Unsigned)
		}
	case f.Match(d.Integer.Flag()):
		return func(i, j int) bool {
			return ds[j].(Integer) == ds[i].(Integer)
		}
	case f.Match(d.Irrational.Flag()):
		return func(i, j int) bool {
			return ds[j].(Irrational) == ds[i].(Irrational)
		}
	}
	return nil
}
func newDataFind(ds dataSorter, praed Data) func(int) bool {
	var f = praed.Flag()
	switch {
	case f.Match(d.Symbolic.Flag()):
		return func(i int) bool {
			return compareSymbolic(ds[i].(Symbolic), praed.(Symbolic)) >= 0

		}
	case f.Match(d.Flag.Flag()):
		return func(i int) bool {
			return ds[i].Flag() >= praed.Flag()

		}
	case f.Match(d.Unsigned.Flag()):
		return func(i int) bool {
			return ds[i].(Unsigned).Uint() >= praed.(Unsigned).Uint()

		}
	case f.Match(d.Integer.Flag()):
		return func(i int) bool {
			return ds[i].(Integer).Int() >= praed.(Integer).Int()

		}
	case f.Match(d.Irrational.Flag()):
		return func(i int) bool {
			return ds[i].(Irrational).Float() >= praed.(Irrational).Float()

		}
	}
	return nil
}

// pair sorter has the methods to search for a pair in-/, and sort slices of
// pairs. pairs will be sorted by the left parameter, since it references the
// accessor (key) in an accessor/value pair.
type pairSorter []Paired

func newPairSorter(p ...Paired) pairSorter { return append(pairSorter{}, p...) }
func (a pairSorter) Empty() bool {
	if len(a) > 0 {
		for _, p := range a {
			if !elemEmpty(p) {
				return false
			}
		}
	}
	return true
}
func (p pairSorter) Len() int      { return len(p) }
func (p pairSorter) Swap(i, j int) { p[j], p[i] = p[i], p[j] }
func (p pairSorter) Sort(f d.Type) {
	less := newPraedLess(p, f)
	sort.Slice(p, less)
}
func (p pairSorter) Search(praed Data) int {
	var idx = sort.Search(len(p), newPraedFind(p, praed))
	// when praedicate is a precedence type encoding bit-flag
	if praed.Flag().Match(d.Flag.Flag()) {
		if d.Type(p[idx].Left().Flag()) == praed {
			return idx
		}
	}
	// otherwise check if key is equal to praedicate
	if idx < len(p) {
		if p[idx].Left() == praed {
			return idx
		}
	}
	return -1
}
func (p pairSorter) Get(praed Data) Paired {
	idx := p.Search(praed)
	if idx != -1 {
		return p[idx]
	}
	return nil
}
func (p pairSorter) Range(praed Data) []Paired {
	var ran = []Paired{}
	idx := p.Search(praed)
	if idx != -1 {
		for pair := p[idx]; pair != nil; {
			ran = append(ran, pair)
		}
	}
	return ran
}

func newPraedLess(accs pairSorter, t d.Type) func(i, j int) bool {
	f := t.Flag()
	switch {
	case f.Match(d.Symbolic.Flag()):
		return func(i, j int) bool {
			chain := accs
			if strings.Compare(
				chain[i].(Paired).Left().String(),
				chain[j].(Paired).Left().String(),
			) <= 0 {
				return true
			}
			return false
		}
	case f.Match(d.Flag.Flag()):
		return func(i, j int) bool { // sort by value-, NOT accessor type
			chain := accs
			if chain[i].(Paired).Right().Flag() <=
				chain[j].(Paired).Right().Flag() {
				return true
			}
			return false
		}
	case f.Match(d.Unsigned.Flag()):
		return func(i, j int) bool {
			chain := accs
			if chain[i].(Paired).Left().(Unsigned).Uint() <=
				chain[j].(Paired).Left().(Unsigned).Uint() {
				return true
			}
			return false
		}
	case f.Match(d.Integer.Flag()):
		return func(i, j int) bool {
			chain := accs
			if chain[i].(Paired).Left().(Integer).Int() <=
				chain[j].(Paired).Left().(Integer).Int() {
				return true
			}
			return false
		}
	case f.Match(d.Irrational.Flag()):
		return func(i, j int) bool {
			chain := accs
			if chain[i].(Paired).Left().(Irrational).Float() <=
				chain[j].(Paired).Left().(Irrational).Float() {
				return true
			}
			return false
		}
	}
	return nil
}
func newPraedFind(accs pairSorter, praed Data) func(i int) bool {
	var f = praed.Flag()
	var fn func(i int) bool
	switch { // parameters are accessor/value pairs to be applyed.
	case f.Match(d.Unsigned.Flag()):
		fn = func(i int) bool {
			return uint(accs[i].(Paired).Left().(Unsigned).Uint()) >
				uint(praed.(Unsigned).Uint())
		}
	case f.Match(d.Integer.Flag()):
		fn = func(i int) bool {
			return int(accs[i].(Paired).Left().(Integer).Int()) >
				int(praed.(Integer).Int())
		}
	case f.Match(d.Irrational.Flag()):
		fn = func(i int) bool {
			return int(accs[i].(Paired).Left().(Irrational).Float()) >
				int(praed.(Irrational).Float())
		}
	case f.Match(d.Symbolic.Flag()):
		fn = func(i int) bool {
			return strings.Compare(
				accs[i].(Paired).Left().String(),
				praed.String()) >= 0
		}
	case f.Match(d.Flag.Flag()):
		fn = func(i int) bool {
			return accs[i].(Paired).Right().(d.BitFlag) >
				praed.(d.BitFlag)
		}
	}
	return fn
}
