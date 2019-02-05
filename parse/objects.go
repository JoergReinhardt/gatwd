/*
TYPE IDENTITY PATTERNS

  patterns.go provides functions to deal with tokenized representation of
  godeep constructs, by implementing the token types and helper functions that
  get used internaly to split, join and shuffle sequences in assisting
  signature generation, parsing and the like.
*/
package parse

import (
	d "github.com/JoergReinhardt/godeep/data"
)

type TyObject d.BitFlag

func (t TyObject) TypeObj() d.BitFlag  { return d.Flag.Flag() }
func (t TyObject) TypePrim() d.BitFlag { return d.Flag.Flag() }
func (t TyObject) Flag() d.BitFlag     { return d.BitFlag(t) }

//go:generate stringer -type=TyObject
const (
	Constructor TyObject = 1 << iota
	FunctionClosure
	Thunk
	SelectorThunk
	PartialApplication
	GenericApplication
	StackApplication
	Indirection
	ByteCodeObject
	BlackHole
	Array
	// impure closures performing side effects
	IOByteStream
)
