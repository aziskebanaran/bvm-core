package p2p

import (
	"github.com/aziskebanaran/BVM.core/pkg/constants"
	"github.com/aziskebanaran/BVM.core/x/events"
)

type HandshakeRequest struct {
	NodeID      string `json:"node_id"`
	Version     string `json:"version"`      // Harus cocok dengan constants.ProjectVersion
	GenesisHash string `json:"genesis_hash"` // Kunci utama: Harus sama!
	BestHeight  int    `json:"best_height"`  // Untuk tahu siapa yang lebih update
	Timestamp   int64  `json:"timestamp"`
}

type HandshakeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	NodeID  string `json:"node_id"`
	Height  int    `json:"height"`
}

// ValidateHandshake: Filter keamanan Sultan
func ValidateHandshake(req HandshakeRequest, myGenesis string) (bool, string) {
	// 1. Cek Versi (Sultan tidak ingin Node jadul merusak jaringan)
	if req.Version != constants.ProjectVersion {
		return false, "Versi aplikasi tidak cocok! Harap update."
	}

	// 2. Cek Genesis Hash (Kunci paling fatal)
	// Jika Genesis beda, berarti mereka berada di "Dunia Paralel" (Fork)
	if req.GenesisHash != myGenesis {
		return false, "Genesis Hash berbeda! Anda bukan bagian dari jaringan BVM ini."
	}

	// 3. Catat ke Event Bus
	events.EmitEvent("P2P_HANDSHAKE", map[string]interface{}{
		"peer_id": req.NodeID,
		"height":  req.BestHeight,
		"status":  "verified",
	})

	return true, "Handshake Berhasil. Selamat datang di BVM Mainnet!"
}
