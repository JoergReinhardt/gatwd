package collections

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestListSerialization(t *testing.T) {
	list := Cons(1, Cons(2, Cons(3, Nil[int]())))

	// Test JSON Text Marshal
	b, err := json.Marshal(list)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	if string(b) != "[1,2,3]" {
		t.Fatalf("Expected [1,2,3], got %s", string(b))
	}

	// Test JSON Text Unmarshal
	var unmarshaledList List[int]
	err = json.Unmarshal(b, &unmarshaledList)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	h1, _ := unmarshaledList.Head()
	if h1 != 1 {
		t.Fatalf("Expected head to be 1, got %v", h1)
	}

	// Test Value (driver.Valuer)
	val, err := list.Value()
	if err != nil {
		t.Fatalf("Value failed: %v", err)
	}

	if !reflect.DeepEqual(val, "[1,2,3]") {
		t.Fatalf("Expected string [1,2,3], got %v", val)
	}

	// Test Scan (sql.Scanner)
	var scannedList List[int]
	err = scannedList.Scan("[4,5,6]")
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	h4, _ := scannedList.Head()
	if h4 != 4 {
		t.Fatalf("Expected scanned head to be 4, got %v", h4)
	}
}

func TestList_FP(t *testing.T) {
	l1 := Cons(1, Cons(2, Nil[int]()))
	l2 := Cons(3, Cons(4, Nil[int]()))

	// Concat
	cat := l1.Concat(l2)
	if cat.Len() != 4 || cat.String() != "[1 2 3 4]" {
		t.Fatalf("expected cat [1 2 3 4], got %v", cat.String())
	}

	// Reverse
	rev := l1.Reverse()
	if rev.String() != "[2 1]" {
		t.Errorf("expected rev [2 1], got %v", rev.String())
	}

	// Equal
	l1Clone := Cons(1, Cons(2, Nil[int]()))
	if !l1.Equal(l1Clone) {
		t.Errorf("expected l1 and l1Clone to be Equal")
	}

	// Map
	mapped := l1.Map(func(x int) int { return x * 10 })
	if mapped.String() != "[10 20]" {
		t.Errorf("expected mapped [10 20], got %v", mapped.String())
	}

	// FoldLeft & FoldRight
	sum := l1.FoldLeft(0, func(acc any, x int) any { return acc.(int) + x })
	if sum != 3 {
		t.Errorf("expected sum 3, got %v", sum)
	}

	// Filter
	filtered := l1.Filter(func(x int) bool { return x > 1 })
	if filtered.String() != "[2]" {
		t.Errorf("expected filtered [2], got %v", filtered.String())
	}
}

func TestTree_FP(t *testing.T) {
	child1 := Leaf(10)
	child2 := Leaf(20)
	children := Cons(child1, Cons(child2, Nil[*Tree[int]]()))
	root := NewTree(5, children)

	// Equal
	rootClone := NewTree(5, Cons(Leaf(10), Cons(Leaf(20), Nil[*Tree[int]]())))
	if !root.Equal(rootClone) {
		t.Errorf("expected root and rootClone to be Equal")
	}

	// Map
	mapped := root.Map(func(x int) int { return x * 2 })
	if mapped.Value() != 10 {
		t.Errorf("expected root mapped value 10, got %d", mapped.Value())
	}
	firstChild, _ := mapped.Children().Head()
	if firstChild.Value() != 20 {
		t.Errorf("expected first child mapped value 20, got %d", firstChild.Value())
	}

	// Fold (catamorphism)
	sumTree := root.Fold(func(val int, childResults []any) any {
		tot := val
		for _, cr := range childResults {
			tot += cr.(int)
		}
		return tot
	})
	if sumTree != 35 {
		t.Errorf("expected tree fold sum 35, got %v", sumTree)
	}

	// AddChild, Size, Depth
	child3 := Leaf(30)
	treeWithChild := root.AddChild(child3)
	if treeWithChild.Size() != 4 {
		t.Errorf("expected tree size 4, got %d", treeWithChild.Size())
	}
	if treeWithChild.Depth() != 2 {
		t.Errorf("expected tree depth 2, got %d", treeWithChild.Depth())
	}

	// Flatten and FlattenChunk
	flatList := treeWithChild.Flatten()
	if flatList.Len() != 4 {
		t.Errorf("expected flatList len 4, got %d", flatList.Len())
	}
	flatChunk := treeWithChild.FlattenChunk()
	if flatChunk.Len() != 4 || flatChunk.Items()[0] != 5 {
		t.Errorf("expected flatChunk len 4 with root 5, got %v", flatChunk.Items())
	}

	// JSON Serialization
	data, err := root.MarshalJSON()
	if err != nil {
		t.Fatalf("tree MarshalJSON failed: %v", err)
	}
	newRoot := Leaf(0)
	if err := newRoot.UnmarshalJSON(data); err != nil {
		t.Fatalf("tree UnmarshalJSON failed: %v", err)
	}
	if !root.Equal(newRoot) {
		t.Errorf("expected deserialized tree to equal original root")
	}
}
