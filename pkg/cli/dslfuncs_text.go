package cli

import (
	"fmt"
	"unicode"

	"github.com/joergreinhardt/gatwd/functions"
)

// registerCharFuncs adds the Char function forms (hardening W2, D1). All are data-last
// and curried; each argument that must be a Char also accepts a one-character String
// (the conversion table in valueops.go).
//
//	char_code c            Char → Int (the Unicode code point)
//	char_from_code n       Int → Char (an error for an invalid code point)
//	char_to_str c          Char → String
//	char_of_str s          String → Maybe Char (Nothing unless s is exactly one character)
//	char_is_digit/alpha/space/upper/lower c   Char → Bool
//	char_upper/lower c     Char → Char
func registerCharFuncs(m map[string]any) {
	defn(m, "char_code", "c", func(a []any) (any, error) {
		c, err := asChar(a[0])
		if err != nil {
			return nil, err
		}
		return int(c), nil
	})
	defn(m, "char_from_code", "n", func(a []any) (any, error) {
		n, err := asInt(a[0])
		if err != nil {
			return nil, err
		}
		c, ok := functions.CharFromCode(int64(n))
		if !ok {
			return nil, fmt.Errorf("%d is not a valid Unicode code point", n)
		}
		return c, nil
	})
	defn(m, "char_to_str", "c", func(a []any) (any, error) {
		c, err := asChar(a[0])
		if err != nil {
			return nil, err
		}
		return c.String(), nil
	})
	defn(m, "char_of_str", "s", func(a []any) (any, error) {
		s, ok := a[0].(string)
		if !ok {
			return nil, fmt.Errorf("want a String, got %T", a[0])
		}
		if c, ok := functions.CharOfString(s); ok {
			return functions.Just[any](c), nil
		}
		return functions.Nothing[any](), nil
	})

	predicates := map[string]func(rune) bool{
		"char_is_digit": unicode.IsDigit,
		"char_is_alpha": unicode.IsLetter,
		"char_is_space": unicode.IsSpace,
		"char_is_upper": unicode.IsUpper,
		"char_is_lower": unicode.IsLower,
	}
	for name, pred := range predicates {
		defn(m, name, "c", func(a []any) (any, error) {
			c, err := asChar(a[0])
			if err != nil {
				return nil, err
			}
			return pred(rune(c)), nil
		})
	}
	defn(m, "char_upper", "c", func(a []any) (any, error) {
		c, err := asChar(a[0])
		if err != nil {
			return nil, err
		}
		return functions.Char(unicode.ToUpper(rune(c))), nil
	})
	defn(m, "char_lower", "c", func(a []any) (any, error) {
		c, err := asChar(a[0])
		if err != nil {
			return nil, err
		}
		return functions.Char(unicode.ToLower(rune(c))), nil
	})
}
