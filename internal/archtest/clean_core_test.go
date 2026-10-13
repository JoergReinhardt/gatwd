package archtest

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// corePackages are the packages Mandate 5 (Clean Core Prevention, AGENTS.md) keeps
// strictly generic: mathematical, category-theoretic and domain-free. Domain
// knowledge (Kubernetes, AWS, a database, a game) lives in adapters/, pkg/plugins/,
// pkg/<domain>/, cmd/ or a .gatwd prelude — never here.
var corePackages = []string{
	"functions", "dsltypes", "collections", "arrows", "comonads", "optics",
	"transducers", "transformers", "evaluator", "parser", "pkg/cli",
}

// coreImportAllowlist is the short, explicit list of third-party imports core may use, keyed by
// the package allowed to use them. Add to it only for infrastructure that carries no domain
// knowledge, and say why.
//
//   - pkg/cli may import cobra and pflag: a command-line framework is plumbing, not domain
//     pollution, and it is what kubectl, Helm and kind are built on (decision 2026-10-13, see
//     docs/hardening_plan.md §4). It is deliberately NOT allowed in the lower layers
//     (functions, collections, evaluator, parser, ...), which stay standard-library only.
//   - pkg/cli may import go.yaml.in/yaml/v3: YAML is a data format like JSON, which the
//     standard library covers and YAML it does not (decision 2026-10-13, docs/hardening_plan.md
//     §1, for yaml_parse/yaml_encode and the visualizer). Only that module, only in pkg/cli.
var coreImportAllowlist = map[string][]string{
	"pkg/cli": {"github.com/spf13/cobra", "github.com/spf13/pflag", "go.yaml.in/yaml/v3"},
}

func allowedThirdParty(pkg, importPath string) bool {
	for _, p := range coreImportAllowlist[pkg] {
		if importPath == p || strings.HasPrefix(importPath, p+"/") {
			return true
		}
	}
	return false
}

// domainWords must not appear in any identifier declared or used in core code.
// Comments and documentation prose may still mention a domain as an EXAMPLE; what is
// forbidden is code that knows about one.
var domainWords = []string{"kubernetes", "k8s", "kube", "conway", "lifegrid", "postgres", "redis", "kafka"}

// TestCleanCoreImportsNoThirdParty: core packages import only the standard library
// and this module, plus the explicit coreImportAllowlist (cobra, pflag and the YAML library in pkg/cli). A
// third-party import in core is the first step of domain pollution; put it behind an
// adapter, plugin or shim and register it with cli.RegisterStep / cli.RegisterType instead.
func TestCleanCoreImportsNoThirdParty(t *testing.T) {
	root := moduleRoot(t)
	for _, pkg := range corePackages {
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
				first, _, _ := strings.Cut(p, "/")
				if strings.HasPrefix(p, modulePath) || !strings.Contains(first, ".") {
					continue // this module, or the standard library
				}
				if allowedThirdParty(pkg, p) {
					continue // explicitly allowlisted for this package
				}
				t.Errorf("%s/%s imports third-party %q: core must stay dependency-free (Mandate 5)", pkg, filepath.Base(path), p)
			}
		}
	}
}

// TestCleanCoreNoDomainIdentifiers: no identifier in core PRODUCTION code names a
// domain. Test files are exempt: demonstrating a generic comonad on Conway's rule is
// fine, shipping a StepConway is not.
func TestCleanCoreNoDomainIdentifiers(t *testing.T) {
	root := moduleRoot(t)
	for _, pkg := range corePackages {
		files, err := filepath.Glob(filepath.Join(root, pkg, "*.go"))
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range files {
			if strings.HasSuffix(path, "_test.go") {
				continue // a test may use a domain (e.g. Conway) as an example of a generic mechanism
			}
			f, err := parser.ParseFile(token.NewFileSet(), path, nil, 0) // comments dropped
			if err != nil {
				t.Fatalf("parse %s: %v", path, err)
			}
			ast.Inspect(f, func(n ast.Node) bool {
				id, ok := n.(*ast.Ident)
				if !ok {
					return true
				}
				low := strings.ToLower(id.Name)
				for _, w := range domainWords {
					if strings.Contains(low, w) {
						t.Errorf("%s/%s: identifier %q names the domain %q: core must stay domain-free (Mandate 5)", pkg, filepath.Base(path), id.Name, w)
					}
				}
				return true
			})
		}
	}
}
