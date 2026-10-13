package cli

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"unicode/utf8"

	"github.com/joergreinhardt/gatwd/functions"
)

// ---- Bytes, encodings and hashes (hardening W2, tier 2) -----------------------------
//
// Bytes is functions.Bytes (an immutable []byte; functions never modify their argument; it
// prints as 0x-hex). A plain []byte from an I/O adapter is accepted wherever Bytes is. It is a
// sequence of Ints 0..255 for length/nth/map/fold, `++` concatenates two Bytes, and `==`
// compares them. A String is NOT silently a Bytes — but wherever a function wants Bytes
// (the encoders and the hashes) a String is accepted as its UTF-8 bytes (see asBytes). The
// other direction is explicit and partial: `utf8_decode` is a Maybe.
//
// Decoders return a Maybe: malformed text is Nothing, never an error. Encoders return a
// String.
//
//	utf8_encode s · utf8_decode b (Maybe String) · bytes_length b · bytes_slice from to b
//	bytes_from_list xs · bytes_to_list b
//	base64_encode/base64_decode · base64url_encode/base64url_decode · hex_encode/hex_decode
//	url_encode s · url_decode s (Maybe String)
//	sha256 b · md5 b            (lowercase hex String; md5 is a fingerprint, not a security measure)

// asBytes reads v as Bytes: a Bytes itself, or a String as its UTF-8 bytes. The result is a
// copy, so callers may not alias the interpreter's values.
func asBytes(v any) (functions.Bytes, error) {
	if b, ok := functions.AsBytes(v); ok {
		return b, nil
	}
	if s, ok := v.(string); ok {
		return functions.Bytes(s), nil
	}
	return nil, fmt.Errorf("want Bytes (or a String), got %s", functions.KindName(v))
}

func registerEncodingFuncs(m map[string]any) {
	defn(m, "utf8_encode", "s", func(a []any) (any, error) {
		s, err := asString(a[0])
		if err != nil {
			return nil, err
		}
		return functions.Bytes(s), nil
	})
	defn(m, "utf8_decode", "b", func(a []any) (any, error) {
		b, ok := functions.AsBytes(a[0])
		if !ok {
			return nil, fmt.Errorf("want Bytes, got %s", functions.KindName(a[0]))
		}
		if !utf8.Valid(b) {
			return functions.Nothing[any](), nil
		}
		return functions.Just[any](string(b)), nil
	})
	defn(m, "bytes_length", "b", func(a []any) (any, error) {
		b, err := asBytes(a[0])
		return len(b), err
	})
	defn(m, "bytes_slice", "from to b", func(a []any) (any, error) {
		from, err := strictInt(a[0])
		if err != nil {
			return nil, err
		}
		to, err := strictInt(a[1])
		if err != nil {
			return nil, err
		}
		b, err := asBytes(a[2])
		if err != nil {
			return nil, err
		}
		if from < 0 || to < from || to > len(b) {
			return nil, fmt.Errorf("slice [%d,%d) out of range for %d bytes", from, to, len(b))
		}
		return functions.Bytes(b[from:to]), nil
	})
	defn(m, "bytes_from_list", "xs", func(a []any) (any, error) {
		items, ok := elementsOf(a[0])
		if !ok {
			return nil, fmt.Errorf("want a list of Ints 0..255, got %s", functions.KindName(a[0]))
		}
		out := make(functions.Bytes, len(items))
		for i, it := range items {
			x, err := functions.ByteOf(it)
			if err != nil {
				return nil, fmt.Errorf("element %d: %w", i, err)
			}
			out[i] = x
		}
		return out, nil
	})
	defn(m, "bytes_to_list", "b", func(a []any) (any, error) {
		b, err := asBytes(a[0])
		if err != nil {
			return nil, err
		}
		out := make([]any, len(b))
		for i, x := range b {
			out[i] = int(x)
		}
		return listOf(out), nil
	})

	for name, enc := range map[string]*base64.Encoding{"base64": base64.StdEncoding, "base64url": base64.URLEncoding} {
		defn(m, name+"_encode", "b", func(a []any) (any, error) {
			b, err := asBytes(a[0])
			if err != nil {
				return nil, err
			}
			return enc.EncodeToString(b), nil
		})
		defn(m, name+"_decode", "s", func(a []any) (any, error) {
			s, err := asString(a[0])
			if err != nil {
				return nil, err
			}
			b, derr := enc.DecodeString(s)
			if derr != nil {
				return functions.Nothing[any](), nil
			}
			return functions.Just[any](functions.Bytes(b)), nil
		})
	}
	defn(m, "hex_encode", "b", func(a []any) (any, error) {
		b, err := asBytes(a[0])
		if err != nil {
			return nil, err
		}
		return hex.EncodeToString(b), nil
	})
	defn(m, "hex_decode", "s", func(a []any) (any, error) {
		s, err := asString(a[0])
		if err != nil {
			return nil, err
		}
		b, derr := hex.DecodeString(s)
		if derr != nil {
			return functions.Nothing[any](), nil
		}
		return functions.Just[any](functions.Bytes(b)), nil
	})
	defn(m, "url_encode", "s", func(a []any) (any, error) {
		s, err := asString(a[0])
		if err != nil {
			return nil, err
		}
		return url.QueryEscape(s), nil
	})
	defn(m, "url_decode", "s", func(a []any) (any, error) {
		s, err := asString(a[0])
		if err != nil {
			return nil, err
		}
		d, derr := url.QueryUnescape(s)
		if derr != nil || !utf8.ValidString(d) {
			return functions.Nothing[any](), nil
		}
		return functions.Just[any](d), nil
	})

	defn(m, "sha256", "b", func(a []any) (any, error) {
		b, err := asBytes(a[0])
		if err != nil {
			return nil, err
		}
		sum := sha256.Sum256(b)
		return hex.EncodeToString(sum[:]), nil
	})
	defn(m, "md5", "b", func(a []any) (any, error) {
		b, err := asBytes(a[0])
		if err != nil {
			return nil, err
		}
		sum := md5.Sum(b)
		return hex.EncodeToString(sum[:]), nil
	})
}
