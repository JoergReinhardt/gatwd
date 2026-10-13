package parser_test

import (
	"github.com/joergreinhardt/gatwd/parser"
	"testing"
)

func TestAdvancedCombinators(t *testing.T) {
	// Test Take
	takeParser := parser.ParseTake()
	count := 0
	for res := range takeParser("take 42") {
		count++
		if res.Fst != 42 {
			t.Fatalf("Expected 42, got %d", res.Fst)
		}
	}
	if count != 1 {
		t.Fatalf("Failed to parse take")
	}

	// Test Zipper
	zipParser := parser.ParseZipperOp()
	for res := range zipParser("  down ") {
		if res.Fst != "down" {
			t.Fatalf("Expected down, got %s", res.Fst)
		}
	}

	// Test FanOut Arrow
	arrowParser := parser.ParseFanOut()
	count = 0
	for res := range arrowParser("(getLogs &&& getMetrics) ") {
		count++
		if res.Fst.Fst != "getLogs" || res.Fst.Snd != "getMetrics" {
			t.Fatalf("Expected getLogs/getMetrics, got %v", res.Fst)
		}
	}
	if count != 1 {
		t.Fatalf("Failed to parse FanOut Kleisli Arrow")
	}

	// Test Split Arrow [ Left *** Right ]
	splitParser := parser.ParseSplit()
	splitCount := 0
	for res := range splitParser("[ left_step *** right_step ]") {
		splitCount++
		if res.Fst.Fst != "left_step" || res.Fst.Snd != "right_step" {
			t.Fatalf("Expected left_step/right_step, got %v", res.Fst)
		}
	}
	if splitCount != 1 {
		t.Fatalf("Failed to parse Split Arrow")
	}

	// Test Match Case "Pattern -> Target"
	matchCaseParser := parser.ParseMatchCase()
	caseCount := 0
	for res := range matchCaseParser("User(active) -> send_email") {
		caseCount++
		if res.Fst.Fst != "User(active)" || res.Fst.Snd != "send_email" {
			t.Fatalf("Expected User(active) -> send_email, got %v", res.Fst)
		}
	}
	if caseCount != 1 {
		t.Fatalf("Failed to parse Match Case")
	}
}

func TestParseMatchCaseSplitsAtTheFirstArrowOutsideLiterals(t *testing.T) {
	for _, row := range []struct{ in, pat, body string }{
		{"-1 -> x", "-1", "x"},
		{"@2026-10-13 -> echo hi", "@2026-10-13", "echo hi"},
		{`"a->b" -> 1`, `"a->b"`, "1"},
		{`'-' -> 1`, `'-'`, "1"},
		{`[m, n] -> let r = f in \x -> x`, "[m, n]", `let r = f in \x -> x`},
	} {
		n := 0
		for res := range parser.ParseMatchCase()(row.in) {
			n++
			if res.Fst.Fst != row.pat || res.Fst.Snd != row.body {
				t.Fatalf("%q = (%q, %q), want (%q, %q)", row.in, res.Fst.Fst, res.Fst.Snd, row.pat, row.body)
			}
		}
		if n != 1 {
			t.Fatalf("%q produced %d results, want 1", row.in, n)
		}
	}
	// No arrow outside a literal: no result, so the caller can report the arm.
	for _, in := range []string{"zzz", `"a->b"`, ""} {
		for res := range parser.ParseMatchCase()(in) {
			t.Fatalf("%q parsed as %v, want no result", in, res)
		}
	}
}
