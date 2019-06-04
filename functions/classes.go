package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

func Add(num ...d.Numeral) d.Numeral {
	return num[0]
}
