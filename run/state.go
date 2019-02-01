/*
uEGISTRY

  data type that holds the runtime state of the type system. Comes with helper
  functions to eanipulate chains of tokens when dealing with signatures during
  type checking, or construction.
*/
package run

import (
	d "github.com/JoergReinhardt/godeep/data"
	f "github.com/JoergReinhardt/godeep/functions"
	p "github.com/JoergReinhardt/godeep/parse"
)

// FRAME
//
// Frame func() (uid int, prop p.Property, poly p.Polymorph, caller StateFn)
// instances form the nodes in the acyclic graph of function definitions value
// declarations and function applications describing the execution state during
// runtime. in functional programming everything is a function, which has the
// benefit, that pretty much all those nodes have the same shape.
//
// every instance has a unique id, propertys that tell the runtime on how to
// call it, and a 'Polymorph', defining the function instanciated and
// containing all it's patterns, implementations, argument sets, return value
// types‥. and last, but not least, a reference to the caller to pass the
// return value eventually computed at some point.
func NewFrame(
	uid int,
	caller int,
	poly p.Polymorph,
	args f.Parameters,
	fnc f.Function,
) Frame {
	return Frame(func() (
		uid int,
		caller int,
		poly p.Polymorph,
		args f.Parameters,
		fnc f.Function,
	) {
		return uid,
			caller,
			poly,
			args,
			fnc
	})
}

type Frame func() (
	uid int,
	caller int,
	poly p.Polymorph,
	args f.Parameters,
	fnc f.Function,
)

func (i Frame) Uid() int           { uid, _, _, _, _ := i(); return uid }
func (i Frame) Caller() int        { _, caller, _, _, _ := i(); return caller }
func (i Frame) Poly() p.Polymorph  { _, _, poly, _, _ := i(); return poly }
func (i Frame) Args() f.Parameters { _, _, _, parms, _ := i(); return parms }
func (i Frame) Fnc() f.Function    { _, _, _, _, fnc := i(); return fnc }
func (i Frame) Flag() d.BitFlag    { return d.Machinery.Flag() }
func (i Frame) Kind() f.BitFlag    { return f.Instance.Flag() }
func (i Frame) String() string     { return i.Poly().String() }
