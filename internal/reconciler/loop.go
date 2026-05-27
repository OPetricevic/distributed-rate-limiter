package reconciler

import (
	"distributed-rate-limiter/internal/gossip"
	store "distributed-rate-limiter/internal/storage"
	"fmt"
	"time"
)

type Reconciler struct {
	state    *gossip.State
	registry *store.Registry
	interval time.Duration
}

func NewReconciler(state *gossip.State, registry *store.Registry, interval time.Duration) (*Reconciler, error) {
	if state == nil {
		return nil, fmt.Errorf("reconciler state is empty %v", state)
	}
	if registry == nil {
		return nil, fmt.Errorf("reconciler registry is empty %v", registry)
	}
	if interval <= 0 {
		return nil, fmt.Errorf("interval is non existant %v", interval)
	}

	return &Reconciler{
		state:    state,
		registry: registry,
		interval: interval,
	}, nil
}

func (r *Reconciler) Run() {
	for {

		for _, key := range r.state.UserKeys() {
			globalTotal := r.state.Total(key)
			bucket := r.registry.Get(key)
			if bucket == nil {
				continue
			}
			capacity := bucket.Capacity()
			remaining := capacity - float64(globalTotal)

			if float64(globalTotal) >= capacity {
				bucket.SetTokens(0.00)
			} else {
				bucket.SetTokens(remaining)
			}

		}
		time.Sleep(r.interval)
	}
}
