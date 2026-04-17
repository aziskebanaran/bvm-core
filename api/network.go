package api

import (
    "github.com/aziskebanaran/bvm-core/x/bvm/keeper"
    "github.com/aziskebanaran/bvm-core/x/bvm/types"
    "encoding/json"
    "net/http"
    "strings"
)

// HandleHeartbeat: Sekarang memanggil menteri P2P lewat Jenderal 'k'
func HandleHeartbeat(k *keeper.Keeper, bc *types.Blockchain) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 1. Ambil IP Pengirim
        ip := r.RemoteAddr
        if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
            ip = strings.Split(forwarded, ",")[0]
        }

        // 2. Tambahkan ke daftar Peer lewat Menteri P2P
        if ip != "" && k.P2P != nil {
            // Karena AddPeer minta (ip, nodeID), kita beri nodeID kosong atau "anonymous" dulu
            k.P2P.AddPeer(ip, "remote-node") 
        }

        // 3. Ambil data MATANG dari querier.go (Status sudah termasuk PeerCount)
        status := k.GetStatus()

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(status)
    }
}

// HandleGetPeers: Sekarang butuh Jenderal 'k' agar bisa akses P2P
func HandleGetPeers(k *keeper.Keeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        
        var peerList []types.Peer
        total := 0

        if k.P2P != nil {
            peerList = k.P2P.GetPeers()
            total = k.P2P.CountActive()
        }

        json.NewEncoder(w).Encode(map[string]interface{}{
            "status": "success",
            "peers":  peerList,
            "total":  total,
        })
    }
}
