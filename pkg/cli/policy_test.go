package cli

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
)

// ---- the Policy type ----

func TestPolicyZeroValueGrantsNothing(t *testing.T) {
	var p Policy
	for _, c := range []Capability{CapExec, CapFS, CapNet, CapEnv, Capability("bogus"), ""} {
		if p.Allows(c) {
			t.Errorf("the zero Policy allows %q", c)
		}
	}
	if DenyAll() != (Policy{}) {
		t.Errorf("DenyAll() = %+v, want the zero Policy", DenyAll())
	}
	if got := p.String(); got != "none" {
		t.Errorf("zero Policy String = %q, want none", got)
	}
}

func TestTrustedPolicyGrantsEveryCapability(t *testing.T) {
	p := TrustedPolicy()
	for _, c := range []Capability{CapExec, CapFS, CapNet, CapEnv} {
		if !p.Allows(c) {
			t.Errorf("TrustedPolicy denies %q", c)
		}
	}
	if p.Allows(Capability("bogus")) {
		t.Error("an unknown capability must never be granted, even by TrustedPolicy")
	}
	if got := p.String(); got != "exec,fs,net,env" {
		t.Errorf("TrustedPolicy String = %q", got)
	}
}

func TestPolicyAllowsIsPerCapability(t *testing.T) {
	cases := []struct {
		p    Policy
		want map[Capability]bool
	}{
		{Policy{AllowExec: true}, map[Capability]bool{CapExec: true}},
		{Policy{AllowFS: true}, map[Capability]bool{CapFS: true}},
		{Policy{AllowNet: true}, map[Capability]bool{CapNet: true}},
		{Policy{AllowEnv: true}, map[Capability]bool{CapEnv: true}},
	}
	for _, tc := range cases {
		for _, c := range []Capability{CapExec, CapFS, CapNet, CapEnv} {
			if got := tc.p.Allows(c); got != tc.want[c] {
				t.Errorf("%v.Allows(%s) = %v, want %v", tc.p, c, got, tc.want[c])
			}
		}
	}
}

func TestPolicyRequireNamesStepAndCapability(t *testing.T) {
	err := DenyAll().Require(CapExec, "printf", "")
	var ce *CapabilityError
	if !errors.As(err, &ce) {
		t.Fatalf("Require returned %T (%v), want *CapabilityError", err, err)
	}
	if ce.Capability != CapExec || ce.Step != "printf" {
		t.Errorf("fields = %+v", ce)
	}
	if want := "step 'printf' requires the exec capability (policy denies it)"; err.Error() != want {
		t.Errorf("message = %q, want %q", err.Error(), want)
	}
	if err := TrustedPolicy().Require(CapExec, "printf", ""); err != nil {
		t.Errorf("a granted capability must return nil, got %v", err)
	}
	imp := DenyAll().Require(CapFS, "import", "rules/extra.gatwd")
	if want := `import "rules/extra.gatwd" requires the fs capability (policy denies it)`; imp.Error() != want {
		t.Errorf("import message = %q, want %q", imp.Error(), want)
	}
}

// spawnPfx starts a pipeline with a source node; spawnX adds one `emit x`.
const spawnPfx = `spawn StrVal: seed |> `
const spawnX = spawnPfx + `emit x |> `

// ---- enforcement at the steps a script can reach ----

// evalStrings evaluates one pipeline and returns every result as text; an error value is
// rendered with the `[ERROR] ` prefix the DSL prints.
func evalStrings(in *Interpreter, line string) []string {
	var out []string
	for _, v := range in.EvaluateToValues(line) {
		if err, ok := v.(error); ok {
			out = append(out, "[ERROR] "+err.Error())
			continue
		}
		out = append(out, fmt.Sprint(v))
	}
	return out
}

func anyContains(lines []string, sub string) bool {
	for _, l := range lines {
		if strings.Contains(l, sub) {
			return true
		}
	}
	return false
}

// loopbackEcho serves one TCP line-echo on 127.0.0.1 (never a real network) and returns its address.
func loopbackEcho(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("no loopback listener: %v", err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				defer conn.Close()
				if line, err := bufio.NewReader(conn).ReadString('\n'); err == nil {
					fmt.Fprintf(conn, "ECHO:%s", line)
				}
			}()
		}
	}()
	return ln.Addr().String()
}

// capabilityRow is one script-reachable capability: a statement that uses it, the capability
// and step the denial must name, and what the same statement yields when it is allowed.
type capabilityRow struct {
	name      string
	statement string
	cap       Capability
	step      string
	allowed   string // substring of the trusted result
}

func capabilityRows(t *testing.T) []capabilityRow {
	dir := t.TempDir()
	data := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(data, []byte("alpha beta\ngamma\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	addr := loopbackEcho(t)
	return []capabilityRow{
		{"exec fallback", spawnX + `echo hello-exec`, CapExec, "echo", "hello-exec"},
		{"exec fallback, a name that is no program", spawnX + `frobnicate-step`, CapExec, "frobnicate-step", "unknown step or command"},
		{"memoize", spawnX + `memoize echo hello-memo`, CapExec, "memoize", "hello-memo"},
		{"zip", spawnX + `zip date`, CapExec, "zip", ":"},
		{"scan_file", spawnX + `scan_file ` + data, CapFS, "scan_file", "alpha beta"},
		{"scan_words", spawnX + `scan_words ` + data, CapFS, "scan_words", "alpha"},
		{"walk_fs", spawnX + `walk_fs .`, CapFS, "walk_fs", "&{."},
		{"write_file", spawnPfx + `emit written-by-policy-test |> write_file ` + filepath.Join(dir, "w.txt"), CapFS, "write_file", ""},
		{"append_file", spawnPfx + `emit appended |> append_file ` + filepath.Join(dir, "a.txt"), CapFS, "append_file", ""},
		{"dial_tcp", spawnPfx + `emit hi |> dial_tcp ` + addr, CapNet, "dial_tcp", "ECHO:hi"},
	}
}

func TestDeniedCapabilityIsAnErrorNamingTheCapability(t *testing.T) {
	for _, row := range capabilityRows(t) {
		t.Run(row.name, func(t *testing.T) {
			in, _ := quietInterpreter()
			in.SetPolicy(DenyAll())
			got := evalStrings(in, row.statement)
			want := fmt.Sprintf("[ERROR] step '%s' requires the %s capability (policy denies it)", row.step, row.cap)
			if len(got) != 1 || got[0] != want {
				t.Fatalf("under DenyAll %q = %q, want exactly [%q]", row.statement, got, want)
			}
		})
	}
}

func TestDeniedWriteFileLeavesNoFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "must-not-exist.txt")
	in, _ := quietInterpreter()
	in.SetPolicy(DenyAll())
	if got := evalStrings(in, spawnPfx+`emit secret |> write_file `+path); !anyContains(got, "requires the fs capability") {
		t.Fatalf("got %q", got)
	}
	if _, err := os.Stat(path); err == nil {
		t.Fatal("write_file ran although the policy denied the fs capability")
	}
}

func TestSameStatementsRunUnderTrustedPolicy(t *testing.T) {
	for _, row := range capabilityRows(t) {
		t.Run(row.name, func(t *testing.T) {
			in, _ := quietInterpreter() // NewInterpreter is the trusted command-line engine
			if in.Policy() != TrustedPolicy() {
				t.Fatalf("NewInterpreter policy = %v, want trusted", in.Policy())
			}
			got := evalStrings(in, row.statement)
			if anyContains(got, "requires the") {
				t.Fatalf("trusted run was denied: %q", got)
			}
			if row.allowed != "" && !anyContains(got, row.allowed) {
				t.Fatalf("trusted %q = %q, want a result containing %q", row.statement, got, row.allowed)
			}
			if row.allowed != "" && row.name != "exec fallback, a name that is no program" && anyContains(got, "[ERROR]") {
				t.Fatalf("trusted %q failed: %q", row.statement, got)
			}
		})
	}
	// the file-writing rows must really have written
	dir := t.TempDir()
	in, _ := quietInterpreter()
	evalStrings(in, spawnPfx+`emit written |> write_file `+filepath.Join(dir, "w.txt"))
	if b, err := os.ReadFile(filepath.Join(dir, "w.txt")); err != nil || !strings.Contains(string(b), "written") {
		t.Fatalf("write_file under TrustedPolicy: %q, %v", b, err)
	}
}

func TestDeniedImportIsReportedAndNotLoaded(t *testing.T) {
	dir := t.TempDir()
	lib := filepath.Join(dir, "lib.gatwd")
	if err := os.WriteFile(lib, []byte("const FROM_IMPORT = 41\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	script := fmt.Sprintf("import %q\n", lib)

	in, buf := quietInterpreter()
	in.SetPolicy(DenyAll())
	in.LoadScriptContent(script, dir)
	want := fmt.Sprintf(`∷ [ERROR] import %q requires the fs capability (policy denies it)`, lib)
	if !strings.Contains(buf.String(), want) {
		t.Fatalf("denied import printed %q, want a line containing %q", buf.String(), want)
	}
	if _, ok := in.Constants["FROM_IMPORT"]; ok {
		t.Fatal("the imported file was loaded although the policy denied the fs capability")
	}

	in2, buf2 := quietInterpreter()
	in2.LoadScriptContent(script, dir)
	if strings.Contains(buf2.String(), "[ERROR]") || in2.Constants["FROM_IMPORT"] == nil {
		t.Fatalf("trusted import failed: out %q, constants %v", buf2.String(), in2.Constants["FROM_IMPORT"])
	}
}

func TestEmbeddedImportsAreNotFileAccess(t *testing.T) {
	// an //go:embed prelude importing its siblings is the author's code, not a script reading disk
	in := NewInterpreter()
	in.SetPolicy(DenyAll())
	buf := &strings.Builder{}
	in.Out = buf
	fsys := fstest.MapFS{
		"main.gatwd": &fstest.MapFile{Data: []byte("import \"sub.gatwd\"\n")},
		"sub.gatwd":  &fstest.MapFile{Data: []byte("const FROM_EMBED = 7\n")},
	}
	if err := in.LoadScriptFS(fsys, "main.gatwd"); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), "[ERROR]") || in.Constants["FROM_EMBED"] == nil {
		t.Fatalf("embedded import under DenyAll: out %q, constant %v", buf.String(), in.Constants["FROM_EMBED"])
	}
}

// ---- the policy reaches every evaluation path ----

func TestDenialReachesForksAndLambdaBodies(t *testing.T) {
	for name, stmt := range map[string]string{
		"parallel fork (||>)":  spawnPfx + `emit a, b ||> echo hello-par`,
		"pipeline lambda body": spawnPfx + `emit a |> map (\x -> x |> echo hello-lambda)`,
	} {
		t.Run(name, func(t *testing.T) {
			denied, _ := quietInterpreter()
			denied.SetPolicy(DenyAll())
			got := evalStrings(denied, stmt)
			if !anyContains(got, "requires the exec capability") || anyContains(got, "hello-") {
				t.Fatalf("under DenyAll %q = %q, want a capability error and no command output", stmt, got)
			}
			trusted, _ := quietInterpreter()
			got = evalStrings(trusted, stmt)
			if !anyContains(got, "hello-") || anyContains(got, "[ERROR]") {
				t.Fatalf("under TrustedPolicy %q = %q, want the command output", stmt, got)
			}
		})
	}
}

func TestDenialReachesDefsCompiledBeforeThePolicyChanged(t *testing.T) {
	in, _ := quietInterpreter()
	in.EvaluateLine(`def greet = echo hello-def`)
	if got := evalStrings(in, spawnX+`greet`); !anyContains(got, "hello-def") {
		t.Fatalf("trusted def call = %q", got) // also warms the compiled cache
	}
	in.SetPolicy(DenyAll())
	got := evalStrings(in, spawnX+`greet`)
	if !anyContains(got, "step 'echo' requires the exec capability") || anyContains(got, "hello-def") {
		t.Fatalf("a cached def ran after the policy was tightened: %q", got)
	}
	in.SetPolicy(TrustedPolicy())
	if got := evalStrings(in, spawnX+`greet`); !anyContains(got, "hello-def") {
		t.Fatalf("restoring the policy must restore the step: %q", got)
	}
}

func TestForkedInterpreterNeverWidensThePolicy(t *testing.T) {
	in := NewInterpreter()
	in.SetPolicy(Policy{AllowFS: true})
	fork := in.fork()
	if fork.Policy() != in.Policy() {
		t.Fatalf("fork policy %v, parent %v", fork.Policy(), in.Policy())
	}
	fork.SetPolicy(DenyAll()) // shared state: tightening the fork tightens the tree
	if in.Policy() != DenyAll() {
		t.Fatal("fork and parent must share one policy")
	}
}

func TestEditionStepsStayAllowedUnderDenyAll(t *testing.T) {
	var ran bool
	in, _ := quietInterpreter()
	in.RegisterStep(MapStep("edition_wired", func(v any) any {
		ran = true
		return fmt.Sprintf("wired:%v", v)
	}))
	in.SetPolicy(DenyAll())
	got := evalStrings(in, spawnPfx+`edition_wired`)
	if !ran || len(got) != 1 || got[0] != "wired:seed" {
		t.Fatalf("an Edition-registered step must run under DenyAll: ran=%v got=%q", ran, got)
	}
}

func TestRequireCapabilityLetsStepsTakePartInThePolicy(t *testing.T) {
	// an Edition step that reads the environment asks for CapEnv, the capability no builtin uses yet
	step := func() StepHandler {
		return func(stepStr string, interp *Interpreter) (evaluator.PipelineStep[functions.Either[error, any]], bool) {
			if stepStr != "getenv_probe" {
				return nil, false
			}
			if err := interp.RequireCapability(CapEnv, "getenv_probe", ""); err != nil {
				return interp.failStep(err), true
			}
			return MapStep("getenv_probe", func(any) any { return "env-read" })("getenv_probe", interp)
		}
	}
	for _, tc := range []struct {
		policy Policy
		want   string
	}{
		{DenyAll(), "[ERROR] step 'getenv_probe' requires the env capability (policy denies it)"},
		{Policy{AllowExec: true, AllowFS: true, AllowNet: true}, "[ERROR] step 'getenv_probe' requires the env capability (policy denies it)"},
		{Policy{AllowEnv: true}, "env-read"},
	} {
		in, _ := quietInterpreter()
		in.RegisterStep(step())
		in.SetPolicy(tc.policy)
		got := evalStrings(in, spawnPfx+`getenv_probe`)
		if len(got) != 1 || got[0] != tc.want {
			t.Errorf("policy %v: got %q, want [%q]", tc.policy, got, tc.want)
		}
	}
}

// ---- nothing in a script can widen the policy ----

func TestNoScriptStatementCanWidenThePolicy(t *testing.T) {
	// There is no DSL name for the policy; every attempt below is an error or a no-op, and the
	// policy in force is unchanged afterwards.
	attempts := []string{
		`const AllowExec = true`,
		`let AllowExec = true in AllowExec`,
		`def policy = TrustedPolicy`,
		`emit 1 |> TrustedPolicy`,
		`emit 1 |> SetPolicy`,
		`emit 1 |> allow_exec`,
		`emit 1 |> policy_allow exec`,
	}
	in, _ := quietInterpreter()
	in.SetPolicy(DenyAll())
	for _, a := range attempts {
		in.EvaluateLine(a)
		if in.Policy() != DenyAll() {
			t.Fatalf("%q widened the policy to %v", a, in.Policy())
		}
	}
	if got := evalStrings(in, spawnX+`echo still-denied`); !anyContains(got, "requires the exec capability") {
		t.Fatalf("after the attempts: %q", got)
	}
}

// ---- guard: every script-reachable side effect in pkg/cli is gated ----

// gatedSites are the Go calls that give a script a side effect. A line that makes one must
// pass through Interpreter.gated within the same statement; the guard fails the build when a
// new step is wired in without the policy.
var gatedSites = regexp.MustCompile(`interactive\.GenericExecAdapter|i\.execAdapter\(|adapters\.(WalkFSAdapter|WriteFile|StepDial|StepListenTCP|StepScanFile|WalkFS|ListenerToSeq|ConnToLines)\b|exec\.Command|net\.Dial|net\.Listen`)

func TestEveryScriptReachableSideEffectIsGated(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	var violations []string
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		lines := strings.Split(string(raw), "\n")
		for n, line := range lines {
			trim := strings.TrimSpace(line)
			if strings.HasPrefix(trim, "//") || !gatedSites.MatchString(line) {
				continue
			}
			if strings.Contains(line, "func (i *Interpreter) execAdapter") ||
				strings.Contains(line, "return evaluator.EitherT(interactive.GenericExecAdapterContext(i.runContext") {
				continue // the definition of the adapter constructor itself
			}
			if strings.Contains(line, `GenericExecAdapter("", flag)`) {
				continue // OnMissingAdapter: an EMPTY command template runs nothing (pinned by TestEmptyExecTemplateRunsNothing)
			}
			// the call and its `i.gated(` wrapper are in one statement: this line or the next two
			window := strings.Join(lines[n:min(n+3, len(lines))], "\n")
			if !strings.Contains(window, "i.gated(") {
				violations = append(violations, fmt.Sprintf("%s:%d: %s", f, n+1, trim))
			}
		}
	}
	if len(violations) > 0 {
		t.Errorf("script-reachable side effects that bypass the execution policy (wrap the step in i.gated):\n  %s", strings.Join(violations, "\n  "))
	}
}

// TestEmptyExecTemplateRunsNothing pins why the OnMissingAdapter fallback needs no gate: the
// adapter it builds has an empty command template, which starts no process.
func TestEmptyExecTemplateRunsNothing(t *testing.T) {
	in, _ := quietInterpreter()
	in.SetPolicy(DenyAll())
	adapter := in.Engine.OnMissingAdapter(1)
	node := dsltypes.NewASTNode(functions.Right[error, any](any("payload")), 1)
	for n := range adapter(node) {
		t.Fatalf("an empty command template produced a node: %v", n)
	}
}
