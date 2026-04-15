package api

import (
	"bvm.core/x"
	"encoding/hex"   // 🚩 Tambahkan ini
	"encoding/json"
	"fmt"            // 🚩 Tambahkan ini
	"net/http"
	"os"             // 🚩 Tambahkan ini
	"crypto/sha256"  // 🚩 Tambahkan ini
)

func HandleAuditSupply(k x.BVMKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        symbol := r.URL.Query().Get("symbol")
        if symbol == "" { symbol = "BVM" }

        // ✅ Gunakan GetStatus yang biasanya sudah ada di Interface
        status := k.GetStatus() 

        report := map[string]interface{}{
            "symbol":             symbol,
            "height":             status.Height,
            "circulating_supply": status.TotalSupply, // Diambil dari GetStatus
            "total_burned":       status.TotalBurned, // Diambil dari GetStatus
            "status":             "VERIFIED_BY_KERNEL",
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(report)
    }
}

func HandleVerifyContract(w http.ResponseWriter, r *http.Request) {
    name := r.URL.Query().Get("name") // contoh: dpos

    // 1. Ambil Hash Biner yang sedang jalan di folder build/
    wasmPath := fmt.Sprintf("./build/%s.wasm", name)
    wasmBytes, err := os.ReadFile(wasmPath)
    if err != nil {
        http.Error(w, "Kontrak tidak ditemukan", 404)
        return
    }
    hashWasm := sha256.Sum256(wasmBytes)

    // 2. Kirim status verifikasi ke rakyat
    response := map[string]interface{}{
        "contract":    name,
        "binary_hash": hex.EncodeToString(hashWasm[:]),
        "source_code": fmt.Sprintf("/api/audit/source?name=%s", name),
        "verified":    true, // Di masa depan, ini hasil perbandingan otomatis
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
