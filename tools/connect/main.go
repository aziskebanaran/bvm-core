package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
	"github.com/aziskebanaran/bvm-core/pkg/p2p"
	"github.com/aziskebanaran/bvm-core/pkg/constants"
)

type PeerInfo struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}

func main() {
	// 1. Ambil Height Lokal
	respStats, err := http.Get("http://localhost:8080/api/stats")
	currentHeight := 0
	if err == nil {
		var stats map[string]interface{}
		json.NewDecoder(respStats.Body).Decode(&stats)
		if network, ok := stats["network"].(map[string]interface{}); ok {
			currentHeight = int(network["height"].(float64))
		}
		respStats.Body.Close()
	}

	// 2. Baca daftar Peers dari JSON
	peerFile, err := os.ReadFile("peers.json")
	if err != nil {
		fmt.Println("❌ Gagal membaca peers.json. Pastikan file ada.")
		return
	}

	var peers []PeerInfo
	json.Unmarshal(peerFile, &peers)

	fmt.Printf("🌐 Memulai Auto-Connect ke %d target...\n", len(peers))
	fmt.Println("---------------------------------------")

	// 3. Looping untuk menyapa setiap Peer
	for _, peer := range peers {
		fmt.Printf("📡 Mengetuk pintu %s (%s)... ", peer.Name, peer.Address)

		req := p2p.HandshakeRequest{
			NodeID:      "BVM_GENESIS_001",
			Version:     constants.ProjectVersion,
			GenesisHash: constants.GenesisHash,
			BestHeight:  currentHeight,
			Timestamp:   time.Now().Unix(),
		}

		jsonData, _ := json.Marshal(req)
		
		// Set timeout agar tidak menunggu terlalu lama jika node mati
		client := http.Client{Timeout: 3 * time.Second}
		resp, err := client.Post(fmt.Sprintf("http://%s/api/peers", peer.Address), "application/json", bytes.NewBuffer(jsonData))

		if err != nil {
			fmt.Println("❌ OFFLINE")
			continue
		}

		var result p2p.HandshakeResponse
		json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()

		if result.Success {
			fmt.Printf("✅ CONNECTED (Height: %d)\n", result.Height)
		} else {
			fmt.Printf("🚫 DITOLAK: %s\n", result.Message)
		}
	}
	fmt.Println("---------------------------------------")
	fmt.Println("🏁 Selesai memindai jaringan.")
}
