package types

type Fixity int8Val

//go:generate stringer -type=Fixity
const (
	PreFix  Fixity = -1
	InFix          = 0
	PostFix        = 1
)

type Narity int8Val

const (
	NArgs    = -3
	BinArgs  = -2
	UniArg   = -1
	constant = 0 // everything is an expression
	Unary    = 1
	BiNary   = 2
	NNary    = 3
)
