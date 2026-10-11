package parser_test

import (
	"testing"
	"github.com/joergreinhardt/gatwd/parser"
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
