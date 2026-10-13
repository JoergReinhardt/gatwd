package cli

import (
	"strings"
	"testing"
)

// failLoudRows pins Mandate 6 (fail loud): each statement used to print nothing,
// `[<nil> <nil>]`, `[]` or the fold seed. Each must now yield an [ERROR] result whose
// text names the cause.
var failLoudRows = []struct {
	id        string
	statement string
	want      string
}{
	{"F1 div", `1 / 0`, "by zero"},
	{"F1 intdiv", `1 // 0`, "by zero"},
	{"F1 mod", `1 % 0`, "by zero"},
	{"F2 bool+int", `true + 1`, "unsupported binary operator"},
	{"F3 map undefined fn", `[1, 2] |> map (\x -> nosuchfn x)`, "nosuchfn"},
	{"F3 map list+int", `[1, 2] |> map (\x -> x + [1])`, "unsupported binary operator"},
	{"F4 filter undefined fn", `[1, 2] |> filter (\x -> nosuchfn x)`, "nosuchfn"},
	{"F5 fold undefined name", `[1, 2] |> fold (\a x -> a + nosuch) 0`, "nosuch"},
	{"graph map_nodes failing lambda", `spawn S: x |> graph_empty |> graph_add_node N a |> graph_map_nodes (\p -> nosuch p)`, "nosuch"},
	{"graph filter_nodes failing lambda", `spawn S: x |> graph_empty |> graph_add_node N a |> graph_filter_nodes (\p -> nosuch p)`, "nosuch"},
	{"graph filter_nodes non-bool", `spawn S: x |> graph_empty |> graph_add_node N a |> graph_filter_nodes (\p -> 1)`, "want bool"},
	{"graph fold_nodes failing lambda", `spawn S: x |> graph_empty |> graph_add_node N a |> graph_fold_nodes (\acc p -> acc + nosuch) 0`, "nosuch"},
	{"F2 char literal malformed", `'ab'`, "exactly one character"},
	{"char_code of a long string", `char_code "ab"`, "one-character"},
	{"char_from_code invalid", `char_from_code 55296`, "not a valid Unicode code point"},
	{"Char arithmetic", `'a' + 1`, "not defined for Char and Int"},
	{"D2 string plus string", `"a" + "b"`, "use ++"},
	{"D2 string plus int", `"a" + 1`, "not defined for String and Int"},
	{"D2 int plus string", `1 + "a"`, "not defined for Int and String"},
	{"D2x string minus int", `"5" - 1`, "str_to_int"},
	{"D2x string times int", `"5" * 2`, "str_to_int"},
	{"D2x string div string", `"6" / "2"`, "String and String"},
	{"D2x string lt int", `"5" < 10`, "str_to_int"},
	{"D2 numeric string plus int", `"1" + 2`, "use ++"},
	{"F2 power operator", `2 ** 3`, "cannot evaluate"},
	{"F2 bang", `!true`, "cannot evaluate"},
	{"not of a non-Bool", `not 1`, "want a Bool"},
	{"and of a non-Bool", `and true 1`, "second argument"},
	{"F2 unary minus", `- 5`, "cannot evaluate"},
	{"undefined bare name", `no_such_binding`, "undefined identifier: no_such_binding"},
	{"F10 exec non-zero exit", `spawn S: x |> sh -c "echo boom >&2; exit 3"`, "exit status 3"},
	{"F10 exec unterminated quote", `spawn S: x |> sh -c "echo`, "unterminated"},
	{"F8 assert on empty stream", `spawn Node: A |> filter "zzz" |> assert "A"`, "empty stream"},
}

func TestFailLoudSweep(t *testing.T) {
	for _, row := range failLoudRows {
		t.Run(row.id, func(t *testing.T) {
			in := NewInterpreter()
			results := in.evaluatePipelineStr(row.statement)
			if len(results) == 0 {
				t.Fatalf("%q produced no output; want an [ERROR] containing %q", row.statement, row.want)
			}
			joined := strings.Join(results, "\n")
			if !strings.Contains(joined, "[ERROR]") {
				t.Fatalf("%q = %v; want an [ERROR] containing %q", row.statement, results, row.want)
			}
			if !strings.Contains(joined, row.want) {
				t.Fatalf("%q error %q does not name the cause %q", row.statement, joined, row.want)
			}
		})
	}
}

// F6: a top-level `let` binds the real value; the pipeline-head text substitution must not
// replace a non-scalar (here a HAMT) with its `%v` text.
func TestLetBoundValueIsNotTextSubstituted(t *testing.T) {
	in := NewInterpreter()
	in.EvaluateLine("let r = spawn StrVal: Start |> emit user:alice, role:admin |> to_hamt")
	got := in.evaluatePipelineStr(`hamt_get "role" r`)
	if len(got) != 1 || got[0] != "admin" {
		t.Fatalf(`hamt_get "role" r = %v, want [admin]`, got)
	}
}

// F7: `||>` runs the following step across the stream on a worker pool; it used to
// route the finished nodes through the engine's adapter table and drop them all.
func TestParallelPipeYieldsResults(t *testing.T) {
	in := NewInterpreter()
	got := in.evaluatePipelineStr(`[1, 2, 3] ||> map (\x -> x + 1)`)
	if len(got) != 1 || got[0] != "[2 3 4]" {
		t.Fatalf("[1,2,3] ||> map = %v, want [[2 3 4]]", got)
	}

	// Several stream nodes: every one is processed, in any order.
	got = in.evaluatePipelineStr(`spawn S: x |> sh -c "printf 'a\nb\nc\n'" ||> sh -c "echo got-%s"`)
	seen := map[string]bool{}
	for _, r := range got {
		seen[r] = true
	}
	if len(got) != 3 || !seen["got-a"] || !seen["got-b"] || !seen["got-c"] {
		t.Fatalf("parallel exec = %v, want got-a, got-b, got-c in any order", got)
	}
}

// F10: quoted arguments reach the program intact, quotes removed as a shell would.
func TestExecStepHonoursQuotes(t *testing.T) {
	in := NewInterpreter()
	got := in.evaluatePipelineStr(`spawn S: x |> sh -c "echo $((6*7))"`)
	if len(got) != 1 || got[0] != "42" {
		t.Fatalf("sh -c \"echo $((6*7))\" = %v, want [42]", got)
	}
}

// A bare literal head (`5 |> f`) is an expression like any other; it used to be handed to
// Parser.Parse, which yields nothing, so the whole pipeline printed nothing.
func TestLiteralHeadIsAnExpression(t *testing.T) {
	in := NewInterpreter()
	in.EvaluateLine(`def lh_add = \x y -> x + y`)
	for stmt, want := range map[string]string{
		`5 |> lh_add 1`:         "6",
		`3 |> + 4`:              "7",
		`"hi" |> map (\c -> c)`: "hi",
	} {
		if got := in.evaluatePipelineStr(stmt); len(got) != 1 || got[0] != want {
			t.Errorf("%s = %v, want [%s]", stmt, got, want)
		}
	}
	got := in.EvaluateToValues(`let lh_inc = lh_add 1 in 5 |> lh_inc`)
	if len(got) != 1 || got[0] != 6 {
		t.Errorf("let-bound partial application = %v, want [6]", got)
	}
}
