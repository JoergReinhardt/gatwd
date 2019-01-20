package functions

import (
	d "github.com/JoergReinhardt/godeep/data"
	l "github.com/JoergReinhardt/godeep/lang"
)

type function parameters

func (f function) get(acc Data) Paired { return parameters(f).Get(acc) }
func newFuncDef(
	uid int, // unique id for this type
	name string,
	prec d.BitFlag,
	kind Kind,
	fixity Property, // PostFix|InFix|PreFix
	lazy Property, // Eager|Lazy
	bound Property, // Left_Bound|Right_Bound
	mutable Property, // Mutable|Imutable
	fnc Function, //  Call(...Data) Data
	retType Flag, // type of return value
	argSig ...Data, // either []Flag, or []Paired
) FncDef {
	var argsAcc = NamedArgs
	var flags = []Flag{}
	var accs = []Data{}
	// infer argument accessor property (access via named-, or positional arguments)
	if len(argSig) > 0 {
		for _, arg := range argSig { // ‥.eiher []Paired ⇒ named arguments
			if pair, ok := arg.(Paired); ok {
				flags = append(flags, pair.Right().(Flag))
				accs = append(accs, pair.Left())
			} else { // ‥. or []Flag ⇒ positional arguments
				if fl, ok := arg.(Flag); ok {
					argsAcc = Positional
					flags = append(flags, fl)
				}
			}
		}
	}
	var flagSet = newFlagSet(flags...)
	var accSet = newArguments(accs...)
	return function(newParameters(
		newPair(d.StrVal("UID"), d.UintVal(uid)),
		newPair(d.StrVal("Name"), d.StrVal(name)),
		newPair(d.StrVal("Precedence Type"), prec),
		newPair(d.StrVal("Kind"), kind),
		newPair(d.StrVal("Fixity"), fixity),
		newPair(d.StrVal("Lazy"), lazy),
		newPair(d.StrVal("Bound"), bound),
		newPair(d.StrVal("Mutable"), mutable),
		newPair(d.StrVal("Return Type"), retType),
		newPair(d.StrVal("Function"), fnc),
		// inferred propertys
		newPair(d.StrVal("Arity"), Arity(len(argSig))),
		newPair(d.StrVal("Access Type"), argsAcc),
		newPair(d.StrVal("Arg Types"), flagSet),
		newPair(d.StrVal("Accessors"), accSet),
	).(parameters))
}
func (f function) Type() Flag           { return newFlag(f.UID(), f.Kind(), f.Prec()) }
func (f function) Flag() d.BitFlag      { return f.Prec() }
func (f function) UID() int             { return f.get(d.StrVal("UID")).Right().(Integer).Int() }
func (f function) Name() string         { return f.get(d.StrVal("Name")).Right().(Symbolic).String() }
func (f function) Prec() d.BitFlag      { return f.get(d.StrVal("PrecedenceT ype")).Right().(d.BitFlag) }
func (f function) Kind() Kind           { return f.get(d.StrVal("Kind")).Right().(Kind) }
func (f function) Arity() Arity         { return f.get(d.StrVal("Arity")).Right().(Arity) }
func (f function) Fix() Property        { return f.get(d.StrVal("Fixity")).Right().(Property) }
func (f function) Lazy() Property       { return f.get(d.StrVal("Lazy")).Right().(Property) }
func (f function) Bound() Property      { return f.get(d.StrVal("Bound")).Right().(Property) }
func (f function) Mutable() Property    { return f.get(d.StrVal("Mutable")).Right().(Property) }
func (f function) AccessType() Property { return f.get(d.StrVal("Access Type")).Right().(Property) }
func (f function) ArgTypes() []Flag     { return f.get(d.StrVal("Arg Types")).Right().(FlagSet)() }
func (f function) Accs() []Argumented   { return f.get(d.StrVal("Accessors")).Right().(arguments).Args() }
func (f function) RetType() Flag        { return f.get(d.StrVal("Return Type")).Right().(Flag) }
func (f function) Fnc() Function        { return f.get(d.StrVal("Function")).Right().(Function) }
func (f function) nameToken() Token     { return newToken(Symbolic_Token, newData(d.StrVal(f.Name()))) }
func (f function) returnToken() Token   { return newToken(Data_Type_Token, f.RetType()) }
func (f function) argumentTokens() []Token {
	var tok = []Token{}
	for _, flag := range f.ArgTypes() {
		// transform flags to tokens
		tok = append(tok, newToken(Data_Type_Token, flag))
	}
	// arg0 → arg1 → argn
	tok = tokJoin(newToken(Syntax_Token, l.RightArrow), tok)
	return tok
}
func (f function) sigTokens() []Token {
	return tokJoin(
		// arguments → name → return-type
		newToken(Syntax_Token, l.RightArrow),
		append(
			append(
				f.argumentTokens(), f.nameToken()), f.returnToken()))
}
