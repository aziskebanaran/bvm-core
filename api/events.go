package api

import (
	"github.com/aziskebanaran/BVM.core/x"
	"bufio"
	"encoding/json"
	"net/http"
	"os"
)

func HandleGetEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 1. Buka file events.log
	file, err := os.Open("events.log")
	if err != nil {
		// Jika file belum ada, kirim array kosong []
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}
	defer file.Close()

	// 2. Baca baris demi baris
	var allEvents []interface{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var ev interface{}
		if err := json.Unmarshal([]byte(scanner.Text()), &ev); err == nil {
			allEvents = append(allEvents, ev)
		}
	}

	// 3. Ambil 10 kejadian terakhir saja agar ringan
	count := len(allEvents)
	start := 0
	if count > 10 {
		start = count - 10
	}

	json.NewEncoder(w).Encode(allEvents[start:])
}

// HandleGetUpgrades: Kita ubah agar mengembalikan http.HandlerFunc
// Dan menggunakan interface x.BVMKeeper agar seragam dengan router.go
func HandleGetUpgrades(k x.BVMKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")

        // Daftar fitur yang ingin Sultan pantau
        features := []string{"WASM_ENGINE", "STAKING_V2", "BURN_FEE_V2", "TOKEN_FACTORY"}

        type UpgradeStatus struct {
            Feature          string `json:"feature"`
            ActivationHeight int64  `json:"activation_height"`
            IsActive         bool   `json:"is_active"`
        }

        var results []UpgradeStatus
        currentHeight := k.GetLastHeight()

        for _, f := range features {
            var actHeight int64
            key := "gov:upgrade:" + f
            err := k.GetStore().Get(key, &actHeight)

            if err == nil {
                results = append(results, UpgradeStatus{
                    Feature:          f,
                    ActivationHeight: actHeight,
                    IsActive:         int64(currentHeight) >= actHeight,
                })
            } else {
                // 🚩 Tambahkan status "Pending" jika data belum ada di DB
                results = append(results, UpgradeStatus{
                    Feature:          f,
                    ActivationHeight: 0, // Belum dijadwalkan
                    IsActive:         false,
                })
            }
        }

        json.NewEncoder(w).Encode(results)
    }
}
