package domain

import "testing"

func TestInventoriable_EmbedsIntoAnimalAndThing(t *testing.T) {
	a := NewMonkey(4, 7)
	if a.InventoryNumber() != 0 {
		t.Fatalf("default InventoryNumber should be 0")
	}
	a.SetInventoryNumber(42)
	if a.InventoryNumber() != 42 {
		t.Fatalf("InventoryNumber=%d, want 42", a.InventoryNumber())
	}

	tbl := NewTable()
	if tbl.InventoryNumber() != 0 {
		t.Fatalf("default InventoryNumber should be 0")
	}
	tbl.SetInventoryNumber(7)
	if tbl.InventoryNumber() != 7 {
		t.Fatalf("InventoryNumber=%d, want 7", tbl.InventoryNumber())
	}
}
