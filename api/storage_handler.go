package api

import (
	"encoding/json"
	"fmt"                // 🚩 Tambahkan ini
	"io"                 // 🚩 Tambahkan ini
	"net/http"
	"os"
        "path/filepath"
        "time"
	"github.com/aziskebanaran/bvm-core/x"
	"github.com/aziskebanaran/bvm-core/x/storage/keeper"
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

// ... HandleAppRegister tetap seperti sebelumnya ...

func HandleAppPut(sk *keeper.StorageKeeper, k x.BVMKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 1. Parsing Multipart
        if err := r.ParseMultipartForm(100 << 20); err != nil {
            http.Error(w, "File terlalu besar", 400)
            return
        }

        file, header, err := r.FormFile("file")
        if err != nil {
            http.Error(w, "Paket file tidak ditemukan", 400)
            return
        }
        defer file.Close()

        owner := r.FormValue("owner")
        appID := r.FormValue("app_id") // 🚩 Sekarang kita gunakan di bawah

        // 2. Billing
        fileSize := header.Size
        burnAmount, err := sk.ProcessAutoBilling(owner, int(fileSize), k)
        if err != nil {
            http.Error(w, "Saldo tidak cukup", 402)
            return
        }

        // 3. Penyimpanan Fisik
        // Kita masukkan appID ke dalam nama file agar rapi
        storageID := fmt.Sprintf("%s-%d-%s", appID, time.Now().Unix(), header.Filename)
        savePath := filepath.Join("data", "apps_storage", storageID)

        dst, err := os.Create(savePath)
        if err != nil {
            http.Error(w, "Gagal membuat gudang data", 500)
            return
        }
        defer dst.Close()

        io.Copy(dst, file)

        json.NewEncoder(w).Encode(map[string]interface{}{
            "status":     "success",
            "storage_id": storageID,
            "burned":     burnAmount,
        })
    }
}

func HandleAppGet(sk *keeper.StorageKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        storageID := r.URL.Query().Get("id")
        if storageID == "" {
            http.Error(w, "ID diperlukan", 400)
            return
        }

        // Jalur ke data/apps_storage
        filePath := filepath.Join("data", "apps_storage", storageID)

        if _, err := os.Stat(filePath); os.IsNotExist(err) {
            http.Error(w, "File tidak ditemukan", 404)
            return
        }

        w.Header().Set("Content-Type", "application/zip")
        http.ServeFile(w, r, filePath)
    }
}


// Helper untuk ambil UserID dari context secara ringkas
func getUserID(r *http.Request) string {
    if uid, ok := r.Context().Value("user_id").(string); ok {
        return uid
    }
    return ""
}

