package p2p

import (
	"bvm.core/x/events"
	"fmt"
	"net"
	"net/http"
	"time"
)

// StartNode: Tetap menggunakan TCP untuk koneksi dasar
func StartNode(port int) {
	address := fmt.Sprintf("0.0.0.0:%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Printf("❌ Gagal menghidupkan P2P: %v\n", err)
		return
	}

	fmt.Printf("📡 P2P Node BVM Aktif di %s\n", address)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				continue
			}
			// Di masa depan, di sini tempat Handshake TCP murni dilakukan
			fmt.Printf("🤝 Koneksi P2P diterima: %s\n", conn.RemoteAddr().String())
			conn.Close() 
		}
	}()
}

// SyncBlocks: Fungsi inti untuk menarik blok yang tertinggal
func SyncBlocks(targetIP string, startHeight int, targetHeight int) {
	fmt.Printf("🔄 [SYNC] Memulai sinkronisasi: %d -> %d dari %s\n", startHeight, targetHeight, targetIP)
	
	client := &http.Client{Timeout: 10 * time.Second}

	for h := startHeight + 1; h <= targetHeight; h++ {
		// 1. Request blok ke API target
		url := fmt.Sprintf("http://%s/api/block?height=%d", targetIP, h)
		resp, err := client.Get(url)
		if err != nil {
			fmt.Printf("⚠️ Gagal mengambil blok #%d: %v\n", h, err)
			break
		}
		
		// 2. Dekode data blok (Bisa diganti MsgPack nanti agar lebih ngebut)
		// Di sini kita asumsikan respon sukses
		if resp.StatusCode == http.StatusOK {
			// Logika panggil x/bvm/backfill.go dimasukkan di sini
			// k.BackfillBlock(decodedBlock)
			fmt.Printf("📥 [SYNC] Blok #%d berhasil diunduh.\n", h)
		}
		resp.Body.Close()
	}

	// 3. Laporkan ke Event Bus Sultan
	events.EmitEvent("SYNC_COMPLETED", map[string]interface{}{
		"final_height": targetHeight,
		"source":       targetIP,
		"timestamp":    time.Now().Unix(),
	})
}

// BroadcastMessage: Mengirim kabar ke semua tetangga
func BroadcastMessage(msg string) {
	fmt.Printf("📢 [P2P Broadcast]: %s\n", msg)
	// Logika iterasi GlobalPeerManager.Peers dan kirim data akan ada di sini
}
