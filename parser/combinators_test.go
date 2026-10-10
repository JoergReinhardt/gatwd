package parser_test

import (
	"testing"
	"github.com/joergreinhardt/gatwd/parser"
)

func TestCombinators(t *testing.T) {
	// Parse a word (sequence of alphabetic chars)
	isAlpha := func(b byte) bool {
		return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
	}
	
	wordParser := parser.Map(
		parser.Many(parser.Satisfy(isAlpha)),
		func(bytes []byte) string { return string(bytes) },
	)

	// Simpler test: Match "spawn " then a word
	spawnParser := parser.Bind(parser.Token("spawn"), func(_ string) parser.Combinator[string] {
		return parser.Bind(parser.Token(""), func(_ string) parser.Combinator[string] {
			return wordParser
		})
	})

	input := "   spawn LinuxHost"
	
	count := 0
	for res := range spawnParser(input) {
		count++
		if res.Fst != "LinuxHost" {
			t.Fatalf("Expected LinuxHost, got %s", res.Fst)
		}
	}

	if count != 1 {
		t.Fatalf("Expected 1 successful parse, got %d", count)
	}
}
