package collections

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestListSerialization(t *testing.T) {
	list := Cons(1, Cons(2, Cons(3, Nil[int]())))
	
	// Test JSON Text Marshal
	b, err := json.Marshal(list)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}
	
	if string(b) != "[1,2,3]" {
		t.Fatalf("Expected [1,2,3], got %s", string(b))
	}
	
	// Test JSON Text Unmarshal
	var unmarshaledList List[int]
	err = json.Unmarshal(b, &unmarshaledList)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	
	h1, _ := unmarshaledList.Head()
	if h1 != 1 {
		t.Fatalf("Expected head to be 1, got %v", h1)
	}
	
	// Test Value (driver.Valuer)
	val, err := list.Value()
	if err != nil {
		t.Fatalf("Value failed: %v", err)
	}
	
	if !reflect.DeepEqual(val, "[1,2,3]") {
		t.Fatalf("Expected string [1,2,3], got %v", val)
	}
	
	// Test Scan (sql.Scanner)
	var scannedList List[int]
	err = scannedList.Scan("[4,5,6]")
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	
	h4, _ := scannedList.Head()
	if h4 != 4 {
		t.Fatalf("Expected scanned head to be 4, got %v", h4)
	}
}
