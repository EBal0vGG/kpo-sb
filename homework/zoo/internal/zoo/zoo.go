package zoo

import (
	"fmt"
	"sort"

	"github.com/ebal0v/hse/zoo/internal/clinic"
	"github.com/ebal0v/hse/zoo/internal/domain"
)

// Service defines zoo operations
type Service interface {
	AddAnimal(a domain.IAlive, inv domain.Inventory) (accepted bool, invNumber int)
	AddThing(t domain.Inventory) int
	TotalAnimals() int
	TotalFoodPerDayKg() int
	ContactZooAnimals() []string
	InventoryReport() []string
	InventoryAnimalsReport() []string
	InventoryThingsReport() []string
}

type service struct {
	checker clinic.Checker
	nextInv int
	animals AnimalRepository
	things  ThingRepository
}

type animalRec struct {
	name     string
	alive    domain.IAlive
	inv      domain.Inventory
	kindness int // cached when applicable
}

type thingRec struct {
	title string
	inv   domain.Inventory
}

func NewService(checker clinic.Checker, aRepo AnimalRepository, tRepo ThingRepository) Service {
	return &service{checker: checker, nextInv: 1, animals: aRepo, things: tRepo}
}

func (s *service) allocateInv(i domain.Inventory) int {
	num := s.nextInv
	s.nextInv++
	i.SetInventoryNumber(num)
	return num
}

func (s *service) AddAnimal(a domain.IAlive, inv domain.Inventory) (bool, int) {
	if !s.checker.Check(a) {
		return false, 0
	}
	invNum := s.allocateInv(inv)
	name := a.Name()
	k, ok := a.Kindness()
	kindness := 0
	if ok {
		kindness = k
	}
	_ = s.animals.Add(animalRec{name: name, alive: a, inv: inv, kindness: kindness})
	return true, invNum
}

func (s *service) AddThing(t domain.Inventory) int {
	invNum := s.allocateInv(t)
	title := "Thing"
	switch v := t.(type) {
	case *domain.Table:
		title = v.Title
	case *domain.Computer:
		title = v.Title
	case *domain.Thing:
		title = v.Title
	default:
		_ = v
	}
	_ = s.things.Add(thingRec{title: title, inv: t})
	return invNum
}

func (s *service) TotalAnimals() int { list, _ := s.animals.List(); return len(list) }

func (s *service) TotalFoodPerDayKg() int {
	var total int
	list, _ := s.animals.List()
	for _, r := range list {
		total += r.alive.FoodKgPerDay()
	}
	return total
}

func (s *service) ContactZooAnimals() []string {
	var res []string
	list, _ := s.animals.List()
	for _, r := range list {
		if _, ok := r.alive.Kindness(); ok && r.kindness > 5 {
			res = append(res, fmt.Sprintf("%s #%d (kindness=%d)", r.name, r.inv.InventoryNumber(), r.kindness))
		}
	}
	sort.Strings(res)
	return res
}

func (s *service) InventoryReport() []string {
	// Concatenate specialized reports to keep single source of rules
	lines := append([]string(nil), s.InventoryAnimalsReport()...)
	lines = append(lines, s.InventoryThingsReport()...)
	sort.Strings(lines)
	return lines
}

func (s *service) InventoryAnimalsReport() []string {
	var lines []string
	al, _ := s.animals.List()
	for _, r := range al {
		lines = append(lines, fmt.Sprintf("%s #%d", r.name, r.inv.InventoryNumber()))
	}
	sort.Strings(lines)
	return lines
}

func (s *service) InventoryThingsReport() []string {
	var lines []string
	tl, _ := s.things.List()
	for _, r := range tl {
		lines = append(lines, fmt.Sprintf("%s #%d", r.title, r.inv.InventoryNumber()))
	}
	sort.Strings(lines)
	return lines
}
