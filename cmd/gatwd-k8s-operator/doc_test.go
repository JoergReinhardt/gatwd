package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// The how-to (docs/k8s_howto.md) shows commands and what they print. This file ties the two
// together: every sample below runs the REAL binaries over the checked-in fixtures, and its
// needles must appear both in the how-to text and in what the binary printed. Edit the
// how-to without the code (or the code without the how-to) and one side stops matching.
//
// A needle is a literal piece, so the how-to may elide the rest with "…"; where the how-to
// spells a path differently from the way the test must run (the copied fixture directory
// "rendered/"), docNeedles and outNeedles differ.

const howtoDoc = "../../docs/k8s_howto.md"

// howtoBinaries builds the edition and the operator once per test.
func howtoBinaries(t *testing.T) (edition, operatorBin string) {
	t.Helper()
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go toolchain not on PATH")
	}
	edition = filepath.Join(t.TempDir(), "gatwd-k8s")
	if out, err := exec.Command(goBin, "build", "-o", edition, "../gatwd-k8s").CombinedOutput(); err != nil {
		t.Fatalf("build gatwd-k8s: %v\n%s", err, out)
	}
	return edition, buildOperator(t)
}

// runIn runs bin in dir with no reachable cluster, a fixed clock-independent environment and
// a time limit, and returns the exit code, stdout and stderr.
func runIn(t *testing.T, dir, bin string, stdin string, args ...string) (int, string, string) {
	t.Helper()
	c := exec.Command(bin, args...)
	c.Dir = dir
	c.Env = append(os.Environ(), "KUBECONFIG="+filepath.Join(t.TempDir(), "absent"), "HOME="+t.TempDir(), "KUBERNETES_SERVICE_HOST=")
	c.Stdin = strings.NewReader(stdin)
	var out, errb bytes.Buffer
	c.Stdout, c.Stderr = &out, &errb
	done := make(chan error, 1)
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	go func() { done <- c.Wait() }()
	var err error
	select {
	case err = <-done:
	case <-time.After(60 * time.Second):
		_ = c.Process.Kill()
		t.Fatalf("%s %v timed out", bin, args)
	}
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		t.Fatal(err)
	}
	return code, out.String(), errb.String()
}

type docSample struct {
	name       string
	which      string // "edition" or "operator"
	dir        string // working directory; "" = the repository root
	args       []string
	code       int
	docNeedles []string // must occur in docs/k8s_howto.md
	outNeedles []string // must occur in stdout+stderr
}

func TestHowToDocSamples(t *testing.T) {
	edition, operatorBin := howtoBinaries(t)
	doc, err := os.ReadFile(howtoDoc)
	if err != nil {
		t.Fatal(err)
	}
	text := squash(string(doc))

	// The duplicate-id sample shows paths relative to a directory called rendered/.
	renderedParent := t.TempDir()
	copyTree(t, "../../pkg/k8sgraph/testdata/duplicates-files", filepath.Join(renderedParent, "rendered"))

	// An override that runs a program, and one that imports a sibling file.
	dir := t.TempDir()
	execRule := filepath.Join(dir, "exec.gatwd")
	mustWrite(t, execRule, "def compliance = printf hello\n")
	mustWrite(t, filepath.Join(dir, "sibling.gatwd"), "def rule_s = \\g -> []\n")
	importer := filepath.Join(dir, "importer.gatwd")
	mustWrite(t, importer, "import \"sibling.gatwd\"\ndef compliance = k8s_graph |> fold (\\_ g -> (rule_s g)) 0\n")

	const cluster = "examples/k8s_howto/cluster"
	const prod = "pkg/k8sgraph/testdata/kustomize/overlays/prod"
	rows := []docSample{
		{name: "the layered override and its report", which: "operator",
			args: []string{"check", "--manifests", cluster, "--rules", "examples/k8s_howto/rules.gatwd"},
			docNeedles: []string{
				"[[IdleDeployment Deployment/default/worker] [UnusedService Service/default/backend-svc] [MissingConfigMap Deployment/default/worker] [DatabaseWithoutCredentials Database/default/audit]]",
				"rules: overridden: compliance",
				"rules: added: infra_graph, rel_hosted_on, rel_uses_database, rule_database_without_credentials, secret_target",
			}},
		{name: "the extended model", which: "edition",
			args: []string{"examples/k8s_howto/infra.gatwd"},
			docNeedles: []string{
				"[Secret/default/orders-db-credentials-UsesDatabase->Database/default/orders]",
				"[Database/default/audit-HostedOn->DBMS/default/pg-prod Database/default/orders-HostedOn->DBMS/default/pg-prod]",
				"[Secret/default/orders-db-credentials Database/default/orders DBMS/default/pg-prod]",
			}},
		{name: "kustomize renders a root", which: "edition",
			args:       []string{"eval", "-e", "spawn StrVal: " + prod + " |> manifests |> k8s_graph |> graph_order"},
			docNeedles: []string{"gatwd-k8s eval -e 'spawn StrVal: " + prod + " |> manifests |> k8s_graph |> graph_order'      # ∷ 4"},
			outNeedles: []string{"∷ 4"}},
		{name: "--no-kustomize reads plain files", which: "edition",
			args:       []string{"--no-kustomize", "eval", "-e", "spawn StrVal: " + prod + " |> manifests |> collect_list |> length"},
			docNeedles: []string{"gatwd-k8s --no-kustomize eval -e 'spawn StrVal: " + prod + " |> manifests |> collect_list |> length'   # ∷ 3"},
			outNeedles: []string{"∷ 3"}},
		{name: "the json report", which: "operator",
			args: []string{"check", "--manifests", cluster, "--format", "json"},
			docNeedles: []string{
				`{"status":"violations","count":3,"violations":[{"rule":"IdleDeployment","resource":"Deployment/default/worker"},`,
				`"errors":[],"diagnostics":[]}`,
			}},
		{name: "exit 3 for violations under --fail-on-violation", which: "operator", code: 3,
			args:       []string{"check", "--manifests", cluster, "--fail-on-violation"},
			docNeedles: []string{`3) echo "policy violations"; exit 1 ;;`},
			outNeedles: []string{"3 violation(s) found"}},
		{name: "--no-kustomize on the operator", which: "operator", code: 1,
			args:       []string{"check", "--manifests", prod, "--no-kustomize"},
			docNeedles: []string{"--manifests DIR --no-kustomize", "key 'kind' not found"},
			outNeedles: []string{"key 'kind' not found"}},
		{name: "--no-kustomize with the live source is a usage error", which: "operator", code: 2,
			args:       []string{"check", "--no-kustomize"},
			docNeedles: []string{"(exit 2) rather than a flag that silently does nothing"},
			outNeedles: []string{"--no-kustomize applies to --manifests"}},
		{name: "a duplicate id names both files", which: "operator", dir: renderedParent, code: 1,
			args: []string{"check", "--manifests", "rendered"},
			docNeedles: []string{
				`duplicate node id "Pod/default/dup": declared as Pod in rendered/a.yaml (`,
				`and again as Pod in rendered/more/b.yaml, document 2 (`,
				"the two declarations are different",
			}},
		{name: "an override that runs a program is denied", which: "operator", code: 1,
			args:       []string{"check", "--manifests", cluster, "--rules", execRule},
			docNeedles: []string{"[ERROR] step 'printf' requires the exec capability (policy denies it)"}},
		{name: "the denial in the json report", which: "operator", code: 1,
			args:       []string{"check", "--manifests", cluster, "--rules", execRule, "--format", "json"},
			docNeedles: []string{`"kind":"capability_denied"`, `"step":"printf","capability":"exec"`},
			outNeedles: []string{`"kind":"capability_denied","message":"step 'printf' requires the exec capability (policy denies it)","step":"printf","capability":"exec"`}},
		{name: "--allow-exec grants it", which: "operator", code: 0,
			args:       []string{"check", "--manifests", cluster, "--rules", execRule, "--allow-exec"},
			docNeedles: []string{"--allow-exec"}, outNeedles: []string{"hello"}},
		{name: "an import needs --allow-fs", which: "operator", code: 1,
			args:       []string{"check", "--manifests", cluster, "--rules", importer},
			docNeedles: []string{`import "…" requires the fs capability`},
			outNeedles: []string{`import "sibling.gatwd" requires the fs capability`}},
		{name: "an import with --allow-fs", which: "operator", code: 0,
			args:       []string{"check", "--manifests", cluster, "--rules", importer, "--allow-fs"},
			docNeedles: []string{"--allow-fs"}, outNeedles: []string{"added: rule_s"}},
		{name: "structured data in resources", which: "edition",
			args:       []string{"examples/k8s_howto/data.gatwd", "--", "examples/k8s_howto/data"},
			docNeedles: []string{"∷ [3 2Gi 3r4 [[WeakPassword Secret/default/weak-secret]] [622493693890 f52fbd32b2b3]]"}},
		{name: "the age rule as a script (the wall clock: only the date-independent part)", which: "edition",
			args:       []string{"examples/k8s_howto/age.gatwd", "--", "examples/k8s_howto/aged"},
			docNeedles: []string{"[StalePod Pod/default/old-pod]", "[UnknownAge Pod/default/bad-stamp-pod] [UnknownAge Pod/default/no-stamp-pod]]"}},
		{name: "the age rule as an operator override", which: "operator",
			args:       []string{"check", "--manifests", "examples/k8s_howto/aged", "--rules", "examples/k8s_howto/age_rules.gatwd"},
			docNeedles: []string{"gatwd-k8s-operator check --manifests examples/k8s_howto/aged --rules examples/k8s_howto/age_rules.gatwd"},
			outNeedles: []string{"[StalePod Pod/default/old-pod]", "rules: added: born, older_than, rule_stale_pod"}},
		{name: "prelude list names the imported files", which: "edition",
			args:       []string{"prelude", "list"},
			docNeedles: []string{"prelude list", "lib/k8s/relations.gatwd"},
			outNeedles: []string{"lib/k8s/prelude.gatwd", "lib/k8s/relations.gatwd"}},
		{name: "a crawl without a cluster names the cause", which: "edition", code: 1,
			args:       []string{"eval", "-e", "spawn StrVal: default |> k8s_pods |> collect_list"},
			docNeedles: []string{"[ERROR] no cluster access ("},
			outNeedles: []string{"[ERROR] no cluster access ("}},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			bin := edition
			if r.which == "operator" {
				bin = operatorBin
			}
			wd := r.dir
			if wd == "" {
				wd = "../.."
			}
			code, stdout, stderr := runIn(t, wd, bin, "", r.args...)
			if code != r.code {
				t.Fatalf("exit %d, want %d\nstdout: %s\nstderr: %s", code, r.code, stdout, stderr)
			}
			out := stdout + stderr
			needles := r.outNeedles
			if len(needles) == 0 {
				needles = r.docNeedles
			}
			for _, want := range needles {
				if !strings.Contains(out, want) {
					t.Errorf("the binary no longer prints %q\nstdout: %s\nstderr: %s", want, stdout, stderr)
				}
			}
			for _, want := range r.docNeedles {
				if !strings.Contains(text, squash(want)) {
					t.Errorf("docs/k8s_howto.md no longer shows %q", want)
				}
			}
		})
	}
}

// TestHowToHasNoStaleClaims: statements that were true before the hardening work and are
// false now must not come back. Each phrase is matched literally (case-insensitive).
func TestHowToHasNoStaleClaims(t *testing.T) {
	doc, err := os.ReadFile(howtoDoc)
	if err != nil {
		t.Fatal(err)
	}
	text := strings.ToLower(string(doc))
	for _, stale := range []string{
		"no base64",                     // base64_encode/decode exist
		"has no base64",                 //
		"ids are `metadata.name`",       // ids are Kind/namespace/name
		"exits 0 always",                // exit codes 0-3
		"replace, not layer",            // overrides layer by default
		"needs an external `kustomize`", // kustomize renders in process
		"requires `kubectl kustomize`",  //
		"has no flag for this yet",      // the operator has --no-kustomize
		"does not attach a source path", // duplicate errors name the file
		"the encodings tier of the hardening plan adds it",
	} {
		if strings.Contains(text, stale) {
			t.Errorf("docs/k8s_howto.md carries a stale claim: %q", stale)
		}
	}
}

// excerptBlock matches a fenced haskell block whose first line names the file it excerpts.
var excerptBlock = regexp.MustCompile("(?s)```haskell\n-- ((?:examples|cmd|pkg)/[^\n ]+\\.gatwd)\n(.*?)```")

// normalizeCode drops a trailing `-- comment` and collapses runs of whitespace.
func normalizeCode(line string) string {
	if i := strings.Index(line, " --"); i >= 0 {
		line = line[:i]
	}
	return strings.Join(strings.Fields(line), " ")
}

// TestHowToExcerptsAreInTheFiles: a how-to code block that says which file it shows must
// quote that file — every non-comment line (up to a "…" elision) is a line of the file.
func TestHowToExcerptsAreInTheFiles(t *testing.T) {
	doc, err := os.ReadFile(howtoDoc)
	if err != nil {
		t.Fatal(err)
	}
	blocks := excerptBlock.FindAllStringSubmatch(string(doc), -1)
	if len(blocks) == 0 {
		t.Fatal("no excerpt blocks found in the how-to: the pattern drifted from the doc")
	}
	for _, b := range blocks {
		file := filepath.Join("../..", b[1])
		src, err := os.ReadFile(file)
		if err != nil {
			t.Errorf("the how-to excerpts %s, which cannot be read: %v", b[1], err)
			continue
		}
		var have []string
		for _, l := range strings.Split(string(src), "\n") {
			have = append(have, normalizeCode(l))
		}
		for _, l := range strings.Split(b[2], "\n") {
			l = normalizeCode(l)
			if l == "" || strings.HasPrefix(l, "--") {
				continue
			}
			if i := strings.Index(l, "…"); i >= 0 {
				l = strings.TrimSpace(l[:i])
			}
			found := false
			for _, h := range have {
				if strings.HasPrefix(h, l) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("the how-to shows %q as part of %s, but no such line is there", l, b[1])
			}
		}
	}
}

// squash collapses every run of whitespace (a wrapped line of the how-to) to one space.
func squash(s string) string { return strings.Join(strings.Fields(s), " ") }

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// copyTree copies the regular files below src into dst.
func copyTree(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, b, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
}
