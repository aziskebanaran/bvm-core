package api

import (
    "encoding/json"
    "net/http"
    "time"
    "fmt"
    "bvm.core/x"
	"bvm.core/x/bvm/types"
)

func HandleJoinValidator(k x.BVMKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "Gunakan POST!", 405)
            return
        }

        var req struct {
            Address string `json:"address"`
            Amount  uint64 `json:"amount"`
        }
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Format data cacat!", 400)
            return
        }

        // 1. Rakit Transaksi Stake
        payload, _ := json.Marshal(map[string]interface{}{"amount": req.Amount})
        tx := types.Transaction{
            Type:      "stake",
            From:      req.Address,
            Amount:    req.Amount,
            Payload:   payload,
            Fee:       1000000,
            Nonce:     k.GetNextNonce(req.Address),
            Timestamp: time.Now().Unix(), // Sultan butuh ini untuk Hash yang unik
        }

        // 🚩 SOLUSI: Gunakan GenerateID() dan simpan ke field ID
        tx.ID = tx.GenerateID()

        // 2. KIRIM KE PROTOKOL
        if err := k.ProcessTransaction(tx); err != nil {
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{
            "status":  "STAKE_PENDING",
            "message": "Permintaan staking masuk antrean blok.",
            "tx_hash": tx.ID, // Gunakan tx.ID
        })
    }
}

func HandleUnstake(k x.BVMKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "Gunakan POST Sultan!", 405)
            return
        }

        var req struct {
            Address string `json:"address"`
            Amount  uint64 `json:"amount"`
        }
        json.NewDecoder(r.Body).Decode(&req)

        // 3. Rakit Transaksi Unstake
        payload, _ := json.Marshal(map[string]interface{}{"amount": req.Amount})
        tx := types.Transaction{
            Type:      "unstake",
            From:      req.Address,
            Amount:    req.Amount,
            Payload:   payload,
            Fee:       1000000,
            Nonce:     k.GetNextNonce(req.Address),
            Timestamp: time.Now().Unix(),
        }

        // 🚩 SOLUSI: Gunakan GenerateID() dan simpan ke field ID
        tx.ID = tx.GenerateID()

        if err := k.ProcessTransaction(tx); err != nil {
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "status":  "PENDING",
            "message": fmt.Sprintf("Penarikan %.4f BVM diproses!", k.FromAtomic(req.Amount)),
            "tx_hash": tx.ID,
        })
    }
}
