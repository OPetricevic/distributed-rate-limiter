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

func NewState(nodeID string) (*State, error) {
	if nodeID == "" {
		return nil, fmt.Errorf("creating state failed, check nodeID %v", nodeID)
	}

	return &State{
		nodeID:   nodeID,
		counters: make(map[string]NodeCounts),
	}, nil
}

func (s *State) Merge(userKey string, incoming NodeCounts) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if userKey == "" || incoming == nil {
		return fmt.Errorf("merging state failed, check userKey and incoming data")
	}

	if s.counters[userKey] == nil {
		s.counters[userKey] = make(NodeCounts)
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

	if s.counters[userKey] == nil {
		s.counters[userKey] = make(NodeCounts)
	}

	s.counters[userKey][s.nodeID] += 1

	return nil
}

func (s *State) Total(userKey string) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var sum int64
	for _, value := range s.counters[userKey] {
		sum += value
	}

	return sum

}

func (s *State) UserKeys() []string {

	var userKeys []string
	for userKey := range s.counters {
		userKeys = append(userKeys, userKey)
	}

	return userKeys
}
