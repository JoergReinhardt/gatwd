package run

import (
	d "github.com/JoergReinhardt/godeep/data"
	f "github.com/JoergReinhardt/godeep/functions"
)

type HeapVal d.SetUint

func (h HeapVal) String() string  { return d.SetUint(h).String() }
func (h HeapVal) Flag() d.BitFlag { return d.SetUint(h).Flag() }
func (h HeapVal) Fetch(uid d.UintVal) (Object, bool) {
	if dat, ok := d.SetUint(h).Get(uid); ok {
		return dat.(f.Functional), ok
	}
	return nil, false
}
func (h HeapVal) Alloc(uid d.UintVal, obj Object) { d.SetUint(h).Set(uid, obj) }
