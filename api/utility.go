package api

import (
	"bvm.core/x"
	"bvm.core/x/bvm/types"
	"encoding/json"
	"fmt"
	"github.com/vmihailenco/msgpack/v5"
	"net/http"
	"os"
	"time"
)


// HandleSearchUser: Mencari data pengguna berdasarkan alamat dompet
func HandleSearchUser(k x.BVMKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        query := r.URL.Query().Get("q")

        if query == "" {
            http.Error(w, `{"error": "Parameter pencarian (q) diperlukan"}`, http.StatusBadRequest)
            return
        }

        // Menggunakan GetSecureBalance agar hasil pencarian sama akuratnya dengan pengecekan saldo
        accountData, found := k.GetSecureBalance(query)

        if !found {
            w.WriteHeader(http.StatusNotFound)
            json.NewEncoder(w).Encode(map[string]string{
                "error": "Alamat dompet tidak ditemukan dalam database",
            })
            return
        }

        json.NewEncoder(w).Encode(accountData)
    }
}

// HandleInspect: Sekarang berdiri sendiri (Hanya butuh string identity)
func HandleInspect(nodeAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity := nodeAddr
		if identity == "" {
			identity = "bvm_anonymous_node"
		}

		tx := types.Transaction{
			From:      identity,
			To:        "bvmf_sample_destination",
			Amount:    1.0,
			Timestamp: time.Now().Unix(),
		}

		jsonData, _ := json.Marshal(tx)
		binData, _ := msgpack.Marshal(tx)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"node_identity":  identity,
			"json_size_byte": len(jsonData),
			"bin_size_byte":  len(binData),
			"efficiency":     fmt.Sprintf("%d%% lebih hemat", 100-(len(binData)*100/len(jsonData))),
			"binary_hex":     fmt.Sprintf("%x", binData),
		})
	}
}

// GetLiveLocation tetap sama karena ini fungsi mandiri
func GetLiveLocation() string {
	cacheFile := "data/last_location.txt"
	if manualLoc := os.Getenv("BVM_LOCATION"); manualLoc != "" {
		return manualLoc
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://ip-api.com/json/")
	if err == nil {
		defer resp.Body.Close()
		var data struct {
			City    string `json:"city"`
			Region  string `json:"regionName"`
			Country string `json:"countryCode"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&data); err == nil {
			locString := fmt.Sprintf("%s, %s, %s", data.City, data.Region, data.Country)
			_ = os.WriteFile(cacheFile, []byte(locString), 0644)
			return locString
		}
	}

	lastKnown, err := os.ReadFile(cacheFile)
	if err == nil {
		return string(lastKnown) + " (Cached)"
	}
	return "Unknown Location (System Offline)"
}

