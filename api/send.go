package api

import (
	"github.com/aziskebanaran/BVM.core/x"
	"github.com/aziskebanaran/BVM.core/x/bvm/types"
	"encoding/json"
	"net/http"

	"github.com/vmihailenco/msgpack/v5"
)


func HandleSend(k x.BVMKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var tx types.Transaction
        var err error

        if r.Header.Get("Content-Type") == "application/x-msgpack" {
            err = msgpack.NewDecoder(r.Body).Decode(&tx)
        } else {
            err = json.NewDecoder(r.Body).Decode(&tx)
        }

        if err != nil {
            http.Error(w, "Gagal bongkar transaksi: "+err.Error(), http.StatusBadRequest)
            return
        }

        // 🚩 PERBAIKAN SULTAN: Jangan Timpa Timestamp & ID!
        // Biarkan Kernel yang memverifikasi ID aslinya.
        if tx.Symbol == "" { tx.Symbol = "BVM" }

        // Opsional: Hanya buat ID jika Wallet lupa mengirimnya
        if tx.ID == "" {
            tx.ID = tx.GenerateID()
        }

        // 🚀 3. EKSEKUSI LEWAT JENDERAL
        err = k.ProcessTransaction(tx)

        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{
                "status":  "Rejected",
                "message": err.Error(),
            })
            return
        }

        // 4. RESPON SUKSES
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "status":   "Pending",
            "message":  "Transaksi diterima dan masuk antrean Mempool",
            "tx_id":    tx.ID,
            "nonce":    tx.Nonce,
        })
    }
}


// --- FUNGSI 2: CEK ANTREAN (NONCE) ---
// 🚩 Taruh tepat di bawah HandleSend
func HandleNonce(k x.BVMKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        address := r.URL.Query().Get("address")
        if address == "" {
            http.Error(w, "Alamat Sultan tidak ditemukan dalam query", http.StatusBadRequest)
            return
        }

        // Tanya Jenderal (Keeper) berapa Nonce berikutnya di Disk/State
        nonce := k.GetNextNonce(address)

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]uint64{
            "nonce": nonce,
        })
    }
}
