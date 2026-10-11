package collections

import (
	"math/bits"
)

type Map[K comparable, V any] struct {
	size int
}

func NewMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{}
}
