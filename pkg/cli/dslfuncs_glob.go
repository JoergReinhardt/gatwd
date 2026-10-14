package cli

import (
	"path"
	"strings"
)

// ---- Path globbing -----------------------------------------------------------------------
//
// glob_match pat s (Bool) · glob_root pat (String)
//
// A pattern with no "*", "?" or "[" is a literal directory: it matches itself and everything
// under it (s == pat, or s starts with pat + "/"), so a plain root behaves exactly as it always
// has — nothing to opt into. A pattern that uses them is matched segment by segment ("/"
// delimited): "*", "?" and "[...]" via the standard single-segment path.Match, and "**" (its own
// segment) matching zero or more whole segments, so it can stand for "any number of directories
// in between" — ".../manifests/**/overlays/staging" reaches every project's staging overlay
// however deep its own structure goes, while ".../manifests/sip*/overlays/staging" narrows to a
// name prefix. glob_root returns the longest literal (wildcard-free) directory prefix of a
// pattern — the one real, listable directory `walk_fs` can start from; glob_match then filters
// what it finds down to what the whole pattern actually names.

// hasGlobChars reports whether s contains any of the wildcard characters glob_match treats
// specially; a pattern without them is a plain literal path.
func hasGlobChars(s string) bool {
	return strings.ContainsAny(s, "*?[")
}

// globMatchSegments matches "/"-delimited path segments against pattern segments, where a
// pattern segment that is exactly "**" consumes zero or more path segments and any other
// pattern segment matches exactly one path segment via path.Match (so "*"/"?"/"[...]" apply
// within a segment, never across a "/").
func globMatchSegments(pat, s []string) bool {
	if len(pat) == 0 {
		return len(s) == 0
	}
	if pat[0] == "**" {
		if globMatchSegments(pat[1:], s) {
			return true
		}
		if len(s) > 0 && globMatchSegments(pat, s[1:]) {
			return true
		}
		return false
	}
	if len(s) == 0 {
		return false
	}
	ok, err := path.Match(pat[0], s[0])
	if err != nil || !ok {
		return false
	}
	return globMatchSegments(pat[1:], s[1:])
}

// globMatch reports whether s matches pattern (see the package-level doc above for the two
// cases: a literal prefix, or a real "*"/"?"/"[...]"/"**" glob).
func globMatch(pattern, s string) bool {
	if !hasGlobChars(pattern) {
		return s == pattern || strings.HasPrefix(s, pattern+"/")
	}
	return globMatchSegments(strings.Split(pattern, "/"), strings.Split(s, "/"))
}

// globRoot returns the longest wildcard-free directory prefix of pattern: pattern itself when
// it has no wildcard, else everything before the path separator preceding the first wildcard
// character ("." when the wildcard is in the very first segment).
func globRoot(pattern string) string {
	idx := strings.IndexAny(pattern, "*?[")
	if idx == -1 {
		return pattern
	}
	prefix := pattern[:idx]
	if i := strings.LastIndex(prefix, "/"); i >= 0 {
		return prefix[:i]
	}
	return "."
}

func registerGlobFuncs(m map[string]any) {
	defn(m, "glob_match", "pat s", func(a []any) (any, error) {
		pat, err := asString(a[0])
		if err != nil {
			return nil, err
		}
		s, err := asString(a[1])
		if err != nil {
			return nil, err
		}
		return globMatch(pat, s), nil
	})
	defn(m, "glob_root", "pat", func(a []any) (any, error) {
		pat, err := asString(a[0])
		if err != nil {
			return nil, err
		}
		return globRoot(pat), nil
	})
}
