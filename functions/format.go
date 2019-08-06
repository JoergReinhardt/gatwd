package functions

import (
	"strings"
)

/// VALUE

func (p PairType) String() string {
	return "(" + p.Left().String() + ", " + p.Right().String() + ")"
}
func (a KeyPairType) String() string {
	return "(" + a.Right().String() + " : " + a.Left().String() + ")"
}
func (a IndexPairType) String() string {
	return "(" + a.Right().String() + " : " + a.Left().String() + ")"
}

//func (r RightBoundFnc) String() string { return "ϝ ← [т‥.]" }

/// VECTOR
func (v VecType) String() string {
	var pairs = []string{}
	for _, pair := range v() {
		pairs = append(pairs, pair.String())
	}
	return "[" + strings.Join(pairs, ", ") + "]"
}

/// ACCESSABLE VECTOR (SLICE OF PAIRS)
func (v PairVecType) String() string {
	var pairs = []string{}
	for _, pair := range v() {
		pairs = append(pairs, pair.String())
	}
	return "[" + strings.Join(pairs, ", ") + "]"
}

/// ASSOCIATIVE SET
////func (v SetCol) String() string {
////	var pairs = []string{}
////	for _, pair := range v.Pairs() {
////		pairs = append(pairs, pair.String())
////	}
////	return "[" + strings.Join(pairs, ", ") + "]"
////}

/// LIST
func (l ListType) String() string {
	var (
		args       = []string{}
		head, list = l()
	)
	for head != nil {
		args = append(args, head.String())
		head, list = list()
	}
	return "(" + strings.Join(args, ", ") + ")"
}
func (l PairListType) String() string {
	var (
		args       = []string{}
		head, list = l()
	)
	for head != nil {
		args = append(args, head.String())
		head, list = list()
	}
	return "(" + strings.Join(args, ", ") + ")"
}
