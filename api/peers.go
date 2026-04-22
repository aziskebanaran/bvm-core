package api

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"
	"github.com/aziskebanaran/bvm-core/x"
	"github.com/aziskebanaran/bvm-core/pkg/p2p"
	"github.com/aziskebanaran/bvm-core/x/bvm/types"
)

func HandlePeers(k x.BVMKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        p2pMgr := k.GetP2P()

        if r.Method == http.MethodGet {
            // Standar Get: Berikan daftar teman yang diketahui
            peers := p2pMgr.GetPeers()
            json.NewEncoder(w).Encode(peers)
        } else if r.Method == http.MethodPost {
            var req p2p.HandshakeRequest
            if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
                http.Error(w, "Bad Request", 400)
                return
            }

            // Jalankan ritual Handshake (Cek Genesis & Versi)
            success, msg := p2p.ValidateHandshake(req, "BVM_GENESIS_001")

            var allPeers []types.Peer // Persiapan untuk kirim balik daftar teman

            if success {
                host, _, _ := net.SplitHostPort(r.RemoteAddr)
                // Filter localhost agar tidak membingungkan peers.json
                if host != "127.0.0.1" && host != "::1" {
                    // 🚩 SEKARANG DATA PORT DIAMBIL DARI REQUEST (req)
                    saveNewPeer(host, req.NodeID, req.P2PPort, req.APIPort)

                    // Masukkan juga ke memori (RAM) Keeper
                    p2pMgr.AddPeer(host, req.NodeID) 
                }

                // Ambil daftar seluruh teman untuk diberikan ke pendatang baru
                allPeers = p2pMgr.GetPeers()
            }

            // Kirim balasan paspor beserta daftar teman
            json.NewEncoder(w).Encode(struct {
                p2p.HandshakeResponse
                Peers []types.Peer `json:"peers"`
            }{
                HandshakeResponse: p2p.HandshakeResponse{
                    Success: success,
                    Message: msg,
                    NodeID:  "BVM_NEXUS_MAIN", // Nama server Sultan
                    Height:  k.GetLastHeight(), // Melaporkan ketinggian blok saat ini
                },
                Peers: allPeers,
            })
        }
    }
}


func saveNewPeer(newIP string, nodeID string, p2pPort int, apiPort int) {
    // Gunakan folder data_nexus jika ini dijalankan di lingkungan Nexus
    const peerDir = "data_nexus" 
    const peerFile = "data_nexus/peers.json"

    // 1. Proteksi Folder
    if _, err := os.Stat(peerDir); os.IsNotExist(err) {
        os.MkdirAll(peerDir, 0755)
    }

    // 2. Baca data lama
    data, err := os.ReadFile(peerFile)
    var peers []map[string]interface{} // Gunakan interface{} agar bisa simpan INT
    if err == nil {
        json.Unmarshal(data, &peers)
    }

    // 3. Cek Duplikasi (Berdasarkan IP dan NodeID)
    for _, p := range peers {
        if p["address"] == newIP && p["node_id"] == nodeID {
            return 
        }
    }

    // 4. Rakit Identitas Baru sesuai struct Peer Sultan
    newPeer := map[string]interface{}{
        "address":   newIP,
        "p2p_port":  p2pPort,
        "api_port":  apiPort,
        "node_id":   nodeID,
        "last_seen": time.Now().Unix(),
        "version":   "BVM_PRO_1.0",
    }

    peers = append(peers, newPeer)

    // 5. Simpan secara permanen
    newData, _ := json.MarshalIndent(peers, "", "    ")
    os.WriteFile(peerFile, newData, 0644)
    
    fmt.Printf("📝 [P2P] Node %s Berhasil Terdaftar (P2P:%d | API:%d)\n", 
        nodeID, p2pPort, apiPort)
}


