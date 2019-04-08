package functions

import (
	"bytes"
	"strings"

	d "github.com/joergreinhardt/gatwd/data"
	l "github.com/joergreinhardt/gatwd/lex"
	"github.com/olekukonko/tablewriter"
)

/// VALUE

func (p PairFnc) String() string {
	var buf = bytes.NewBuffer([]byte{})
	buf.WriteString(p.Left().String())
	buf.WriteString(":")
	buf.WriteString(l.Blank.Syntax())
	buf.WriteString(p.Right().String())
	return buf.String()
}
func (c ConstFnc) String() string      { return "ϝ → т" }
func (g FunctorFnc) String() string    { return "ϝ → Fₙ [F]" }
func (g ApplicapleFnc) String() string { return "ϝ → Aₙₘ [A]" }
func (g MonadicFnc) String() string    { return "ϝ → Mₙₘ [M]" }

//func (r RightBoundFnc) String() string { return "ϝ ← [т‥.]" }
func (u UnaryFnc) String() string  { return "т → ϝ → т" }
func (b BinaryFnc) String() string { return "т → т → ϝ → т" }
func (n NaryFnc) String() string   { return "[т‥.] → ϝ → т" }

/// VECTOR
func (v VecFnc) String() string {
	var slice []d.Native
	for _, dat := range v() {
		slice = append(slice, dat)
	}
	return d.StringSlice("∙", "[", "]", slice...)
}

/// ACCESSABLE VECTOR (SLICE OF PAIRS)
func (v RecordFnc) String() string {
	var slice []d.Native
	for _, dat := range v() {
		slice = append(slice, dat)
	}
	return d.StringSlice("∙", "[", "]", slice...)
}

/// ASSOCIATIVE SET
func (v AssocSetFnc) String() string {
	var strb = &strings.Builder{}
	var tab = tablewriter.NewWriter(strb)

	for _, pair := range v.Pairs() {
		var row = []string{pair.Left().String(), pair.Right().String()}
		tab.Append(row)
	}
	tab.Render()
	return strb.String()
}

/// LIST
func (l ListFnc) String() string {
	var h, t = l()
	if t != nil {
	}
	return h.String()
}

/// RECORD

/// TOKEN
func (t tokens) String() string {
	var str string
	for _, tok := range t {
		str = str + " " + tok.String() + "\n"
	}
	return str
}
