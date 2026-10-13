package cli

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/joergreinhardt/gatwd/functions"
)

func TestRegexValues(t *testing.T) {
	in := NewInterpreter()
	for _, tc := range []struct{ src, want string }{
		{`re_match "^v[0-9]+$" "v12"`, "true"},
		{`re_match "^v[0-9]+$" "v12x"`, "false"},
		{`re_match "\\d+" "abc 42"`, "true"},
		{`re_match "" "anything"`, "true"},
		{`re_match "[(]" "f(x)"`, "true"},
		{`re_find "[a-z]+" "123abc456def"`, "Just(abc)"},
		{`re_find "z" "abc"`, "Nothing"},
		{`re_find "a*" "baaa"`, "Just()"}, // the leftmost match may be empty
		{`re_find_all "\\d+" "a1b22c333"`, "[1 22 333]"},
		{`re_find_all "\\d+" "abc"`, "[]"},
		{`re_groups "(\\w+)@(\\w+)\\.com" "bob@example.com"`, "Just([bob@example.com bob example])"},
		{`re_groups "(a)|(b)" "b"`, "Just([b  b])"}, // the group that took no part is ""
		{`re_groups "x" "y"`, "Nothing"},
		{`re_groups "(?P<k>\\w+)=(?P<v>\\w+)" "a=b"`, "Just([a=b a b])"},
		{`re_replace "(\\w+)@(\\w+)" "$2 at $1" "bob@example"`, "example at bob"},
		{`re_replace "\\s+" " " "a   b \t c"`, "a b c"},
		{`re_replace "x" "$$" "axb"`, "a$b"},
		{`re_replace "z" "y" "abc"`, "abc"},
		{`re_split "[,;]\\s*" "a, b;c"`, "[a b c]"},
		{`re_split "," ""`, "[]"},
		{`re_quote "a.b*c"`, `a\.b\*c`},
		{`re_match (re_quote "a.b") "a.b"`, "true"},
		{`re_match (re_quote "a.b") "axb"`, "false"},
		{`re_match "é" "café"`, "true"},
		{`"v1.2" |> re_find "\\d+\\.\\d+"`, "Just(1.2)"},
	} {
		if got := displayText(evalOne(t, in, tc.src)); got != tc.want {
			t.Errorf("%s = %q, want %q", tc.src, got, tc.want)
		}
	}
}

func TestRegexStructural(t *testing.T) {
	in := NewInterpreter()
	items := itemsOf(t, evalOne(t, in, `re_find_all "\\d+" "a1b22c333"`))
	if len(items) != 3 || items[0] != "1" || items[2] != "333" {
		t.Errorf("re_find_all = %#v", items)
	}
	mb, ok := evalOne(t, in, `re_find "[a-z]+" "123abc"`).(functions.Maybe[any])
	if v, _ := mb.MaybeValue(); !ok || v != "abc" {
		t.Errorf("re_find = %#v", v)
	}
}

func TestRegexFailsLoud(t *testing.T) {
	in := NewInterpreter()
	for _, tc := range []struct{ src, want string }{
		{`re_match "(" "x"`, "invalid pattern"},
		{`re_match "[a" "x"`, "invalid pattern"},
		{`re_match "(a)\\1" "aa"`, "invalid pattern"}, // RE2 has no backreferences
		{`re_match "(?=a)a" "a"`, "invalid pattern"},  // ... and no lookahead
		{`re_match 5 "x"`, "pattern: want a String"},
		{`re_match "a" 5`, "subject: want a String"},
		{`re_replace "(" "x" "y"`, "invalid pattern"},
		{`re_replace "a" 5 "y"`, "replacement"},
		{`re_split "*" "x"`, "invalid pattern"},
		{`re_find_all "a{1001}" "x"`, "invalid pattern"},
	} {
		res := in.evaluatePipelineStr(tc.src)
		if len(res) != 1 || !strings.Contains(res[0], "[ERROR]") || !strings.Contains(res[0], tc.want) {
			t.Errorf("%s = %v, want an [ERROR] containing %q", tc.src, res, tc.want)
		}
	}
	long := strings.Repeat("a", maxPatternLen+1)
	if _, err := compileRegex(long); err == nil || !strings.Contains(err.Error(), "longer than") {
		t.Errorf("an over-long pattern must be rejected, got %v", err)
	}
}

// RE2 runs in time linear in the input: the classic catastrophic-backtracking pattern on a
// long non-matching subject finishes at once instead of hanging.
func TestRegexIsLinearTime(t *testing.T) {
	in := NewInterpreter()
	subject := strings.Repeat("a", 200000) + "b"
	in.Constants["SUBJ"] = subject
	start := time.Now()
	got := in.EvaluateToValues(`re_match "^(a+)+$" SUBJ`)
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("the catastrophic pattern took %v; the engine must be linear-time", elapsed)
	}
	if len(got) != 1 || got[0] != false {
		t.Errorf("(a+)+$ on a...ab = %v, want false", got)
	}
}

func TestRegexCacheIsBounded(t *testing.T) {
	for i := 0; i < 600; i++ {
		if _, err := compileRegex("x" + strconv.Itoa(i)); err != nil {
			t.Fatal(err)
		}
	}
	reCacheMu.Lock()
	n := len(reCache)
	reCacheMu.Unlock()
	if n > 256 {
		t.Errorf("the cache grew to %d entries, want at most 256", n)
	}
}

func TestRegexAsSteps(t *testing.T) {
	in := NewInterpreter()
	for _, c := range [][2]string{
		{`"abc 42" |> re_find "\\d+"`, `re_find "\\d+" "abc 42"`},
		{`"a1b2" |> re_find_all "\\d"`, `re_find_all "\\d" "a1b2"`},
		{`"a  b" |> re_replace "\\s+" "-"`, `re_replace "\\s+" "-" "a  b"`},
	} {
		step := strings.Join(in.evaluatePipelineStr(c[0]), ";")
		fn := strings.Join(in.evaluatePipelineStr(c[1]), ";")
		if step == "" || step != fn || strings.Contains(step, "ERROR") {
			t.Errorf("%s = %q but %s = %q", c[0], step, c[1], fn)
		}
	}
}

// Statement assembly ignores delimiters inside String and Char literals, so a pattern such as
// "(" or "[(]" does not swallow the following lines.
func TestScanLine(t *testing.T) {
	for _, tc := range []struct {
		s    string
		open bool
	}{
		{`re_match "(" "x"`, false},
		{`re_match "[(]" "x"`, false},
		{`'(' == x`, false},
		{`'\"' == x`, false},
		{`'"' == "(" `, false},
		{`f (g 1`, true},
		{`[1, 2`, true},
		{`xs |> map (\x -> x`, true},
		{`"open (`, true},
		{`x = {`, true},
		{`f (g 1))`, false},
		{`"a\"(b"`, false},
		{`'\n' == c`, false},
	} {
		if got := scanLine(tc.s).open(); got != tc.open {
			t.Errorf("scanLine(%s).open() = %v, want %v", tc.s, got, tc.open)
		}
	}
	lines := logicalLines("re_match \"(\" \"x\"\nlength \"abc\"\nre_match \"[(]\" \"f(\"\nxs = [1,\n2]\n")
	if len(lines) != 4 {
		t.Errorf("got %d logical lines %v, want 4 (a real multi-line list must still join)", len(lines), lines)
	}
}

// The pipeline and match-arm splitters treat String and Char literals as text: a `(`, `[` or
// `|>` inside one is not structure.
func TestSplittersRespectLiterals(t *testing.T) {
	for _, tc := range []struct {
		line string
		want []string
	}{
		{`re_match "[(]" "f(x)" |> assert true`, []string{`re_match "[(]" "f(x)"`, `assert true`}},
		{`xs |> str_split "|>" |> length`, []string{`xs`, `str_split "|>"`, `length`}},
		{`'(' == c |> assert true`, []string{`'(' == c`, `assert true`}},
		{`'\n' == c |> f`, []string{`'\n' == c`, `f`}},
		{`"a\"(b" |> length`, []string{`"a\"(b"`, `length`}},
		{`f "unterminated (`, []string{`f "unterminated (`}},
		{`a |> (b |> c) |> d`, []string{`a`, `(b |> c)`, `d`}},
	} {
		got, _ := splitPipelineOps(tc.line)
		if strings.Join(got, "§") != strings.Join(tc.want, "§") {
			t.Errorf("splitPipelineOps(%s) = %q, want %q", tc.line, got, tc.want)
		}
	}
	arms := splitCasesAware(`"a|b" -> echo x | 'c' -> echo y | _ -> echo "(" `)
	if len(arms) != 3 {
		t.Errorf("splitCasesAware must split on the two arm separators only, got %q", arms)
	}
}
