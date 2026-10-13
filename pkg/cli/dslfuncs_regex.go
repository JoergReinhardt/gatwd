package cli

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/joergreinhardt/gatwd/functions"
)

// ---- Regular expressions (hardening W2, tier 2) -----------------------------------------
//
// Patterns are Go RE2 syntax: the engine runs in time linear in the input, so a pattern from
// an untrusted source cannot cause catastrophic backtracking (and, consequently, has no
// backreferences or lookaround — such a pattern is a compile error). A pattern is a String
// (write `\d` as "\\d" in a String literal). Matching is unanchored. An invalid pattern is an
// error naming the pattern; "no match" is a Maybe or an empty list, never an error.
//
//	re_match pat s (Bool) · re_find pat s (Maybe String) · re_find_all pat s (List String)
//	re_groups pat s (Maybe (List String): the whole match, then each group; a group that
//	took no part is "") · re_replace pat repl s (repl expands $1 and ${name}; $$ is a literal $)
//	re_split pat s (List String) · re_quote s (escape every regex metacharacter)

const maxPatternLen = 1 << 12

var (
	reCacheMu sync.Mutex
	reCache   = map[string]*regexp.Regexp{}
)

// compileRegex compiles pat, caching the result. The cache is bounded (it is dropped when it
// reaches 256 patterns) so a script generating patterns cannot grow it without limit.
func compileRegex(pat string) (*regexp.Regexp, error) {
	if len(pat) > maxPatternLen {
		return nil, fmt.Errorf("pattern longer than %d bytes", maxPatternLen)
	}
	reCacheMu.Lock()
	defer reCacheMu.Unlock()
	if re, ok := reCache[pat]; ok {
		return re, nil
	}
	re, err := regexp.Compile(pat)
	if err != nil {
		return nil, fmt.Errorf("invalid pattern %q: %w", pat, err)
	}
	if len(reCache) >= 256 {
		reCache = map[string]*regexp.Regexp{}
	}
	reCache[pat] = re
	return re, nil
}

func registerRegexFuncs(m map[string]any) {
	// withRegex reads (pattern, subject) and hands the compiled pattern to f.
	withRegex := func(name string, f func(re *regexp.Regexp, s string) any) {
		defn(m, name, "pat s", func(a []any) (any, error) {
			pat, err := asString(a[0])
			if err != nil {
				return nil, fmt.Errorf("pattern: %w", err)
			}
			s, err := asString(a[1])
			if err != nil {
				return nil, fmt.Errorf("subject: %w", err)
			}
			re, err := compileRegex(pat)
			if err != nil {
				return nil, err
			}
			return f(re, s), nil
		})
	}
	strList := func(ss []string) any {
		out := make([]any, len(ss))
		for i, s := range ss {
			out[i] = s
		}
		return listOf(out)
	}
	withRegex("re_match", func(re *regexp.Regexp, s string) any { return re.MatchString(s) })
	withRegex("re_find", func(re *regexp.Regexp, s string) any {
		loc := re.FindStringIndex(s)
		if loc == nil {
			return functions.Nothing[any]()
		}
		return functions.Just[any](s[loc[0]:loc[1]])
	})
	withRegex("re_find_all", func(re *regexp.Regexp, s string) any { return strList(re.FindAllString(s, -1)) })
	withRegex("re_groups", func(re *regexp.Regexp, s string) any {
		g := re.FindStringSubmatch(s)
		if g == nil {
			return functions.Nothing[any]()
		}
		return functions.Just[any](strList(g))
	})
	withRegex("re_split", func(re *regexp.Regexp, s string) any { return strList(re.Split(s, -1)) })

	defn(m, "re_replace", "pat repl s", func(a []any) (any, error) {
		pat, err := asString(a[0])
		if err != nil {
			return nil, fmt.Errorf("pattern: %w", err)
		}
		repl, err := asString(a[1])
		if err != nil {
			return nil, fmt.Errorf("replacement: %w", err)
		}
		s, err := asString(a[2])
		if err != nil {
			return nil, fmt.Errorf("subject: %w", err)
		}
		re, err := compileRegex(pat)
		if err != nil {
			return nil, err
		}
		return re.ReplaceAllString(s, repl), nil
	})
	defn(m, "re_quote", "s", func(a []any) (any, error) {
		s, err := asString(a[0])
		if err != nil {
			return nil, err
		}
		return regexp.QuoteMeta(s), nil
	})
}
