package cli

import (
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
)

// F18: every function form is also a pipeline step, applied data-last to the whole
// incoming value. Step and function form must agree (Mandate 2).
func TestFunctionFormIsAlsoAStep(t *testing.T) {
	in := NewInterpreter()
	for _, c := range [][2]string{
		{`[3, 1, 2] |> reverse`, `reverse [3, 1, 2]`},
		{`[3, 1, 2] |> length`, `length [3, 1, 2]`},
		{`[3, 1, 2] |> member 1`, `member 1 [3, 1, 2]`},
		{`[3, 1, 2] |> nth 1`, `nth 1 [3, 1, 2]`},
		{`[3, 1, 2] |> is_empty`, `is_empty [3, 1, 2]`},
		{`[3, 1, 2] |> cons 9`, `cons 9 [3, 1, 2]`},
		{`[3, 1, 2] |> concat [4]`, `concat [4] [3, 1, 2]`},
		{`3 |> max 5`, `max 5 3`},
		{`'a' |> char_upper`, `char_upper 'a'`},
		{`"é" |> char_code`, `char_code "é"`},
		{`"z" |> char_of_str`, `char_of_str "z"`},
		{`5 |> pair 1`, `pair 1 5`},
		{`(just 3) |> maybe_default 0`, `maybe_default 0 (just 3)`},
		{`[1, 2] |> maybe_default 0`, `maybe_default 0 [1, 2]`}, // the same error either way
		{`'a' |> char_upper |> char_code`, `char_code (char_upper 'a')`},
	} {
		step := strings.Join(in.evaluatePipelineStr(c[0]), " ; ")
		fn := strings.Join(in.evaluatePipelineStr(c[1]), " ; ")
		if step == "" || strings.Contains(step, "unknown step") {
			t.Errorf("%s = %q: not a step", c[0], step)
		}
		if step != fn {
			t.Errorf("%s = %q but %s = %q", c[0], step, c[1], fn)
		}
	}
}

// A builtin is applied to the WHOLE value (data-last), not mapped over its elements.
func TestBuiltinStepAppliesToTheWholeValue(t *testing.T) {
	in := NewInterpreter()
	if got := evalOne(t, in, `[1, 2, 3] |> length`); got != 3 {
		t.Errorf("[1,2,3] |> length = %#v, want 3", got)
	}
	got := in.evaluatePipelineStr(`['a', 'b'] |> char_upper`)
	if len(got) != 1 || !strings.Contains(got[0], "[ERROR]") || !strings.Contains(got[0], "want a Char") {
		t.Errorf(`['a','b'] |> char_upper = %v; the step is not mapped over the elements, so it must be an error`, got)
	}
}

// The fallback sits after the step chain: existing steps keep their meaning, and an
// unknown word is still an unknown step.
func TestBuiltinStepKeepsExistingStepsAndUnknowns(t *testing.T) {
	in := NewInterpreter()
	if got := in.evaluatePipelineStr(`[3, 1, 2] |> head`); len(got) != 1 || got[0] != "[3 1 2]" {
		t.Errorf("the head step changed meaning: %v", got)
	}
	got := in.evaluatePipelineStr(`[1] |> no_such_function_anywhere`)
	if len(got) != 1 || !strings.Contains(got[0], "unknown step or command") {
		t.Errorf("an unknown step must still fail loudly, got %v", got)
	}
}

// No registry builtin may fall through to the OS-exec fallback: each name resolves as a
// step (an existing one or the F18 fallback).
func TestEveryBuiltinNameResolvesAsAStep(t *testing.T) {
	in := NewInterpreter()
	n := 0
	for name, v := range dslBuiltinFuncs {
		if !functions.IsCallable(v) {
			continue
		}
		n++
		for _, line := range in.evaluatePipelineStr(`[1, 2] |> ` + name) {
			if strings.Contains(line, "unknown step or command") {
				t.Errorf("builtin %q is not usable as a step: %s", name, line)
			}
		}
	}
	if n < 100 {
		t.Fatalf("only %d callable builtins surveyed; the registry changed shape", n)
	}
}
