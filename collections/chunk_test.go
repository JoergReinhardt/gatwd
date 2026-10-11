package collections

import (
	"testing"
)

func TestChunk(t *testing.T) {
	chunk := NewChunk[int](3)

	if chunk.Capacity() != 3 {
		t.Errorf("Expected capacity 3, got %d", chunk.Capacity())
	}
	if chunk.Len() != 0 {
		t.Errorf("Expected len 0, got %d", chunk.Len())
	}
	if !chunk.IsEmpty() {
		t.Errorf("Expected empty chunk")
	}
	if chunk.IsFull() {
		t.Errorf("Expected non-full chunk")
	}

	if !chunk.Append(1) {
		t.Errorf("Failed to append 1")
	}
	if !chunk.Append(2) {
		t.Errorf("Failed to append 2")
	}
	if chunk.Len() != 2 {
		t.Errorf("Expected len 2, got %d", chunk.Len())
	}
	if chunk.IsEmpty() {
		t.Errorf("Expected non-empty chunk")
	}
	if chunk.IsFull() {
		t.Errorf("Expected non-full chunk")
	}

	if !chunk.Append(3) {
		t.Errorf("Failed to append 3")
	}
	if chunk.Len() != 3 {
		t.Errorf("Expected len 3, got %d", chunk.Len())
	}
	if !chunk.IsFull() {
		t.Errorf("Expected full chunk")
	}

	if chunk.Append(4) {
		t.Errorf("Appended 4 to full chunk")
	}
	if chunk.Len() != 3 {
		t.Errorf("Expected len 3 after failed append, got %d", chunk.Len())
	}

	items := chunk.Items()
	if len(items) != 3 || items[0] != 1 || items[1] != 2 || items[2] != 3 {
		t.Errorf("Unexpected items: %v", items)
	}

	// Test Seq
	seqItems := []int{}
	for item := range chunk.Seq() {
		seqItems = append(seqItems, item)
	}
	if len(seqItems) != 3 || seqItems[0] != 1 || seqItems[1] != 2 || seqItems[2] != 3 {
		t.Errorf("Unexpected seq items: %v", seqItems)
	}

	chunk.Clear()
	if !chunk.IsEmpty() {
		t.Errorf("Expected empty chunk after clear")
	}
	if chunk.Len() != 0 {
		t.Errorf("Expected len 0 after clear, got %d", chunk.Len())
	}
}
