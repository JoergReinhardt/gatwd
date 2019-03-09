package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

var pred = NewPredicate(func(scrut ...Parametric) bool {
	if scrut[0].TypeNat().Flag().Match(d.Int) {
		return true
	}
	return false
})
