package main

import (
	"fmt"
	"net"
	"time"

	"github.com/joergreinhardt/gatwd/pkg/cli"
)

func main() {
	interpreter := cli.NewInterpreter()
	
	interpreter.RegisterType("GoTime", func(s string) any {
		t, _ := time.Parse(time.RFC3339, s)
		return t
	})

	interpreter.RegisterType("GoIP", func(s string) any {
		return net.ParseIP(s)
	})

	fmt.Println("=== Starting Custom gatwd VM ===")
	interpreter.LoadScript("custom.gatwd")
}
