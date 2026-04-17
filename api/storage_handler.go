package api

import (
	"encoding/json"
	"fmt"                // 🚩 Tambahkan ini
	"io"                 // 🚩 Tambahkan ini
	"net/http"
	"github.com/aziskebanaran/BVM.core/x"
	"github.com/aziskebanaran/BVM.core/x/storage/keeper"
	"github.com/aziskebanaran/BVM.core/x/storage/types"
)


// HandleAppRegister: Tempat pengembang mendaftar & membayar untuk API Key
func HandleAppRegister(sk *keeper.StorageKeeper, k x.BVMKeeper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			AppID string                 `json:"app_id"`
			Owner string                 `json:"owner"`
			Rules map[string]interface{} `json:"rules"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Format JSON Salah", 400)
			return
		}

		// 🚩 LOGIKA BISNIS SULTAN: Pajak Pendaftaran (Misal: 10 BVM)
		// Kita potong saldo si Owner sebelum memberikan Key
		regCost := k.ToAtomic(10.0)
		if k.GetBalanceBVM(req.Owner) < regCost {
			http.Error(w, "Saldo BVM Tidak Cukup untuk Biaya Pendaftaran (10 BVM)", 402)
			return
		}

		// Eksekusi Pemotongan (Burning/Transfer ke Treasury)
		// k.SubBalanceBVM(req.Owner, regCost, nil) 

		// Daftarkan di Cloud Storage
		apiKey, err := sk.RegisterApp(req.Owner, req.AppID, req.Rules)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "SUCCESS",
			"app_id":  req.AppID,
			"api_key": apiKey,
			"message": "Simpan API Key Anda baik-baik!",
		})
	}
}

func HandleAppPut(sk *keeper.StorageKeeper, k x.BVMKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 1. Ambil data dari context dengan Safe Assertion 🛡️
        rawApp   := r.Context().Value("app_metadata")
        rawUser  := r.Context().Value("user_id")
        // app_owner biasanya ada di dalam app_metadata (AppContainer)

        if rawApp == nil || rawUser == nil {
            http.Error(w, `{"error": "Sesi tidak valid, silakan login ulang"}`, 401)
            return
        }

        app, ok1 := rawApp.(types.AppContainer)
        userID, ok2 := rawUser.(string)
        if !ok1 || !ok2 {
            http.Error(w, `{"error": "Gagal memproses otentikasi cloud"}`, 500)
            return
        }

        // 2. Baca Body (Gunakan LimitReader agar tidak terkena serangan Flood)
        bodyBytes, err := io.ReadAll(r.Body)
        if err != nil {
            http.Error(w, "Gagal membaca data", 400)
            return
        }

        var req struct {
            Key   string `json:"key"`
            Value string `json:"value"`
        }
        if err := json.Unmarshal(bodyBytes, &req); err != nil {
            http.Error(w, "Format JSON salah", 400)
            return
        }

        // 3. Namespacing: AppID:UserID:Key
        finalKey := fmt.Sprintf("%s:%s:%s", app.AppID, userID, req.Key)

        // 4. Proses Billing (Bakar saldo Owner Aplikasi) 🔥
        // Kita gunakan app.Owner yang didapat dari database metadata
        burnAmount, err := sk.ProcessAutoBilling(app.Owner, len(bodyBytes), k)
        if err != nil {
            http.Error(w, "Saldo Owner App Tidak Cukup untuk Biaya Simpan (Burn)", 402)
            return
        }

        // 5. Pahat ke Disk melalui SafePut
        userData := types.UserData{
            AppID: app.AppID,
            Key:   finalKey, 
            Value: req.Value,
        }

        // Gunakan owner aplikasi sebagai pemegang otoritas SafePut
        err = sk.SafePut(app.AppID, userData, app.Owner)
        if err != nil {
            http.Error(w, "BVM-GUARD: "+err.Error(), 403)
            return
        }

        // 6. Response Sukses
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "status": "Data Abadi di Cloud",
            "path":   finalKey,
            "burned": burnAmount,
        })
    }
}


// Helper untuk ambil UserID dari context secara ringkas
func getUserID(r *http.Request) string {
    if uid, ok := r.Context().Value("user_id").(string); ok {
        return uid
    }
    return ""
}

// HandleAppGet: Versi Ramping & Sinkron 🚀
func HandleAppGet(sk *keeper.StorageKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 1. Ambil Metadata & UserID
        rawApp, _ := r.Context().Value("app_metadata").(types.AppContainer)
        userID := getUserID(r)
        key := r.URL.Query().Get("key")

        if rawApp.AppID == "" || userID == "" || key == "" {
            http.Error(w, "Parameter atau Sesi tidak lengkap", 400)
            return
        }

        // 2. 🚩 SINKRONISASI: Gunakan format yang SAMA dengan HandleAppPut
        finalKey := fmt.Sprintf("%s:%s:%s", rawApp.AppID, userID, key)

        // 3. Eksekusi Ambil Data
        appDB, err := sk.GetAppStore(rawApp.AppID)
        if err != nil {
            http.Error(w, "DB Error", 500); return
        }
        defer appDB.Close()

        var val string
        if err := appDB.Get(finalKey, &val); err != nil {
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(map[string]string{
                "key": key, "value": "", "status": "NOT_FOUND", "path": finalKey,
            })
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{
            "key": key, "value": val, "status": "SUCCESS",
        })
    }
}
