package zoo

import (
	"testing"

	"github.com/ebal0v/hse/zoo/internal/domain"
)

func TestMemRepos_AddAndList(t *testing.T) {
	ar := NewMemAnimalRepo()
	tr := NewMemThingRepo()

	m := domain.NewMonkey(3, 6)
	_ = ar.Add(animalRec{name: m.Name(), alive: m, inv: &m.Herbo.Animal, kindness: 6})

	tbl := domain.NewTable()
	_ = tr.Add(thingRec{title: tbl.Title, inv: &tbl.Thing})

	al, _ := ar.List()
	if len(al) != 1 {
		t.Fatalf("animals len=%d, want 1", len(al))
	}

	tl, _ := tr.List()
	if len(tl) != 1 {
		t.Fatalf("things len=%d, want 1", len(tl))
	}
}
