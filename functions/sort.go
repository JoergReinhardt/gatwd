/*
SORT & SEARCH

  implements golang sort and search slices of data and pairs of data. since
  'data' can be of collection type, this implements search and sort for pretty
  much every type thinkable of. generalizes over contained types by using gatwds
  capabilitys.
*/
package functions

import (
	"fmt"
	"math/big"
	"sort"
	"strings"

	d "github.com/joergreinhardt/gatwd/data"
)

// type class based comparison functions
func CompareText(a, b Text) int {
	return strings.Compare(a.String(), b.String())
}
func CompareRational(a, b d.RatioVal) int {
	return ((*big.Rat)(&a)).Cmp((*big.Rat)(&b))
}
func CompareNatural(a, b Natural) int {
	if a.Uint() < b.Uint() {
		return -1
	}
	if a.Uint() > b.Uint() {
		return 1
	}
	return 0
}
func CompareInteger(a, b Integer) int {
	if a.Int() < b.Int() {
		return -1
	}
	if a.Int() > b.Int() {
		return 1
	}
	return 0
}
func CopareReal(a, b Real) int {
	if a.Float() < b.Float() {
		return -1
	}
	if a.Float() > b.Float() {
		return 1
	}
	return 0
}
func CompareFlag(a, b d.BitFlag) int {
	if a.TypeNat() < b.TypeNat() {
		return -1
	}
	if a.TypeNat() > b.TypeNat() {
		return 1
	}
	return 0
}
func compInt2BooIncl(i int) bool { return i >= 0 }
func compInt2BooExcl(i int) bool { return i > 0 }

////////////////////////////////////////////////////////////////////////////
// type to sort slices of data
type dataSorter []Callable

func (d dataSorter) Empty() bool {
	if len(d) > 0 {
		for _, dat := range d {
			if dat != nil {
				return false

			}
		}
	}
	return true
}

func newDataSorter(dat ...Callable) dataSorter { return dataSorter(dat) }

func (d dataSorter) Len() int { return len(d) }

func (d dataSorter) Swap(i, j int) { d[i], d[j] = d[j], d[i] }

func (ds dataSorter) Sort(argType d.TyNative) {
	sort.Slice(ds, newDataLess(argType, ds))
}

func (ds dataSorter) Search(pred Callable) int {
	var idx = sort.Search(len(ds), newDataFind(ds, pred))
	if idx < len(ds) {
		if strings.Compare(ds[idx].String(), pred.String()) == 0 {
			return idx
		}
	}
	return -1
}

func newDataLess(argType d.TyNative, ds dataSorter) func(i, j int) bool {
	var f = argType.TypeNat().Flag()
	switch {
	case f.Match(d.Letters.TypeNat()):
		return func(i, j int) bool {
			if strings.Compare(ds[j].String(), ds[i].String()) >=
				0 {
				return true
			}
			return false
		}

	case f.Match(d.Flag.TypeNat()):
		return func(i, j int) bool {
			return ds[j].Eval().TypeNat() >=
				ds[i].Eval().TypeNat()
		}

	case f.Match(d.Naturals.TypeNat()):
		return func(i, j int) bool {
			return ds[j].Eval().(Natural).Uint() >=
				ds[i].Eval().(Natural).Uint()
		}

	case f.Match(d.Integers.TypeNat()):
		return func(i, j int) bool {
			return ds[j].Eval().(Integer).Int() >=
				ds[i].Eval().(Integer).Int()
		}

	case f.Match(d.Reals.TypeNat()):
		return func(i, j int) bool {
			return ds[j].Eval().(Real).Float() >=
				ds[i].Eval().(Real).Float()
		}
	}
	return nil
}

func newDataFind(ds dataSorter, pred Callable) func(int) bool {

	// preallocate function
	var fn func(int) bool
	// predicate native type flag
	var f = pred.Eval().TypeNat().Flag()

	switch {
	case f.Match(d.Letters.TypeNat()):
		fn = func(i int) bool {
			fmt.Printf("%s %s", ds[i], pred.String())
			return strings.Compare(
				ds[i].(d.Native).String(),
				pred.(d.Native).String(),
			) >= 0
		}

	case f.Match(d.Flag.TypeNat()):
		fn = func(i int) bool {
			return ds[i].(d.Native).TypeNat() >=
				pred.(d.Native).TypeNat()
		}

	case f.Match(d.Naturals.TypeNat()):
		fn = func(i int) bool {
			return ds[i].(Natural).Uint() >=
				pred.(Natural).Uint()
		}

	case f.Match(d.Integers.TypeNat()):
		fn = func(i int) bool {
			return ds[i].(Integer).Int() >=
				pred.(Integer).Int()
		}

	case f.Match(d.Reals.TypeNat()):
		fn = func(i int) bool {
			return ds[i].(Real).Float() >=
				pred.(Real).Float()
		}
	}
	return fn
}

// pair sorter has the methods to search for a pair in-/, and sort slices of
// pairs. pairs will be sorted by the left parameter, since it references the
// accessor (key) in an accessor/value pair.
type pairSorter []PairFnc

func newPairSorter(p ...PairFnc) pairSorter {
	return append(pairSorter{}, p...)
}

func (a pairSorter) ValueSorter() pairSorter {
	return NewRecord(a...).SwitchedPairs()
}

func (a pairSorter) AppendKeyValue(key Callable, val Callable) {
	a = append(a, NewPair(key, val))
}

func (a pairSorter) Empty() bool {
	if len(a) > 0 {
		for _, p := range a {
			if p != nil {
				return false
			}
		}
	}
	return true
}
func (p pairSorter) Len() int      { return len(p) }
func (p pairSorter) Swap(i, j int) { p[j], p[i] = p[i], p[j] }

func (p pairSorter) Sort(f d.TyNative) {
	less := newpredLess(p, f)
	sort.Slice(p, less)
}

func (p pairSorter) SortByValue(f d.TyNative) {
	var ps = pairSorter(
		NewRecord(
			p...,
		).SwitchedPairs(),
	)

	ps.Sort(f)

	p = NewRecord(ps...).SwitchedPairs()
}

func (p pairSorter) Search(pred Callable) int {
	var idx = sort.Search(len(p), newpredFind(p, pred))
	// when predicate is a precedence type encoding bit-flag
	if idx != -1 {
		if pred.TypeNat().Flag().Match(d.Flag.TypeNat()) {
			if p[idx].Left().TypeNat() == pred.TypeNat() {
				return idx
			}
		}
		// otherwise check if key is equal to predicate
		if idx < len(p) {
			if p[idx].Left().Eval() == pred.Eval() {
				return idx
			}
		}
	}
	return -1
}

func (p pairSorter) SearchByValue(pred Callable) int {
	return pairSorter(
		NewRecord(p...).SwitchedPairs(),
	).Search(pred)
}

func (p pairSorter) Get(pred Callable) PairFnc {
	idx := p.Search(pred)
	if idx != -1 {
		return p[idx]
	}
	return NewPair(New(d.NilVal{}), New(d.NilVal{}))
}

func (p pairSorter) GetByValue(pred Callable) PairFnc {
	return pairSorter(
		NewRecord(p...).SwitchedPairs(),
	).Get(pred)
}

func (p pairSorter) Range(pred Callable) []PairFnc {
	var ran = []PairFnc{}
	idx := p.Search(pred)
	if idx != -1 {
		for pair := p[idx]; pair != nil; {
			ran = append(ran, pair)
		}
	}
	return ran
}

func (p pairSorter) RangeByValue(pred Callable) []PairFnc {
	return pairSorter(
		NewRecord(p...).SwitchedPairs(),
	).Range(pred)
}

func newpredLess(accs pairSorter, t d.TyNative) func(i, j int) bool {
	f := t.TypeNat().Flag()
	switch {
	case f.Match(d.Letters.TypeNat()):
		return func(i, j int) bool {
			chain := accs
			if strings.Compare(
				chain[i].Left().Eval().String(),
				chain[j].Left().Eval().String(),
			) <= 0 {
				return true
			}
			return false
		}

	case f.Match(d.Flag.TypeNat()):
		return func(i, j int) bool { // sort by value-, NOT accessor type
			chain := accs
			if chain[i].Right().Eval().TypeNat() <=
				chain[j].Right().Eval().TypeNat() {
				return true
			}
			return false
		}

	case f.Match(d.Naturals.TypeNat()):
		return func(i, j int) bool {
			chain := accs
			if chain[i].Left().Eval().(Natural).Uint() <=
				chain[j].Left().Eval().(Natural).Uint() {
				return true
			}
			return false
		}

	case f.Match(d.Integers.TypeNat()):
		return func(i, j int) bool {
			chain := accs
			if chain[i].Left().Eval().(Integer).Int() <=
				chain[j].Left().Eval().(Integer).Int() {
				return true
			}
			return false
		}

	case f.Match(d.Reals.TypeNat()):
		return func(i, j int) bool {
			chain := accs
			if chain[i].Left().Eval().(Real).Float() <=
				chain[j].Left().Eval().(Real).Float() {
				return true
			}
			return false
		}
	}
	return nil
}

func newpredFind(accs pairSorter, pred Callable) func(i int) bool {
	var f = pred.Eval().TypeNat().Flag()
	var fn func(i int) bool
	switch { // parameters are accessor/value pairs to be applyed.

	case f.Match(d.Naturals.TypeNat()):
		fn = func(i int) bool {
			return uint(accs[i].Left().Eval().(Natural).Uint()) >=
				uint(pred.Eval().(Natural).Uint())
		}

	case f.Match(d.Integers.TypeNat()):
		fn = func(i int) bool {
			return int(accs[i].Left().Eval().(Integer).Int()) >=
				int(pred.Eval().(Integer).Int())
		}

	case f.Match(d.Reals.TypeNat()):
		fn = func(i int) bool {
			return int(accs[i].Left().Eval().(Real).Float()) >=
				int(pred.Eval().(Real).Float())
		}

	case f.Match(d.Letters.TypeNat()):
		fn = func(i int) bool {
			return strings.Compare(
				accs[i].Left().Eval().String(),
				pred.Eval().String()) >= 0
		}

	case f.Match(d.Flag.TypeNat()):
		fn = func(i int) bool {
			return accs[i].Right().Eval().(d.BitFlag) >=
				pred.Eval().(d.BitFlag)
		}
	}
	return fn
}
