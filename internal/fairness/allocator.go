package fairness

import (
	"fmt"
)

type Allocator struct {
	maxCapacity float64
	tiers       map[string]Tier
}

// Allocate returns tierBaseRate and tierCapacity
func (a *Allocator) Allocate(tierName string) (float64, float64, error) {
	tier, exists := a.tiers[tierName]
	if !exists {
		return 0, 0, fmt.Errorf("tier not found %s", tierName)
	}

	var (
		totalWeight int64
		totalDemand float64
	)

	for _, tier := range a.tiers {
		totalWeight += tier.weight
		totalDemand += tier.baseRate
	}

	return tier.baseRate, tier.capacity, nil
}

func NewAllocator(maxCapacity float64, tiers map[string]Tier) (*Allocator, error) {
	if maxCapacity <= 0 {
		return nil, fmt.Errorf("maxCapacity must be positive, got: %d", maxCapacity)
	}
	if len(tiers) == 0 {
		return nil, fmt.Errorf("tiers are empty, got: %v", tiers)
	}

	return &Allocator{
		maxCapacity: maxCapacity,
		tiers:       tiers,
	}, nil
}
