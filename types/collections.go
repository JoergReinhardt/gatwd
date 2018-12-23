package types

func newSlice(val ...Evaluable) slice {
	l := make([]Evaluable, 0, len(val))
	l = append(l, val...)
	return l
}

type Function func(...Evaluable) Evaluable

// normalize narity
type Nullary Function
type Unary Function
type Binary Function
type Nary Function

func (Nullary) Arity() int { return 0 }
func (Unary) Arity() int   { return 1 }
func (Binary) Arity() int  { return 2 }
func (Nary) Arity() int    { return -1 }

func composeNullary(n func() Evaluable) Callable {
	return Nullary(func(...Evaluable) Evaluable {
		return n()
	})
}
func composeUnary(u func(e Evaluable) Evaluable) Callable {
	return Unary(func(e ...Evaluable) Evaluable {
		return u(e[0])
	})
}
func composeBinary(b func(a, b Evaluable) Evaluable) Callable {
	return Binary(func(e ...Evaluable) Evaluable {
		return b(e[0], e[1])
	})
}
func composeNary(n func(e ...Evaluable) Evaluable) Callable {
	return Nary(func(e ...Evaluable) Evaluable {
		return n(e...)
	})
}

///////////////////////////////////////////////////////////////
// CELL
//
