package zoo

// In-memory repositories

type memAnimalRepo struct{ items []animalRec }

func NewMemAnimalRepo() AnimalRepository { return &memAnimalRepo{} }

func (m *memAnimalRepo) Add(rec animalRec) error { m.items = append(m.items, rec); return nil }

func (m *memAnimalRepo) List() ([]animalRec, error) { return append([]animalRec(nil), m.items...), nil }

type memThingRepo struct{ items []thingRec }

func NewMemThingRepo() ThingRepository { return &memThingRepo{} }

func (m *memThingRepo) Add(rec thingRec) error { m.items = append(m.items, rec); return nil }

func (m *memThingRepo) List() ([]thingRec, error) { return append([]thingRec(nil), m.items...), nil }
