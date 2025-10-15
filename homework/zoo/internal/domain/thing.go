package domain

// Thing is a base inventory item
type Thing struct {
	Title string
	Inventoriable
}

func NewThing(title string) Thing { return Thing{Title: title} }

// Table concrete thing
type Table struct{ Thing }

func NewTable() *Table { return &Table{Thing: NewThing("Table")} }

// Computer concrete thing
type Computer struct{ Thing }

func NewComputer() *Computer { return &Computer{Thing: NewThing("Computer")} }
