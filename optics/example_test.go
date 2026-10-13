package optics_test

import (
	"fmt"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/optics"
	"github.com/joergreinhardt/gatwd/visualizer"
)

type ServerConfig struct {
	Host string
	Port int
}

type LensOutput struct {
	OriginalPort int
	UpdatedPort  int
	IsImmutable  bool
}

func ExampleLens() {
	opts := visualizer.ParseOptions("--curved --border")

	// Lens focusing on the Port field of ServerConfig
	portLens := optics.Lens[ServerConfig, int]{
		Get: func(c ServerConfig) int { return c.Port },
		Set: func(c ServerConfig, p int) ServerConfig {
			return ServerConfig{Host: c.Host, Port: p}
		},
	}

	orig := ServerConfig{Host: "localhost", Port: 8080}
	updated := portLens.Set(orig, 9090)

	out := LensOutput{
		OriginalPort: portLens.Get(orig),
		UpdatedPort:  portLens.Get(updated),
		IsImmutable:  orig.Port == 8080,
	}

	fmt.Print(visualizer.Format(out, opts))

	// Output:
	// ╭───────────────────────────╮
	// │ ╰─ LensOutput             │
	// │     ├─ OriginalPort: 8080 │
	// │     ├─ UpdatedPort: 9090  │
	// │     ╰─ IsImmutable: true  │
	// ╰───────────────────────────╯
}

type RecordLensOutput struct {
	OriginalUser any
	UpdatedUser  any
}

func ExampleRecordLens() {
	opts := visualizer.ParseOptions("--curved --border")

	userLens := optics.RecordLens("user")
	record := map[string]any{"user": "alice", "role": "admin"}
	updated := userLens.Set(record, "bob")

	out := RecordLensOutput{
		OriginalUser: userLens.Get(record),
		UpdatedUser:  userLens.Get(updated),
	}

	fmt.Print(visualizer.Format(out, opts))

	// Output:
	// ╭────────────────────────────╮
	// │ ╰─ RecordLensOutput        │
	// │     ├─ OriginalUser: alice │
	// │     ╰─ UpdatedUser: bob    │
	// ╰────────────────────────────╯
}

// ExampleHAMTPathLens shows deep, composable field access over nested persistent
// HAMTs — the functional equivalent of a chain of struct field accessors
// (s.Spec.Selector.App) for a dynamically-typed, open-schema payload. The Get is
// an Optional "preview" (Maybe); the Set rebuilds only the spine along the path,
// so the original graph/resource is untouched (structural sharing).
func ExampleHAMTPathLens() {
	selector := collections.NewHAMT[string, any]().Set("app", "nginx")
	spec := collections.NewHAMT[string, any]().Set("selector", selector)
	root := collections.NewHAMT[string, any]().Set("spec", spec)

	pathL := optics.HAMTPathLens("spec", "selector", "app")

	got, _ := pathL.Get(root).Unwrap()
	fmt.Println("get:", got)

	// Focused rewrite of the deep leaf; only spec→selector→app is copied.
	updated := pathL.Set(root, functions.Just[any]("redis"))
	newVal, _ := pathL.Get(updated).Unwrap()
	oldVal, _ := pathL.Get(root).Unwrap()
	fmt.Println("updated:", newVal)
	fmt.Println("original unchanged:", oldVal)

	// Output:
	// get: nginx
	// updated: redis
	// original unchanged: nginx
}
