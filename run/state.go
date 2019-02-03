package run

import (
	d "github.com/JoergReinhardt/godeep/data"
)

type StateVal func() (
	heap Heap,
	stack Stack,
	types d.SetString,
	symbols d.SetString,
	statefnc StateFnc,
)

func (s StateVal) String() string       { return "State" }
func (s StateVal) Flag() d.BitFlag      { return d.Object.Flag() }
func (s StateVal) Heap() Heap           { heap, _, _, _, _ := s(); return heap }
func (s StateVal) Stack() Stack         { _, stack, _, _, _ := s(); return stack }
func (s StateVal) types() d.SetString   { _, _, types, _, _ := s(); return types }
func (s StateVal) symbols() d.SetString { _, _, _, symbols, _ := s(); return symbols }
func (s StateVal) Next() StateFnc       { _, _, _, _, stf := s(); return stf }
func (s StateVal) Lookup(name d.StrVal) Object {
	var obj Object
	return obj
}
func (s StateVal) SetNext(stf StateFnc) { s = newStateFnc(s, stf) }

func newStateFnc(s StateVal, stf StateFnc) StateVal {
	h, st, t, y, _ := s()
	return func() (
		heap Heap,
		stack Stack,
		types d.SetString,
		symbols d.SetString,
		statefnc StateFnc,
	) {
		return h, st, t, y, stf
	}

}
func NewState() State {
	var h = HeapVal{}
	var st = StackVal{}
	var t = d.SetString{}
	var y = d.SetString{}

	return StateVal(func() (
		heap Heap,
		stack Stack,
		types d.SetString,
		symbols d.SetString,
		statefnc StateFnc,
	) {
		return h, st, t, y, nil
	})
}
