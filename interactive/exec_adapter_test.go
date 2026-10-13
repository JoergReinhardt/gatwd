package interactive

import (
	"reflect"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/dsltypes"
)

func TestSplitCommandLine(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{`echo hi`, []string{"echo", "hi"}},
		{`sh -c "echo hi there"`, []string{"sh", "-c", "echo hi there"}},
		{`sh -c 'echo $((6*7))'`, []string{"sh", "-c", "echo $((6*7))"}},
		{`printf "a \"b\" c"`, []string{"printf", `a "b" c`}},
		{`echo ""`, []string{"echo", ""}},
		{`  spaced   out  `, []string{"spaced", "out"}},
	}
	for _, c := range cases {
		got, err := SplitCommandLine(c.in)
		if err != nil || !reflect.DeepEqual(got, c.want) {
			t.Errorf("SplitCommandLine(%q) = %q, %v; want %q", c.in, got, err, c.want)
		}
	}
	if _, err := SplitCommandLine(`echo "open`); err == nil {
		t.Error("an unterminated quote must be an error")
	}
}

func run(t *testing.T, cmd string) []any {
	t.Helper()
	var out []any
	for n := range GenericExecAdapter(cmd, dsltypes.BitFlag(0))(dsltypes.NewASTNode[any]("x", dsltypes.BitFlag(0))) {
		out = append(out, n.Value)
	}
	return out
}

func TestExecAdapterQuotedArgumentsReachTheProgram(t *testing.T) {
	got := run(t, `sh -c "echo $((6*7))"`)
	if len(got) != 1 || got[0] != "42" {
		t.Fatalf("got %v, want [42]", got)
	}
}

func TestExecAdapterNonZeroExitIsAnError(t *testing.T) {
	got := run(t, `sh -c "echo oops >&2; exit 3"`)
	if len(got) != 1 {
		t.Fatalf("got %v, want one error", got)
	}
	err, ok := got[0].(error)
	if !ok || !strings.Contains(err.Error(), "exit status 3") || !strings.Contains(err.Error(), "oops") {
		t.Fatalf("got %v, want an error naming the exit status and stderr", got[0])
	}
}
