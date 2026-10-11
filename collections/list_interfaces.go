package collections

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// MarshalJSON implements json.Marshaler.
func (l *List[T]) MarshalJSON() ([]byte, error) {
	if l == nil || l.IsEmpty() {
		return []byte("[]"), nil
	}

	var elements []T
	current := l
	for !current.IsEmpty() {
		head, _ := current.Head()
		elements = append(elements, head)
		current, _ = current.Tail()
	}

	return json.Marshal(elements)
}

// UnmarshalJSON implements json.Unmarshaler.
func (l *List[T]) UnmarshalJSON(text []byte) error {
	var elements []T
	if err := json.Unmarshal(text, &elements); err != nil {
		return err
	}

	var newList *List[T] = Nil[T]()
	for i := len(elements) - 1; i >= 0; i-- {
		newList = Cons(elements[i], newList)
	}

	l.head = newList.head
	l.tail = newList.tail
	l.empty = newList.empty

	return nil
}

// Value implements driver.Valuer.
func (l *List[T]) Value() (driver.Value, error) {
	b, err := l.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// Scan implements sql.Scanner.
func (l *List[T]) Scan(src any) error {
	var source []byte
	switch v := src.(type) {
	case string:
		source = []byte(v)
	case []byte:
		source = v
	case nil:
		return l.UnmarshalJSON([]byte("[]"))
	default:
		return fmt.Errorf("incompatible type for List Scan: %T", src)
	}
	return l.UnmarshalJSON(source)
}
