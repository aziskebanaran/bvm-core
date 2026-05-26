package api

import (
	"github.com/aziskebanaran/bvm-core/x" // 🚩 Gunakan Interface Pusat
	"github.com/aziskebanaran/bvm-core/x/bvm/types"
	"encoding/json"
	"net/http"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

func HandleMempool(k x.BVMKeeper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Tarik data langsung dari Jenderal (Mempool RAM)
		pendingTxs := k.GetMempool().GetPendingTransactions()

		if pendingTxs == nil {
			pendingTxs = []types.Transaction{}
		}

		// Support Msgpack (VVIP Mode)
		if r.Header.Get("Accept") == "application/x-msgpack" {
			w.Header().Set("Content-Type", "application/x-msgpack")
			msgpack.NewEncoder(w).Encode(pendingTxs)
			return
		}

		// Support JSON (CLI/Web Mode)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"count":   len(pendingTxs),
			"txs":     pendingTxs,
			"height":  k.GetLastHeight(),
			"status":  "Synced",
		})
	}
}


// HandleMempoolStats: Statistik antrean untuk Dashboard Sultan
func HandleMempoolStats(k x.BVMKeeper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🚩 PERBAIKAN: Ambil menteri mempool melalui Jenderal
		mp := k.GetMempool()
		txs := mp.GetPendingTransactions()

		var totalFee uint64
		for _, tx := range txs {
			totalFee += tx.Fee
		}

		stats := map[string]interface{}{
			"count":     len(txs),
			"total_fee": totalFee,
			"is_busy":   len(txs) > 100,
			"timestamp": time.Now().Unix(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}
}

// HandleMempoolPing: Loket khusus menerima sinyal detak jantung (ping) External Miner
// POST /api/mempool/ping
func HandleMempoolPing(k x.BVMKeeper) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
                if r.Method != http.MethodPost {
                        http.Error(w, "❌ Method not allowed", http.StatusMethodNotAllowed)
                        return
                }

                var req struct {
                        Address string `json:"address"`
                }

                if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
                        http.Error(w, "🚨 Bad request payload", http.StatusBadRequest)
                        return
                }

                if req.Address == "" {
                        http.Error(w, "⚠️ Address is required", http.StatusBadRequest)
                        return
                }

                store := k.GetStore()
                batch := store.NewBatch()
                timestampNow := time.Now().Unix()

                // 1. Pahat Sejarah Abadi Ke Disk LevelDB
                store.PutToBatch(batch, "mempool_active:"+req.Address, timestampNow)
                if err := store.WriteBatch(batch); err != nil {
                        http.Error(w, "🚨 Internal server error", http.StatusInternalServerError)
                        return
                }

                // 2. 🚀 MASUKKAN LANGSUNG KE KANTUNG RAM VIA INTERFACE RESMI SULTAN
                k.UpdateActiveMiner(req.Address, timestampNow)

                w.Header().Set("Content-Type", "application/json")
                json.NewEncoder(w).Encode(map[string]string{
                        "status":  "success",
                        "message": "💓 Heartbeat acknowledged by Core L1 Jenderal",
                })
        }
}
