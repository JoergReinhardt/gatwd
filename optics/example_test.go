package optics_test

import (
	"fmt"

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
