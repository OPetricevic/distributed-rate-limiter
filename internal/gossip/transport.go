package gossip

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"time"
)

type Transport struct {
	udp         net.PacketConn
	state       *State
	peerManager *PeerManager
	interval    time.Duration
}

func NewTransport(address string, state *State, peerManager *PeerManager, interval time.Duration) (*Transport, error) {
	if address == "" {
		return nil, fmt.Errorf("creating new transport, address value is missing %v", address)
	}

	if state == nil {
		return nil, fmt.Errorf("creating new transport, statue is missing %v", state)
	}

	if peerManager == nil {
		return nil, fmt.Errorf("creating new transport, peer manager is missing %v", peerManager)
	}

	if interval <= 0 {
		return nil, fmt.Errorf("creating new transport, interval is missing %v", interval)
	}

	conn, err := net.ListenPacket("udp", address)
	if err != nil {
		return nil, fmt.Errorf("listening packet failed with address %v", address)
	}

	return &Transport{
		udp:         conn,
		state:       state,
		peerManager: peerManager,
		interval:    interval,
	}, nil
}

func (t *Transport) Send() error {

	peers := t.peerManager.LivePeers()
	if len(peers) == 0 {
		return fmt.Errorf("there are no peers to send to %v", peers)
	}
	peer := peers[rand.Intn(len(peers))]

	addr, err := net.ResolveUDPAddr("udp", peer.address)
	if err != nil {
		return fmt.Errorf("resolving address failed for peer %v", peer)
	}

	data, err := json.Marshal(t.state.counters)
	if err != nil {
		return fmt.Errorf("failed data construction while for peer %v", peer)
	}

	_, err = t.udp.WriteTo(data, addr)
	if err != nil {
		return fmt.Errorf("failed writing to %v", addr)
	}

	return nil
}

func (t *Transport) Listen() error {

	buffer := make([]byte, 65535)
	n, _, err := t.udp.ReadFrom(buffer)
	if err != nil {
		return fmt.Errorf("failed reading from a buffer %v", buffer)
	}

	var receivedData map[string]NodeCounts

	err = json.Unmarshal(buffer[:n], &receivedData)
	if err != nil {
		return fmt.Errorf("unmarshalling buffers received data failed, %v", buffer[:n])
	}

	for userKey, nodeCounts := range receivedData {
		t.state.Merge(userKey, nodeCounts)
	}

	return nil
}
