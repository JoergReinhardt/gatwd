package functions

import (
	"math/bits"
	"strconv"

	d "github.com/JoergReinhardt/godeep/data"
	l "github.com/JoergReinhardt/godeep/lang"
)

/// VALUE
func (dat value) String() string { return dat().(d.Data).String() }

/// PAIR
func (p pair) String() string { l, r := p(); return l.String() + " " + r.String() }

/// ARGUMENTS
func (p argument) String() string {
	d, _ := p()
	return d.Flag().String() +
		" " +
		d.String()
}
func (a arguments) String() string {
	var strdat = [][]d.Data{}
	for i, dat := range a.Args() {
		strdat = append(strdat, []d.Data{})
		strdat[i] = append(strdat[i], d.New(i), d.New(dat.String()))
	}
	return d.StringChainTable(strdat...)
}

/// PRAEDICATES
func (p parameter) String() string {
	l, r := p.Both()
	return l.String() + ": " + r.String()
}
func (a parameters) String() string {
	var strout = [][]d.Data{}
	for i, pa := range a.Accs() {
		strout = append(strout, []d.Data{})
		strout[i] = append(
			strout[i],
			d.New(i),
			d.New(pa.Left().String()),
			d.New(pa.Right().String()))
	}
	return d.StringChainTable(strout...)
}

/// CONSTANT
func (c constant) String() string { return c().(d.Data).String() }

/// VECTOR
func (v vector) String() string {
	var slice []d.Data
	for _, dat := range v() {
		slice = append(slice, dat)
	}
	return d.StringSlice("∙", "[", "]", slice...)
}

/// LIST
func (l list) String() string {
	var h, t = l()
	if t != nil {
		return h.String() + ", " + t.String()
	}
	return h.String()
}

/// TUPLE
func (t tuple) String() string {
	var slice []d.Data
	var v, _ = t()
	for _, dat := range v.(vector)() {
		slice = append(slice, dat)
	}
	return d.StringSlice(", ", "(", ")", slice...)
}

/// RECORD
func (r record) String() string {
	var str = "{"
	var l = r.Len()
	for i, pair := range r.Slice() {
		str = str + pair.(Paired).Left().String() + "::" +
			pair.(Paired).Right().String()
		if i < l-1 {
			str = str + ", "
		}
	}
	return str + "}"
}

//// FLAG
func (t Flag) String() string {
	u, k, p := t()
	var str = "[" + strconv.Itoa(u) + "] "
	if bits.OnesCount(k.Uint()) > 1 {
		var flags = d.FlagDecompose(d.BitFlag(k))
		for i, f := range flags {
			str = str + Kind(f).String()
			if i < len(flags)-1 {
				str = str + "|"
			}
		}
	} else {
		str = Kind(k).String()
	}
	return str + ":" + d.StringBitFlag(d.BitFlag(p))
}
func (f FlagSet) String() string {
	var str = "["
	var l = len(f())
	for i, flag := range f() {
		str = str + flag.String()
		if i < l-1 {
			str = str + ", "
		}
	}
	return str + "]"
}

///// PATTERNS MONOID
func (s pattern) String() string {
	var str string
	for i, tok := range s.Tokens() {
		str = str + tok.String()
		if i < len(s.Tokens())-1 {
			str = str + " "
		}
	}
	return strconv.Itoa(s.Id()) + str
}
func (s monoid) String() string {
	return strconv.Itoa(s.Id()) + " " + tokens(s.Tokens()).String()
}
func (s polymorph) String() string {
	var str string
	for _, mon := range s.Monom() {
		str = str + tokens(mon.Tokens()).String() + "\n"
	}
	return strconv.Itoa(s.Id()) + " " + str
}

//// TOKENS
func (t tokens) String() string {
	var str string
	for _, tok := range t {
		str = str + " " + tok.String()
	}
	return str
}
func (t tokenSlice) String() string {
	var str string
	for _, s := range t {
		str = str + tokens(s).String() + "\n"
	}
	return str
}
func (t token) String() string {
	var str string
	switch t.typ {
	case Syntax_Token:
		str = t.flag.(l.TokType).Syntax()
	case Data_Type_Token:
		str = d.StringBitFlag(t.flag.(d.Type).Flag())
	case Kind_Token:
		str = d.StringBitFlag(t.flag.(Kind).Flag())
	default:
		str = "Don't know how to print this token"
	}
	return str
}
func (t dataToken) String() string {
	var str string
	switch t.typ {
	case Data_Value_Token:
		str = t.d.(d.Data).String()
	case Argument_Token:
		str = "Arg: " + d.Type(t.Flag()).String()
	case Return_Token:
		str = "Ret: " + d.Type(t.Flag()).String()
	}
	return str
}

/// FUNCTION DEFINITION
func (f function) String() string {
	var rows = [][]d.Data{}
	pairs, _ := f()
	props := pairs[0:12]
	for i, pair := range props {
		rows = append(rows, []d.Data{})
		rows[i] = append(rows[i],
			d.IntVal(i),
			d.StrVal(pair.Left().String()),
			d.StrVal(pair.Right().String()),
		)
	}

	var row12a = [][]d.Data{}
	flags := f.ArgTypes()
	for i, flag := range flags {
		row12a = append(row12a, []d.Data{})
		row12a[i] = append(row12a[i],
			d.StrVal(strconv.Itoa(i)),
			d.StrVal(flag.String()),
		)
	}

	var row12b = [][]d.Data{}
	args := f.Accs()
	for i, arg := range args {
		row12b = append(row12b, []d.Data{})
		row12b[i] = append(row12b[i],
			d.StrVal(strconv.Itoa(i)),
			d.StrVal(arg.Flag().String()),
			d.StrVal(arg.(Argumented).Data().String()),
		)
	}

	var lastRow = [][]d.Data{}
	for i, _ := range flags {
		if f.AccessType() == NamedArgs {
			lastRow = append(lastRow, []d.Data{d.IntVal(len(rows)), row12b[i][1], row12b[i][2]})
		} else {
			lastRow = append(lastRow, []d.Data{d.IntVal(len(rows)), d.IntVal(i), flags[i]})
		}
	}

	rows = append(rows, lastRow...)
	return d.StringChainTable(rows...)
}
