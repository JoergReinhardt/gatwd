package gatw

type (
	Elem interface {
		Id() int
		Uniq()
	}

	Cls interface {
		Elem
		Cons(...Elem) Elem
	}
	CnsCls func(Cls, ...Elem) Cls

	Id    func() int
	CnsId func(i int) Id

	Ids    func() []int
	CnsIds func(Ids, ...Id) Ids

	Flg    func() uint
	CnsFlg func(uint) Flg

	Fst    func() uint
	CnsFst func(Flgs, ...Flg) Fst

	Flgs    func() []uint
	CnsFlgs func(Flgs, ...Flg) Flgs

	Sym    func() string
	CnsSym func(string) Sym

	Syms    func() []string
	CnsSyms func(Syms, ...Sym) Syms

	Cnst   func() Elem
	CnsCnt func(Elem) func() Elem

	Pair    func() (l, r Elem)
	CnsPair func(Pair, ...Elem) Pair

	Lkn    func() (Elem, Pair)
	CnsLkn func(Lkn, ...Elem) Lkn

	Tpl    func() []Elem
	CnsTpl func(Tpl, ...Elem) Tpl

	UOp func(Elem) Elem
	BOp func(x, y Elem) Elem
	NOp func(...Elem) Elem
	// fnc for every number of variadic arguments
	Arns    func() (Cnst, UOp, BOp, NOp)
	CnsArns func(Cnst, UOp, BOp, NOp) Arns

	// fnc name, args tuple & expression
	Def    func() (name Sym, args Tpl, expr NOp)
	CnsDef func(name Sym, args Tpl, expr NOp) Def

	Defs    func(...Elem) (Def, Defs)
	CnsDefs func(Defs, ...Def) Defs

	Acc func(e Elem) (Acc, Elem)
	Gen func() (Elem, Gen)

	Cons func(...Elem) (Elem, Cons)
)
