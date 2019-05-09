package functions

import d "github.com/joergreinhardt/gatwd/data"

type (
	BufferMon func(...Callable) (Callable, MonadicVal)
)

func (b BufferMon) Call(args ...Callable) Callable { return MonadicVal(b).Call(args...) }
func (b BufferMon) Eval(args ...d.Native) d.Native { return MonadicVal(b).Eval(args...) }
func (b BufferMon) Ident() Callable                { return MonadicVal(b).Ident() }
func (b BufferMon) Head() Callable                 { return MonadicVal(b).Head() }
func (b BufferMon) Tail() Consumeable              { return MonadicVal(b).Tail() }
func (b BufferMon) TypeNat() d.TyNat               { return MonadicVal(b).TypeNat() }
func (b BufferMon) TypeFnc() TyFnc                 { return Monad | IO | Buffer }
func (b BufferMon) Buffer() *d.BufferVal {
	var data, _ = b()
	return data.(DataVal)().(*d.BufferVal)
}
