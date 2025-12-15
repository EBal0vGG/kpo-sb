package clinic

import (
	"testing"
)

type dummyAlive struct{ f int }

func (d dummyAlive) FoodKgPerDay() int     { return d.f }
func (d dummyAlive) Name() string          { return "dummy" }
func (d dummyAlive) Kindness() (int, bool) { return 0, false }

func TestVetClinic_Check_DeterministicBySeed(t *testing.T) {
	cases := []struct {
		seed int64
		food int
		want bool
	}{
		{seed: 1, food: 0, want: false},
		{seed: 1, food: 20, want: true},
		{seed: 42, food: 5, want: true},
		{seed: 42, food: 15, want: true},
	}
	for i, c := range cases {
		vc := NewVetClinicWithSeed(c.seed)
		got := vc.Check(dummyAlive{f: c.food})
		if got != c.want {
			t.Fatalf("case %d: seed=%d food=%d => got %v, want %v", i, c.seed, c.food, got, c.want)
		}
	}
}
