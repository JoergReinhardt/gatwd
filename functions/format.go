package functions

import (
	"bytes"

	d "github.com/JoergReinhardt/godeep/data"
	l "github.com/JoergReinhardt/godeep/lex"
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

////////////////////
func (dat DataVal) String() string {
	var buf = bytes.NewBuffer([]byte{})
	buf.WriteString(l.Lambda.Syntax())
	buf.WriteString(l.Blank.Syntax())
	buf.WriteString(l.RightArrow.Syntax())
	buf.WriteString(l.Blank.Syntax())
	buf.WriteString(dat.Flag().String())
	buf.WriteString(l.Blank.Syntax())
	buf.WriteString(dat.Eval().String())
	return buf.String()
}

func (p PairVal) String() string {
	var buf = bytes.NewBuffer([]byte{})
	buf.WriteString(p.Left().String())
	buf.WriteString(l.Colon.Syntax())
	buf.WriteString(l.Blank.Syntax())
	buf.WriteString(p.Right().String())
	return buf.String()
}

func (p ArgVal) String() string { return p.Arg().String() }

func (p ArgSet) String() string {
	var buf = bytes.NewBuffer([]byte{})
	buf.WriteString(l.LeftBra.Syntax())
	var args = p.Data()
	var length = len(args) - 1
	for i, arg := range args {
		buf.WriteString(arg.String())
		if i < length {
			buf.WriteString(l.Comma.Syntax())
			buf.WriteString(l.Blank.Syntax())
		}
	}
	buf.WriteString(l.RightBra.Syntax())
	return buf.String()
}

func (p ParamVal) String() string { return p.Pair().String() }

func (p ParamSet) String() string {
	var buf = bytes.NewBuffer([]byte{})
	buf.WriteString(l.LeftBra.Syntax())
	var parms = p.Parms()
	var length = len(parms) - 1
	for i, parm := range parms {
		buf.WriteString(parm.String())
		if i < length {
			buf.WriteString(l.Comma.Syntax())
			buf.WriteString(l.Blank.Syntax())
		}
	}
	buf.WriteString(l.RightBra.Syntax())
	return buf.String()
}
