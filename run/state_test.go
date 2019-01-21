package run

import (
	"fmt"
	"testing"
)

func TestState(t *testing.T) {
	state := initState()
	fmt.Println(state())
}
