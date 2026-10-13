package parser_test

import (
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/parser"
)

func TestCharLiteral(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want functions.Char
	}{
		{`'a'`, 'a'},
		{`'é'`, 'é'},
		{`'😀'`, '😀'},
		{`' '`, ' '},
		{`'\n'`, '\n'},
		{`'\t'`, '\t'},
		{`'\r'`, '\r'},
		{`'\0'`, 0},
		{`'\\'`, '\\'},
		{`'\''`, '\''},
		{`'\"'`, '"'},
		{`'\x41'`, 'A'},
		{`'ὠ'`, 'ὠ'},
	} {
		val, err := parser.EvalInfixString(tc.src, nil)
		if err != nil {
			t.Errorf("%s: %v", tc.src, err)
			continue
		}
		if got, ok := val.(functions.Char); !ok || got != tc.want {
			t.Errorf("%s = %#v (%T), want Char %q", tc.src, val, val, rune(tc.want))
		}
	}
}

func TestCharLiteralErrors(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{`'`, "unterminated"},
		{`'a`, "exactly one character"},
		{`'ab'`, "exactly one character"},
		{`''`, "empty"},
		{`'\q'`, "unknown escape"},
		{`'\x4'`, "invalid"},
		{`'\x4`, "unterminated"},
		{`'\xZZ'`, "invalid"},
		{`'\uD800'`, "invalid code point"},
		{`'\`, "unterminated"},
	} {
		_, err := parser.EvalInfixString(tc.src, nil)
		if err == nil {
			t.Errorf("%s: want an error containing %q", tc.src, tc.want)
			continue
		}
		if !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: error %q does not contain %q", tc.src, err, tc.want)
		}
	}
}

// A Char literal is an atom: it can be an argument, a list element and an operand.
func TestCharLiteralInContext(t *testing.T) {
	lookup := func(name string) (any, bool) {
		if name == "code" {
			return functions.NewClosure([]string{"c"}, func(a []any) (any, error) {
				return int(a[0].(functions.Char)), nil
			}), true
		}
		return nil, false
	}
	if v, err := parser.EvalInfixString(`code 'a'`, lookup); err != nil || v != 97 {
		t.Errorf("code 'a' = %v, %v; want 97", v, err)
	}
	if v, err := parser.EvalInfixString(`'a' < 'b' && 'b' == 'b'`, nil); err != nil || v != true {
		t.Errorf("char comparison = %v, %v; want true", v, err)
	}
	v, err := parser.EvalInfixString(`['a', 'b']`, nil)
	if err != nil {
		t.Fatal(err)
	}
	items, _ := functions.ToSliceAny(v)
	if len(items) != 2 || items[0] != functions.Char('a') || items[1] != functions.Char('b') {
		t.Errorf("['a', 'b'] = %#v", items)
	}
}

func TestStringLiteralEscapes(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{`"plain"`, "plain"},
		{`"a\nb"`, "a\nb"},
		{`"tab\there"`, "tab\there"},
		{`"cr\r"`, "cr\r"},
		{`"nul\0"`, "nul\x00"},
		{`"back\\slash"`, `back\slash`},
		{`"say \"hi\""`, `say "hi"`},
		{`"it\'s"`, "it's"},
		{`"A\x41é"`, "AAé"},
		{`"é😀"`, "é😀"},
		{`""`, ""},
		{`"\\"`, `\`},
		{`"\\n"`, `\n`}, // an escaped backslash followed by n: two characters
	} {
		val, err := parser.EvalInfixString(tc.src, nil)
		if err != nil {
			t.Errorf("%s: %v", tc.src, err)
			continue
		}
		if val != tc.want {
			t.Errorf("%s = %q, want %q", tc.src, val, tc.want)
		}
	}
}

func TestStringLiteralEscapeErrors(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{`"bad\q"`, "unknown escape"},
		{`"open`, "unterminated"},
		{`"open\"`, "unterminated"},
		{`"x\`, "unterminated"},
		{`"\x4"`, "invalid"},
		{`"\x4`, "unterminated"},
		{`"\uD800"`, "invalid code point"},
		{`"\xZZ"`, "invalid"},
	} {
		_, err := parser.EvalInfixString(tc.src, nil)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: error %v, want one containing %q", tc.src, err, tc.want)
		}
	}
}

// A Char and a String share one escape table.
func TestCharAndStringShareEscapes(t *testing.T) {
	for _, esc := range []string{`\n`, `\t`, `\r`, `\0`, `\\`, `\'`, `\"`, `\x41`, `é`} {
		c, err := parser.EvalInfixString(`'`+esc+`'`, nil)
		if err != nil {
			t.Fatalf("char %s: %v", esc, err)
		}
		s, err := parser.EvalInfixString(`"`+esc+`"`, nil)
		if err != nil {
			t.Fatalf("string %s: %v", esc, err)
		}
		if string(rune(c.(functions.Char))) != s {
			t.Errorf("%s: char %q and string %q disagree", esc, c, s)
		}
	}
}

// Ordering is declared first in every Parser, so its Variants are equal across Parsers and
// to OrderingVariant — the value `compare` returns equals the constant a script writes.
func TestOrderingIsSharedAcrossParsers(t *testing.T) {
	for _, tag := range []string{"LT", "EQ", "GT"} {
		want, ok := parser.OrderingVariant(tag)
		if !ok {
			t.Fatalf("OrderingVariant(%q) not found", tag)
		}
		got, ok := parser.NewParser().MakeVariant("Ordering", tag)
		if !ok || !got.Equal(want) {
			t.Errorf("%s: a fresh Parser built %#v, OrderingVariant %#v", tag, got, want)
		}
		if !parser.NewParser().IsConstructor(tag) {
			t.Errorf("%s must be a registered constructor", tag)
		}
	}
	if _, ok := parser.OrderingVariant("MAYBE"); ok {
		t.Error("OrderingVariant must reject an unknown tag")
	}
}
