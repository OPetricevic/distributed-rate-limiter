package main

import (
	"distributed-rate-limiter/internal/fairness"
	store "distributed-rate-limiter/internal/storage"
	"log"
	"net/http"
)

type Server struct {
	registry  *store.Registry
	allocator *fairness.Allocator
}

func (s *Server) CheckHandler(w http.ResponseWriter, r *http.Request) {
	userKey := r.Header.Get("X-API-Key")
	tierName := r.Header.Get("X-Tier")

	baseTierRate, baseCapacity, err := s.allocator.Allocate(tierName)
	if err != nil {
		http.Error(w, "CheckHandler failed at allocating tier", http.StatusBadRequest)
		return
	}

	bucket, err := s.registry.GetOrCreate(userKey, baseCapacity, baseTierRate)
	if err != nil {
		http.Error(w, "CheckHandler failed at fetching or creating bucket", http.StatusInternalServerError)
		return
	}

	if !bucket.Allow(1) {
		http.Error(w, "CheckHandler too many requests, try again later", http.StatusTooManyRequests)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	allocator, err := fairness.NewAllocator(100000.00, fairness.Tiers)
	if err != nil {
		log.Fatal("creating allocator errored out with", err)
	}

	s := &Server{
		registry:  store.NewRegistry(),
		allocator: allocator,
	}

	http.HandleFunc("/check", s.CheckHandler)
	http.ListenAndServe(":8080", nil)
}
