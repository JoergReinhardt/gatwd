package main

import (
	"fmt"

	"github.com/joergreinhardt/gatwd/optics"
	"github.com/joergreinhardt/gatwd/visualizer"
)

type NetworkConfig struct {
	IP   string
	Port int
}

type ServerConfig struct {
	Name    string
	Network NetworkConfig
}

func main() {
	header := visualizer.Format("Operator: Optics (Functional Lenses)", visualizer.ParseOptions("--heavy --border --color=cyan"))
	fmt.Print(header)

	originalServer := ServerConfig{
		Name:    "prod-db",
		Network: NetworkConfig{IP: "192.168.1.100", Port: 5432},
	}

	// Define a Lens focusing deeply on the IP field
	ipLens := optics.Lens[ServerConfig, string]{
		Get: func(s ServerConfig) string { return s.Network.IP },
		Set: func(s ServerConfig, ip string) ServerConfig {
			s.Network.IP = ip
			return s // Returns a mutated COPY, ensuring functional immutability
		},
	}

	fmt.Println("\n-- 1. Original Immutably Stored Config --")
	fmt.Print(visualizer.Format(originalServer, visualizer.ParseOptions("--curved --border --color=blue")))

	// Use Lens.Set to functionally update the deeply nested property
	newServer := ipLens.Set(originalServer, "10.0.0.99")

	fmt.Println("\n-- 2. New Immutable Config via Lens.Set --")
	fmt.Print(visualizer.Format(newServer, visualizer.ParseOptions("--curved --border --color=green")))

	if originalServer.Network.IP == newServer.Network.IP {
		panic("Lenses failed to guarantee immutability!")
	}
}
