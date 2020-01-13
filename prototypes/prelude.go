package prototypes

import (
	d "github.com/joergreinhardt/gatwd/data"
	f "github.com/joergreinhardt/gatwd/functions"
)

type (
	TypeCons func(types ...d.Typed) (f.TyComp, []DataCons)
	DataCons func(args ...f.Expression) f.FuncVal
)
