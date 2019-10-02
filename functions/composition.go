package functions

func Map(function, data Expression) Expression {
	switch {
	case data.TypeFnc().Match(Collections):
	case data.TypeFnc().Match(Pair):
	case data.TypeFnc().Match(None):
	}
	return function.Call(data)
}
func Bind(function, data Expression) Expression {
	var result Expression
	return result
}
func Fold(function, init, data Expression) Expression {
	var result Expression
	return result
}
func Pass(function, data Expression) Expression {
	var result Expression
	return result
}
func Filter(function, data Expression) Expression {
	var result Expression
	return result
}
