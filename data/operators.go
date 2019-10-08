package data

type TyOp uint8

//go:generate stringer -type=TyOp
const (
	Not TyOp = 0 + iota
	And
	Or
	Xor
	AndNot
	ShiftL
	ShiftR
	Negate
	Add
	Substract
	Multiply
	QuoRatio
	Quotient
	Power
	Greater
	Lesser
	Equal
)
