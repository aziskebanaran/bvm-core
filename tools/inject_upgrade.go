package main

import (
	"fmt"
	"github.com/aziskebanaran/BVM.core/pkg/storage"
)

func main() {
	// 1. Koneksi ke Database (Node Wajib OFF)
	store, err := storage.NewLevelDBStore("./data/blockchain_db")
	if err != nil {
		fmt.Printf("❌ Gagal membuka DB: %v\n", err)
		return
	}

	// 2. Daftar Fitur yang Akan Diaktifkan
	upgrades := map[string]uint64{
		"WASM_ENGINE":   1800, // Aktif di blok 1800
		"TOKEN_FACTORY": 1850, // Aktif sedikit lebih lambat, di blok 1850
		"STAKING_V2": 2300,
	}

	// 3. Proses Suntikan Massal
	for feature, height := range upgrades {
		key := "gov:upgrade:" + feature
		err = store.Put(key, height)
		if err != nil {
			fmt.Printf("❌ Gagal menyuntik %s: %v\n", feature, err)
			continue
		}
		fmt.Printf("✅ %s dijadwalkan pada blok #%d\n", feature, height)
	}

	fmt.Println("\n🚀 Semua jadwal upgrade telah diperbarui dalam Konstitusi BVM.")
}
