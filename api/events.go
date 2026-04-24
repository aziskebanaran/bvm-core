package api

import (
	"github.com/aziskebanaran/bvm-core/x"
	"bufio"
	"encoding/json"
	"net/http"
	"strings"
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

func HandleGetUpgrades(k x.BVMKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")

        type UpgradeStatus struct {
            Feature          string `json:"feature"`
            ActivationHeight uint64 `json:"activation_height"`
            IsActive         bool   `json:"is_active"`
            Source           string `json:"source"` // Tambahan info biar Sultan tahu asal usulnya
        }

        var results []UpgradeStatus
        currentHeight := uint64(k.GetLastHeight())

        // 1. 🔍 SCAN DATABASE SECARA DINAMIS
        // Cari semua yang berawalan "gov:upgrade:"
        dynamicData, err := k.GetStore().Scan("gov:upgrade:")
        
        if err == nil {
            for key, val := range dynamicData {
                // Bersihkan prefix untuk mendapatkan nama fiturnya (misal: GOV_V2)
                featureName := strings.TrimPrefix(key, "gov:upgrade:")
                
                // Pastikan konversi tipe data aman
                actHeight, _ := val.(uint64)

                results = append(results, UpgradeStatus{
                    Feature:          featureName,
                    ActivationHeight: actHeight,
                    IsActive:         currentHeight >= actHeight,
                    Source:           "WASM_Governance",
                })
            }
        }

        // 2. Kirim ke Sultan
        if len(results) == 0 {
            // Jika kosong, kirim array kosong bukan null
            json.NewEncoder(w).Encode([]UpgradeStatus{})
            return
        }
        
        json.NewEncoder(w).Encode(results)
    }
}
