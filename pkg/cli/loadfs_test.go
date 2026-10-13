package cli

import (
	"testing"
	"testing/fstest"
)

// TestLoadScriptFSResolvesImportsInFS pins that `import` inside a script loaded
// from a virtual filesystem is resolved within that filesystem (not the OS), so an
// embedded multi-file prelude can import its sibling modules.
func TestLoadScriptFSResolvesImportsInFS(t *testing.T) {
	fsys := fstest.MapFS{
		"lib/main.gatwd": {Data: []byte("import \"dep.gatwd\"\ndef twice = map (\\x -> x * 2)\n")},
		"lib/dep.gatwd":  {Data: []byte("def inc_all = map (\\x -> x + 1)\n")},
	}
	interp := NewInterpreter()
	if err := interp.LoadScriptFS(fsys, "lib/main.gatwd"); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"twice", "inc_all"} {
		if _, ok := interp.Funcs[name]; !ok {
			t.Errorf("%s not defined: imports were not resolved inside the FS", name)
		}
	}
}
