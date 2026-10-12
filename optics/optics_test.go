package optics_test

import (
	"testing"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/optics"
)

type Address struct {
	City string
}

type User struct {
	Name    string
	Address Address
}

func TestLens(t *testing.T) {
	// Lens for User -> Address
	userAddressLens := optics.Lens[User, Address]{
		Get: func(u User) Address { return u.Address },
		Set: func(u User, a Address) User { 
			return User{Name: u.Name, Address: a} // Brand new copy
		},
	}

	// Lens for Address -> City string
	addressCityLens := optics.Lens[Address, string]{
		Get: func(a Address) string { return a.City },
		Set: func(a Address, c string) Address { 
			return Address{City: c} 
		},
	}

	// Compose them! User -> City
	userCityLens := optics.ComposeLenses(userAddressLens, addressCityLens)

	u1 := User{Name: "Alice", Address: Address{City: "London"}}
	
	// Get the deep value
	city := userCityLens.Get(u1)
	if city != "London" {
		t.Fatalf("Expected London, got %s", city)
	}

	// Set the deep value IMMUTABLY
	u2 := userCityLens.Set(u1, "Paris")
	
	if u1.Address.City != "London" {
		t.Fatal("Original user was mutated! Lenses must be pure.")
	}
	if u2.Address.City != "Paris" {
		t.Fatal("Lens failed to set the new value.")
	}
}

func TestRecordLens(t *testing.T) {
	orig := map[string]any{"user": "alice", "score": 42}
	userLens := optics.RecordLens("user")

	if userLens.Get(orig) != "alice" {
		t.Fatalf("expected 'alice', got %v", userLens.Get(orig))
	}

	updated := userLens.Set(orig, "bob")
	if updated["user"] != "bob" {
		t.Fatalf("expected updated 'bob', got %v", updated["user"])
	}
	if orig["user"] != "alice" {
		t.Fatalf("original map mutated! expected 'alice', got %v", orig["user"])
	}
}

func TestMapLens(t *testing.T) {
	orig := map[string]int{"alpha": 1, "beta": 2}
	alphaLens := optics.MapLens[string, int]("alpha")

	if alphaLens.Get(orig) != 1 {
		t.Fatalf("expected 1, got %d", alphaLens.Get(orig))
	}

	updated := alphaLens.Set(orig, 100)
	if updated["alpha"] != 100 {
		t.Fatalf("expected 100, got %d", updated["alpha"])
	}
	if orig["alpha"] != 1 {
		t.Fatalf("original map mutated! expected 1, got %d", orig["alpha"])
	}
}

func TestVariantPrism(t *testing.T) {
	activeFlag := dsltypes.BitFlag(1)
	pausedFlag := dsltypes.BitFlag(2)

	activePrism := optics.VariantPrism[int]("Active", activeFlag)
	pausedPrism := optics.VariantPrism[string]("Paused", pausedFlag)

	vActive := activePrism.Review(42)
	if !vActive.Is("Active") || vActive.Value != 42 {
		t.Fatalf("expected Active(42), got %v", vActive)
	}

	// Preview success
	mActive := activePrism.Preview(vActive)
	if !mActive.IsJust() || mActive.UnwrapOr(0) != 42 {
		t.Fatalf("expected Just(42), got %v", mActive)
	}

	// Preview mismatch
	mPaused := pausedPrism.Preview(vActive)
	if !mPaused.IsNothing() {
		t.Fatalf("expected Nothing for mismatched branch, got %v", mPaused)
	}
}

func TestEitherPrisms(t *testing.T) {
	rightPrism := optics.EitherRightPrism[error, int]()
	leftPrism := optics.EitherLeftPrism[string, int]()

	eRight := rightPrism.Review(100)
	if !eRight.IsRight() {
		t.Fatalf("expected Right(100), got %v", eRight)
	}
	mR := rightPrism.Preview(eRight)
	if !mR.IsJust() || mR.UnwrapOr(0) != 100 {
		t.Fatalf("expected Just(100), got %v", mR)
	}

	eLeft := leftPrism.Review("failure")
	if !eLeft.IsLeft() {
		t.Fatalf("expected Left('failure'), got %v", eLeft)
	}
	mL := leftPrism.Preview(eLeft)
	if !mL.IsJust() || mL.UnwrapOr("") != "failure" {
		t.Fatalf("expected Just('failure'), got %v", mL)
	}
}

func TestHAMTLens(t *testing.T) {
	h := collections.NewHAMT[string, int]().Set("score", 95)
	scoreLens := optics.HAMTLens[string, int]("score")

	// Get
	mVal := scoreLens.Get(h)
	if !mVal.IsJust() || mVal.UnwrapOr(0) != 95 {
		t.Fatalf("expected Just(95), got %v", mVal)
	}

	// Set immutably
	h2 := scoreLens.Set(h, functions.Just(100))
	mVal2 := scoreLens.Get(h2)
	if !mVal2.IsJust() || mVal2.UnwrapOr(0) != 100 {
		t.Fatalf("expected Just(100) in h2, got %v", mVal2)
	}

	// Verify original h not mutated
	mOrig := scoreLens.Get(h)
	if !mOrig.IsJust() || mOrig.UnwrapOr(0) != 95 {
		t.Fatalf("expected original h score 95, got %v", mOrig)
	}
}
