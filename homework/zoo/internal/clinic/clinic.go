package clinic

import (
	"math/rand"
	"time"

	"github.com/ebal0v/hse/zoo/internal/domain"
)

const (
	baseAccept = 0.5
	maxBoost   = 0.3
)

// Checker abstracts vet checks (healthy/unhealthy)
type Checker interface {
	Check(a domain.IAlive) bool
}

// VetClinic implements Checker with deterministic seed-able PRNG for demo.
type VetClinic struct {
	r *rand.Rand
}

func NewVetClinic() *VetClinic {
	return &VetClinic{r: rand.New(rand.NewSource(time.Now().UnixNano()))}
}

// NewVetClinicWithSeed allows deterministic behavior for tests
func NewVetClinicWithSeed(seed int64) *VetClinic {
	return &VetClinic{r: rand.New(rand.NewSource(seed))}
}

// Check returns true if accepted to zoo.
// Simple heuristic: higher food consumption => slightly higher chance to be healthy
func (v *VetClinic) Check(a domain.IAlive) bool {
	food := a.FoodKgPerDay()
	base := baseAccept
	boost := float64(food) * 0.02
	if boost > maxBoost {
		boost = maxBoost
	}
	return v.r.Float64() < base+boost
}
