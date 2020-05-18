package gatw

import (
	"math/bits"
)

type TC uint8

//go:generate stringer -type TC
const (
	N TC = 1<<iota - 1
	Type
	Func
	Symb

	C = N | Type | Func | Symb
)

func (t TC) Ident() Elem { return t }
func (t TC) uint() uint  { return uint(t) }
func (t TC) Flag() uint  { return t.uint() }
func (t TC) Id() int     { return bits.Len8(uint8(t)) }

func Has(f Flg) bool { return uint(C)&^f() != 0 }

func New(...Elem) (Unit, Let, Define) {
	var ( // lacal values and cathegory operations are enclosed by cathegory
		names Record // names of function definitions and values
		cons  = []Obj{}
		vals  = []Val{}
		fncs  = []Define{}

		def = Define(func(
			name string, expr Elem, args ...Elem,
		) Definition {
			var i = len(fncs)
			return Definition(func() (Val, []Elem) {
				return ConVal(name, ConObj(i, expr)), args
			})
		})

		let = Let(func(name string, e Elem) Val {
			var (
				id        = len(cons)
				o  ObjFnc = func() (int, Elem) {
					// element exists → lookup
					if o, ok := names[Name(name)]; ok {
						return id, o
					}
					return 0, N
				}
			)
			names[Name(name)] = o
			var v = ConVal(name, o)
			vals = append(vals, v)
			return v
		})

		unit = func(e Elem) Val {
			var (
				f  Flg
				ok bool
			)
			// element is a type flag
			if f, ok = e.(Flg); ok {
				// element exists → lookup
				if len(cons) == f.Id()+1 {
					var o = ConObj(f.Id(), e)
					cons = append(cons, o)
					return ConVal(f.String(), o)
				}
			}
			return ConVal("⊥", N)
		}
	)
	return unit, let, def
}
