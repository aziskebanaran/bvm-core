package types

// Peer: Representasi satu teman di jaringan BVM dengan standar High-Availability
type Peer struct {
    Address   string `json:"address"`    // IP Address (e.g., "192.168.1.5")
    P2PPort   int    `json:"p2p_port"`   // Port untuk Handshake/Sync (9090/9091)
    APIPort   int    `json:"api_port"`   // Port untuk Kirim TX/Mempool (8080/9092)
    NodeID    string `json:"node_id"`
    LastSeen  int64  `json:"last_seen"`
    Version   string `json:"version"`
    FailCount int    `json:"-"`          // Internal counter, tidak diekspos ke JSON
}
