package api

import (
	"github.com/aziskebanaran/bvm-core/x"
	"github.com/aziskebanaran/bvm-core/x/bvm/types"
	"encoding/json"
	"fmt"
	"github.com/vmihailenco/msgpack/v5"
	"net/http"
	"os"
	"time"
"strings"
)

func HandleSearchUser(k x.BVMKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")

        // 🚩 DETEKSI AGRESIF: Coba ambil dari parameter 'q', 'username', atau 'address'
        rawQuery := r.URL.Query().Get("q")
        if rawQuery == "" { rawQuery = r.URL.Query().Get("username") }
        if rawQuery == "" { rawQuery = r.URL.Query().Get("address") }

        query := strings.ToLower(strings.TrimSpace(rawQuery))
        query = strings.TrimPrefix(query, "@") 

        // 🚩 LOG RADAR: CEK DI TAB NODE (TAB 1) SAAT SULTAN SEARCH!
        fmt.Printf("\n--- [RADAR SCAN] ---\n")
        fmt.Printf("Raw Query: '%s'\n", rawQuery)
        fmt.Printf("Clean Query: '%s'\n", query)
        fmt.Printf("Full URL: %s\n", r.URL.String())
        fmt.Printf("--------------------\n")

        if query == "" {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]string{"error": "Server menerima query kosong dari CLI"})
            return
        }

        var targetAddress string
        profile, err := k.GetAuth().GetProfile(query)

        if err != nil || profile.Address == "" {
            if strings.HasPrefix(query, "bvmf") {
                targetAddress = query
            } else {
                w.WriteHeader(http.StatusNotFound)
                json.NewEncoder(w).Encode(map[string]string{
                    "error": fmt.Sprintf("Username '%s' tidak terdaftar di database", query),
                })
                return
            }
        } else {
            targetAddress = profile.Address
        }

        finalState, ok := k.GetSecureBalance(targetAddress)
        if !ok {
            w.WriteHeader(http.StatusNotFound)
            json.NewEncoder(w).Encode(map[string]string{"error": "Wallet ada tapi saldo gagal dimuat"})
            return
        }

        json.NewEncoder(w).Encode(finalState)
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

func HandleLogin(k x.BVMKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")

        var req struct {
            Username  string `json:"username"`
            Signature string `json:"signature"`
            Message   string `json:"message"`
        }

        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]string{"status": "ERROR", "error": "Format JSON Request Salah"})
            return
        }

        // 🚩 PERBAIKAN: Gunakan req.Username, bukan query!
        loginQuery := strings.ToLower(strings.TrimSpace(req.Username))

        // 1. Cari profil berdasarkan username yang mau login
        profile, err := k.GetAuth().GetProfile(loginQuery)

        // Debug log untuk memantau siapa yang mencoba login
        fmt.Printf("[LOGIN-DEBUG] User: %s | Found Addr: %s | Err: %v\n", loginQuery, profile.Address, err)

        if err != nil || profile.Address == "" {
            w.WriteHeader(http.StatusNotFound)
            json.NewEncoder(w).Encode(map[string]string{
                "status": "ERROR", 
                "error":  "User @" + req.Username + " belum terdaftar di blok!",
            })
            return
        }

        // 2. Verifikasi Tanda Tangan
        isValid := k.GetAuth().VerifyManualSignature(profile.Address, req.Message, req.Signature)
        if !isValid {
            w.WriteHeader(http.StatusUnauthorized)
            json.NewEncoder(w).Encode(map[string]string{"status": "ERROR", "error": "Tanda tangan digital tidak sah!"})
            return
        }

        // 3. Generate Token
        token, err := k.GetAuth().GenerateUserToken(profile.Username, profile.Address)
        if err != nil || token == "" {
            w.WriteHeader(http.StatusInternalServerError)
            json.NewEncoder(w).Encode(map[string]string{"status": "ERROR", "error": "Gagal membuat sesi token"})
            return
        }

        // 4. RESPON SUKSES
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{
            "status": "LOGIN_SUCCESS",
            "token":  token,
            "type":   "Bearer",
        })
    }
}
