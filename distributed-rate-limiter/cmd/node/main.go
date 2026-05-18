package main

import (
	store "distributed-rate-limiter/internal/storage"
	"net/http"
)

type Server struct {
	registry *store.Registry
}

const (
	capacity  = 5
	refilRate = 1
)

func (s *Server) CheckHandler(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("X-API-Key")

	bucket, err := s.registry.GetOrCreate(key, capacity, refilRate)
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
	s := &Server{registry: store.NewRegistry()}

	http.HandleFunc("/check", s.CheckHandler)
	http.ListenAndServe(":8080", nil)
}
