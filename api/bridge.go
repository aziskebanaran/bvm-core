package api

import (
    "encoding/json"
    "net/http"
    "github.com/aziskebanaran/bvm-core/x"
    "github.com/aziskebanaran/bvm-core/x/bvm/types"
)

// PINTU KEBERANGKATAN (LOCK ASET DI VAULT)
func HandleBridgeOut(k x.BVMKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var tx types.Transaction
        if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
            http.Error(w, "Format transaksi bridge rusak", http.StatusBadRequest)
            return
        }

        // 1. Validasi Tipe Transaksi
        if tx.Type != "bridge_out" {
            http.Error(w, "Bukan transaksi Bridge Out", http.StatusBadRequest)
            return
        }

        // 🚀 AUTO-INJECT PASPOR NEGARA (Mencegah salah dimensi)
        if tx.ChainID == 0 {
            tx.ChainID = k.GetParamsData().GetChainID()
        }

        // 2. Lempar ke Jenderal untuk diproses
        err := k.ProcessTransaction(tx)
        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            // 🚩 PERBAIKAN: Masukkan err.Error() langsung ke dalam "message"
            json.NewEncoder(w).Encode(map[string]string{
                "message": "DITOLAK: " + err.Error(), 
            })
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{
            "status": "success",
            "message": "Aset sedang dikunci. Menunggu Relayer.",
            "tx_id": tx.ID,
        })
    }
}

// PINTU KEDATANGAN & VALIDASI (Menerima bridge_in DAN submit_proof)
func HandleBridgeIn(k x.BVMKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var tx types.Transaction
        if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
            http.Error(w, "Format transaksi bridge rusak", http.StatusBadRequest)
            return
        }

        // 🚩 PERBAIKAN: Izinkan "bridge_in" DAN "submit_proof" masuk ke sini
        if tx.Type != "bridge_in" && tx.Type != "submit_proof" {
            http.Error(w, "Tipe transaksi bridge tidak dikenal", http.StatusBadRequest)
            return
        }

        if tx.ChainID == 0 {
            tx.ChainID = k.GetParamsData().GetChainID()
        }

        // Lempar ke Jenderal untuk diproses (Keeper sudah punya logic switch case-nya)
        err := k.ProcessTransaction(tx)
        if err != nil {
            http.Error(w, "Gagal diproses: "+err.Error(), http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{
            "status": "success",
            "message": "Transaksi bridge diterima.",
            "tx_id": tx.ID,
        })
    }
}
