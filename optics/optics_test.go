package optics_test

import (
	"testing"
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
