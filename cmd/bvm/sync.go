package main

import (
	"fmt"
	"time"

	"github.com/aziskebanaran/bvm-core/pkg/client"
	"github.com/aziskebanaran/bvm-core/pkg/storage"
	"github.com/aziskebanaran/bvm-core/pkg/logger"
)

// RunSyncEngine adalah "Mata" Core untuk Nexus
func RunSyncEngine(nexusURL string, store storage.BVMStore) {
	fmt.Println("🛰️ [CORE-SYNC] Engine Sinkronisasi Agresif Aktif...")
	nexusClient := client.NewBVMClient(nexusURL)

	for {
		// 1. Tanya kondisi Nexus
		info, err := nexusClient.GetNetworkInfo()
		if err != nil {
			logger.Error("SYNC", "Nexus tidak merespon, menunggu...")
			time.Sleep(10 * time.Second)
			continue
		}

		// 2. Cek tinggi blok lokal
		var localHeight uint64
		store.Get("m:height", &localHeight)

		// 3. Jika tertinggal, eksekusi FastSync
		if uint64(info.Height) > localHeight {
			fmt.Printf("📥 [SYNC] Nexus di #%d | Lokal di #%d. Menyedot data...\n", info.Height, localHeight)

			err := nexusClient.FastSync(localHeight, uint64(info.Height), store)
			if err != nil {
				logger.Error("SYNC", fmt.Sprintf("Gagal menyedot data: %v", err))
			} else {
				fmt.Println("✅ [SYNC] Core sudah sinkron dengan Nexus!")
			}
		}

		// Ritme detak jantung (sesuaikan dengan kecepatan network Anda)
		time.Sleep(5 * time.Second)
	}
}
