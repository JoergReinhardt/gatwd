package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

var isInteger = NewTruth(func(arg Callable) bool { return d.Integers.Flag().Match(arg.TypeNat()) })

var isFloat = NewTruth(func(arg Callable) bool { return d.Rationals.Flag().Match(arg.TypeNat()) })
