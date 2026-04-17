package api

import (
	"github.com/aziskebanaran/BVM.core/x" // 🚩 Gunakan Interface Pusat
	"github.com/aziskebanaran/BVM.core/x/bvm/types"
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
