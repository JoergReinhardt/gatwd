package main

import (
	"fmt"
	"github.com/joergreinhardt/gatwd/optics"
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
	fmt.Println("=== Operator: Optics (Lenses) ===")
	
	originalServer := ServerConfig{
		Name: "prod-db",
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

	fmt.Printf("Original Graph: %+v\n", originalServer)

	// Use Lens.Set to functionally update the deeply nested property
	newServer := ipLens.Set(originalServer, "10.0.0.99")
	
	fmt.Printf("New Immutable Graph: %+v\n", newServer)
	
	if originalServer.Network.IP == newServer.Network.IP {
		panic("Lenses failed to guarantee immutability!")
	}
	fmt.Println("Notice how the Original Graph remains completely unchanged!")
}
