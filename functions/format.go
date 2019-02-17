package functions

import (
	"bytes"

	d "github.com/JoergReinhardt/gatwd/data"
	l "github.com/JoergReinhardt/gatwd/lex"
)

/// VALUE

func (p PairVal) String() string {
	var buf = bytes.NewBuffer([]byte{})
	buf.WriteString(p.Left().String())
	buf.WriteString(l.Colon.Syntax())
	buf.WriteString(l.Blank.Syntax())
	buf.WriteString(p.Right().String())
	return buf.String()
}
func (c ConstFnc) String() string  { return "ϝ → т" }
func (u UnaryFnc) String() string  { return "т → ϝ → т" }
func (b BinaryFnc) String() string { return "т → т → ϝ → т" }
func (n NaryFnc) String() string   { return "[т...] → ϝ → т" }

/// VECTOR
func (v VecFnc) String() string {
	var slice []d.Native
	for _, dat := range v() {
		slice = append(slice, dat)
	}
	return d.StringSlice("∙", "[", "]", slice...)
}

/// ACCESSABLE VECTOR (SLICE OF PAIRS)
func (v AssocVecFnc) String() string {
	var slice []d.Native
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

/// RECORD

/// ARGUMENT
func (p ArgVal) String() string { return p.Arg().String() }

/// ARGUMENTS
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

//// PARAMETER
func (p ParamVal) String() string { return p.Pair().String() }

//// PARAMETERS
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
