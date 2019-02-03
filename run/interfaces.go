package run

import (
	d "github.com/JoergReinhardt/godeep/data"
	f "github.com/JoergReinhardt/godeep/functions"
)

type StateFnc func() StateFnc

type Frame interface {
	f.Functional
}
type Stack interface {
	d.Data
	Push(Frame)
	Pull() Frame
}
type Object interface {
	f.Functional
}
type Heap interface {
	d.Data
	Fetch(uid d.UintVal) (Object, bool)
	Alloc(uid d.UintVal, obj Object)
}
type State interface {
	d.Data
	Heap() Heap
	Stack() Stack
	Lookup(name d.StrVal) Object
	SetNext(StateFnc)
	Next() StateFnc
}
