package functions

import (
	//	s "strings"
	//	u "unicode"

	d "github.com/joergreinhardt/gatwd/data"
)

type TyTok uint8

func (t TyTok) Type() Typed                   { return t }
func (t TyTok) TypeFnc() TyFnc                { return Type }
func (t TyTok) Call(...Expression) Expression { return t }
func (t TyTok) Flag() d.BitFlag               { return d.BitFlag(uint(t)) }
func (t TyTok) FlagType() d.Uint8Val          { return Flag_Token.U() }
func (t TyTok) TypeName() string              { return t.String() }
func (t TyTok) Match(typ d.Typed) bool {
	if typ.FlagType() == Flag_Token.U() {
		return t == typ.(TyTok)
	}
	return false
}

//go:generate stringer -type=TyTok
const (
	Tok_Text TyTok = 0 + iota
	Tok_FncType
	Tok_NatType
	Tok_ParamType
	Tok_ClassType
)

type runetext d.RuneVec

func newRuneTest(str string) runetext {
	var runes = d.RuneVec{}
	for _, r := range []rune(str) {
		runes = append(runes, r)
	}
	return runetext(runes)
}
func (t runetext) RuneVec() d.RuneVec          { return d.RuneVec(t) }
func (t runetext) Get(i Native) d.Native       { return t.RuneVec().Get(i) }
func (t runetext) GetInt(i int) d.Native       { return t.RuneVec().GetInt(i) }
func (t runetext) Range(i, j int) d.Sliceable  { return t.RuneVec().Range(i, j) }
func (t runetext) Native(i int) rune           { return t.RuneVec().Native(i) }
func (t runetext) RangeNative(i, j int) []rune { return t.RuneVec().RangeNative(i, j) }
func (t runetext) TypeNat() d.TyNat            { return t.RuneVec().TypeNat() }
func (t runetext) Slice() []d.Native           { return t.RuneVec().Slice() }
func (t runetext) TypeName() string            { return t.RuneVec().TypeName() }
func (t runetext) String() string              { return t.RuneVec().String() }
func (t runetext) ElemType() d.TyNat           { return t.RuneVec().ElemType() }
func (t runetext) Null() d.Native              { return t.RuneVec().Null() }
func (t runetext) Empty() bool                 { return t.RuneVec().Empty() }
func (t runetext) Copy() d.Native              { return t.RuneVec().Copy() }
func (t runetext) Len() int                    { return t.RuneVec().Len() }

func (t runetext) Peek(i int) (rune, bool) {
	if i < len(t)-1 {
		return t.Native(i), true
	}
	return ' ', false
}
func (t runetext) Head() (rune, bool) {
	if t.Len() == 0 {
		return []rune(t)[0], false
	}
	if t.Len() > 0 {
		return []rune(t)[0], true
	}
	return ' ', false
}
func (t runetext) Tail() ([]rune, bool) {
	if t.Len() > 1 {
		return []rune(t)[1:], true
	}
	return nil, false
}
func (t runetext) Consume() (rune, runetext, bool) {
	var head, hok = t.Head()
	var tail, tok = t.Tail()
	if hok && tok {
		return head, runetext(tail), true
	}
	return head, nil, false
}
func (t runetext) Take(n int) ([]rune, runetext, bool) {
	if n < t.Len()-1 {
		return []rune(t[:n]), runetext(t[n:]), true
	}
	if n == t.Len()-1 {
		return []rune(t), runetext([]rune{}), false
	}
	return nil, t, false
}

type token KeyPair

func newToken(text string, kind TyTok, types ...Typed) token {
	var vec = NewVector(NewNative(kind))
	if len(types) > 0 {
		for _, typ := range types {
			vec = vec.Append(typedToExpression(typ))
		}
	}
	return token(NewKeyPair(text, vec))
}

func (t token) TypeFnc() TyFnc                 { return Key }
func (t token) KeyType() TyFnc                 { return Key.TypeFnc() }
func (t token) TypeNat() d.TyNat               { return d.Function }
func (t token) KeyPair() KeyPair               { return KeyPair(t) }
func (t token) FlagType() d.Uint8Val           { return Flag_Functional.U() }
func (t token) KeyStr() string                 { return t.KeyPair().KeyStr() }
func (t token) String() string                 { return t.KeyPair().String() }
func (t token) Value() Expression              { return t.KeyPair().Value() }
func (t token) Left() Expression               { return t.KeyPair().Value() }
func (t token) Right() Expression              { return t.KeyPair().Right() }
func (t token) Both() (Expression, Expression) { return t.Left(), t.Right() }
func (t token) Pair() Paired                   { return t.KeyPair().Pair() }
func (t token) Pairs() []Paired                { return t.KeyPair().Pairs() }
func (t token) Key() Expression                { return t.KeyPair().Right() }
func (t token) Swap() (Expression, Expression) { return t.KeyPair().Swap() }
func (t token) SwappedPair() Paired            { return t.KeyPair().SwappedPair() }
func (t token) Empty() bool                    { return t.KeyPair().Empty() }
func (t token) Type() Typed                    { return t.KeyPair().Type() }
func (t token) TypeName() string               { return t.KeyPair().TypeName() }
func (t token) ValType() TyFnc                 { return t.KeyPair().Value().TypeFnc() }
func (t token) Call(args ...Expression) Expression {
	return t.KeyPair().Value().Call(args...)
}

func (t token) vector() VecCol   { return t.KeyPair().Right().(VecCol) }
func (t token) Text() string     { return t.KeyPair().KeyStr() }
func (t token) TokenType() TyTok { return t.vector().Head().(TyTok) }
func (t token) TypeFlags() []Typed {
	var types = make([]Typed, 0, t.vector().Len()-1)
	if t.vector().Len() > 1 {
		for _, expr := range t.vector()()[1:] {
			types = append(types, expr.(Typed))
		}
	}
	return types
}

type tokVec PairVec

func (v tokVec) VecCol() tokVec                           { return tokVec(v) }
func (v tokVec) TypeFnc() TyFnc                           { return Vector }
func (v tokVec) TypeNat() d.TyNat                         { return d.Function }
func (v tokVec) FlagType() d.Uint8Val                     { return Flag_Functional.U() }
func (v tokVec) Con(args ...Expression) PairVec           { return v.VecCol().Con(args...) }
func (v tokVec) ConPairs(pairs ...Paired) PairVec         { return v.VecCol().ConPairs(pairs...) }
func (v tokVec) Consume() (Expression, Consumeable)       { return v.VecCol().Consume() }
func (v tokVec) ConsumePairVec() (Paired, PairVec)        { return v.VecCol().ConsumePairVec() }
func (v tokVec) Pairs() []Paired                          { return v.VecCol().Pairs() }
func (v tokVec) ConsumePair() (Paired, ConsumeablePaired) { return v.VecCol().ConsumePair() }
func (v tokVec) SwitchedPairs() []Paired                  { return v.VecCol().SwitchedPairs() }
func (v tokVec) Slice() []Expression                      { return v.VecCol().Slice() }
func (v tokVec) HeadPair() Paired                         { return v.VecCol().HeadPair() }
func (v tokVec) Head() Expression                         { return v.VecCol().Head() }
func (v tokVec) TailPairs() ConsumeablePaired             { return v.VecCol().TailPairs() }
func (v tokVec) Tail() Consumeable                        { return v.VecCol().Tail() }
func (v tokVec) Call(args ...Expression) Expression       { return v.VecCol().Call(args...) }
func (v tokVec) Empty() bool                              { return v.VecCol().Empty() }
func (v tokVec) TypeElem() TyFnc                          { return v.VecCol().TypeElem() }
func (v tokVec) KeyType() TyFnc                           { return v.VecCol().KeyType() }
func (v tokVec) ValType() TyFnc                           { return v.VecCol().ValType() }
func (v tokVec) TypeName() string                         { return v.VecCol().TypeName() }
func (v tokVec) Type() Typed                              { return v.VecCol().Type() }
func (v tokVec) Len() int                                 { return v.VecCol().Len() }

func (v tokVec) Tokens() []token {
	if v.Len() > 1 {
		var tokens = make([]token, 0, v.Len()-1)
		for _, pair := range v.Pairs()[1:] {
			tokens = append(tokens, token(pair.(KeyPair)))
		}
		return tokens
	}
	return []token{}
}
func (v tokVec) Put(toks ...token) tokVec {
	var pairs []Paired
	if len(toks) == 1 {
		return tokVec(v.VecCol().ConPairs(toks[0]))
	}
	if len(toks) > 1 {
		pairs = make([]Paired, 0, len(toks))
		for _, tok := range toks {
			pairs = append(pairs, tok)
		}
	}
	return tokVec(v.VecCol().ConPairs(pairs...))
}
func (v tokVec) Pop() (token, tokVec) {
	var pair, vec = v.VecCol().ConsumePairVec()
	return token(pair.(KeyPair)), tokVec(vec)
}
func (v tokVec) Peek() token {
	return token(v.VecCol().HeadPair().(KeyPair))
}
func (v tokVec) PeekN(n int) token {
	if n <= v.Len()-1 {
		return token(v.VecCol().Tokens()[n-1])
	}
	return token(NewKeyPair("peek index greater remaining tokens", NewNone()))
}
func (v tokVec) TakeN(n int) []token {
	var tokens = v.Tokens()
	if n <= len(tokens)-2 {
		return tokens[:n+1]
	}
	if n == len(tokens)-1 {
		return tokens
	}
	return []token{}
}

type symTab d.SetString

func newSymbolTable() symTab {
	return symTab(d.NewStringSet().(d.SetString))
}
func (s symTab) strset() d.SetString { return d.SetString(s) }
func (s symTab) Eval() d.Native      { return d.SetString(s) }
func (s symTab) First() d.Paired     { return s.strset().First() }
func (s symTab) Keys() []d.Native    { return s.strset().Keys() }
func (s symTab) Data() []d.Native    { return s.strset().Data() }
func (s symTab) Slice() []d.Native   { return s.strset().Slice() }
func (s symTab) Fields() []d.Paired  { return s.strset().Fields() }
func (s symTab) TypeNat() d.TyNat    { return s.strset().TypeNat() }
func (s symTab) KeyType() d.TyNat    { return s.strset().KeyType() }
func (s symTab) TypeName() string    { return s.strset().TypeName() }
func (s symTab) ValType() d.TyNat {
	return s.strset().First().Right().TypeNat()
}
func (s symTab) Has(acc d.Native) bool {
	return s.strset().Has(acc)
}
func (s symTab) HasStr(key string) bool {
	return s.strset().HasStr(key)
}
func (s symTab) Get(acc d.Native) (d.Native, bool) {
	return s.strset().Get(acc)
}
func (s symTab) Set(acc d.Native, dat d.Native) d.Mapped {
	return s.strset().Set(acc, dat)
}
func (s symTab) GetStr(key string) (d.Native, bool) {
	return s.strset().GetStr(key)
}
func (s symTab) SetStr(key string, dat d.Native) d.Mapped {
	return s.strset().SetStr(key, dat)
}
func (s symTab) DeleteNat(acc d.Native) bool {
	return s.strset().Delete(acc)
}
func (s symTab) Len() int { return s.strset().Len() }

func (s symTab) Tokens() []token {
	var tokens = make([]token, 0, s.Len())
	for _, field := range s.Fields() {
		tokens = append(tokens, token(NewKeyPair(
			field.Left().String(),
			field.Right().(DataExpr)().(VecCol),
		)))
	}
	return tokens
}
func (s symTab) AddToken(tok token) {
	var text = tok.KeyStr()
	var elems = tok.vector()
	s = symTab(s.Set(
		d.StrVal(text),
		NewNative(elems),
	).(d.SetString))
}
func (s symTab) GetToken(str string) (token, bool) {
	var nat, ok = s.GetStr(str)
	if ok {
		var vec = nat.(DataExpr)().(VecCol)
		return token(NewKeyPair(str, vec)), true
	}
	return token(NewKeyPair("empty", NewNone())), false
}

//func parseSig(tm tokenMap, sig, tokens VecCol) (tokenMap, VecCol, VecCol) {
//	var stack = NewVector()
//	var popped Expression
//	if sig.Len() > 0 {
//		var tok = sig.Head().String()
//		if s.ContainsAny(tok, "()") {
//			_, sig = sig.ConsumeVec()
//			if s.Contains(tok, "(") {
//				stack = stack.Append(tokens)
//				if tok == "(" {
//					tm, sig, tokens = parseSig(tm, sig, NewVector())
//				}
//				if s.HasPrefix(tok, "(") {
//					_, sig = sig.ConsumeVec()
//					tok = s.TrimLeft(tok, "(")
//					sig = sig.Append(NewData(d.StrVal(tok)))
//					tm, sig, tokens = parseSig(tm, sig, NewVector())
//				}
//			}
//			if s.Contains(tok, ")") {
//				popped, stack = stack.ConsumeVec()
//				if tok == ")" {
//					popped, stack = stack.ConsumeVec()
//					tokens = tokens.Append(popped.(VecCol))
//				}
//				if s.HasSuffix(tok, ")") {
//					tok = s.TrimRight(tok, ")")
//					sig = sig.Append(NewData(d.StrVal(tok)))
//					popped, stack = stack.ConsumeVec()
//					tm, sig, tokens = parseSig(tm, sig, popped.(VecCol))
//				}
//			}
//		}
//		tm, sig, tokens = parseElem(tm, sig, tokens)
//	}
//	return tm, sig, tokens
//}

//func parseElem(tm tokenMap, sig, elems VecCol) (tokenMap, VecCol, VecCol) {
//	if sig.Len() > 0 {
//		var val token
//		var tok = sig.Head().String()
//		if s.ContainsAny(tok, "∷:->→=>⇒") {
//			// pop the arrow token
//			_, sig = sig.ConsumeVec()
//		}
//		if u.IsUpper([]rune(tok)[0]) {
//			if nat, ok := searchNatType(tok); ok {
//				_, sig = sig.ConsumeVec()
//				elems = elems.Append(NewData(d.StrVal(tok)))
//				val = newTval(NativeType, nat, None)
//				tm[tok] = val
//				return tm, sig, elems
//			}
//			if fnc, ok := searchFncType(tok); ok {
//				_, sig = sig.ConsumeVec()
//				elems = elems.Append(NewData(d.StrVal(tok)))
//				val = newTval(FunctionType, d.Nil, fnc)
//				tm[tok] = val
//				return tm, sig, elems
//			}
//			_, sig = sig.ConsumeVec()
//			elems = elems.Append(NewData(d.StrVal(tok)))
//			val = newTval(ParamType, d.Nil, None)
//			tm[tok] = val
//			return tm, sig, elems
//		}
//		if u.IsLower([]rune(tok)[0]) {
//			_, sig = sig.ConsumeVec()
//			elems = elems.Append(NewData(d.StrVal(tok)))
//			val = newTval(FunctionName, d.Nil, None)
//			tm[tok] = val
//			return tm, sig, elems
//		}
//		if sig.Len() > 0 {
//			tm, sig, elems = parseSig(tm, sig, elems)
//		}
//	}
//	return tm, sig, elems
//}
