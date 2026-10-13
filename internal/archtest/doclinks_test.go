package archtest

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// markdownLink matches [text](target). Code spans and fenced blocks are stripped
// first, so Go snippets such as `[]T](` and DSL examples cannot produce false hits.
var markdownLink = regexp.MustCompile(`\[[^\]\n]*\]\(([^)\s]+)(?:\s+"[^"]*")?\)`)

var (
	fencedBlock = regexp.MustCompile("(?s)```.*?```")
	codeSpan    = regexp.MustCompile("`[^`\n]*`")
)

// docFiles are the documents contributors and agents are told to follow. A broken
// link in them silently sends a reader (or an agent) to nothing.
func docFiles(t *testing.T, root string) []string {
	t.Helper()
	var files []string
	add := func(pattern string) {
		m, err := filepath.Glob(filepath.Join(root, pattern))
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, m...)
	}
	for _, p := range []string{"*.md", "docs/*.md", ".agents/*.md", ".agents/prompts/*.md", ".agents/skills/*/SKILL.md", "examples/new_features/*.md"} {
		add(p)
	}
	return files
}

// TestDocLinksResolve fails on a relative markdown link whose target does not exist,
// and on a link that hard-codes a machine-specific absolute path (file:///home/…,
// /home/…), which only ever resolves on one person's checkout.
func TestDocLinksResolve(t *testing.T) {
	root := moduleRoot(t)
	for _, path := range docFiles(t, root) {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		text := codeSpan.ReplaceAllString(fencedBlock.ReplaceAllString(string(raw), ""), "")
		rel, _ := filepath.Rel(root, path)
		for _, m := range markdownLink.FindAllStringSubmatch(text, -1) {
			target := m[1]
			switch {
			case strings.HasPrefix(target, "http://"), strings.HasPrefix(target, "https://"), strings.HasPrefix(target, "mailto:"), strings.HasPrefix(target, "#"):
				continue
			case strings.HasPrefix(target, "file://"), strings.HasPrefix(target, "/home/"), strings.HasPrefix(target, "/Users/"):
				t.Errorf("%s: link %q hard-codes a machine-specific absolute path; use a relative link", rel, target)
				continue
			}
			file, _, _ := strings.Cut(target, "#")
			if file == "" {
				continue
			}
			if _, err := os.Stat(filepath.Join(filepath.Dir(path), file)); err != nil {
				t.Errorf("%s: broken relative link %q", rel, target)
			}
		}
	}
}
