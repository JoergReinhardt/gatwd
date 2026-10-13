package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
)

// Pins written BEFORE the HAMT key change (hardening, user decision 2026-10-13: a String
// stops equalling the printed form of a collection). They hold on both sides of it.

// conwayDef is the committed pure-DSL Game of Life rule (examples/game_of_life_pure_dsl).
func conwayDef(t *testing.T) string {
	t.Helper()
	src, err := os.ReadFile("../../examples/game_of_life_pure_dsl/game_of_life.gatwd")
	if err != nil {
		t.Fatal(err)
	}
	text := string(src)
	start := strings.Index(text, "def conway_step = match {")
	end := strings.Index(text, "-- Iterate conway_step")
	if start < 0 || end < 0 {
		t.Fatal("conway_step moved in the example script")
	}
	// the definition spans lines; the interpreter's statement joiner reads it as one
	return text[start:end]
}

func printedCells(t *testing.T, v any) []string {
	t.Helper()
	items, ok := functions.ToSliceAny(v)
	if !ok {
		t.Fatalf("want a list of cells, got %#v", v)
	}
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = fmt.Sprint(it)
	}
	sort.Strings(out)
	return out
}

// The blinker flips between a horizontal and a vertical bar; the glider keeps five cells.
func TestGameOfLifeRulePinned(t *testing.T) {
	in, _ := quietInterpreter()
	in.EvaluateScript(conwayDef(t))
	got := printedCells(t, evalOne(t, in, `[[[2, 1], [2, 2], [2, 3]], 8, 8] |> conway_step`))
	if want := []string{"[1 2]", "[2 2]", "[3 2]"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("blinker after one generation = %v, want %v", got, want)
	}
	got = printedCells(t, evalOne(t, in, `[[[2, 1], [3, 2], [1, 3], [2, 3], [3, 3]], 8, 8] |> conway_step`))
	if want := []string{"[1 2]", "[2 3]", "[2 4]", "[3 2]", "[3 3]"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("glider after one generation = %v, want %v", got, want)
	}
}

// The committed example keeps its output: both invariants hold after 20 and 40 generations.
func TestGameOfLifeExampleOutputPinned(t *testing.T) {
	in, buf := quietInterpreter()
	if _, err := in.Run(RunOptions{Script: "../../examples/game_of_life_pure_dsl/game_of_life.gatwd"}); err != nil {
		t.Fatal(err)
	}
	var vals []string
	for _, line := range strings.Split(buf.String(), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			vals = append(vals, line)
		}
	}
	if len(vals) == 0 || strings.Contains(buf.String(), "[ERROR]") {
		t.Fatalf("the example printed %q", buf.String())
	}
	if got := strings.Join(vals, "|"); !strings.HasSuffix(got, "∷ 3|∷ 5") && !strings.HasSuffix(got, "3|5") {
		t.Fatalf("blinker count 3 then glider count 5 expected at the end, got %q", got)
	}
}

// String keys — every record read from JSON/YAML/manifests — stay exactly themselves.
func TestHAMTStringKeysUnchanged(t *testing.T) {
	in := NewInterpreter()
	for src, want := range map[string]any{
		`hamt_get "a" (hamt_set "a" 1 hamt_empty)`:          1,
		`hamt_has "a" (hamt_set "a" 1 hamt_empty)`:          true,
		`hamt_has "b" (hamt_set "a" 1 hamt_empty)`:          false,
		`hamt_keys (hamt_set "a" 5 hamt_empty)`:             nil,
		`hamt_get_or 0 "zz" (hamt_from [["a", 1]])`:         0,
		`length (hamt_from [["a", 1], ["a", 2], ["b", 3]])`: 2,
		`hamt_get "a" (hamt_from [["a", 1], ["a", 2]])`:     2,
		`hamt_get "ab" (hamt_set "ab" 2 hamt_empty)`:        2,
	} {
		got := evalOne(t, in, src)
		if want != nil && !functions.StructuralEqual(got, want) {
			t.Errorf("%s = %#v, want %#v", src, got, want)
		}
	}
	keys := evalOne(t, in, `hamt_keys (hamt_set "a" 5 hamt_empty)`)
	items, _ := functions.ToSliceAny(keys)
	if len(items) != 1 {
		t.Fatalf("hamt_keys = %#v", keys)
	}
	if s, ok := items[0].(string); !ok || s != "a" {
		t.Errorf("a String key reads back as the String, got %#v", items[0])
	}
}
