package api

import (
	"github.com/aziskebanaran/BVM.core/x"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

var (
	activeValidators = make(map[string]int64)
	validatorMu      sync.Mutex
)

func HandleStats(k x.BVMKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")

        const BurnAddress = "bvmf000000000000000000000000000000000000burn"
        symbol := r.URL.Query().Get("symbol")
        if symbol == "" { symbol = "BVM" }

        nodeStatus := k.GetStatus()
        chain := k.GetChain()
        bank := k.GetBank()

        // 1. HITUNG SUPPLY (Sumber: GetAllBalances yang baru diperbaiki)
        allAccounts := bank.GetAllBalances()
        var actualSupply uint64
        for _, acc := range allAccounts {
            actualSupply += acc[symbol]
        }

        totalBurned := bank.GetBalance(BurnAddress, symbol)

        // 2. SINKRONISASI VALIDATOR (Gunakan data riil dari Staking Keeper)
        // Ambil dari Staking, bukan dari map heartbeat yang sering amnesia
        realValidators := k.GetStaking().QueryTopValidators(100)
        validatorCount := len(realValidators)

        // 3. ESTIMASI HASHRATE
        var hashRate float64
        if len(chain) > 10 {
            // Logika Sultan tetap dipertahankan
            timeDiff := chain[len(chain)-1].Timestamp - chain[len(chain)-10].Timestamp
            if timeDiff > 0 {
                hashRate = (float64(nodeStatus.Difficulty) * 10) / float64(timeDiff)
            }
        }

        // 4. SUSUN DASHBOARD
        stats := map[string]interface{}{
            "network": map[string]interface{}{
                "symbol":             symbol,
                "height":             nodeStatus.Height,
                "latest_hash":        nodeStatus.LatestHash,
                // 🚩 TIPS: Gunakan FormatDisplay jika Sultan ingin 24199.99
                // "circulating_supply": k.GetMining().FromAtomic(actualSupply),
                "circulating_supply": actualSupply, 
                "total_burned":       totalBurned,
                "difficulty":         nodeStatus.Difficulty,
                "hashrate_est":       hashRate,
                "validator_count":    validatorCount, // 🚩 Sekarang muncul 1 (Sultan)
            },
            "mempool": map[string]interface{}{
                "count":  len(k.GetPendingTransactions()),
                "status": "Active",
            },
            "timestamp": time.Now().Unix(),
        }

        json.NewEncoder(w).Encode(stats)
    }
}


func HandleStatus(k x.BVMKeeper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		minerAddr := r.URL.Query().Get("address")
		status := k.GetStatus()

		if minerAddr != "" {
			fmt.Printf("📡 Miner connected: %s | Height: %d\n", minerAddr[:12]+"...", status.Height)
		}

		json.NewEncoder(w).Encode(status)
	}
}

func GetActiveValidatorCount() int {
	validatorMu.Lock()
	defer validatorMu.Unlock()
	now := time.Now().Unix()
	count := 0
	for addr, lastSeen := range activeValidators {
		if now-lastSeen < 60 {
			count++
		} else {
			delete(activeValidators, addr)
		}
	}
	return count
}


func HandleValidators(k x.BVMKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")

        // Ambil langsung dari engine tanpa lewat keeper yang berbelit
        vals := k.GetStaking().QueryTopValidators(100)

        if vals == nil {
            // Jangan biarkan jadi null, kirim array kosong
            json.NewEncoder(w).Encode([]string{})
            return
        }
        json.NewEncoder(w).Encode(vals)
    }
}
