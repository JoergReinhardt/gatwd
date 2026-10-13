package collections

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"hash/maphash"
	"iter"
	"math/bits"
	"reflect"
)

var defaultHAMTSeed = maphash.MakeSeed()

// node interface represents either a leaf entry, a collision bucket, or an internal branch node.
type hamtNode[K comparable, V any] interface {
	get(key K, hash uint64, shift uint) (V, bool)
	set(key K, val V, hash uint64, shift uint) (hamtNode[K, V], bool) // returns new node, and whether size increased
	delete(key K, hash uint64, shift uint) (hamtNode[K, V], bool)     // returns new node, and whether an element was removed
	iter(yield func(K, V) bool) bool
}

// leafNode represents a single key-value entry.
type leafNode[K comparable, V any] struct {
	key  K
	val  V
	hash uint64
}

func (l *leafNode[K, V]) get(key K, hash uint64, shift uint) (V, bool) {
	if l.key == key {
		return l.val, true
	}
	var zero V
	return zero, false
}

func (l *leafNode[K, V]) set(key K, val V, hash uint64, shift uint) (hamtNode[K, V], bool) {
	if l.key == key {
		return &leafNode[K, V]{key: key, val: val, hash: hash}, false
	}
	if l.hash == hash {
		// Hash collision at this level
		return &collisionNode[K, V]{
			hash:    hash,
			entries: []entry[K, V]{{key: l.key, val: l.val}, {key: key, val: val}},
		}, true
	}
	// Expand into an internal branch node
	var n hamtNode[K, V] = &branchNode[K, V]{}
	n, _ = n.set(l.key, l.val, l.hash, shift)
	return n.set(key, val, hash, shift)
}

func (l *leafNode[K, V]) delete(key K, hash uint64, shift uint) (hamtNode[K, V], bool) {
	if l.key == key {
		return nil, true
	}
	return l, false
}

func (l *leafNode[K, V]) iter(yield func(K, V) bool) bool {
	return yield(l.key, l.val)
}

type entry[K comparable, V any] struct {
	key K
	val V
}

// collisionNode handles entries with identical 64-bit hashes.
type collisionNode[K comparable, V any] struct {
	hash    uint64
	entries []entry[K, V]
}

func (c *collisionNode[K, V]) get(key K, hash uint64, shift uint) (V, bool) {
	if c.hash != hash {
		var zero V
		return zero, false
	}
	for _, e := range c.entries {
		if e.key == key {
			return e.val, true
		}
	}
	var zero V
	return zero, false
}

func (c *collisionNode[K, V]) set(key K, val V, hash uint64, shift uint) (hamtNode[K, V], bool) {
	if c.hash == hash {
		newEntries := make([]entry[K, V], len(c.entries))
		copy(newEntries, c.entries)
		for i, e := range newEntries {
			if e.key == key {
				newEntries[i] = entry[K, V]{key: key, val: val}
				return &collisionNode[K, V]{hash: hash, entries: newEntries}, false
			}
		}
		newEntries = append(newEntries, entry[K, V]{key: key, val: val})
		return &collisionNode[K, V]{hash: hash, entries: newEntries}, true
	}
	var n hamtNode[K, V] = &branchNode[K, V]{}
	for _, e := range c.entries {
		n, _ = n.set(e.key, e.val, c.hash, shift)
	}
	return n.set(key, val, hash, shift)
}

func (c *collisionNode[K, V]) delete(key K, hash uint64, shift uint) (hamtNode[K, V], bool) {
	for i, e := range c.entries {
		if e.key == key {
			if len(c.entries) == 2 {
				// Degrade back to leafNode
				other := 1 - i
				return &leafNode[K, V]{key: c.entries[other].key, val: c.entries[other].val, hash: c.hash}, true
			}
			newEntries := make([]entry[K, V], 0, len(c.entries)-1)
			newEntries = append(newEntries, c.entries[:i]...)
			newEntries = append(newEntries, c.entries[i+1:]...)
			return &collisionNode[K, V]{hash: c.hash, entries: newEntries}, true
		}
	}
	return c, false
}

func (c *collisionNode[K, V]) iter(yield func(K, V) bool) bool {
	for _, e := range c.entries {
		if !yield(e.key, e.val) {
			return false
		}
	}
	return true
}

// branchNode is a 32-way branching node with bitmap compression.
type branchNode[K comparable, V any] struct {
	bitmap   uint32
	children []hamtNode[K, V]
}

func (b *branchNode[K, V]) get(key K, hash uint64, shift uint) (V, bool) {
	idx := uint32((hash >> shift) & 0x1F)
	mask := uint32(1 << idx)
	if (b.bitmap & mask) == 0 {
		var zero V
		return zero, false
	}
	pos := bits.OnesCount32(b.bitmap & (mask - 1))
	return b.children[pos].get(key, hash, shift+5)
}

func (b *branchNode[K, V]) set(key K, val V, hash uint64, shift uint) (hamtNode[K, V], bool) {
	idx := uint32((hash >> shift) & 0x1F)
	mask := uint32(1 << idx)
	pos := bits.OnesCount32(b.bitmap & (mask - 1))

	if (b.bitmap & mask) == 0 {
		// New slot
		newChildren := make([]hamtNode[K, V], len(b.children)+1)
		copy(newChildren[:pos], b.children[:pos])
		newChildren[pos] = &leafNode[K, V]{key: key, val: val, hash: hash}
		copy(newChildren[pos+1:], b.children[pos:])
		return &branchNode[K, V]{
			bitmap:   b.bitmap | mask,
			children: newChildren,
		}, true
	}

	// Slot exists, delegate recursively
	childRes, added := b.children[pos].set(key, val, hash, shift+5)
	newChildren := make([]hamtNode[K, V], len(b.children))
	copy(newChildren, b.children)
	newChildren[pos] = childRes
	return &branchNode[K, V]{
		bitmap:   b.bitmap,
		children: newChildren,
	}, added
}

func (b *branchNode[K, V]) delete(key K, hash uint64, shift uint) (hamtNode[K, V], bool) {
	idx := uint32((hash >> shift) & 0x1F)
	mask := uint32(1 << idx)
	if (b.bitmap & mask) == 0 {
		return b, false
	}
	pos := bits.OnesCount32(b.bitmap & (mask - 1))
	childRes, removed := b.children[pos].delete(key, hash, shift+5)
	if !removed {
		return b, false
	}

	if childRes == nil {
		// Child was completely removed
		if len(b.children) == 1 {
			return nil, true
		}
		newChildren := make([]hamtNode[K, V], len(b.children)-1)
		copy(newChildren[:pos], b.children[:pos])
		copy(newChildren[pos:], b.children[pos+1:])
		newBranch := &branchNode[K, V]{
			bitmap:   b.bitmap ^ mask,
			children: newChildren,
		}
		// If only 1 child remains and it's a leaf, compress node
		if len(newBranch.children) == 1 {
			if leaf, ok := newBranch.children[0].(*leafNode[K, V]); ok {
				return leaf, true
			}
		}
		return newBranch, true
	}

	newChildren := make([]hamtNode[K, V], len(b.children))
	copy(newChildren, b.children)
	newChildren[pos] = childRes
	return &branchNode[K, V]{
		bitmap:   b.bitmap,
		children: newChildren,
	}, true
}

func (b *branchNode[K, V]) iter(yield func(K, V) bool) bool {
	for _, child := range b.children {
		if !child.iter(yield) {
			return false
		}
	}
	return true
}

// HAMT represents an immutable, persistent Hash Array Mapped Trie.
// It provides O(log32 N) (~O(1)) lookups, insertions, and deletions with structural sharing.
//
// A HAMT may carry one Source (WithSource): provenance for error messages that is not part
// of the map's contents. See Source for what it does and does not affect.
type HAMT[K comparable, V any] struct {
	root hamtNode[K, V]
	size int
	// src is the optional provenance; never consulted by Get, iteration, Equal or JSON.
	src *Source
}

// NewHAMT creates a new empty persistent HAMT map.
func NewHAMT[K comparable, V any]() *HAMT[K, V] {
	return &HAMT[K, V]{
		root: nil,
		size: 0,
	}
}

// Len returns the number of key-value pairs in the HAMT.
func (h *HAMT[K, V]) Len() int {
	if h == nil {
		return 0
	}
	return h.size
}

// IsEmpty returns true if the HAMT has no elements.
func (h *HAMT[K, V]) IsEmpty() bool {
	return h.Len() == 0
}

func hashKey[K comparable](k K) uint64 {
	return maphash.Comparable(defaultHAMTSeed, k)
}

// Get looks up a key in the HAMT.
func (h *HAMT[K, V]) Get(key K) (V, bool) {
	if h == nil || h.root == nil {
		var zero V
		return zero, false
	}
	return h.root.get(key, hashKey(key), 0)
}

// Set returns a new HAMT with the key mapped to the value.
// The existing HAMT is not modified.
func (h *HAMT[K, V]) Set(key K, val V) *HAMT[K, V] {
	if h == nil || h.root == nil {
		res := &HAMT[K, V]{
			root: &leafNode[K, V]{key: key, val: val, hash: hashKey(key)},
			size: 1,
		}
		if h != nil {
			res.src = h.src
		}
		return res
	}
	newRoot, added := h.root.set(key, val, hashKey(key), 0)
	newSize := h.size
	if added {
		newSize++
	}
	return &HAMT[K, V]{
		root: newRoot,
		size: newSize,
		src:  h.src,
	}
}

// Delete returns a new HAMT with the key removed.
// The existing HAMT is not modified.
func (h *HAMT[K, V]) Delete(key K) *HAMT[K, V] {
	if h == nil || h.root == nil {
		return h
	}
	newRoot, removed := h.root.delete(key, hashKey(key), 0)
	if !removed {
		return h
	}
	return &HAMT[K, V]{
		root: newRoot,
		size: h.size - 1,
		src:  h.src,
	}
}

// WithSource returns a HAMT with the same contents that remembers where it came from. The
// receiver is not modified and the structure is shared. Set and Delete keep the Source (the
// result is still "that record, edited"), and so do the operations that are built from them
// starting at the receiver (Merge and Diff keep the receiver's); the ones that build a fresh
// map (Filter, Reject, MapValues, Intersect, ...) do not. The Source is invisible to
// equality, iteration and JSON: see Source.
func (h *HAMT[K, V]) WithSource(src Source) *HAMT[K, V] {
	res := &HAMT[K, V]{src: &src}
	if h != nil {
		res.root, res.size = h.root, h.size
	}
	return res
}

// Source returns the Source attached with WithSource, and whether there is one.
func (h *HAMT[K, V]) Source() (Source, bool) {
	if h == nil || h.src == nil {
		return Source{}, false
	}
	return *h.src, true
}

// Seq returns a Go 1.23 standard iter.Seq2 of (key, value) pairs.
func (h *HAMT[K, V]) Seq() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		if h == nil || h.root == nil {
			return
		}
		h.root.iter(yield)
	}
}

// Keys returns an iter.Seq over all keys.
func (h *HAMT[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		for k := range h.Seq() {
			if !yield(k) {
				return
			}
		}
	}
}

// Values returns an iter.Seq over all values.
func (h *HAMT[K, V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, v := range h.Seq() {
			if !yield(v) {
				return
			}
		}
	}
}

// KeySet returns a persistent Set containing all keys of the HAMT.
func (h *HAMT[K, V]) KeySet() *Set[K] {
	s := NewSet[K]()
	if h == nil || h.root == nil {
		return s
	}
	for k := range h.Keys() {
		s = s.Add(k)
	}
	return s
}

// ValueChunk returns a contiguous bounded Chunk containing all values of the HAMT.
func (h *HAMT[K, V]) ValueChunk() *Chunk[V] {
	if h == nil || h.root == nil {
		return NewChunk[V](0)
	}
	c := NewChunk[V](h.size)
	for v := range h.Values() {
		c.Append(v)
	}
	return c
}

// ValueList returns a persistent List containing all values of the HAMT.
func (h *HAMT[K, V]) ValueList() *List[V] {
	if h == nil || h.root == nil {
		return Nil[V]()
	}
	var items []V
	for v := range h.Values() {
		items = append(items, v)
	}
	res := Nil[V]()
	for i := len(items) - 1; i >= 0; i-- {
		res = Cons(items[i], res)
	}
	return res
}

// Merge returns a new HAMT containing all mappings from h and other.
// If keys collide, values from other take precedence (right-biased merge).
func (h *HAMT[K, V]) Merge(other *HAMT[K, V]) *HAMT[K, V] {
	if h == nil || h.root == nil {
		if other == nil {
			return NewHAMT[K, V]()
		}
		return other
	}
	if other == nil || other.root == nil {
		return h
	}
	res := h
	for k, v := range other.Seq() {
		res = res.Set(k, v)
	}
	return res
}

// Intersect returns a new HAMT containing only the keys present in both h and
// other, taking the values from h. This is the keyed analogue of Set.Intersect
// (H₁ ∩ H₂), completing the set-algebra row for HAMT alongside Merge (union).
func (h *HAMT[K, V]) Intersect(other *HAMT[K, V]) *HAMT[K, V] {
	res := NewHAMT[K, V]()
	if h == nil || h.root == nil || other == nil || other.root == nil {
		return res
	}
	for k, v := range h.Seq() {
		if _, ok := other.Get(k); ok {
			res = res.Set(k, v)
		}
	}
	return res
}

// Diff returns a new HAMT containing the keys of h that are absent from other
// (H₁ \ H₂), preserving h's values. This is the keyed analogue of Set.Diff.
func (h *HAMT[K, V]) Diff(other *HAMT[K, V]) *HAMT[K, V] {
	if h == nil || h.root == nil {
		return NewHAMT[K, V]()
	}
	if other == nil || other.root == nil {
		return h
	}
	res := h
	for k := range other.Seq() {
		res = res.Delete(k)
	}
	return res
}

// Reject returns a new HAMT containing only the entries for which predicate is
// false — the complement of Filter.
func (h *HAMT[K, V]) Reject(predicate func(K, V) bool) *HAMT[K, V] {
	return h.Filter(func(k K, v V) bool { return !predicate(k, v) })
}

// Equal returns true if h and other have identical keys and values.
func (h *HAMT[K, V]) Equal(other *HAMT[K, V]) bool {
	if h == other {
		return true
	}
	hLen := 0
	if h != nil {
		hLen = h.Len()
	}
	oLen := 0
	if other != nil {
		oLen = other.Len()
	}
	if hLen != oLen {
		return false
	}
	if hLen == 0 {
		return true
	}
	for k, v := range h.Seq() {
		otherVal, exists := other.Get(k)
		if !exists {
			return false
		}
		if !reflect.DeepEqual(v, otherVal) {
			return false
		}
	}
	return true
}

// MapValues returns a new HAMT where each value has been transformed by f.
func (h *HAMT[K, V]) MapValues(f func(V) V) *HAMT[K, V] {
	res := NewHAMT[K, V]()
	if h == nil || h.root == nil {
		return res
	}
	for k, v := range h.Seq() {
		res = res.Set(k, f(v))
	}
	return res
}

// Fold reduces all key-value pairs of the HAMT using the accumulator function f.
func (h *HAMT[K, V]) Fold(initial any, f func(acc any, k K, v V) any) any {
	acc := initial
	if h == nil || h.root == nil {
		return acc
	}
	for k, v := range h.Seq() {
		acc = f(acc, k, v)
	}
	return acc
}

// Filter returns a new HAMT containing only the key-value pairs satisfying predicate.
func (h *HAMT[K, V]) Filter(predicate func(K, V) bool) *HAMT[K, V] {
	res := NewHAMT[K, V]()
	if h == nil || h.root == nil {
		return res
	}
	for k, v := range h.Seq() {
		if predicate(k, v) {
			res = res.Set(k, v)
		}
	}
	return res
}

// IsHAMT implements a marker method for type introspection.
func (h *HAMT[K, V]) IsHAMT() bool {
	return true
}

// MarshalJSON implements json.Marshaler.
func (h *HAMT[K, V]) MarshalJSON() ([]byte, error) {
	m := make(map[string]V)
	if h != nil && h.root != nil {
		for k, v := range h.Seq() {
			m[fmt.Sprintf("%v", k)] = v
		}
	}
	return json.Marshal(m)
}

// UnmarshalJSON implements json.Unmarshaler for string-keyed HAMTs.
func (h *HAMT[K, V]) UnmarshalJSON(data []byte) error {
	var m map[string]V
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	newH := NewHAMT[K, V]()
	for k, v := range m {
		// Only works if K is string
		if kAny, ok := any(k).(K); ok {
			newH = newH.Set(kAny, v)
		}
	}
	h.root = newH.root
	h.size = newH.size
	return nil
}

// Value implements driver.Valuer.
func (h *HAMT[K, V]) Value() (driver.Value, error) {
	b, err := h.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// Scan implements sql.Scanner.
func (h *HAMT[K, V]) Scan(src any) error {
	var source []byte
	switch v := src.(type) {
	case string:
		source = []byte(v)
	case []byte:
		source = v
	case nil:
		return h.UnmarshalJSON([]byte("{}"))
	default:
		return fmt.Errorf("incompatible type for HAMT Scan: %T", src)
	}
	return h.UnmarshalJSON(source)
}

// Graph exposes the contents of the HAMT as a standard map for visualizer tree/DOT formatting.
func (h *HAMT[K, V]) Graph() any {
	if h == nil || h.root == nil || h.size == 0 {
		return "<empty HAMT>"
	}
	res := make(map[string]V, h.size)
	for k, v := range h.Seq() {
		res[fmt.Sprintf("%v", k)] = v
	}
	return res
}
