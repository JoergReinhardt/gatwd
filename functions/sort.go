/*
SORT & SEARCH

  implements golang sort and search slices of data and pairs of data. since
  'data' can be of collection type, this implements search and sort for pretty
  much every type thinkable of. generalizes over contained types by using gatwds
  capabilitys.
*/
package functions

//import (
//	"fmt"
//	"sort"
//	"strings"
//
//	d "github.com/joergreinhardt/gatwd/data"
//)

////////////////////////////////////////////////////////////////////////////
// type to sort slices of data
//type SortedData []Expression
//
//func (d SortedData) Empty() bool {
//	if len(d) > 0 {
//		for _, dat := range d {
//			if dat != nil {
//				return false
//
//			}
//		}
//	}
//	return true
//}
//
//func Sort(dat ...Expression) SortedData { return SortedData(dat) }
//
//func (d SortedData) Len() int { return len(d) }
//
//func (d SortedData) Swap(i, j int) { d[i], d[j] = d[j], d[i] }
//
//func (ds SortedData) Sort(argType TyFnc) {
//	sort.Slice(ds, ConsLess(argType, ds))
//}
//
//func (ds SortedData) Search(pred Expression) int {
//	var idx = sort.Search(len(ds), ConsFind(ds, pred))
//	if idx < len(ds) {
//		if strings.Compare(ds[idx].String(), pred.String()) == 0 {
//			return idx
//		}
//	}
//	return -1
//}
//
////func ConsLess(argType TyFnc, ds SortedData) func(i, j int) bool {
////	var f = argType.TypeNat().Flag()
////	switch {
////	case f.Match(d.Letters.TypeNat()):
////		return func(i, j int) bool {
////			if strings.Compare(ds[j].String(), ds[i].String()) >=
////				0 {
////				return true
////			}
////			return false
////		}
////
////		//		//	case f.Match(d.Type.TypeNat()):
////		//		//		return func(i, j int) bool {
////		//		//			return ds[j].Eval().TypeNat() >=
////		//		//				ds[i].Eval().TypeNat()
////		//		//		}
////
////		//	case f.Match(d.Naturals.TypeNat()):
////		//		return func(i, j int) bool {
////		//			return ds[j].Eval().(Natural).GoUint() >=
////		//				ds[i].Eval().(Natural).GoUint()
////		//		}
////
////	case f.Match(d.Integers.TypeNat()):
////		return func(i, j int) bool {
////			return ds[j].Eval().(Integer).Idx() >=
////				ds[i].Eval().(Integer).Idx()
////		}
////
////		////	case f.Match(d.Reals.TypeNat()):
////		////		return func(i, j int) bool {
////		////			return ds[j].Eval().(Real).GoFlt() >=
////		////				ds[i].Eval().(Real).GoFlt()
////		////		}
////	}
////	return nil
////}
////
////func ConsFind(ds SortedData, pred Expression) func(int) bool {
////
////	// preallocate function
////	var fn func(int) bool
////	// predicate native type flag
////	var f = pred.Eval().TypeNat().Flag()
////
////	switch {
////	case f.Match(d.Letters.TypeNat()):
////		fn = func(i int) bool {
////			fmt.Printf("%s %s", ds[i], pred.String())
////			return strings.Compare(
////				ds[i].(d.Native).String(),
////				pred.(d.Native).String(),
////			) >= 0
////		}
////
////	case f.Match(d.Type.TypeNat()):
////		fn = func(i int) bool {
////			return ds[i].(d.Native).TypeNat() >=
////				pred.(d.Native).TypeNat()
////		}
////
////	case f.Match(d.Naturals.TypeNat()):
////		fn = func(i int) bool {
////			return ds[i].(Natural).GoUint() >=
////				pred.(Natural).GoUint()
////		}
////
////	case f.Match(d.Integers.TypeNat()):
////		fn = func(i int) bool {
////			return ds[i].(Integer).Idx() >=
////				pred.(Integer).Idx()
////		}
////
////	case f.Match(d.Reals.TypeNat()):
////		fn = func(i int) bool {
////			return ds[i].(Real).GoFlt() >=
////				pred.(Real).GoFlt()
////		}
////	}
////	return fn
////}
////
////// pair sorter has the methods to search for a pair in-/, and sort slices of
////// pairs. pairs will be sorted by the left parameter, since it references the
////// accessor (key) in an accessor/value pair.
//////type SortedPairs []Paired
//////
//////func SortPairs(p ...Paired) SortedPairs {
//////	return append(SortedPairs{}, p...)
//////}
//////
//////func (a SortedPairs) ValueSorter() SortedPairs {
//////	return NewPairVectorFromPairs(a...).SwitchedPairs()
//////}
//////
//////func (a SortedPairs) AppendKeyValue(key Expression, val Expression) {
//////	a = append(a, NewPair(key, val))
//////}
//////
//////func (a SortedPairs) Empty() bool {
//////	if len(a) > 0 {
//////		for _, p := range a {
//////			if p != nil {
//////				return false
//////			}
//////		}
//////	}
//////	return true
//////}
//////func (p SortedPairs) Len() int      { return len(p) }
//////func (p SortedPairs) Swap(i, j int) { p[j], p[i] = p[i], p[j] }
//////
//////func (p SortedPairs) Sort(f d.TyNat) {
//////	less := ConsPairLess(p, f)
//////	sort.Slice(p, less)
//////}
//////
//////func (p SortedPairs) SortByValue(f d.TyNat) {
//////	var ps = SortedPairs(
//////		NewPairVectorFromPairs(
//////			p...,
//////		).SwitchedPairs(),
//////	)
//////
//////	ps.Sort(f)
//////
//////	p = NewPairVectorFromPairs(ps...).SwitchedPairs()
//////}
//////
//////func (p SortedPairs) Search(pred Expression) int {
//////	var idx = sort.Search(len(p), ConsPairFind(p, pred))
//////	// when predicate is a precedence type encoding bit-flag
//////	if idx != -1 {
//////		if pred.TypeNat().Flag().Match(d.Type.TypeNat()) {
//////			if p[idx].Left().TypeNat() == pred.TypeNat() {
//////				return idx
//////			}
//////		}
//////		// otherwise check if key is equal to predicate
//////		if idx < len(p) {
//////			if p[idx].Left().Eval() == pred.Eval() {
//////				return idx
//////			}
//////		}
//////	}
//////	return -1
//////}
//////
//////func (p SortedPairs) SearchByValue(pred Expression) int {
//////	return SortedPairs(
//////		NewPairVectorFromPairs(p...).SwitchedPairs(),
//////	).Search(pred)
//////}
//////
//////func (p SortedPairs) Get(pred Expression) Paired {
//////	idx := p.Search(pred)
//////	if idx != -1 {
//////		return p[idx]
//////	}
//////	return NewPair(New(d.NewNil()), New(d.NewNil()))
//////}
//////
//////func (p SortedPairs) GetByValue(pred Expression) Paired {
//////	return SortedPairs(
//////		NewPairVectorFromPairs(p...).SwitchedPairs(),
//////	).Get(pred)
//////}
//////
//////func (p SortedPairs) Range(pred Expression) []Paired {
//////	var ran = []Paired{}
//////	idx := p.Search(pred)
//////	if idx != -1 {
//////		for pair := p[idx]; pair != nil; {
//////			ran = append(ran, pair)
//////		}
//////	}
//////	return ran
//////}
//////
//////func (p SortedPairs) RangeByValue(pred Expression) []Paired {
//////	return SortedPairs(
//////		NewPairVectorFromPairs(p...).SwitchedPairs(),
//////	).Range(pred)
//////}
//////
//////func ConsPairLess(accs SortedPairs, t d.TyNat) func(i, j int) bool {
//////	f := t.TypeNat().Flag()
//////	switch {
//////	case f.Match(d.Letters.TypeNat()):
//////		return func(i, j int) bool {
//////			chain := accs
//////			if strings.Compare(
//////				chain[i].Left().Eval().String(),
//////				chain[j].Left().Eval().String(),
//////			) <= 0 {
//////				return true
//////			}
//////			return false
//////		}
//////
//////	case f.Match(d.Type.TypeNat()):
//////		return func(i, j int) bool { // sort by value-, NOT accessor type
//////			chain := accs
//////			if chain[i].Right().Eval().TypeNat() <=
//////				chain[j].Right().Eval().TypeNat() {
//////				return true
//////			}
//////			return false
//////		}
//////
//////	case f.Match(d.Naturals.TypeNat()):
//////		return func(i, j int) bool {
//////			chain := accs
//////			if chain[i].Left().Eval().(Natural).GoUint() <=
//////				chain[j].Left().Eval().(Natural).GoUint() {
//////				return true
//////			}
//////			return false
//////		}
//////
//////	case f.Match(d.Integers.TypeNat()):
//////		return func(i, j int) bool {
//////			chain := accs
//////			if chain[i].Left().Eval().(Integer).Idx() <=
//////				chain[j].Left().Eval().(Integer).Idx() {
//////				return true
//////			}
//////			return false
//////		}
//////
//////	case f.Match(d.Reals.TypeNat()):
//////		return func(i, j int) bool {
//////			chain := accs
//////			if chain[i].Left().Eval().(Real).GoFlt() <=
//////				chain[j].Left().Eval().(Real).GoFlt() {
//////				return true
//////			}
//////			return false
//////		}
//////	}
//////	return nil
//////}
//////
//////func ConsPairFind(accs SortedPairs, pred Expression) func(i int) bool {
//////	var f = pred.Eval().TypeNat().Flag()
//////	var fn func(i int) bool
//////	switch { // parameters are accessor/value pairs to be applyed.
//////
//////	case f.Match(d.Naturals.TypeNat()):
//////		fn = func(i int) bool {
//////			return uint(accs[i].Left().Eval().(Natural).GoUint()) >=
//////				uint(pred.Eval().(Natural).GoUint())
//////		}
//////
//////	case f.Match(d.Integers.TypeNat()):
//////		fn = func(i int) bool {
//////			return int(accs[i].Left().Eval().(Integer).Idx()) >=
//////				int(pred.Eval().(Integer).Idx())
//////		}
//////
//////	case f.Match(d.Reals.TypeNat()):
//////		fn = func(i int) bool {
//////			return int(accs[i].Left().Eval().(Real).GoFlt()) >=
//////				int(pred.Eval().(Real).GoFlt())
//////		}
//////
//////	case f.Match(d.Letters.TypeNat()):
//////		fn = func(i int) bool {
//////			return strings.Compare(
//////				accs[i].Left().Eval().String(),
//////				pred.Eval().String()) >= 0
//////		}
//////
//////	case f.Match(d.Type.TypeNat()):
//////		fn = func(i int) bool {
//////			return accs[i].Right().Eval().(d.BitFlag) >=
//////				pred.Eval().(d.BitFlag)
//////		}
//////	}
//////	return fn
//////}
