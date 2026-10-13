// Package archtest enforces architectural boundaries with ordinary Go tests.
package archtest

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const modulePath = "github.com/joergreinhardt/gatwd/"

// libraryRank orders the packages that are meant to be reusable as a
// standalone functional Go library (docs/library_extraction_plan.md, L1).
// A package may import only packages of strictly lower rank.
var libraryRank = map[string]int{
	"functions":    0,
	"collections":  1,
	"arrows":       2,
	"comonads":     2,
	"optics":       2,
	"transducers":  2,
	"transformers": 2,
}

// knownViolations lists boundary breaches that exist today. Each entry is
// "<dir>/<file> -> <import>". External test packages (package x_test) may
// import any library package. The list may only shrink: the test fails both
// on new violations and on entries that no longer occur.
var knownViolations = map[string]string{
	"optics/prism.go -> dsltypes":                     "L3: move DSL type flags behind a facade",
	"transformers/statet.go -> dsltypes":              "L3: move DSL type flags behind a facade",
	"functions/example_test.go -> visualizer":         "L1: library tests must not use the visualizer",
	"collections/example_test.go -> visualizer":       "L1: library tests must not use the visualizer",
	"collections/visual_pretty_test.go -> visualizer": "L1: library tests must not use the visualizer",
	"optics/example_test.go -> visualizer":            "L1: library tests must not use the visualizer",
	"optics/optics_test.go -> dsltypes":               "L3: move DSL type flags behind a facade",
	"transformers/statet_test.go -> dsltypes":         "L3: move DSL type flags behind a facade",
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("module root not found at %s: %v", root, err)
	}
	return root
}

func TestLibraryLayersBoundaries(t *testing.T) {
	root := moduleRoot(t)
	seen := map[string]bool{}
	var unexpected []string

	for pkg, rank := range libraryRank {
		files, err := filepath.Glob(filepath.Join(root, pkg, "*.go"))
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range files {
			f, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
			if err != nil {
				t.Fatalf("parse %s: %v", path, err)
			}
			for _, imp := range f.Imports {
				p, _ := strconv.Unquote(imp.Path.Value)
				if !strings.HasPrefix(p, modulePath) {
					continue
				}
				dep := strings.TrimPrefix(p, modulePath)
				depRank, isLib := libraryRank[dep]
				if isLib && (dep == pkg || depRank < rank || strings.HasSuffix(f.Name.Name, "_test")) {
					continue
				}
				key := pkg + "/" + filepath.Base(path) + " -> " + dep
				if _, ok := knownViolations[key]; ok {
					seen[key] = true
					continue
				}
				unexpected = append(unexpected, key)
			}
		}
	}

	var stale []string
	for key := range knownViolations {
		if !seen[key] {
			stale = append(stale, key)
		}
	}
	sort.Strings(unexpected)
	sort.Strings(stale)
	for _, k := range unexpected {
		t.Errorf("new boundary violation: %s (library layers may import only lower library layers)", k)
	}
	for _, k := range stale {
		t.Errorf("stale allowlist entry, remove it from knownViolations: %s", k)
	}
}
