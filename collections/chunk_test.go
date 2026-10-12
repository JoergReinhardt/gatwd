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

func TestChunk_FP(t *testing.T) {
	c1 := NewChunk[int](3)
	c1.Append(1)
	c1.Append(2)

	c2 := NewChunk[int](3)
	c2.Append(3)
	c2.Append(4)

	// Concat
	cat := c1.Concat(c2)
	if cat.Len() != 4 || cat.Items()[0] != 1 || cat.Items()[3] != 4 {
		t.Fatalf("expected cat len=4 [1 2 3 4], got %v", cat.Items())
	}

	// Equal
	c1Clone := NewChunk[int](2)
	c1Clone.Append(1)
	c1Clone.Append(2)
	if !c1.Equal(c1Clone) {
		t.Errorf("expected c1 and c1Clone to be Equal")
	}

	// Map
	mapped := c1.Map(func(x int) int { return x * 10 })
	if mapped.Len() != 2 || mapped.Items()[0] != 10 || mapped.Items()[1] != 20 {
		t.Errorf("expected mapped [10 20], got %v", mapped.Items())
	}

	// Fold
	sum := c1.Fold(0, func(acc any, x int) any { return acc.(int) + x })
	if sum != 3 {
		t.Errorf("expected sum 3, got %v", sum)
	}

	// Filter
	filtered := c1.Filter(func(x int) bool { return x > 1 })
	if filtered.Len() != 1 || filtered.Items()[0] != 2 {
		t.Errorf("expected filtered [2], got %v", filtered.Items())
	}
}

func TestChunk_Reverse_Slice_Serialization(t *testing.T) {
	c := NewChunk[int](4)
	c.Append(10)
	c.Append(20)
	c.Append(30)

	// Reverse
	rev := c.Reverse()
	if rev.Len() != 3 || rev.Items()[0] != 30 || rev.Items()[1] != 20 || rev.Items()[2] != 10 {
		t.Errorf("expected reversed [30 20 10], got %v", rev.Items())
	}

	// Slice
	sub := c.Slice(1, 3)
	if sub.Len() != 2 || sub.Items()[0] != 20 || sub.Items()[1] != 30 {
		t.Errorf("expected sliced [20 30], got %v", sub.Items())
	}

	// JSON
	data, err := c.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}
	newC := NewChunk[int](0)
	if err := newC.UnmarshalJSON(data); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if !c.Equal(newC) {
		t.Errorf("expected deserialized chunk to equal original")
	}

	// SQL Valuer and Scanner
	val, err := c.Value()
	if err != nil {
		t.Fatalf("Value() failed: %v", err)
	}
	scanned := NewChunk[int](0)
	if err := scanned.Scan(val); err != nil {
		t.Fatalf("Scan() failed: %v", err)
	}
	if !c.Equal(scanned) {
		t.Errorf("expected scanned chunk to equal original")
	}
}
