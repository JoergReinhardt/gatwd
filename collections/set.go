package collections

// Set represents a naive immutable Set.
// In functional programming, immutability is key. Operations on this Set 
// return a completely new Set rather than mutating the original.
type Set[T comparable] struct {
	elements map[T]struct{}
}

// NewSet initializes an empty Set.
func NewSet[T comparable]() *Set[T] {
	return &Set[T]{
		elements: make(map[T]struct{}),
	}
}

// Add returns a new Set containing the new element, preserving the original.
func (s *Set[T]) Add(elem T) *Set[T] {
	newMap := make(map[T]struct{})
	for k := range s.elements {
		newMap[k] = struct{}{}
	}
	newMap[elem] = struct{}{}
	return &Set[T]{elements: newMap}
}

// Contains checks if the element exists in the Set.
func (s *Set[T]) Contains(elem T) bool {
	_, exists := s.elements[elem]
	return exists
}
