package evaluator

import (
	"fmt"
	"math"
	"math/big"
	"sort"
	"strconv"
	"strings"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
)

// This file is the key design of the DSL's HAMT: how a VALUE (a number, a coordinate
// pair, a Maybe, ...) becomes the string key of a *collections.HAMT[string, any] and back.
//
// What is the concept? A dictionary needs one notion of "the same key". For gatwd it is
// the language's own `==`: two keys are the same entry exactly when `a == b`. That has
// two halves. Equal keys must meet (`1` and `1.0`, `[1, 2]` and the Pair 1,2 are `==`,
// so they are one entry) and different keys must never meet (`[1, 2]` and `"[1 2]"` are
// different values, so they are two entries; two different collections never share one).
//
// Why is it needed? The HAMT is keyed by Go strings (the JSON/YAML/manifest records
// that fill it, the optics and every adapter, all use plain string keys). The DSL used
// to coerce a non-String key to its printed form, which is lossy (a list of Ints and a
// list of Strings print alike) and which is why a String once had to equal the printed
// form of a collection. EncodeKey replaces the coercion by an encoding that is
// canonical (equal keys encode alike, see below), injective (different keys encode
// differently) and invertible (DecodeKey returns a value `==` the original), so a key
// read back from a HAMT is a real value and not text.
//
// How is it implemented? A String key stays exactly itself, so every record read from
// JSON, YAML or a manifest is unchanged. Any other value is written as keyMark (the NUL
// byte, which no real name starts with) plus a tagged, length-prefixed canonical text:
//
//	String, Char      the text itself (a text that starts with keyMark is escaped: keyMark "s" text)
//	Bool              b0 / b1
//	number            i<decimal>  whole numbers (CanonNumber: Int, big integer, whole Float or Ratio)
//	                  f<g>        a fraction a double represents exactly, and ±Inf
//	                  r<n>/<d>    any other Ratio
//	                  c<re>,<im>  a Complex with a non-zero imaginary part
//	List Chunk Pair   l<n>:<len>:<elem>...         (ordered; all sequences that are == share it)
//	slice Bytes
//	Set               S<n>:<len>:<elem>...         (elements sorted: a Set is unordered)
//	HAMT              M<n>:<len>:<key><len>:<elem>... (entries sorted by key)
//	Tree              T<len>:<value><n>:<len>:<child>...
//	Maybe, Either     N | J<elem> | L<elem> | R<elem>
//	Variant           V<len>:<tag><flag>:<elem>
//
// where <elem> is the same grammar again, so a key may be any nesting of these. A
// String element is always tagged inside a composite, so `["a"]` and `[1]` differ.
// Values with no such form — a function, a Graph, NaN (equal to nothing), a Go value
// this file does not know — are an error, never a coerced guess.
//
// Two deliberate edges: a Complex number is compared by `==` through doubles, so a big
// integer and a Complex that only round to the same double are `==` but different keys
// (the key is stricter, never lossy); and a Set or HAMT key holding values that are not
// comparable cannot be rebuilt as a Set by DecodeKey and decodes to the same text.

// keyMark starts the encoding of every key that is not a plain String. NUL cannot occur
// in a name (Kubernetes ids, JSON keys, file names), and a String that does start with it
// is escaped, so a plain key and an encoded one never collide.
const keyMark = "\x00"

// EncodeKey turns a DSL value into the string key of a *collections.HAMT[string, any]. A String is
// its own key; every other value gets the canonical encoding documented above, so two
// values have the same key exactly when they are `==`. It reports an error for a value
// that cannot be a key (a function, a Graph, NaN, an unknown Go type).
func EncodeKey(v any) (string, error) {
	switch x := v.(type) {
	case string:
		return plainKey(x), nil
	case functions.Char:
		return plainKey(string(rune(x))), nil
	}
	enc, err := encodeKeyValue(v)
	if err != nil {
		return "", err
	}
	return keyMark + enc, nil
}

// plainKey is a String as a key: itself, except that a text starting with keyMark is
// escaped so it cannot be mistaken for an encoded key.
func plainKey(s string) string {
	if strings.HasPrefix(s, keyMark) {
		return keyMark + "s" + s
	}
	return s
}

// DecodeKey is the inverse of EncodeKey: the value a key stands for. A plain key is the
// String it spells; an encoded one decodes to a value that is `==` the one it was made
// from (a Chunk or a Pair decodes to a List, a whole Float to an Int, ...). A key
// that starts with keyMark but is not valid encoding — it was not made by EncodeKey —
// decodes to itself as a String rather than failing.
func DecodeKey(key string) any {
	if !strings.HasPrefix(key, keyMark) {
		return key
	}
	v, err := decodeKeyValue(key[len(keyMark):])
	if err != nil {
		return key
	}
	return v
}

// lengthPrefixed writes s as <len>:<s>, which makes a concatenation of elements
// unambiguous whatever bytes the elements contain.
func lengthPrefixed(s string) string { return strconv.Itoa(len(s)) + ":" + s }

func encodeKeyValue(v any) (string, error) {
	switch x := v.(type) {
	case nil:
		return "", fmt.Errorf("nil cannot be a HAMT key")
	case string:
		return "s" + x, nil
	case functions.Char:
		return "s" + string(rune(x)), nil
	case bool:
		if x {
			return "b1", nil
		}
		return "b0", nil
	case dsltypes.Variant:
		inner := "z" // a unit variant has no payload, which is not the same as a Nothing payload
		if x.Value != nil {
			enc, err := encodeKeyValue(x.Value)
			if err != nil {
				return "", err
			}
			inner = enc
		}
		return "V" + lengthPrefixed(x.Tag) + strconv.FormatUint(uint64(x.Flag), 10) + ":" + inner, nil
	case functions.Duration:
		return "d" + strconv.FormatInt(int64(x), 10), nil
	case functions.Time:
		return "t" + x.UTC().String(), nil // an instant: the display zone is not part of the key
	case functions.Date:
		return "D" + x.String(), nil
	case *collections.Tree[any]:
		return encodeTreeKey(x)
	case *collections.HAMT[string, any]:
		return encodeHAMTKey(x)
	}
	if functions.IsNumber(v) {
		return encodeNumberKey(v)
	}
	if m, ok := v.(interface{ MaybeValue() (any, bool) }); ok {
		inner, isJust := m.MaybeValue()
		if !isJust {
			return "N", nil
		}
		enc, err := encodeKeyValue(inner)
		return "J" + enc, err
	}
	if e, ok := v.(interface{ EitherValue() (any, any, bool) }); ok {
		l, r, isLeft := e.EitherValue()
		if isLeft {
			enc, err := encodeKeyValue(l)
			return "L" + enc, err
		}
		enc, err := encodeKeyValue(r)
		return "R" + enc, err
	}
	if s, ok := v.(interface {
		IsSet() bool
		ElementsAny() []any
	}); ok {
		return encodeSetKey(s.ElementsAny())
	}
	if elems, ok := functions.ToSliceAny(v); ok {
		parts := make([]string, len(elems))
		for i, e := range elems {
			enc, err := encodeKeyValue(e)
			if err != nil {
				return "", err
			}
			parts[i] = lengthPrefixed(enc)
		}
		return "l" + strconv.Itoa(len(elems)) + ":" + strings.Join(parts, ""), nil
	}
	return "", fmt.Errorf("a %s cannot be a HAMT key (keys are scalars, sequences, Sets, HAMTs, Trees, Maybes, Eithers and Variants)", functions.KindName(v))
}

func encodeNumberKey(v any) (string, error) {
	c, err := functions.CanonNumber(v)
	if err != nil {
		return "", err
	}
	switch x := c.(type) {
	case float64:
		return "f" + strconv.FormatFloat(x, 'g', -1, 64), nil
	case *functions.Ratio:
		return "r" + x.Num().String() + "/" + x.Denom().String(), nil
	case complex128:
		return "c" + strconv.FormatFloat(real(x), 'g', -1, 64) + "," + strconv.FormatFloat(imag(x), 'g', -1, 64), nil
	}
	b, ok := functions.ToBigInt(c)
	if !ok {
		return "", fmt.Errorf("%s has no key encoding", functions.KindName(v))
	}
	return "i" + b.String(), nil
}

func encodeSetKey(elems []any) (string, error) {
	parts := make([]string, len(elems))
	for i, e := range elems {
		enc, err := encodeKeyValue(e)
		if err != nil {
			return "", err
		}
		parts[i] = lengthPrefixed(enc)
	}
	sort.Strings(parts)
	return "S" + strconv.Itoa(len(parts)) + ":" + strings.Join(parts, ""), nil
}

func encodeHAMTKey(h *collections.HAMT[string, any]) (string, error) {
	type entry struct{ k, v string }
	var entries []entry
	var firstErr error
	for k, val := range h.Seq() {
		enc, err := encodeKeyValue(val)
		if err != nil && firstErr == nil {
			firstErr = err
		}
		entries = append(entries, entry{k, enc})
	}
	if firstErr != nil {
		return "", firstErr
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].k < entries[j].k })
	var b strings.Builder
	b.WriteString("M" + strconv.Itoa(len(entries)) + ":")
	for _, e := range entries {
		b.WriteString(lengthPrefixed(e.k))
		b.WriteString(lengthPrefixed(e.v))
	}
	return b.String(), nil
}

func encodeTreeKey(t *collections.Tree[any]) (string, error) {
	val, err := encodeKeyValue(t.Value())
	if err != nil {
		return "", err
	}
	var kids []string
	for _, k := range t.Children().Elements() {
		enc, err := encodeTreeKey(k.(*collections.Tree[any]))
		if err != nil {
			return "", err
		}
		kids = append(kids, lengthPrefixed(enc))
	}
	return "T" + lengthPrefixed(val) + strconv.Itoa(len(kids)) + ":" + strings.Join(kids, ""), nil
}

// ---- decoding ----------------------------------------------------------------------

// keyReader walks an encoded key; every read advances it, and any malformed input is an
// error (DecodeKey turns that into the raw String).
type keyReader struct{ s string }

// number reads decimal digits up to sep.
func (r *keyReader) number(sep byte) (int, error) {
	i := strings.IndexByte(r.s, sep)
	if i <= 0 {
		return 0, fmt.Errorf("malformed key")
	}
	n, err := strconv.Atoi(r.s[:i])
	if err != nil || n < 0 {
		return 0, fmt.Errorf("malformed key")
	}
	r.s = r.s[i+1:]
	return n, nil
}

// chunk reads one length-prefixed piece.
func (r *keyReader) chunk() (string, error) {
	n, err := r.number(':')
	if err != nil || n > len(r.s) {
		return "", fmt.Errorf("malformed key")
	}
	out := r.s[:n]
	r.s = r.s[n:]
	return out, nil
}

// elements reads <n>: followed by n length-prefixed encoded values.
func (r *keyReader) elements() ([]any, error) {
	n, err := r.number(':')
	if err != nil {
		return nil, err
	}
	out := make([]any, 0, n)
	for i := 0; i < n; i++ {
		enc, err := r.chunk()
		if err != nil {
			return nil, err
		}
		v, err := decodeKeyValue(enc)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

func decodeKeyValue(s string) (any, error) {
	if s == "" {
		return nil, fmt.Errorf("malformed key")
	}
	tag, rest := s[0], s[1:]
	switch tag {
	case 's':
		return rest, nil
	case 'b':
		return rest == "1", nil
	case 'd':
		n, err := strconv.ParseInt(rest, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("malformed key")
		}
		return functions.Duration(n), nil
	case 't':
		return functions.ParseTime(rest)
	case 'D':
		return functions.ParseDate(rest)
	case 'i':
		b, ok := new(big.Int).SetString(rest, 10)
		if !ok {
			return nil, fmt.Errorf("malformed key")
		}
		return functions.CanonInt(b), nil
	case 'f':
		f, err := strconv.ParseFloat(rest, 64)
		if err != nil || math.IsNaN(f) {
			return nil, fmt.Errorf("malformed key")
		}
		return f, nil
	case 'r':
		num, den, ok := strings.Cut(rest, "/")
		n, nok := new(big.Int).SetString(num, 10)
		d, dok := new(big.Int).SetString(den, 10)
		if !ok || !nok || !dok || d.Sign() == 0 {
			return nil, fmt.Errorf("malformed key")
		}
		return functions.NewRatio(n, d)
	case 'c':
		re, im, ok := strings.Cut(rest, ",")
		rf, rerr := strconv.ParseFloat(re, 64)
		imf, ierr := strconv.ParseFloat(im, 64)
		if !ok || rerr != nil || ierr != nil {
			return nil, fmt.Errorf("malformed key")
		}
		return complex(rf, imf), nil
	case 'l':
		r := &keyReader{rest}
		elems, err := r.elements()
		if err != nil {
			return nil, err
		}
		list := collections.Nil[any]()
		for i := len(elems) - 1; i >= 0; i-- {
			list = collections.Cons(elems[i], list)
		}
		return list, nil
	case 'S':
		r := &keyReader{rest}
		elems, err := r.elements()
		if err != nil {
			return nil, err
		}
		set := collections.NewSet[any]()
		for _, e := range elems {
			if err := tryAdd(&set, e); err != nil {
				return nil, err
			}
		}
		return set, nil
	case 'M':
		r := &keyReader{rest}
		n, err := r.number(':')
		if err != nil {
			return nil, err
		}
		h := collections.NewHAMT[string, any]()
		for i := 0; i < n; i++ {
			k, err := r.chunk()
			if err != nil {
				return nil, err
			}
			enc, err := r.chunk()
			if err != nil {
				return nil, err
			}
			v, err := decodeKeyValue(enc)
			if err != nil {
				return nil, err
			}
			h = h.Set(k, v)
		}
		return h, nil
	case 'T':
		r := &keyReader{rest}
		enc, err := r.chunk()
		if err != nil {
			return nil, err
		}
		val, err := decodeKeyValue(enc)
		if err != nil {
			return nil, err
		}
		kids, err := r.elements()
		if err != nil {
			return nil, err
		}
		list := collections.Nil[*collections.Tree[any]]()
		for i := len(kids) - 1; i >= 0; i-- {
			list = collections.Cons(kids[i].(*collections.Tree[any]), list)
		}
		return collections.NewTree[any](val, list), nil
	case 'N':
		return functions.Nothing[any](), nil
	case 'J':
		v, err := decodeKeyValue(rest)
		return functions.Just[any](v), err
	case 'L':
		v, err := decodeKeyValue(rest)
		return functions.Left[any, any](v), err
	case 'R':
		v, err := decodeKeyValue(rest)
		return functions.Right[any, any](v), err
	case 'V':
		r := &keyReader{rest}
		tagName, err := r.chunk()
		if err != nil {
			return nil, err
		}
		digits, body, ok := strings.Cut(r.s, ":")
		flag, ferr := strconv.ParseUint(digits, 10, 64)
		if !ok || ferr != nil {
			return nil, fmt.Errorf("malformed key")
		}
		var payload any
		if body != "z" {
			if payload, err = decodeKeyValue(body); err != nil {
				return nil, err
			}
		}
		return dsltypes.NewVariant(tagName, dsltypes.BitFlag(flag), payload), nil
	}
	return nil, fmt.Errorf("malformed key")
}

// tryAdd adds e to the set, reporting (instead of panicking on) an element Go cannot hash.
func tryAdd(s **collections.Set[any], e any) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("a set element of this kind cannot be hashed")
		}
	}()
	*s = (*s).Add(e)
	return nil
}
