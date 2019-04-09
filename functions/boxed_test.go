package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

var pred = NewTruthFunction(NaryFnc(func(scrut ...Callable) Callable {
	if scrut[0].TypeNat().Flag().Match(d.Int) {
		return New(true)
	}
	return New(false)
}))
