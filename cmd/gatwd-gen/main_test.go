package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/pkg/cli"
	"github.com/joergreinhardt/gatwd/spec"
)

func root(t *testing.T) string {
	t.Helper()
	return filepath.Join("..", "..")
}

func read(t *testing.T, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root(t), rel))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func fullSpec() spec.GrammarSpec {
	var names []string
	params := map[string][]string{}
	for _, f := range cli.FunctionForms() {
		names = append(names, f.Name)
		params[f.Name] = f.Params
	}
	return spec.DefaultSpec().WithFunctionForms(names, params)
}

// Every function form the interpreter resolves must reach the generated tables.
func TestSpecCoversEveryFunctionForm(t *testing.T) {
	forms := cli.FunctionForms()
	if len(forms) < 100 {
		t.Fatalf("function registry suspiciously small: %d", len(forms))
	}
	have := map[string]bool{}
	for _, b := range fullSpec().Builtins {
		have[b.Name] = true
	}
	for _, f := range forms {
		if !have[f.Name] {
			t.Errorf("function form %q missing from spec", f.Name)
		}
	}
}

// Every pipeline step the interpreter parses (HasPrefix chains and the graph switch)
// must be listed in the spec; a new step without a spec entry fails here.
func TestSpecCoversEveryPipelineStep(t *testing.T) {
	src := read(t, "pkg/cli/cli.go")
	re := regexp.MustCompile(`(?:HasPrefix\(stepStr, |stepStr == |^\tcase )"([a-z][a-z_0-9]*)`)
	have := map[string]bool{}
	for _, b := range fullSpec().Builtins {
		have[b.Name] = true
	}
	// Names that are statements, aliases or operators, not user-facing step builtins.
	skip := map[string]bool{"match": true, "seq": true, "echo": true, "empty": true, "date": true, "uname": true, "ls": true, "whoami": true,
		"custom_multiply": true}
	found := 0
	for _, line := range strings.Split(src, "\n") {
		for _, m := range re.FindAllStringSubmatch(line, -1) {
			found++
			if !have[m[1]] && !skip[m[1]] {
				t.Errorf("pipeline step %q (cli.go) missing from spec", m[1])
			}
		}
	}
	if found < 100 {
		t.Fatalf("scanner found only %d step names; the pattern is stale", found)
	}
}

// Every keyword and operator the spec declares must be documented in the EBNF.
func TestGrammarDocMentionsSpecSymbols(t *testing.T) {
	doc := read(t, "docs/grammar.md")
	s := spec.DefaultSpec()
	for _, k := range s.Keywords {
		if k.Name == "in" || k.Name == "empty" || k.Name == "seq" {
			continue // documented as part of let / prose
		}
		if !strings.Contains(doc, k.Name) {
			t.Errorf("keyword %q not in docs/grammar.md", k.Name)
		}
	}
	for _, o := range s.Operators {
		if !strings.Contains(doc, o.Symbol) {
			t.Errorf("operator %q not in docs/grammar.md", o.Symbol)
		}
	}
}

// Every non-terminal used in the EBNF block must be defined there.
func TestGrammarNonTerminalsDefined(t *testing.T) {
	doc := read(t, "docs/grammar.md")
	start := strings.Index(doc, "```ebnf")
	end := strings.Index(doc[start+7:], "```")
	block := doc[start+7 : start+7+end]
	block = regexp.MustCompile(`(?s)\(\*.*?\*\)`).ReplaceAllString(block, "")
	block = regexp.MustCompile(`"(?:[^"\\]|\\.)*"`).ReplaceAllString(block, "")
	def := map[string]bool{"ANY_CHAR": true}
	for _, m := range regexp.MustCompile(`(?m)^([A-Z][A-Za-z]*)\s*::=`).FindAllStringSubmatch(block, -1) {
		def[m[1]] = true
	}
	if len(def) < 20 {
		t.Fatalf("parsed only %d rules", len(def))
	}
	for _, m := range regexp.MustCompile(`\b[A-Z][a-z][A-Za-z]*\b`).FindAllString(block, -1) {
		if !def[m] {
			t.Errorf("non-terminal %q used but never defined in docs/grammar.md", m)
		}
	}
}
