package functions

import (
	d "github.com/JoergReinhardt/godeep/data"
)

/// VALUE

/// PAIR

/// VECTOR
func (v VecFnc) String() string {
	var slice []d.Data
	for _, dat := range v() {
		slice = append(slice, dat)
	}
	return d.StringSlice("∙", "[", "]", slice...)
}

/// LIST
func (l ListFnc) String() string {
	var h, t = l()
	if t != nil {
		return h.String() + ", " + t.String()
	}
	return h.String()
}

/// TUPLE
func (t TupleFnc) String() string {
	var slice []d.Data
	var v, _ = t()
	for _, dat := range v.(VecFnc)() {
		slice = append(slice, dat)
	}
	return d.StringSlice(", ", "(", ")", slice...)
}

/// RECORD
func (r RecordFnc) String() string {
	var str = "{"
	var l = r.Len()
	for i, pair := range r.Slice() {
		str = str + pair.(Paired).Left().String() + "::" +
			pair.(Paired).Right().String()
		if i < l-1 {
			str = str + ", "
		}
	}
	return str + "}"
}
