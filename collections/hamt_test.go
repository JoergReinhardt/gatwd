package collections

import (
	"testing"
)

func TestMap(t *testing.T) {
	m := NewMap[string, int]()
	if m == nil {
		t.Fatal("Expected map to not be nil")
	}
}
