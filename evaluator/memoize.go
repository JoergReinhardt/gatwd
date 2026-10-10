package evaluator

import (
	"fmt"
	"iter"
	"sync"
	"github.com/joergreinhardt/gatwd/dsltypes"
)

// StepMemoize caches the results of an Adapter to prevent redundant computation
// for the same input Node. It uses a thread-safe sync.Map.
func StepMemoize[T any](adapter Adapter[T]) PipelineStep[T] {
	var cache sync.Map
	
	return func(stream iter.Seq[dsltypes.ASTNode[T]]) iter.Seq[dsltypes.ASTNode[T]] {
		return func(yield func(dsltypes.ASTNode[T]) bool) {
			for node := range stream {
				hash := fmt.Sprintf("%v_%d", node.Value, node.DSLType)
				
				if cachedObj, hit := cache.Load(hash); hit {
					cachedResults := cachedObj.([]dsltypes.ASTNode[T])
					for _, res := range cachedResults {
						if !yield(res) { return }
					}
					continue
				}
				
				// Execute and Cache
				var results []dsltypes.ASTNode[T]
				for res := range adapter(node) {
					results = append(results, res)
					if !yield(res) { return } // Yield on the fly
				}
				cache.Store(hash, results)
			}
		}
	}
}
