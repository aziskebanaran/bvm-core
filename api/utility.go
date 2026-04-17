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
)

// HandleSearchUser: Mencari data pengguna berdasarkan Alamat atau Username
func HandleSearchUser(k x.BVMKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        query := r.URL.Query().Get("q")

        if query == "" {
            http.Error(w, `{"error": "Parameter pencarian (q) diperlukan"}`, http.StatusBadRequest)
            return
        }

        var profile interface{}
        var err error

        // 🚩 LOGIKA DETEKSI: 
        // Jika dimulai dengan 'bvmf', cari berdasarkan Alamat.
        // Jika tidak, cari berdasarkan Username melalui AuthKeeper.
        if len(query) >= 4 && query[:4] == "bvmf" {
            // Cari data saldo/akun standar
            profile, _ = k.GetSecureBalance(query)
        } else {
            // Cari data Profil dari Registry yang kita buat di x/auth
            // Kita panggil GetProfile dari AuthKeeper Sultan
            profile, err = k.GetAuth().GetProfile(query)
            if err != nil {
                w.WriteHeader(http.StatusNotFound)
                json.NewEncoder(w).Encode(map[string]string{
                    "error": fmt.Sprintf("Username @%s tidak ditemukan", query),
                })
                return
            }
        }

        // Jika data kosong (untuk pencarian alamat)
        if profile == nil {
            w.WriteHeader(http.StatusNotFound)
            json.NewEncoder(w).Encode(map[string]string{
                "error": "Data tidak ditemukan",
            })
            return
        }

        json.NewEncoder(w).Encode(profile)
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
        var req struct {
            Username  string `json:"username"`
            Signature string `json:"signature"`
            Message   string `json:"message"` // Misal: "Login ke BVM Cloud pada [Timestamp]"
        }

        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Format JSON Salah", 400)
            return
        }

        // 1. Cari profil berdasarkan username
        profile, err := k.GetAuth().GetProfile(req.Username)
        if err != nil {
            http.Error(w, "User belum terdaftar", 404)
            return
        }

        // 2. VERIFIKASI TANDA TANGAN (Identity Check)
        // Kita gunakan logika Verifikasi yang sudah ada di AuthKeeper Sultan
        isValid := k.GetAuth().VerifyManualSignature(profile.Address, req.Message, req.Signature)
        if !isValid {
            http.Error(w, "Tanda tangan tidak sah! Anda bukan pemilik akun ini.", 401)
            return
        }

        // 3. Jika sah, berikan Token
        token, _ := k.GetAuth().GenerateUserToken(profile.Username, profile.Address)

        json.NewEncoder(w).Encode(map[string]string{
            "status": "LOGIN_SUCCESS",
            "token":  token,
            "type":   "Bearer",
        })
    }
}
