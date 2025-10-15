package domain

// Inventory marks a type as an item tracked by inventory with a unique number.
type Inventory interface {
	InventoryNumber() int
	SetInventoryNumber(num int)
}

// Inventoriable is an embeddable implementation of Inventory
type Inventoriable struct{ inventoryNum int }

func (i *Inventoriable) InventoryNumber() int       { return i.inventoryNum }
func (i *Inventoriable) SetInventoryNumber(num int) { i.inventoryNum = num }
