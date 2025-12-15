package zoo

import (
	"testing"

	"github.com/ebal0v/hse/zoo/internal/domain"
)

type acceptAllChecker struct{}

func (acceptAllChecker) Check(a domain.IAlive) bool { return true }

type rejectAllChecker struct{}

func (rejectAllChecker) Check(a domain.IAlive) bool { return false }

func TestAddAnimal_Accepted_FlowsThrough(t *testing.T) {
	svc := NewService(acceptAllChecker{}, NewMemAnimalRepo(), NewMemThingRepo())

	m := domain.NewMonkey(5, 7)
	accepted, inv := svc.AddAnimal(m, &m.Herbo.Animal)
	if !accepted {
		t.Fatalf("expected accepted=true")
	}
	if inv == 0 {
		t.Fatalf("expected non-zero inventory number")
	}

	if got := svc.TotalAnimals(); got != 1 {
		t.Fatalf("TotalAnimals = %d, want 1", got)
	}
	if got := svc.TotalFoodPerDayKg(); got != 5 {
		t.Fatalf("TotalFoodPerDayKg = %d, want 5", got)
	}

	cz := svc.ContactZooAnimals()
	if len(cz) != 1 {
		t.Fatalf("ContactZooAnimals len = %d, want 1", len(cz))
	}

	invA := svc.InventoryAnimalsReport()
	if len(invA) != 1 {
		t.Fatalf("InventoryAnimalsReport len = %d, want 1", len(invA))
	}
	invT := svc.InventoryThingsReport()
	if len(invT) != 0 {
		t.Fatalf("InventoryThingsReport len = %d, want 0", len(invT))
	}
	invAll := svc.InventoryReport()
	if len(invAll) != 1 {
		t.Fatalf("InventoryReport len = %d, want 1", len(invAll))
	}
}

func TestAddAnimal_Rejected_NotAdded(t *testing.T) {
	svc := NewService(rejectAllChecker{}, NewMemAnimalRepo(), NewMemThingRepo())
	r := domain.NewRabbit(3, 9)
	accepted, inv := svc.AddAnimal(r, &r.Herbo.Animal)
	if accepted || inv != 0 {
		t.Fatalf("expected rejected (accepted=false, inv=0), got accepted=%v inv=%d", accepted, inv)
	}
	if svc.TotalAnimals() != 0 {
		t.Fatalf("expected no animals in zoo after rejection")
	}
}

func TestAddThing_InventoryReports(t *testing.T) {
	svc := NewService(acceptAllChecker{}, NewMemAnimalRepo(), NewMemThingRepo())

	// No animals, add two things
	tbl := domain.NewTable()
	num1 := svc.AddThing(&tbl.Thing)
	if num1 == 0 {
		t.Fatalf("expected non-zero inv for table")
	}

	pc := domain.NewComputer()
	num2 := svc.AddThing(&pc.Thing)
	if num2 == 0 || num2 == num1 {
		t.Fatalf("expected distinct non-zero inv numbers")
	}

	if svc.TotalAnimals() != 0 {
		t.Fatalf("expected 0 animals")
	}
	if svc.TotalFoodPerDayKg() != 0 {
		t.Fatalf("expected 0 food kg/day")
	}

	invT := svc.InventoryThingsReport()
	if len(invT) != 2 {
		t.Fatalf("InventoryThingsReport len=%d, want 2", len(invT))
	}
	invA := svc.InventoryAnimalsReport()
	if len(invA) != 0 {
		t.Fatalf("InventoryAnimalsReport len=%d, want 0", len(invA))
	}
	invAll := svc.InventoryReport()
	if len(invAll) != 2 {
		t.Fatalf("InventoryReport len=%d, want 2", len(invAll))
	}
}

func TestContactZoo_FilterByKindness(t *testing.T) {
	svc := NewService(acceptAllChecker{}, NewMemAnimalRepo(), NewMemThingRepo())

	// kindness <= 5 should not be listed
	m1 := domain.NewMonkey(4, 5)
	svc.AddAnimal(m1, &m1.Herbo.Animal)
	// kindness > 5 should be listed
	m2 := domain.NewMonkey(4, 6)
	svc.AddAnimal(m2, &m2.Herbo.Animal)

	cz := svc.ContactZooAnimals()
	if len(cz) != 1 {
		t.Fatalf("ContactZooAnimals len=%d, want 1", len(cz))
	}
}
