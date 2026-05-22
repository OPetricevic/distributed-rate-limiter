package gossip

import (
	"fmt"
	"sync"
	"time"
)

type Peer struct {
	id       string
	address  string
	alive    bool
	lastSeen time.Time
}

type PeerManager struct {
	peerList map[string]Peer
	mu       sync.RWMutex
}

func NewPeerManager() *PeerManager {
	return &PeerManager{
		peerList: make(map[string]Peer),
	}
}

func (pm *PeerManager) AddPeer(id, address string) (*Peer, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if (id == "") || (address == "") {
		return nil, fmt.Errorf("adding peer encountered an error, at id value %v, with address %v", id, address)
	}

	newPeer := Peer{
		id:       id,
		address:  address,
		alive:    true,
		lastSeen: time.Now(),
	}
	pm.peerList[id] = newPeer

	return &newPeer, nil
}

func (pm *PeerManager) RemovePeer(id string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if id == "" {
		return fmt.Errorf("removing peer from list, encountered an error with value of %v", id)
	}

	_, ok := pm.peerList[id]
	if !ok {
		return fmt.Errorf("peer not found with provided id %v", id)
	}

	delete(pm.peerList, id)

	return nil
}

func (pm *PeerManager) MarkAlive(id string) (bool, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if id == "" {
		return false, fmt.Errorf("marking peer alive failed with id %v", id)
	}

	peer, ok := pm.peerList[id]
	if !ok {
		return false, fmt.Errorf("marking peer alive, failed locating peer in list with id %v", id)
	}

	peer.alive = true
	peer.lastSeen = time.Now()
	pm.peerList[id] = peer

	return peer.alive, nil
}

func (pm *PeerManager) MarkDead(id string) (bool, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if id == "" {
		return false, fmt.Errorf("marking peer dead, encountered error with id %v", id)
	}

	peer, ok := pm.peerList[id]
	if !ok {
		return false, fmt.Errorf("marking peer dead, failed locating peer in list with id %v", id)
	}
	peer.alive = false
	pm.peerList[id] = peer

	return peer.alive, nil
}

func (pm *PeerManager) LivePeers() []Peer {
	var alivePeers []Peer
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	for _, peer := range pm.peerList {
		if peer.alive {
			alivePeers = append(alivePeers, peer)
		}
	}
	return alivePeers
}
