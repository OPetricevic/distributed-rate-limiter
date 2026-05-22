package gossip

import (
	"fmt"
	"sync"
)

type NodeCounts map[string]int64

type State struct {
	nodeID   string
	counters map[string]NodeCounts
	mu       sync.RWMutex
}

func (s *State) Merge(userKey string, incoming NodeCounts) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if userKey == "" || incoming == nil {
		return fmt.Errorf("merging state failed, check userKey and incoming data")
	}

	for nodeID := range incoming {
		if incoming[nodeID] > s.counters[userKey][nodeID] {
			s.counters[userKey][nodeID] = incoming[nodeID]
		}
	}

	return nil
}

func (s *State) Record(userKey string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if userKey == "" {
		return fmt.Errorf("recording state failed, check userKey")
	}

	s.counters[userKey][s.nodeID] += 1

	return nil
}

func (s *State) Total(userKey string) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

}
