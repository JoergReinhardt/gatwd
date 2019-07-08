package functions

import (
	//	s "strings"
	//	u "unicode"

	d "github.com/joergreinhardt/gatwd/data"
)

type token KeyPair

func (t token) TypeFnc() TyFnc                     { return Key }
func (t token) KeyType() TyFnc                     { return Key.TypeFnc() }
func (t token) TypeNat() d.TyNat                   { return d.Function }
func (t token) KeyPair() KeyPair                   { return KeyPair(t) }
func (t token) FlagType() d.Uint8Val               { return Flag_Functional.U() }
func (t token) KeyStr() string                     { return t.KeyPair().KeyStr() }
func (t token) Value() Expression                  { return t.KeyPair().Value() }
func (t token) Left() Expression                   { return t.KeyPair().Value() }
func (t token) Right() Expression                  { return t.KeyPair().Right() }
func (t token) Both() (Expression, Expression)     { return t.Left(), t.Right() }
func (t token) Pair() Paired                       { return t.KeyPair().Pair() }
func (t token) Pairs() []Paired                    { return t.KeyPair().Pairs() }
func (t token) Key() Expression                    { return t.KeyPair().Right() }
func (t token) Call(args ...Expression) Expression { return t.KeyPair().Value().Call(args...) }
func (t token) ValType() TyFnc                     { return t.KeyPair().Value().TypeFnc() }
func (t token) TypeName() string                   { return t.KeyPair().TypeName() }
func (t token) Type() TyFnc                        { return t.KeyPair().TypeFnc() }
func (t token) Swap() (Expression, Expression)     { return t.KeyPair().Swap() }
func (t token) SwappedPair() Paired                { return t.KeyPair().SwappedPair() }
func (t token) Empty() bool                        { return t.KeyPair().Empty() }

type tokenMap d.SetString

func newTokenMap() tokenMap {
	return tokenMap(d.NewStringSet().(d.SetString))
}
func (s tokenMap) strset() d.SetString { return d.SetString(s) }
func (s tokenMap) eval() d.Native      { return d.SetString(s) }
func (s tokenMap) first() d.Paired     { return s.strset().First() }
func (s tokenMap) typeName() string    { return s.strset().TypeName() }
func (s tokenMap) keys() []d.Native    { return s.strset().Keys() }
func (s tokenMap) data() []d.Native    { return s.strset().Data() }
func (s tokenMap) slice() []d.Native   { return s.strset().Slice() }
func (s tokenMap) fields() []d.Paired  { return s.strset().Fields() }
func (s tokenMap) typeNat() d.TyNat    { return s.strset().TypeNat() }
func (s tokenMap) keyType() d.TyNat    { return s.strset().KeyType() }
func (s tokenMap) valType() d.TyNat {
	return s.strset().First().Right().TypeNat()
}
func (s tokenMap) get(acc d.Native) (d.Native, bool) {
	return s.strset().Get(acc)
}
func (s tokenMap) set(acc d.Native, dat d.Native) d.Mapped {
	return s.strset().Set(acc, dat)
}
func (s tokenMap) delteNat(acc d.Native) bool {
	return s.strset().Delete(acc)
}
func (s tokenMap) NewToken(text string, kind TvKind, flags ...Typed) {
	var nkind = d.Uint8Val(kind)
	var ntext = d.StrVal(text)
	if len(flags) == 0 {
		s = tokenMap(s.set(ntext, nkind).(d.SetString))
	}
	if len(flags) == 1 {
		s = tokenMap(s.set(ntext, d.NewPair(
			nkind, flags[0].(Flagged).Flag(),
		)).(d.SetString))
	}
	if len(flags) > 1 {
		var slice = []d.Native{}
		for _, flag := range flags {
			if Flag_Native.Match(flag.FlagType()) {
				slice = append(slice, d.TyNat(
					flag.(Flagged).Flag()))
				continue
			}
			slice = append(slice, NewData(flag.(TyFnc)))
		}
		s = tokenMap(s.set(ntext, d.NewPair(
			ntext, d.NewPair(nkind, d.NewSlice(slice...)),
		)).(d.SetString))
	}
}

func (s tokenMap) Len() int { return s.strset().Len() }

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
