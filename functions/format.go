package functions

import (
	"strings"

	d "github.com/joergreinhardt/gatwd/data"
	"github.com/olekukonko/tablewriter"
)

/// VALUE

func (p PairVal) String() string   { return p.Print() }
func (a AssocPair) String() string { return a.Print() }
func (a IndexPair) String() string { return a.Print() }

func (p PairVal) Print() string {
	return "(" + p.Left().String() + " " + p.Right().String() + ")"
}
func (a AssocPair) Print() string {
	return a.Left().String() + ":: " + a.Right().String()
}
func (a IndexPair) Print() string {
	return a.Left().String() + ": " + a.Right().String()
}

func (c ConstantExpr) String() string { return c().String() }

//func (r RightBoundFnc) String() string { return "ϝ ← [т‥.]" }
func (u UnaryExpr) String() string  { return "Unary" }
func (b BinaryExpr) String() string { return "Binary" }
func (n NaryExpr) String() string   { return "Nary" }

/// VECTOR
func (v VecVal) String() string {
	var args = []string{}
	for _, arg := range v() {
		args = append(args, arg.String())
	}
	return "[" + strings.Join(args, ", ") + "]"
}

/// ACCESSABLE VECTOR (SLICE OF PAIRS)
func (v AssocVec) String() string {
	var slice []d.Native
	for _, dat := range v() {
		slice = append(slice, dat)
	}
	return d.StringSlice("∙", "[", "]", slice...)
}

/// ASSOCIATIVE SET
func (v SetVal) String() string {
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
func (l ListVal) String() string {
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
