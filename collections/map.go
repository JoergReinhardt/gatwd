package collections

// Map is an immutable-ish generic dictionary.
type Map[K comparable, V any] struct {
	elements map[K]V
}

func NewMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{elements: make(map[K]V)}
}

func (m *Map[K, V]) Set(k K, v V) *Map[K, V] {
	// Create a structural copy to maintain immutability
	newMap := NewMap[K, V]()
	for ek, ev := range m.elements {
		newMap.elements[ek] = ev
	}
	newMap.elements[k] = v
	return newMap
}

func (m *Map[K, V]) Get(k K) (V, bool) {
	v, exists := m.elements[k]
	return v, exists
}

func (m *Map[K, V]) String() string {
	return "Map"
}
