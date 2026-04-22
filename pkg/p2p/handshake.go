package p2p

import (
	"github.com/aziskebanaran/bvm-core/pkg/constants"
	"github.com/aziskebanaran/bvm-core/x/events"
)

type HandshakeRequest struct {
        NodeID      string `json:"node_id"`
        Version     string `json:"version"`
        GenesisHash string `json:"genesis_hash"`
        BestHeight  int    `json:"best_height"`
        // 🚩 TAMBAHKAN INI: Agar Nexus tahu jalur komunikasi balik
        P2PPort     int    `json:"p2p_port"`     // Default: 9090/9091
        APIPort     int    `json:"api_port"`     // Default: 8080/9092
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
