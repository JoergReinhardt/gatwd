package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	NumberFnc  func() d.Native
	ArritUnaOp func(Callable) NumberFnc
	ArritBinOp func(a, b Callable) NumberFnc
	ArritNOp   func(a, b Callable) NumberFnc
)
