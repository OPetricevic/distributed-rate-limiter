package bucket

import (
	"fmt"
	"sync"
	"time"
)

type Bucket struct {
	capacity   float64
	tokenCount float64
	refillRate float64
	lastRefill time.Time
	mu         sync.Mutex
}

func NewBucket(capacity, refillRate float64) (*Bucket, error) {
	if capacity <= 0 || refillRate <= 0 {
		return nil, fmt.Errorf("capacity must be positive, got: %f", capacity)
	}

	return &Bucket{
		capacity:   capacity,
		tokenCount: capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}, nil
}

func (b *Bucket) Allow(cost int) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Add Tokens based on elapsed time
	b.refill()

	if b.tokenCount >= float64(cost) {
		b.tokenCount -= float64(cost)
		return true
	}

	return false
}

func (b *Bucket) Tokens() float64 {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.refill()

	return b.tokenCount
}

func (b *Bucket) refill() {
	elapsed := time.Since(b.lastRefill).Seconds()
	b.tokenCount += elapsed * b.refillRate
	if b.tokenCount > b.capacity {
		b.tokenCount = b.capacity
	}

	b.lastRefill = time.Now()
}
