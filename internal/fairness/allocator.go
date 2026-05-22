package fairness

import (
	"fmt"
)

type Allocator struct {
	maxCapacity float64
	tiers       map[string]Tier
}

// Allocate returns tierBaseRate and tierCapacity
// When total demand is lower than max capacity it provides the tier.base - optimal
// When demand is high and it can provide more then 1, it provides a proportionally reduced rate based on weight.
// When demand is high and it cannot guarantee minimal tier rate and that is lower than 1, it returns at least 1 rate back to all
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

	if totalDemand <= a.maxCapacity {
		return tier.baseRate, tier.capacity, nil
	}

	demandTierRate := (float64(tier.weight) / float64(totalWeight)) * a.maxCapacity

	canGuaranteeMinimum := a.maxCapacity >= float64(len(a.tiers))
	if canGuaranteeMinimum && demandTierRate < 1 {
		return 1, tier.capacity, nil
	}

	return demandTierRate, tier.capacity, nil
}

// NewAllocator takes maxCapcity and passes tiers map, returns a new Allocator
func NewAllocator(maxCapacity float64, tiers map[string]Tier) (*Allocator, error) {
	if maxCapacity <= 0 {
		return nil, fmt.Errorf("maxCapacity must be positive, got: %v", maxCapacity)
	}
	if len(tiers) == 0 {
		return nil, fmt.Errorf("tiers are empty, got: %v", tiers)
	}

	return &Allocator{
		maxCapacity: maxCapacity,
		tiers:       tiers,
	}, nil
}
