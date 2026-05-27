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

// NewBucket takes capacity and refil rate, and creates a new bucket
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

// Allow returns a boolean, after it checks if a requester is able to ping
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

// Tokens returns the tokenCount from a bucket
func (b *Bucket) Tokens() float64 {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.refill()

	return b.tokenCount
}

func (b *Bucket) Capacity() float64 {
	return b.capacity
}

// refill fills the bucket token count based on elapsed time
func (b *Bucket) refill() {
	elapsed := time.Since(b.lastRefill).Seconds()
	b.tokenCount += elapsed * b.refillRate
	if b.tokenCount > b.capacity {
		b.tokenCount = b.capacity
	}

	b.lastRefill = time.Now()
}

func (b *Bucket) SetTokens(amount float64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.tokenCount = amount

}
