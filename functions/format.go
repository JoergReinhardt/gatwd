package functions

import (
	"strings"
)

/// VALUE

func (p PairVal) String() string   { return p.Print() }
func (a KeyPair) String() string   { return a.Print() }
func (a IndexPair) String() string { return a.Print() }

func (p PairVal) Print() string {
	return "(" + p.Left().String() + " " + p.Right().String() + ")"
}
func (a KeyPair) Print() string {
	return a.Left().String() + ":: " + a.Right().String()
}
func (a IndexPair) Print() string {
	return a.Left().String() + ": " + a.Right().String()
}

func (c ConstantExpr) String() string { return c().String() }

//func (r RightBoundFnc) String() string { return "ϝ ← [т‥.]" }
func (u UnaryExpr) String() string  { return "T → ϝ → T" }
func (b BinaryExpr) String() string { return "(T,T) → ϝ → T" }
func (n NaryExpr) String() string   { return "[т‥.] → ϝ → T" }

/// VECTOR
func (v VecVal) String() string {
	var args = []string{}
	for _, arg := range v() {
		args = append(args, arg.String())
	}
	return "[" + strings.Join(args, ", ") + "]"
}

/// ACCESSABLE VECTOR (SLICE OF PAIRS)
func (v PairVec) String() string {
	var args = []string{}
	for _, arg := range v() {
		args = append(args, arg.String())
	}
	return "[" + strings.Join(args, ", ") + "]"
}

/// ASSOCIATIVE SET
func (v SetVal) String() string {
	var args = []string{}
	for _, arg := range v.Pairs() {
		args = append(args, arg.String())
	}
	return "[" + strings.Join(args, ", ") + "]"
}

/// LIST
func (l ListVal) String() string {
	var args = []string{}
	var head, list = l()
	for head != nil {
		args = append(args, head.String())
		head, list = list()
	}
	return "(" + strings.Join(args, ", ") + ")"
}

/// TOKEN
func (t tokens) String() string {
	var str string
	var l = len(t)
	for i, tok := range t {
		str = str + " " + tok.TypeTok().String()
		if i < l-2 {
			str = str + ", "
		}
	}
	return str
}
