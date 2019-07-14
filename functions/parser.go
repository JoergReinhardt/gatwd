package functions

import (
	//	s "strings"
	//	u "unicode"

	d "github.com/joergreinhardt/gatwd/data"
)

type TyTok uint8

func (t TyTok) Type() d.Typed                 { return t }
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
	Tok_Blank TyTok = 0 + iota
	Tok_Text
	Tok_Number
	Tok_Symbol
	Tok_Keyword
	Tok_Operator
	Tok_Delimiter
	Tok_Seperator
	Tok_Punktation
	Tok_DefType
	Tok_NatType
	Tok_FncType
	Tok_Context
)

type runetext d.RuneVec

func newRuneText(str string) runetext {
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
func (t runetext) TypeName() string            { return t.RuneVec().Type().TypeName() }
func (t runetext) String() string              { return t.RuneVec().String() }
func (t runetext) ElemType() d.Typed           { return t.RuneVec().ElemType() }
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
