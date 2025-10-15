package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"go.uber.org/dig"

	"github.com/ebal0v/hse/zoo/internal/clinic"
	"github.com/ebal0v/hse/zoo/internal/cmdutil"
	"github.com/ebal0v/hse/zoo/internal/domain"
	"github.com/ebal0v/hse/zoo/internal/zoo"
)

func buildContainer() *dig.Container {
	c := dig.New()
	_ = c.Provide(func() clinic.Checker { return clinic.NewVetClinic() })
	_ = c.Provide(func() zoo.AnimalRepository { return zoo.NewMemAnimalRepo() })
	_ = c.Provide(func() zoo.ThingRepository { return zoo.NewMemThingRepo() })
	_ = c.Provide(func(ch clinic.Checker, ar zoo.AnimalRepository, tr zoo.ThingRepository) zoo.Service {
		return zoo.NewService(ch, ar, tr)
	})
	return c
}

func main() {
	c := buildContainer()
	_ = c.Invoke(func(s zoo.Service) {
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Println("1) Add Monkey  2) Add Rabbit  3) Add Tiger  4) Add Wolf  5) Add Table  6) Add Computer  7) Report Food  8) Contact Zoo  9) Inventory (all)  a) Inventory (animals)  b) Inventory (things)  0) Exit")
			fmt.Print("> ")
			line, err := reader.ReadString('\n')
			if err != nil {
				// корректное завершение при EOF (Ctrl+D)
				return
			}
			line = strings.TrimSpace(line)
			if line == "0" {
				return
			}
			switch line {
			case "1":
				food := cmdutil.AskIntMin(reader, "Food kg/day (>=0): ", 0)
				kind := cmdutil.AskIntRange(reader, "Kindness 0-10: ", 0, 10)
				m := domain.NewMonkey(food, kind)
				accepted, num := s.AddAnimal(m, &m.Herbo.Animal)
				fmt.Println(resultStr("Monkey", accepted, num))
			case "2":
				food := cmdutil.AskIntMin(reader, "Food kg/day (>=0): ", 0)
				kind := cmdutil.AskIntRange(reader, "Kindness 0-10: ", 0, 10)
				r := domain.NewRabbit(food, kind)
				accepted, num := s.AddAnimal(r, &r.Herbo.Animal)
				fmt.Println(resultStr("Rabbit", accepted, num))
			case "3":
				food := cmdutil.AskIntMin(reader, "Food kg/day (>=0): ", 0)
				t := domain.NewTiger(food)
				accepted, num := s.AddAnimal(t, &t.Predator.Animal)
				fmt.Println(resultStr("Tiger", accepted, num))
			case "4":
				food := cmdutil.AskIntMin(reader, "Food kg/day (>=0): ", 0)
				w := domain.NewWolf(food)
				accepted, num := s.AddAnimal(w, &w.Predator.Animal)
				fmt.Println(resultStr("Wolf", accepted, num))
			case "5":
				t := domain.NewTable()
				num := s.AddThing(&t.Thing)
				fmt.Printf("Added Table #%d\n", num)
			case "6":
				c := domain.NewComputer()
				num := s.AddThing(&c.Thing)
				fmt.Printf("Added Computer #%d\n", num)
			case "7":
				fmt.Printf("Animals: %d, Food per day: %d kg\n", s.TotalAnimals(), s.TotalFoodPerDayKg())
			case "8":
				list := s.ContactZooAnimals()
				if len(list) == 0 {
					fmt.Println("No animals eligible for contact zoo yet")
				} else {
					fmt.Println(strings.Join(list, "\n"))
				}
			case "9":
				rep := s.InventoryReport()
				if len(rep) == 0 {
					fmt.Println("No inventory yet")
				} else {
					fmt.Println(strings.Join(rep, "\n"))
				}
			case "a":
				rep := s.InventoryAnimalsReport()
				if len(rep) == 0 {
					fmt.Println("No animals in inventory yet")
				} else {
					fmt.Println(strings.Join(rep, "\n"))
				}
			case "b":
				rep := s.InventoryThingsReport()
				if len(rep) == 0 {
					fmt.Println("No things in inventory yet")
				} else {
					fmt.Println(strings.Join(rep, "\n"))
				}
			default:
				fmt.Println("Unknown option")
			}
		}
	})
}

func resultStr(name string, accepted bool, inv int) string {
	if accepted {
		return fmt.Sprintf("Accepted %s, inventory #%d", name, inv)
	}
	return fmt.Sprintf("Rejected %s by vet clinic", name)
}
