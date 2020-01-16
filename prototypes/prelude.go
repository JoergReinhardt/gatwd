package prototypes

import (
	d "github.com/joergreinhardt/gatwd/data"
	f "github.com/joergreinhardt/gatwd/functions"
)

type (
	TypeCons func(types ...d.Typed) (f.TyDef, []DataCons)
	DataCons func(args ...f.Functor) f.FuncVal
)
