package collections

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestHAMT_BasicOperations(t *testing.T) {
	h0 := NewHAMT[string, int]()
	if !h0.IsEmpty() || h0.Len() != 0 {
		t.Fatalf("expected empty HAMT, got len=%d", h0.Len())
	}

	h1 := h0.Set("a", 1)
	if h0.Len() != 0 {
		t.Fatalf("h0 was mutated! expected len=0, got %d", h0.Len())
	}
	if h1.Len() != 1 {
		t.Fatalf("h1 expected len=1, got %d", h1.Len())
	}

	val, ok := h1.Get("a")
	if !ok || val != 1 {
		t.Fatalf("expected key 'a' -> 1, got %v (ok=%v)", val, ok)
	}

	// Update key
	h2 := h1.Set("a", 100)
	if h2.Len() != 1 {
		t.Fatalf("h2 expected len=1 on update, got %d", h2.Len())
	}
	val, ok = h2.Get("a")
	if !ok || val != 100 {
		t.Fatalf("expected key 'a' -> 100, got %v", val)
	}
	val, _ = h1.Get("a")
	if val != 1 {
		t.Fatalf("h1 mutated on update of h2! got %v", val)
	}

	// Delete key
	h3 := h2.Delete("a")
	if h3.Len() != 0 {
		t.Fatalf("h3 expected len=0 after delete, got %d", h3.Len())
	}
	_, ok = h3.Get("a")
	if ok {
		t.Fatalf("key 'a' still present after delete")
	}
}

func TestHAMT_MultiLevelBranching(t *testing.T) {
	const n = 1000
	h := NewHAMT[int, string]()

	for i := 0; i < n; i++ {
		h = h.Set(i, fmt.Sprintf("val-%d", i))
	}

	if h.Len() != n {
		t.Fatalf("expected len=%d, got %d", n, h.Len())
	}

	// Verify all keys exist
	for i := 0; i < n; i++ {
		val, ok := h.Get(i)
		expected := fmt.Sprintf("val-%d", i)
		if !ok || val != expected {
			t.Fatalf("key %d: expected %q, got %q (ok=%v)", i, expected, val, ok)
		}
	}

	// Non-existent key
	_, ok := h.Get(n + 100)
	if ok {
		t.Fatalf("expected key %d to not exist", n+100)
	}

	// Delete half
	hDeleted := h
	for i := 0; i < n/2; i++ {
		hDeleted = hDeleted.Delete(i)
	}
	if hDeleted.Len() != n/2 {
		t.Fatalf("expected len=%d after partial deletion, got %d", n/2, hDeleted.Len())
	}

	for i := 0; i < n/2; i++ {
		_, ok := hDeleted.Get(i)
		if ok {
			t.Fatalf("key %d should have been deleted", i)
		}
	}
	for i := n / 2; i < n; i++ {
		val, ok := hDeleted.Get(i)
		if !ok || val != fmt.Sprintf("val-%d", i) {
			t.Fatalf("key %d should still exist", i)
		}
	}

	// Ensure original h is completely intact
	if h.Len() != n {
		t.Fatalf("original h mutated! expected len=%d, got %d", n, h.Len())
	}
}

func TestHAMT_Iterators(t *testing.T) {
	h := NewHAMT[string, int]()
	keys := []string{"foo", "bar", "baz", "qux"}
	for i, k := range keys {
		h = h.Set(k, (i+1)*10)
	}

	visited := make(map[string]int)
	for k, v := range h.Seq() {
		visited[k] = v
	}

	if len(visited) != 4 {
		t.Fatalf("expected 4 items in iterator, got %d", len(visited))
	}
	for i, k := range keys {
		if visited[k] != (i+1)*10 {
			t.Fatalf("key %q: expected %d, got %d", k, (i+1)*10, visited[k])
		}
	}

	// Test Keys()
	var keyList []string
	for k := range h.Keys() {
		keyList = append(keyList, k)
	}
	if len(keyList) != 4 {
		t.Fatalf("expected 4 keys, got %d", len(keyList))
	}

	// Test Values()
	var valList []int
	for v := range h.Values() {
		valList = append(valList, v)
	}
	if len(valList) != 4 {
		t.Fatalf("expected 4 values, got %d", len(valList))
	}
}

func TestHAMT_JSONAndSQL(t *testing.T) {
	h := NewHAMT[string, int]()
	h = h.Set("apple", 5).Set("banana", 8)

	data, err := json.Marshal(h)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	var h2 HAMT[string, int]
	if err := json.Unmarshal(data, &h2); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	if h2.Len() != 2 {
		t.Fatalf("expected unmarshaled len=2, got %d", h2.Len())
	}
	val, ok := h2.Get("apple")
	if !ok || val != 5 {
		t.Fatalf("expected apple=5, got %v (ok=%v)", val, ok)
	}

	// Test SQL Value and Scan
	sqlVal, err := h.Value()
	if err != nil {
		t.Fatalf("Value() failed: %v", err)
	}

	var h3 HAMT[string, int]
	if err := h3.Scan(sqlVal); err != nil {
		t.Fatalf("Scan() failed: %v", err)
	}
	if h3.Len() != 2 {
		t.Fatalf("expected scanned len=2, got %d", h3.Len())
	}
	val, ok = h3.Get("banana")
	if !ok || val != 8 {
		t.Fatalf("expected banana=8, got %v", val)
	}
}

func TestHAMT_MergeAndFP(t *testing.T) {
	h1 := NewHAMT[string, int]().Set("a", 1).Set("b", 2)
	h2 := NewHAMT[string, int]().Set("b", 20).Set("c", 30)

	// Merge (right-biased)
	merged := h1.Merge(h2)
	if merged.Len() != 3 {
		t.Fatalf("expected merged len=3, got %d", merged.Len())
	}
	vB, _ := merged.Get("b")
	if vB != 20 {
		t.Errorf("expected right-biased value 20 for 'b', got %d", vB)
	}

	// Equal
	h1Clone := NewHAMT[string, int]().Set("b", 2).Set("a", 1)
	if !h1.Equal(h1Clone) {
		t.Errorf("expected h1 and h1Clone to be Equal")
	}
	if h1.Equal(merged) {
		t.Errorf("expected h1 and merged not to be Equal")
	}

	// MapValues
	mapped := h1.MapValues(func(v int) int { return v * 10 })
	vA, _ := mapped.Get("a")
	if vA != 10 {
		t.Errorf("expected mapped 'a'=10, got %d", vA)
	}

	// Fold
	sum := h1.Fold(0, func(acc any, k string, v int) any { return acc.(int) + v })
	if sum != 3 {
		t.Errorf("expected sum 3, got %v", sum)
	}

	// Filter
	filtered := h1.Filter(func(k string, v int) bool { return v > 1 })
	if filtered.Len() != 1 || !filtered.Equal(NewHAMT[string, int]().Set("b", 2)) {
		t.Errorf("expected filtered to have only 'b': 2, got len %d", filtered.Len())
	}

	// KeySet
	keySet := h1.KeySet()
	if keySet.Len() != 2 || !keySet.Contains("a") || !keySet.Contains("b") {
		t.Errorf("expected keySet to have 'a' and 'b', got %v", keySet.Elements())
	}

	// ValueChunk
	valChunk := h1.ValueChunk()
	if valChunk.Len() != 2 {
		t.Errorf("expected valChunk len 2, got %d", valChunk.Len())
	}

	// ValueList
	valList := h1.ValueList()
	if valList.Len() != 2 {
		t.Errorf("expected valList len 2, got %d", valList.Len())
	}
}
