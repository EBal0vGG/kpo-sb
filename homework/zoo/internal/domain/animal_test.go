package domain

import "testing"

func TestHerbo_NameAndKindness(t *testing.T) {
	r := NewRabbit(3, 8)
	if got := r.Name(); got != "Rabbit" {
		t.Fatalf("Name()=%q, want Rabbit", got)
	}
	if kg := r.FoodKgPerDay(); kg != 3 {
		t.Fatalf("FoodKgPerDay()=%d, want 3", kg)
	}
	if k, ok := r.Kindness(); !ok || k != 8 {
		t.Fatalf("Kindness()=(%d,%v), want (8,true)", k, ok)
	}
}

func TestPredator_NameAndKindness(t *testing.T) {
	w := NewWolf(6)
	if got := w.Name(); got != "Wolf" {
		t.Fatalf("Name()=%q, want Wolf", got)
	}
	if kg := w.FoodKgPerDay(); kg != 6 {
		t.Fatalf("FoodKgPerDay()=%d, want 6", kg)
	}
	if _, ok := w.Kindness(); ok {
		t.Fatalf("Kindness() ok=true for predator, want false")
	}
}
