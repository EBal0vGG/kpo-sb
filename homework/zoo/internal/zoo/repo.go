package zoo

// AnimalRepository abstracts storage for animals
type AnimalRepository interface {
	Add(rec animalRec) error
	List() ([]animalRec, error)
}

// ThingRepository abstracts storage for things
type ThingRepository interface {
	Add(rec thingRec) error
	List() ([]thingRec, error)
}
