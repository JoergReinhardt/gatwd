package evaluator

import (
	"context"
	"fmt"
	"iter"
	"sync"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
)

// CheckpointStore defines how continuations/watermarks are saved and loaded.
// It can be backed by a SQL database, a file, or Redis.
type CheckpointStore interface {
	Save(ctx context.Context, id string, offset uint64) error
	Load(ctx context.Context, id string) (uint64, error)
}

// MemoryCheckpointStore is a fast, pure functional in-memory implementation backed by an immutable HAMT.
type MemoryCheckpointStore struct {
	mu    sync.RWMutex
	store *collections.HAMT[string, uint64]
}

// NewMemoryCheckpointStore constructs an empty, ready-to-use MemoryCheckpointStore
// backed by a fresh immutable HAMT. It returns a pointer so that the embedded
// mutex and the HAMT pointer are shared across callers of the same store.
func NewMemoryCheckpointStore() *MemoryCheckpointStore {
	return &MemoryCheckpointStore{store: collections.NewHAMT[string, uint64]()}
}

// Save records offset as the latest checkpoint for the continuation identified
// by id, replacing any previous value. Because the underlying HAMT is immutable,
// the write swaps in a new root under a write lock; ctx is accepted to satisfy
// the CheckpointStore interface and is unused by this in-memory implementation.
// It always returns a nil error.
func (m *MemoryCheckpointStore) Save(ctx context.Context, id string, offset uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store = m.store.Set(id, offset)
	return nil
}

// Load returns the most recently saved offset for the continuation identified
// by id. If no checkpoint exists for id it returns 0 and a non-nil error. The
// read is guarded by a read lock; ctx is accepted to satisfy the CheckpointStore
// interface and is unused by this in-memory implementation.
func (m *MemoryCheckpointStore) Load(ctx context.Context, id string) (uint64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.store.Get(id)
	if !ok {
		return 0, fmt.Errorf("checkpoint not found: %s", id)
	}
	return val, nil
}

// StepCheckpoint is a transparent transducer that periodically saves the current stream offset.
// This allows the computation state to be serialized and resumed without degrading performance
// for every single element.
func StepCheckpoint(ctx context.Context, store CheckpointStore, id string, interval uint64) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {

			// Try to resume from an existing checkpoint
			resumeOffset, err := store.Load(ctx, id)
			var currentOffset uint64 = 0
			if err == nil {
				currentOffset = resumeOffset
			}

			var processed uint64 = 0

			for node := range stream {
				// Fast-forward phase: if we are resuming, skip until we reach the offset
				if currentOffset > 0 {
					currentOffset--
					continue
				}

				// Normal execution
				if !yield(node) {
					// If the stream is terminated early, try to save exactly where we stopped
					store.Save(ctx, id, processed+resumeOffset)
					return
				}

				processed++

				// Periodically checkpoint in the background or asynchronously to avoid degrading performance
				// In a real high-throughput system, this would drop onto a non-blocking channel.
				// For structural purity here, we save synchronously every N elements.
				if processed%interval == 0 {
					// Use atomic or direct save
					store.Save(ctx, id, processed+resumeOffset)

					// Check for context cancellation (allows pausing the pipeline externally)
					select {
					case <-ctx.Done():
						return
					default:
					}
				}
			}

			// Final checkpoint when the stream naturally depletes
			store.Save(ctx, id, processed+resumeOffset)
		}
	}
}
