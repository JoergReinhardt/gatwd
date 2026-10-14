package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Every ```gatwd block in docs/reference.md is a self-contained script that must run clean:
// no [ERROR], no silent statement, no --check diagnostic. Blocks tagged ```text are exempt.
func TestReferenceExamplesRunClean(t *testing.T) {
	doc, err := os.ReadFile("../../docs/reference.md")
	if err != nil {
		t.Fatal(err)
	}
	blocks := regexp.MustCompile("(?s)```gatwd\n(.*?)```").FindAllStringSubmatch(string(doc), -1)
	if len(blocks) < 20 {
		t.Fatalf("found only %d gatwd blocks; the fence pattern is stale", len(blocks))
	}
	dir := t.TempDir()
	for n, b := range blocks {
		file := filepath.Join(dir, "block.gatwd")
		if err := os.WriteFile(file, []byte(b[1]), 0o600); err != nil {
			t.Fatal(err)
		}
		if problems := cheatsheetProblems(t, file); len(problems) != 0 {
			t.Errorf("block %d:\n%s\n--> %s", n+1, b[1], strings.Join(problems, "\n    "))
		}
	}
}
