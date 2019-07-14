package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	//// NONE VALUE CONSTRUCTOR
	NoneVal func()
)

//// NONE VALUE CONSTRUCTOR
///
// none represens the abscence of a value of any type. implements consumeable,
// key-, index & generic pair interface to be returneable as such.
func NewNone() NoneVal { return func() {} }

func (n NoneVal) Ident() Expression             { return n }
func (n NoneVal) Head() Expression              { return n }
func (n NoneVal) Tail() Consumeable             { return n }
func (n NoneVal) Len() d.IntVal                 { return 0 }
func (n NoneVal) String() string                { return "⊥" }
func (n NoneVal) Call(...Expression) Expression { return nil }
func (n NoneVal) Key() Expression               { return nil }
func (n NoneVal) Index() Expression             { return nil }
func (n NoneVal) Left() Expression              { return nil }
func (n NoneVal) Right() Expression             { return nil }
func (n NoneVal) Both() Expression              { return nil }
func (n NoneVal) Value() Expression             { return nil }
func (n NoneVal) Empty() d.BoolVal              { return true }
func (n NoneVal) Flag() d.BitFlag               { return d.BitFlag(None) }
func (n NoneVal) TypeFnc() TyFnc                { return None }
func (n NoneVal) TypeNat() d.TyNat              { return d.Nil }
func (n NoneVal) TypeElem() d.Typed             { return None }
func (n NoneVal) TypeName() string              { return n.String() }
func (n NoneVal) FlagType() d.Uint8Val          { return Flag_Function.U() }
func (n NoneVal) Type() d.Typed                 { return None }
func (n NoneVal) Consume() (Expression, Consumeable) {
	return NewNone(), NewNone()
}
