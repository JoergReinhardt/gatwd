package parse

import (
	d "github.com/JoergReinhardt/godeep/data"
	f "github.com/JoergReinhardt/godeep/functions"
)

type Node func() (parent Node, tree Tree, fnc f.Functional)
type Tree func() (Node, Tree)

func (t Tree) Push(parent Node, fnc f.Functional) Node {
	return func() (Node, Tree, f.Functional) { return parent, t, fnc }
}
func NewTree() Tree {
	var tree Tree
	var root Node
	var members = f.NewPair(f.NewConstant(d.StrVal("members: ")), f.NewVector())
	root = func() (Node, Tree, f.Functional) {
		return root, tree, members
	}
	tree = func() (Node, Tree) {
		return tree.Push(root, members), nil
	}
	return tree
}
