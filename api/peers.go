package api

import (
	"github.com/aziskebanaran/bvm-core/x"
	"github.com/aziskebanaran/bvm-core/pkg/p2p"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
)

// HandlePeers: Gabungan fungsi Discovery (GET) dan Handshake (POST)
func HandlePeers(k x.BVMKeeper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Akses Menteri P2P lewat Jenderal
		p2pMgr := k.GetP2P()

		if r.Method == http.MethodGet {
			// 1. DISCOVERY: Mengambil daftar teman dari Jenderal
			peers := p2pMgr.GetPeers()
			json.NewEncoder(w).Encode(peers)

		} else if r.Method == http.MethodPost {
			// 2. HANDSHAKE: Node baru mendaftarkan diri
			var req p2p.HandshakeRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Bad Request", 400)
				return
			}

			// Validasi Genesis (Cek apakah versinya sama dengan Sultan)
			success, msg := p2p.ValidateHandshake(req, "BVM_GENESIS_001")

			if success {
				// LOGIKA OTOMATIS SULTAN: Catat IP pengirim
				host, _, _ := net.SplitHostPort(r.RemoteAddr)

				if host != "127.0.0.1" && host != "::1" {
					targetIP := host + ":8080"
					// Panggil Menteri P2P untuk mencatat di buku tamu (AddPeer)
					p2pMgr.AddPeer(targetIP, req.NodeID)
				}
			}

			json.NewEncoder(w).Encode(p2p.HandshakeResponse{
				Success: success,
				Message: msg,
				NodeID:  "BVM_NODE_SULTAN_001", 
			})
		}
	}
}


// Fungsi Internal untuk mencatat buku tamu
func saveNewPeer(newIP string, nodeID string) {
	const peerFile = "peers.json"

	// Baca data lama
	data, err := os.ReadFile(peerFile)
	var peers []map[string]string
	if err == nil {
		json.Unmarshal(data, &peers)
	}

	// Cek Duplikat: Agar peers.json tidak membengkak dengan IP yang sama
	for _, p := range peers {
		if p["address"] == newIP {
			return
		}
	}

	// Tambahkan ke list
	peers = append(peers, map[string]string{
		"address": newIP,
		"name":    nodeID,
	})

	// Tulis kembali ke file peers.json di folder root github.com/aziskebanaran/bvm-core/
	newData, _ := json.MarshalIndent(peers, "", "    ")
	os.WriteFile(peerFile, newData, 0644)
	fmt.Printf("📝 [P2P] %s Berhasil Terdaftar secara Otomatis!\n", nodeID)
}
