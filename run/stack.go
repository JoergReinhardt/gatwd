package run

import (
	d "github.com/JoergReinhardt/godeep/data"
)

type StackVal d.DataSlice

func (h StackVal) String() string  { return d.DataSlice(h).String() }
func (h StackVal) Flag() d.BitFlag { return d.DataSlice(h).Flag() }
func (h StackVal) Push(f Frame)    { d.SlicePush(d.DataSlice(h), f) }
func (h StackVal) Pull() Frame {
	dat, slice := d.SlicePull(d.DataSlice(h))
	h = StackVal(slice)
	return dat.(Frame)
}
