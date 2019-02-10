package run

import (
	"testing"
)

func TestState(t *testing.T) {
	main := newSymbolDeclaration(
		"main",
		newThunkObject(
			Default,
		),
	)
	state := initState(
		main,
	)
}
