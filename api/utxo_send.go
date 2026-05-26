package api

import (
	"fmt"
    "encoding/json"
    "net/http"
    "github.com/aziskebanaran/bvm-core/x"
    "github.com/aziskebanaran/bvm-core/x/bvm/types"
    "github.com/aziskebanaran/bvm-core/pkg/logger" // Pastikan logger diimpor
)

// HandleUTXOSend: Endpoint Khusus Jalur VIP UTXO
func HandleUTXOSend(k x.BVMKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var tx types.Transaction

        // 1. Decode paket kiriman dari Gateway
        if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]string{"error": "Format transaksi UTXO rusak"})
            return
        }

        // 🚀 BARIKADE POS VIP UTXO:
        // Tolak mentah-mentah di level pintu API jika transaksi UTXO membawa Paspor ChainID yang salah!
        if tx.ChainID > 0 && tx.ChainID != k.GetParamsData().GetChainID() {
            logger.Error("API", fmt.Sprintf("❌ UTXO Rejected: Peretas mencoba melakukan Replay Attack dari ChainID: %d", tx.ChainID))
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{
                "status": "FAILED",
                "error":  "🚨 ILLEGAL CROSS-CHAIN REPLAY: Transaksi ini milik jaringan lain!",
            })
            return
        }

        // 🚩 2. EKSEKUSI LEWAT MESIN KHUSUS UTXO
        // Ini akan memanggil k.ProcessUTXOTransaction(tx) di keeper.go
        err := k.ProcessUTXOTransaction(tx) 

        if err != nil {
            logger.Error("API", "❌ UTXO Rejected: "+err.Error())
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{
                "status": "FAILED", 
                "error": err.Error(),
            })
            return
        }

        // 3. RESPON SUKSES JALUR VIP
        logger.Success("API", "📥 UTXO Move Diterima: "+tx.ID[:10])
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "status":  "SUCCESS", 
            "tx_id":   tx.ID,
            "message": "Kepingan aset masuk antrean blok Jalur VIP",
        })
    }
}
