package main

import (
    "fmt"
    "log"

    "github.com/aziskebanaran/bvm-core/pkg/storage"
    staketypes "github.com/aziskebanaran/bvm-core/x/staking/types"
    "github.com/vmihailenco/msgpack/v5" // 🚩 Gunakan MsgPack sesuai Engine Sultan!
)

func main() {
    // 1. Koneksi ke Database dengan Cache 64MB
    db, err := storage.NewLevelDBStore("data/blockchain_db", 64)
    if err != nil {
        log.Fatalf("🚨 Gagal Akses DB: %v", err)
    }
    defer db.Close()

    // 2. Gunakan Alamat Mnemonic Baru Sultan
    sultanAddr := "bvmfcfcadd6f4dd0cd16f1cb" 

    valData := staketypes.Validator{
        Address:      sultanAddr,
        PubKey:       "ed25519-sultan-key-secure",
        StakedAmount: 100000000000, // 100.000 BVM
        Power:        100000,
        Status:       "ACTIVE",
        IsActive:     true,
    }

    // 3. Pahat menggunakan MsgPack (Sesuai storage/engine.go Sultan)
    data, err := msgpack.Marshal(valData)
    if err != nil {
        log.Fatalf("❌ Gagal Marshal: %v", err)
    }

    // Key sesuai standar Staking Sultan
    key := "s:acc:" + sultanAddr

    // Gunakan db.GetDB().Put karena kita sudah punya data bytes hasil msgpack
    err = db.GetDB().Put([]byte(key), data, nil)
    if err != nil {
        fmt.Printf("❌ Gagal Injeksi: %v\n", err)
    } else {
        // 4. Update Index agar Validator di-load saat startup
        indexKey := "s:index"
        db.GetDB().Put([]byte(indexKey), []byte(sultanAddr), nil)

        fmt.Println("---------------------------------------")
        fmt.Println("✅ INJEKSI SUKSES (MSGPACK FORMAT)")
        fmt.Printf("📍 Validator Baru: %s\n", sultanAddr)
        fmt.Println("---------------------------------------")
    }
}
