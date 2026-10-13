package functions

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

// Char is a single Unicode code point: the element type of a String.
//
// # The concept
//
// A text is a sequence of characters. Keeping the element type distinct from both
// Int and a one-character String lets the language say what it means: `'a'` is not 97
// and not "a", although the three convert on request (`char_code 'a'` is 97,
// `char_to_str 'a'` is "a"). Haskell's Char and Elm's Char are the model.
//
// # Why a named type
//
// Char is `type Char rune`, not a bare rune (int32). Go's type switches distinguish a
// named type from its underlying type, so a Char never matches the `case int32` arms of
// the numeric tower and silently turns into a number: arithmetic on a Char is an error,
// not a surprise.
//
// # Example
//
//	c := functions.Char('é')
//	c.String() // "é"
type Char rune

// String renders the character itself (no quotes), the way a String prints.
func (c Char) String() string { return string(rune(c)) }

// IsChar reports whether v is a Char.
func IsChar(v any) bool { _, ok := v.(Char); return ok }

// CharFromCode returns the Char with the given Unicode code point, or false when code is
// not a valid code point (negative, a surrogate half, or beyond U+10FFFF).
func CharFromCode(code int64) (Char, bool) {
	if code < 0 || code > unicode.MaxRune || (code >= 0xD800 && code <= 0xDFFF) {
		return 0, false
	}
	return Char(rune(code)), true
}

// CharOfString returns the Char a one-character string denotes. It is the String → Char
// half of the on-demand conversion: it is false for the empty string, for a string of two
// or more characters, and for invalid UTF-8.
func CharOfString(s string) (Char, bool) {
	r, size := utf8.DecodeRuneInString(s)
	if size == 0 || size != len(s) || (r == utf8.RuneError && size == 1) {
		return 0, false
	}
	return Char(r), true
}

// charPair resolves a comparison in which at least one side is a Char: the other side
// may be a Char or a one-character String (the D1 on-demand conversion). It is false when
// neither side is a Char or the other side does not convert.
func charPair(a, b any) (Char, Char, bool) {
	ca, aIsChar := a.(Char)
	cb, bIsChar := b.(Char)
	switch {
	case aIsChar && bIsChar:
		return ca, cb, true
	case aIsChar:
		if s, ok := b.(string); ok {
			if c, ok := CharOfString(s); ok {
				return ca, c, true
			}
		}
	case bIsChar:
		if s, ok := a.(string); ok {
			if c, ok := CharOfString(s); ok {
				return c, cb, true
			}
		}
	}
	return 0, 0, false
}

// CharsOfString splits a String into its Chars, the String → collection-of-Chars half of
// the on-demand conversion. Invalid UTF-8 is returned as an error rather than silently
// becoming U+FFFD.
func CharsOfString(s string) ([]Char, error) {
	out := make([]Char, 0, len(s))
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			return nil, fmt.Errorf("invalid UTF-8 at byte %d", i)
		}
		out = append(out, Char(r))
		i += size
	}
	return out, nil
}

// StringOfChars joins Chars into a String, the collection-of-Chars → String half of the
// on-demand conversion.
func StringOfChars(cs []Char) string {
	rs := make([]rune, len(cs))
	for i, c := range cs {
		rs[i] = rune(c)
	}
	return string(rs)
}
