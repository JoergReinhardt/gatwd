package functions

import (
	d "github.com/JoergReinhardt/godeep/data"
)

/// VALUE

/// PAIR
func (p Pair) String() string { l, r := p(); return l.String() + " " + r.String() }

/// ARGUMENTS
func (p Args) String() string {
	d, _ := p()
	return d.Flag().String() +
		" " +
		d.String()
}
func (a ArgSet) String() string {
	var strdat = [][]d.Data{}
	for i, dat := range a.Args() {
		strdat = append(strdat, []d.Data{})
		strdat[i] = append(strdat[i], d.New(i), d.New(dat.String()))
	}
	return d.StringChainTable(strdat...)
}

/// PRAEDICATES
func (p Param) String() string {
	l, r := p.Both()
	return l.String() + ": " + r.String()
}
func (a ParamSet) String() string {
	var strout = [][]d.Data{}
	for i, pa := range a.Pairs() {
		strout = append(strout, []d.Data{})
		strout[i] = append(
			strout[i],
			d.New(i),
			d.New(pa.Left().String()),
			d.New(pa.Right().String()))
	}
	return d.StringChainTable(strout...)
}

/// CONSTANT
func (c ConstFnc) String() string { return c().(d.Data).String() }

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
