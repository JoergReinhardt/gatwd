package dsltypes

import "testing"

// TestVariant_FunctorMonad covers the Functor (Map), catamorphism (Fold), and
// Monad (Bind) instances added to Variant so a tagged union is no longer a
// second-class ADT. Map preserves the tag/flag and transforms only the payload;
// Fold eliminates the variant; Bind re-tags based on the payload.
func TestVariant_FunctorMonad(t *testing.T) {
	just := NewVariant("Just", BitFlag(2), 21)

	// Map transforms the payload, keeping the branch identity.
	mapped := just.Map(func(p any) any { return p.(int) * 2 })
	if mapped.Tag != "Just" || mapped.Flag != BitFlag(2) {
		t.Errorf("Map must preserve tag/flag, got %s/%d", mapped.Tag, mapped.Flag)
	}
	if mapped.Value.(int) != 42 {
		t.Errorf("Map payload = %v, want 42", mapped.Value)
	}

	// Fold eliminates to a single value using tag + payload.
	folded := just.Fold(func(tag string, p any) any {
		if tag == "Just" {
			return p
		}
		return 0
	})
	if folded.(int) != 21 {
		t.Errorf("Fold = %v, want 21", folded)
	}

	// Bind re-tags based on the payload.
	bound := just.Bind(func(p any) Variant {
		if p.(int) > 0 {
			return NewVariant("Positive", BitFlag(4), p)
		}
		return NewVariant("NonPositive", BitFlag(8), p)
	})
	if bound.Tag != "Positive" || bound.Value.(int) != 21 {
		t.Errorf("Bind = %s(%v), want Positive(21)", bound.Tag, bound.Value)
	}
}
