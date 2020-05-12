package gatw

import (
	"math/bits"
)

type Cat uint8

//go:generate stringer -type Cat
const (
	N Cat = 1<<iota - 1
	Type
	Func
	Symb

	C = N | Type | Func | Symb
)

func (c Cat) Uniq()      {}
func (c Cat) uint() uint { return uint(c) }
func (c Cat) Id() int    { return bits.Len8(uint8(c)) }
func (c Cat) Flag() uint { return uint(c) }
func (c Cat) Kind() Elem { return c }

func iTf(e Elem) uint    { return uint(1 << uint(e.Id())) }
func ofKind(e Elem) bool { return C.uint()&^iTf(e) != 0 }

// category monad initialization
func initCat() Cons {

	var ( // enclosed values
		econ, pcon Cons // sub type constructor
		cat        = make([]Cons, 4, 4)
		e          Elem
	)

	///////////////////////////////////
	// define category type constructor
	econ = func(es ...Elem) (Elem, Cons) {
		if len(es) > 0 {
			e = es[0]      // access head element…
			if ofKind(e) { // …is an element of this category?
				// if not yet initialized…

				if pcon = cat[e.Id()]; pcon == nil {
					// initialize category, return elem &
					// type → recursive construction
					if len(es) > 1 {
						es = es[1:]
					} else {
						es = es[:0]
					}
					_, pcon = econ(es...)
					cat[e.Id()] = pcon
					return e, pcon
				}

				// construct elem and constructor from prior,
				// ot along the way initialized categorys
				return pcon(es...)
			}
		}
		// return empty element & empty category, if no element has
		// matched.
		return N, econ
	}

	///////////////////////////
	// return constructor monad
	return econ
}
