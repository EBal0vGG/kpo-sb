package domain

// Animal is a base for all animals. It implements inventory and alive behaviors via composition.
type Animal struct {
	name         string
	foodKgPerDay int
	Inventoriable
}

func NewAnimal(name string, foodKgPerDay int) Animal {
	return Animal{name: name, foodKgPerDay: foodKgPerDay}
}

func (a *Animal) FoodKgPerDay() int { return a.foodKgPerDay }

func (a *Animal) Name() string          { return a.name }
func (a *Animal) Kindness() (int, bool) { return 0, false }

// Herbo is an herbivore with kindness level [0..10].
type Herbo struct {
	Animal
	kindnessLevel int // 0..10
}

func NewHerbo(name string, foodKgPerDay int, kindness int) *Herbo {
	return &Herbo{Animal: NewAnimal(name, foodKgPerDay), kindnessLevel: kindness}
}

// Predator represents carnivorous animals
type Predator struct {
	Animal
}

func NewPredator(name string, foodKgPerDay int) *Predator {
	return &Predator{Animal: NewAnimal(name, foodKgPerDay)}
}

// Monkey is a concrete animal (herbivore rules not specified,
// treat as herbo-like without kindness constraint)
type Monkey struct{ Herbo }

func NewMonkey(foodKgPerDay int, kindness int) *Monkey {
	return &Monkey{Herbo: *NewHerbo("Monkey", foodKgPerDay, kindness)}
}

// Rabbit is herbivore
type Rabbit struct{ Herbo }

func NewRabbit(foodKgPerDay int, kindness int) *Rabbit {
	return &Rabbit{Herbo: *NewHerbo("Rabbit", foodKgPerDay, kindness)}
}

// Tiger is predator
type Tiger struct{ Predator }

func NewTiger(foodKgPerDay int) *Tiger {
	return &Tiger{Predator: *NewPredator("Tiger", foodKgPerDay)}
}

// Wolf is predator
type Wolf struct{ Predator }

func NewWolf(foodKgPerDay int) *Wolf {
	return &Wolf{Predator: *NewPredator("Wolf", foodKgPerDay)}
}

// Implement IAlive kindness exposure for embedded types
func (h *Herbo) Kindness() (int, bool)    { return h.kindnessLevel, true }
func (p *Predator) Kindness() (int, bool) { return 0, false }
