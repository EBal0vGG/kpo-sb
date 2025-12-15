package domain

// IAlive marks a type as a living entity that consumes food daily.
// FoodKgPerDay returns how many kilograms of food the entity consumes per day.
type IAlive interface {
	FoodKgPerDay() int
	Name() string
	// Kindness returns level and true for herbivores; false if not applicable
	Kindness() (int, bool)
}
