package functions

import (
	"strings"
)

/// VALUE

func (p PairVal) String() string {
	return "(" + p.Left().String() + ", " + p.Right().String() + ")"
}
func (a KeyPair) String() string {
	return "(" + a.Right().String() + " : " + a.Left().String() + ")"
}
func (a IndexPair) String() string {
	return "(" + a.Right().String() + " : " + a.Left().String() + ")"
}

//func (r RightBoundFnc) String() string { return "ϝ ← [т‥.]" }

/// VECTOR
func (v ColVec) String() string {
	var pairs = []string{}
	for _, pair := range v() {
		pairs = append(pairs, pair.String())
	}
	return "[" + strings.Join(pairs, ", ") + "]"
}

/// ACCESSABLE VECTOR (SLICE OF PAIRS)
func (v ColPairV) String() string {
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
func (l ColList) String() string {
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
func (l ColPairL) String() string {
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
