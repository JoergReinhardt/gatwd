package functions

import d "github.com/JoergReinhardt/godeep/data"

type function praedicates

func (f function) Type() Flag          { return newFlag(f.UID(), f.Kind(), f.Prec()) }
func (f function) Flag() d.BitFlag     { return f.Type().Prec() }
func (f function) String() string      { return praedicates(f).String() }
func (f function) get(acc Data) Paired { return praedicates(f).Get(acc) }
func newFuncDef(
	uid int, // unique id for this type
	prec d.BitFlag,
	kind Kind,
	fixity Property, // PostFix|InFix|PreFix
	lazy Property, // Eager|Lazy
	bound Property, // Left_Bound|Right_Bound
	mutable Property, // Mutable|Imutable
	pure Property, // Pure|Effected
	argSig []Data, // either []Flag, or []Paired
	retType Flag, // type of return value
	fnc Function, //  Call(...Data) Data
) FncDef {
	var prop = AccArg
	var flags = []Flag{}
	var accs = []Data{}
	if len(argSig) > 0 {
		for _, arg := range argSig {
			if pair, ok := arg.(Paired); ok {
				flags = append(flags, pair.Right().(Flag))
				accs = append(accs, pair.Left())
			} else {
				if fl, ok := arg.(Flag); ok {
					prop = Positional
					flags = append(flags, fl)
				}
			}
		}
	}
	var flagSet = newFlagSet(flags...)
	var accSet = newArguments(accs...)
	return function(newPraedicates(
		newPair(d.StrVal("UID"), d.UintVal(uid)),
		newPair(d.StrVal("Prec"), prec),
		newPair(d.StrVal("Kind"), kind),
		newPair(d.StrVal("Fixity"), fixity),
		newPair(d.StrVal("Lazy"), lazy),
		newPair(d.StrVal("Bound"), bound),
		newPair(d.StrVal("Mutable"), mutable),
		newPair(d.StrVal("ArgProp"), prop),
		newPair(d.StrVal("Arity"), Arity(len(argSig))),
		newPair(d.StrVal("ArgSig"), flagSet),
		newPair(d.StrVal("Accs"), accSet),
		newPair(d.StrVal("RetType"), retType),
		newPair(d.StrVal("Fnc"), fnc),
	).(praedicates))
}
func (f function) UID() int           { return f.get(d.StrVal("UID")).Right().(Integer).Int() }
func (f function) Prec() d.BitFlag    { return f.get(d.StrVal("Prec")).Right().(d.BitFlag) }
func (f function) Kind() Kind         { return f.get(d.StrVal("Kind")).Right().(Kind) }
func (f function) Arity() Arity       { return f.get(d.StrVal("Arity")).Right().(Arity) }
func (f function) Fix() Property      { return f.get(d.StrVal("Fixity")).Right().(Property) }
func (f function) Lazy() Property     { return f.get(d.StrVal("Lazy")).Right().(Property) }
func (f function) Bound() Property    { return f.get(d.StrVal("Bound")).Right().(Property) }
func (f function) Mutable() Property  { return f.get(d.StrVal("Mutable")).Right().(Property) }
func (f function) Pure() Property     { return f.get(d.StrVal("Pure")).Right().(Property) }
func (f function) ArgProp() Property  { return f.get(d.StrVal("ArgProp")).Right().(Property) }
func (f function) ArgTypes() []Flag   { return f.get(d.StrVal("ArgTypes")).Right().(FlagSet)() }
func (f function) Accs() []Argumented { return f.get(d.StrVal("Accs")).Right().(arguments).Args() }
func (f function) RetType() Flag      { return f.get(d.StrVal("RetType")).Right().(Flag) }
func (f function) Fnc() Function      { return f.get(d.StrVal("Fnc")).Right().(Function) }
