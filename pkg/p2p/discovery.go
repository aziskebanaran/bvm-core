package p2p

import (
	"bvm.core/x/events"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type Peer struct {
	ID       string `json:"id"`
	Address  string `json:"address"` // Contoh: "192.168.1.5:9000"
	LastSeen int64  `json:"last_seen"`
}

type PeerManager struct {
	Peers map[string]Peer
	mu    sync.RWMutex
}

func NewPeerManager() *PeerManager {
	return &PeerManager{
		Peers: make(map[string]Peer),
	}
}

// AddPeer: Mendaftarkan teman baru ke jaringan Sultan
func (pm *PeerManager) AddPeer(p Peer) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.Peers[p.ID]; !exists {
		pm.Peers[p.ID] = p
		// 📢 Beritahu Event Bus bahwa ada tetangga baru!
		events.EmitEvent("PEER_JOINED", map[string]interface{}{
			"peer_id": p.ID,
			"addr":    p.Address,
		})
	}
}

// Discover: Menanyakan daftar teman ke Node lain (Standard Industry Seed)
func (pm *PeerManager) Discover(seedAddr string) {
	resp, err := http.Get(fmt.Sprintf("http://%s/api/peers", seedAddr))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var remotePeers []Peer
	json.NewDecoder(resp.Body).Decode(&remotePeers)

	for _, p := range remotePeers {
		pm.AddPeer(p)
	}
}
