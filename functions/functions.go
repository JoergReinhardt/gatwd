package functions

func MapFnc(expr Functional, list Consumeable) Consumeable {
	var elem Functional
	var vec = []Functional{}
	for elem, list = list.DeCap(); !ElemEmpty(elem); {
		vec = append(vec, expr.Call(elem))
	}
	return NewRecursiveList(vec...)
}

func FoldLFnc(expr Functional, list Consumeable, init Functional) Functional {
	var elem Functional
	for elem, list = list.DeCap(); !ElemEmpty(elem); {
		init = init.Call(expr.Call(elem))
	}
	return init
}
