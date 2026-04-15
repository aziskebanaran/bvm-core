package main

import (
	"encoding/json"
	"fmt"
	"log"

	"bvm.core/pkg/storage"
	staketypes "bvm.core/x/staking/types"
)

func main() {
	// 1. Koneksi ke Database
	db, err := storage.NewLevelDBStore("data/blockchain_db") 
	if err != nil {
		log.Fatalf("🚨 Gagal: %v", err)
	}
	defer db.Close() 

	// 2. Siapkan Profil Sultan sebagai Validator
	sultanAddr := "bvmfa15608e2b0225f96b915"
	valData := staketypes.Validator{
		Address:      sultanAddr,
		PubKey:       "ed25519-sultan-key-secure",
		StakedAmount: 1000000000, // 1000 BVM
		Power:        1000,
		Status:       "ACTIVE",
		IsActive:     true,
	}

	// 3. Pahat ke DB dengan KEY yang dikenali StakingEngine Sultan
	// Biasanya StakingEngine menggunakan prefix "s:acc:" atau "stake:"
	// Berdasarkan pola Sultan, mari kita coba "s:acc:" + address
	data, _ := json.Marshal(valData)
	key := "s:acc:" + sultanAddr 

	err = db.Put(key, data)
	if err != nil {
		fmt.Printf("❌ Gagal Injeksi: %v\n", err)
	} else {
		// 🚩 TAMBAHKAN INI: Sultan butuh INDEX agar Engine tahu siapa saja yang harus di-load
		// Kita simpan alamat Sultan ke daftar index validator
		indexKey := "s:index"
		db.Put(indexKey, []byte(sultanAddr))

		fmt.Println("✅ INJEKSI SUKSES: Data Validator Sultan telah dipahat!")
		fmt.Println("📍 Address:", sultanAddr)
	}
}
