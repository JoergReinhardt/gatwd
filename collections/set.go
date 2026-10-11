package collections

// Set is an immutable-ish generic hash set over comparable types.
type Set[T comparable] struct {
	elements map[T]struct{}
}

func NewSet[T comparable](items ...T) *Set[T] {
	s := &Set[T]{elements: make(map[T]struct{})}
	for _, item := range items {
		s.elements[item] = struct{}{}
	}
	return s
}

func (s *Set[T]) Contains(item T) bool {
	_, exists := s.elements[item]
	return exists
}

func (s *Set[T]) Elements() []T {
	keys := make([]T, 0, len(s.elements))
	for k := range s.elements {
		keys = append(keys, k)
	}
	return keys
}

func (s *Set[T]) String() string {
	return "Set"
}

func (s *Set[T]) Add(item T) *Set[T] {
	newSet := NewSet[T]()
	for k := range s.elements {
		newSet.elements[k] = struct{}{}
	}
	newSet.elements[item] = struct{}{}
	return newSet
}
