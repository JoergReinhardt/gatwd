// Package collections provides purely functional, persistent data structures with structural sharing.
//
// In functional programming, collections are immutable. Operations like Cons, Set, Add, or Delete
// do not mutate existing memory; instead, they return a new collection that shares unchanged sub-branches
// with the previous version. This guarantees complete thread safety, eliminates side-effects, and
// allows cheap snapshots and versioning.
//
// Key Data Structures:
//
//   - HAMT[K, V]: A 32-way branching Hash Array Mapped Trie using bitmap index compression
//     (math/bits.OnesCount32) and Go's native hash/maphash.Comparable. Provides O(log32 N) (~O(1))
//     lookups, insertions, and deletions.
//
//   - List[T]: A persistent singly-linked Cons list. Adding an element (Cons) creates a new head
//     pointing to the existing list as its tail in O(1) time and memory.
//
//   - Tree[T]: A multi-way Rose Tree for modeling hierarchical structures and directory graphs.
//
//   - Set[T]: An immutable hash set backed by HAMT[T, struct{}], offering O(log32 N) additions
//     and membership checks with structural sharing.
//
//   - Map[K, V]: An immutable dictionary backed by HAMT[K, V].
//
//   - Chunk[T]: An eager, contiguous slice wrapper providing bounds-checked spatial locality
//     to maximize CPU L1/L2 cache prefetching when processing stream batches.
//
// All primary collections implement Go's standard serialization and persistence interfaces:
// json.Marshaler, json.Unmarshaler, database/sql/driver.Valuer, and database/sql.Scanner.
package collections
