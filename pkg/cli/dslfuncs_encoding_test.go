package cli

import (
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
)

func TestEncodingValues(t *testing.T) {
	in := NewInterpreter()
	for _, tc := range []struct {
		src  string
		want string // displayText of the result
	}{
		// RFC 4648 section 10 test vectors
		{`base64_encode ""`, ""},
		{`base64_encode "f"`, "Zg=="},
		{`base64_encode "fo"`, "Zm8="},
		{`base64_encode "foo"`, "Zm9v"},
		{`base64_encode "foob"`, "Zm9vYg=="},
		{`base64_encode "fooba"`, "Zm9vYmE="},
		{`base64_encode "foobar"`, "Zm9vYmFy"},
		{`base64url_encode (bytes_from_list [251, 255])`, "-_8="},
		{`base64_encode (bytes_from_list [251, 255])`, "+/8="},
		{`base64_decode "Zm9vYmFy"`, "Just(0x666f6f626172)"},
		{`base64_decode "Zg=="`, "Just(0x66)"},
		{`base64_decode "Zg"`, "Nothing"}, // padding is required
		{`base64_decode "!!!!"`, "Nothing"},
		{`base64url_decode "-_8="`, "Just(0xfbff)"},
		{`base64url_decode "+/8="`, "Nothing"},
		{`hex_encode "hi"`, "6869"},
		{`hex_encode (bytes_from_list [0, 255, 16])`, "00ff10"},
		{`hex_decode "6869"`, "Just(0x6869)"},
		{`hex_decode "6"`, "Nothing"},
		{`hex_decode "zz"`, "Nothing"},
		{`hex_decode ""`, "Just(0x)"},
		{`url_encode "a b&c=é"`, "a+b%26c%3D%C3%A9"},
		{`url_decode "a+b%26c"`, "Just(a b&c)"},
		{`url_decode "%zz"`, "Nothing"},
		{`url_decode "%ff"`, "Nothing"}, // not valid UTF-8
		{`utf8_decode (utf8_encode "héllo")`, "Just(héllo)"},
		{`utf8_decode (bytes_from_list [255, 254])`, "Nothing"},
		{`utf8_decode (bytes_from_list [])`, "Just()"},
		// published digests of "abc" and ""
		{`sha256 "abc"`, "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"},
		{`sha256 ""`, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{`md5 "abc"`, "900150983cd24fb0d6963f7d28e17f72"},
		{`md5 ""`, "d41d8cd98f00b204e9800998ecf8427e"},
		{`sha256 (utf8_encode "abc")`, "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"},
		{`bytes_length (utf8_encode "héllo")`, "6"},
		{`bytes_slice 1 3 (utf8_encode "hello")`, "0x656c"},
		{`bytes_slice 2 2 (utf8_encode "hello")`, "0x"},
		{`bytes_to_list (utf8_encode "AB")`, "[65 66]"},
		{`utf8_encode "ab" ++ utf8_encode "cd"`, "0x61626364"},
	} {
		if got := displayText(evalOne(t, in, tc.src)); got != tc.want {
			t.Errorf("%s = %q, want %q", tc.src, got, tc.want)
		}
	}
	// round trip: any bytes survive base64/hex
	for _, enc := range [][2]string{{"base64_encode", "base64_decode"}, {"base64url_encode", "base64url_decode"}, {"hex_encode", "hex_decode"}} {
		src := `maybe_default (utf8_encode "") (` + enc[1] + ` (` + enc[0] + ` (bytes_from_list [0, 1, 127, 128, 254, 255])))`
		got, ok := evalOne(t, in, src).(functions.Bytes)
		if !ok || len(got) != 6 || got[0] != 0 || got[5] != 255 {
			t.Errorf("%s/%s round trip = %#v", enc[0], enc[1], got)
		}
	}
}

func TestBytesTypesAreStructural(t *testing.T) {
	in := NewInterpreter()
	if _, ok := evalOne(t, in, `utf8_encode "hi"`).(functions.Bytes); !ok {
		t.Error("utf8_encode must yield functions.Bytes")
	}
	mb, ok := evalOne(t, in, `hex_decode "6869"`).(functions.Maybe[any])
	if v, _ := mb.MaybeValue(); !ok || !functions.IsBytes(v) {
		t.Errorf("hex_decode must yield a Maybe holding Bytes, got %#v", v)
	}
	if got := evalOne(t, in, `hex_encode "x"`); functions.KindName(got) != "String" {
		t.Errorf("hex_encode must yield a String, got %#v", got)
	}
}

// Bytes is a collection row of Ints (like String is of Chars); the closed operations answer
// with Bytes, map and flatmap with a List.
func TestBytesIsACollectionRow(t *testing.T) {
	in := NewInterpreter()
	for _, tc := range []struct {
		src  string
		want string
	}{
		{`length (utf8_encode "héllo")`, "6"},
		{`reverse (utf8_encode "abc")`, "0x636261"},
		{`tail (utf8_encode "abc")`, "0x6263"},
		{`cons 65 (utf8_encode "bc")`, "0x416263"},
		{`filter (\x -> x > 97) (utf8_encode "abc")`, "0x6263"},
		{`reject (\x -> x > 97) (utf8_encode "abc")`, "0x61"},
		{`fold (\a x -> a + x) 0 (utf8_encode "ab")`, "195"},
		{`member 98 (utf8_encode "abc")`, "true"},
		{`member 5 (utf8_encode "abc")`, "false"},
		{`head (utf8_encode "abc")`, "97"},
		{`nth 1 (utf8_encode "abc")`, "98"},
		{`to_list (utf8_encode "ab")`, "[97 98]"},
		{`map (\x -> x + 1) (utf8_encode "AB")`, "[66 67]"},
		{`flatmap (\x -> [x, x]) (utf8_encode "ab")`, "[97 97 98 98]"},
		{`is_empty (bytes_from_list [])`, "true"},
	} {
		if got := displayText(evalOne(t, in, tc.src)); got != tc.want {
			t.Errorf("%s = %q, want %q", tc.src, got, tc.want)
		}
	}
	if _, ok := evalOne(t, in, `reverse (utf8_encode "abc")`).(functions.Bytes); !ok {
		t.Error("reverse on Bytes must stay Bytes")
	}
}

func TestEncodingFailsLoud(t *testing.T) {
	in := NewInterpreter()
	for _, tc := range []struct{ src, want string }{
		{`base64_encode 5`, "want Bytes"},
		{`base64_decode 5`, "want a String"},
		{`hex_encode [1, 2]`, "want Bytes"},
		{`sha256 5`, "want Bytes"},
		{`utf8_decode "text"`, "want Bytes"},
		{`utf8_encode 5`, "want a String"},
		{`bytes_slice 2 9 (utf8_encode "hi")`, "out of range"},
		{`bytes_slice 2 1 (utf8_encode "hi")`, "out of range"},
		{`bytes_from_list [1, 300]`, "not a byte"},
		{`bytes_from_list ["a"]`, "not a byte"},
		{`bytes_from_list 5`, "list of Ints"},
		{`cons "x" (utf8_encode "bc")`, "cannot cons onto Bytes"},
		{`cons 300 (utf8_encode "bc")`, "cannot cons onto Bytes"},
		{`head (bytes_from_list [])`, "empty"},
	} {
		res := in.evaluatePipelineStr(tc.src)
		if len(res) != 1 || !strings.Contains(res[0], "[ERROR]") || !strings.Contains(res[0], tc.want) {
			t.Errorf("%s = %v, want an [ERROR] containing %q", tc.src, res, tc.want)
		}
	}
}

// Every encoder is also a step (F18), applied to the whole value.
func TestEncodingAsSteps(t *testing.T) {
	in := NewInterpreter()
	for _, c := range [][2]string{
		{`"hello" |> base64_encode`, `base64_encode "hello"`},
		{`"hello" |> sha256`, `sha256 "hello"`},
		{`"hi" |> utf8_encode |> hex_encode`, `hex_encode (utf8_encode "hi")`},
		{`"aGk=" |> base64_decode`, `base64_decode "aGk="`},
	} {
		step := strings.Join(in.evaluatePipelineStr(c[0]), ";")
		fn := strings.Join(in.evaluatePipelineStr(c[1]), ";")
		if step == "" || step != fn || strings.Contains(step, "ERROR") {
			t.Errorf("%s = %q but %s = %q", c[0], step, c[1], fn)
		}
	}
}
