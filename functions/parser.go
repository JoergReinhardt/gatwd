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

type token KeyPair

func (t token) TypeFnc() TyFnc                 { return Key }
func (t token) KeyType() TyFnc                 { return Key.TypeFnc() }
func (t token) TypeNat() d.TyNat               { return d.Function }
func (t token) KeyPair() KeyPair               { return KeyPair(t) }
func (t token) FlagType() d.Uint8Val           { return Flag_Functional.U() }
func (t token) KeyStr() string                 { return t.KeyPair().KeyStr() }
func (t token) Value() Expression              { return t.KeyPair().Value() }
func (t token) Left() Expression               { return t.KeyPair().Value() }
func (t token) Right() Expression              { return t.KeyPair().Right() }
func (t token) Both() (Expression, Expression) { return t.Left(), t.Right() }
func (t token) Pair() Paired                   { return t.KeyPair().Pair() }
func (t token) Pairs() []Paired                { return t.KeyPair().Pairs() }
func (t token) Key() Expression                { return t.KeyPair().Right() }
func (t token) ValType() TyFnc                 { return t.KeyPair().Value().TypeFnc() }
func (t token) TypeName() string               { return t.KeyPair().TypeName() }
func (t token) Type() TyFnc                    { return t.KeyPair().TypeFnc() }
func (t token) Swap() (Expression, Expression) { return t.KeyPair().Swap() }
func (t token) SwappedPair() Paired            { return t.KeyPair().SwappedPair() }
func (t token) Empty() bool                    { return t.KeyPair().Empty() }
func (t token) Elems() VecCol {
	return t.KeyPair().Right().(VecCol)
}
func (t token) Call(args ...Expression) Expression {
	return t.KeyPair().Value().Call(args...)
}

type symbolMap d.SetString

func newTokenMap() symbolMap {
	return symbolMap(d.NewStringSet().(d.SetString))
}
func (s symbolMap) strset() d.SetString { return d.SetString(s) }
func (s symbolMap) eval() d.Native      { return d.SetString(s) }
func (s symbolMap) first() d.Paired     { return s.strset().First() }
func (s symbolMap) typeName() string    { return s.strset().TypeName() }
func (s symbolMap) keys() []d.Native    { return s.strset().Keys() }
func (s symbolMap) data() []d.Native    { return s.strset().Data() }
func (s symbolMap) slice() []d.Native   { return s.strset().Slice() }
func (s symbolMap) fields() []d.Paired  { return s.strset().Fields() }
func (s symbolMap) typeNat() d.TyNat    { return s.strset().TypeNat() }
func (s symbolMap) keyType() d.TyNat    { return s.strset().KeyType() }
func (s symbolMap) valType() d.TyNat {
	return s.strset().First().Right().TypeNat()
}
func (s symbolMap) has(acc d.Native) bool {
	return s.strset().Has(acc)
}
func (s symbolMap) hasStr(key string) bool {
	return s.strset().HasStr(key)
}
func (s symbolMap) get(acc d.Native) (d.Native, bool) {
	return s.strset().Get(acc)
}
func (s symbolMap) set(acc d.Native, dat d.Native) d.Mapped {
	return s.strset().Set(acc, dat)
}
func (s symbolMap) getStr(key string) (d.Native, bool) {
	return s.strset().GetStr(key)
}
func (s symbolMap) setStr(key string, dat d.Native) d.Mapped {
	return s.strset().SetStr(key, dat)
}
func (s symbolMap) deleteNat(acc d.Native) bool {
	return s.strset().Delete(acc)
}
func (s symbolMap) Len() int { return s.strset().Len() }
func (s symbolMap) AddToken(tok token) {
	var text = tok.KeyStr()
	s = symbolMap(s.set(d.StrVal(text),
		vectorToSlice(tok.Elems()),
	).(d.SetString))
}
func (s symbolMap) GetToken(str string) (token, bool) {
	var nat, ok = s.getStr(str)
	if ok {
		var vec = sliceToVec(nat.(d.DataSlice))
		return token(NewKeyPair(str, vec)), true
	}
	return token(NewKeyPair("empty", NewNone())), false
}
func (s symbolMap) Tokens() []token {
	var tokens = make([]token, 0, s.Len())
	for _, field := range s.fields() {
		tokens = append(tokens,
			token(NewKeyPair(
				field.Left().String(),
				sliceToVec(
					field.Right().(d.DataSlice),
				))))
	}
	return tokens
}

func newToken(text string, kind TyTok, types ...Typed) token {
	var vec = NewVector(NewData(d.Uint8Val(kind)))
	if len(types) > 0 {
		for _, typ := range types {
			vec = vec.Append(typedToExpression(typ))
		}
	}
	return token(NewKeyPair(text, vec))
}
func typedToExpression(typ Typed) Expression {
	var expr Expression
	switch {
	case Flag_Native.Match(typ.FlagType()):
		expr = NewData(typ.(d.TyNat))
	case Flag_Functional.Match(typ.FlagType()):
		expr = typ.(TyFnc)
	case Flag_Arity.Match(typ.FlagType()):
		expr = typ.(Arity)
	case Flag_Prop.Match(typ.FlagType()):
		expr = typ.(Propertys)
	}
	return expr
}
func vectorToSlice(vec VecCol) d.DataSlice {
	var slice = d.NewSlice()
	for _, expr := range vec() {
		if expr.TypeFnc().Match(Data) {
			slice = append(slice, expr.(DataConst))
			continue
		}
		if expr.TypeFnc().Match(Type) {
			slice = append(slice, NewNativeExpression(expr))
		}
	}
	return slice
}
func sliceToVec(slice d.DataSlice) VecCol {
	var head d.Native
	var vec = NewVector()
	head, slice = slice.Shift()
	vec = vec.Append(TyTok(head.(d.Uint8Val)))
	for _, nat := range slice.Slice() {
		vec = vec.Append(typedToExpression(nat.(Typed)))
	}
	return vec
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
