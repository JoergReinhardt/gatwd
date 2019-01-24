/*
TREE

  parse package implements a very rudimentary tree data type, to bootstrap the
  typesystem to the point, where further parsing can be defined in terms of
  godeep itself.
*/
package parse

import (
	d "github.com/JoergReinhardt/godeep/data"
)

// node is a closure type that yields reference to a parent node, a root tree
// instance, all nodes are part of and an instance of arbitrary functional
// data, which can be a collection, parameter-set, or the like. that way node
// and tree are utilizeable as flexible base types to form arbitrary graphs
// from tokens.
type Node func() (parent Node, tok Token)

func NewNode(parent Node, tok Token) Node {
	return func() (Node, Token) {
		return parent, tok
	}
}
func (n Node) String() string   { return n.String() }
func (n Node) TokType() TokType { return n.Token().TokType() }
func (n Node) Flag() d.BitFlag  { return n.Token().Flag() }
func (n Node) Parent() Node {
	parent, _ := n()
	return parent
}
func (n Node) Token() Token {
	_, tok := n()
	return tok
}

// while the default behaviour of the tree function is to pop the current node
// and the resulting reduced tree, push provides a way to extend the tree by a
// new layer of references. branch and case nodes are implementet by passing a
// collection of tokens as the contained functional instance. pushing a node
// generates a new top node instance with current top as it's parent.
func (t Node) Push(tok Token) Tree {
	var parent, _ = t()
	return NewNode(parent, tok)
}

// forwards the default behaviour as method.
func (t Node) Pop() (Node, Token) { return t() }
