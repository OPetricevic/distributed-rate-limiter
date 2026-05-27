package store

import (
	"distributed-rate-limiter/internal/bucket"
	"fmt"
	"sync"
)

type Registry struct {
	buckets map[string]*bucket.Bucket
	mu      sync.RWMutex
}

func NewRegistry() *Registry {
	return &Registry{
		buckets: make(map[string]*bucket.Bucket),
	}
}

// GetOrCreate fetches a bucket in the registry, if not found it creates a bucket in the registry and returns it.
func (r *Registry) GetOrCreate(bucketKey string, capacity, refillRate float64) (*bucket.Bucket, error) {
	r.mu.RLock()
	b, exists := r.buckets[bucketKey]
	r.mu.RUnlock()
	if exists {
		return b, nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	// Double-check after acquiring write lock
	b, exists = r.buckets[bucketKey]
	if exists {
		return b, nil
	}

	newBucket, err := bucket.NewBucket(capacity, refillRate)
	if err != nil {
		return nil, fmt.Errorf("getOrCreate failed at creating a bucket %v", err)
	}
	r.buckets[bucketKey] = newBucket

	return newBucket, nil
}

func (r *Registry) Get(key string) *bucket.Bucket {
	r.mu.RLock()
	defer r.mu.RUnlock()
	b, exists := r.buckets[key]

	if exists {
		return b
	}
	return nil
}
