package collections

import (
	"fmt"
	"testing"
)

// TestBind_Collections covers the monadic flatMap (Bind) added to the collection
// types so they are no longer second-class monads next to Maybe and streams.
func TestBind_Collections(t *testing.T) {
	// List: duplicate each element, preserving order.
	l := Nil[int]()
	for i := 3; i >= 1; i-- {
		l = Cons(i, l) // 1 2 3
	}
	got := l.Bind(func(n int) *List[int] { return Cons(n, Cons(n, Nil[int]())) })
	if s := fmt.Sprint(got); s != "[1 1 2 2 3 3]" {
		t.Errorf("List.Bind = %s, want [1 1 2 2 3 3]", s)
	}

	// Chunk: flatMap appends in order.
	c := NewChunk[int](8)
	for _, n := range []int{1, 2} {
		c.Append(n)
	}
	cb := c.Bind(func(n int) *Chunk[int] {
		r := NewChunk[int](2)
		r.Append(n)
		r.Append(n * 10)
		return r
	})
	if cb.Len() != 4 {
		t.Errorf("Chunk.Bind len = %d, want 4", cb.Len())
	}

	// Set: flatMap unions results (idempotent — overlaps collapse).
	s := NewSet[int]().Add(1).Add(2)
	sb := s.Bind(func(n int) *Set[int] { return NewSet[int]().Add(n).Add(0) })
	if sb.Len() != 3 { // {1,2,0}
		t.Errorf("Set.Bind len = %d, want 3", sb.Len())
	}
}

// TestHAMT_IntersectDiff covers the keyed set-algebra added to HAMT so it reaches
// parity with Set (which already had Union/Intersect/Diff).
func TestHAMT_IntersectDiff(t *testing.T) {
	a := NewHAMT[string, int]().Set("a", 1).Set("b", 2).Set("c", 3)
	b := NewHAMT[string, int]().Set("b", 20).Set("c", 30).Set("d", 40)

	inter := a.Intersect(b)
	if inter.Len() != 2 {
		t.Fatalf("Intersect len = %d, want 2", inter.Len())
	}
	if v, ok := inter.Get("b"); !ok || v != 2 { // value taken from the receiver
		t.Errorf("Intersect should keep b=2 (receiver's value), got %v ok=%v", v, ok)
	}
	if _, ok := inter.Get("a"); ok {
		t.Errorf("Intersect should not contain a (only in receiver)")
	}

	diff := a.Diff(b)
	if diff.Len() != 1 {
		t.Fatalf("Diff len = %d, want 1", diff.Len())
	}
	if _, ok := diff.Get("a"); !ok {
		t.Errorf("Diff should contain a")
	}
	if _, ok := diff.Get("b"); ok {
		t.Errorf("Diff should not contain b")
	}
	// Immutability: originals unchanged.
	if a.Len() != 3 || b.Len() != 3 {
		t.Errorf("Intersect/Diff mutated an operand: a=%d b=%d", a.Len(), b.Len())
	}
}

// TestReject_AllCollections verifies the negated-filter operation is present and
// correct across every collection type (no second-class citizens).
func TestReject_AllCollections(t *testing.T) {
	isEven := func(n int) bool { return n%2 == 0 }

	l := Nil[int]()
	for i := 5; i >= 1; i-- {
		l = Cons(i, l) // 1 2 3 4 5
	}
	if got := l.Reject(isEven); got.Len() != 3 { // keeps odds 1,3,5
		t.Errorf("List.Reject len = %d, want 3", got.Len())
	}

	c := NewChunk[int](8)
	for _, n := range []int{1, 2, 3, 4} {
		c.Append(n)
	}
	if got := c.Reject(isEven); got.Len() != 2 { // keeps 1,3
		t.Errorf("Chunk.Reject len = %d, want 2", got.Len())
	}

	s := NewSet[int]().Add(1).Add(2).Add(3).Add(4)
	if got := s.Reject(isEven); got.Len() != 2 {
		t.Errorf("Set.Reject len = %d, want 2", got.Len())
	}

	h := NewHAMT[string, int]().Set("a", 1).Set("b", 2).Set("c", 3)
	if got := h.Reject(func(_ string, v int) bool { return isEven(v) }); got.Len() != 2 {
		t.Errorf("HAMT.Reject len = %d, want 2", got.Len())
	}
}

// TestTree_FilterReject covers the Filter/Reject row for Tree: a child subtree is
// pruned (with its descendants) when the predicate rejects its root value.
func TestTree_FilterReject(t *testing.T) {
	root := Leaf(0).
		AddChild(Leaf(1)).
		AddChild(Leaf(5).AddChild(Leaf(6))). // whole subtree should be pruned
		AddChild(Leaf(2))

	kept := root.Filter(func(n int) bool { return n < 4 })
	if kept.Size() != 3 { // root(0) + 1 + 2
		t.Fatalf("Tree.Filter size = %d, want 3 (root,1,2)", kept.Size())
	}
	if kept.Children().Len() != 2 {
		t.Errorf("Tree.Filter should keep 2 children, got %d", kept.Children().Len())
	}

	rej := root.Reject(func(n int) bool { return n >= 4 })
	if rej.Size() != 3 {
		t.Errorf("Tree.Reject size = %d, want 3", rej.Size())
	}
	// Root is always retained even if it would fail the predicate.
	if root.Filter(func(int) bool { return false }).Value() != 0 {
		t.Errorf("Tree.Filter must retain the root value")
	}
}
