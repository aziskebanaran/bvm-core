package api

import (
	"encoding/json"
	"fmt"                // 🚩 Tambahkan ini
	"io"                 // 🚩 Tambahkan ini
	"net/http"
	"bvm.core/pkg/logger" // 🚩 Tambahkan ini
	"bvm.core/x"
	"bvm.core/x/storage/keeper"
	"bvm.core/x/storage/types"
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
		// 1. Ambil Identitas Aplikasi
		app, ok := r.Context().Value("app_metadata").(types.AppContainer)
		if !ok {
			http.Error(w, "Unauthorized", 401)
			return
		}

		// 2. Baca Seluruh Body untuk hitung ukuran (Billing Dasar)
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Gagal membaca data", 400)
			return
		}

		// 🚩 3. EKSEKUSI PEMBAKARAN (100% BURN)
		// Fungsi ini memotong saldo Owner & mengirimnya ke Burn Address
		burnAmount, err := sk.ProcessAutoBilling(app.Owner, len(bodyBytes), k)
		if err != nil {
			http.Error(w, "Billing Gagal: "+err.Error(), 402)
			return
		}

		// 4. Parsing Data JSON
		var data types.UserData
		if err := json.Unmarshal(bodyBytes, &data); err != nil {
			// Jika format JSON salah, kita sudah terlanjur bakar fee? 
			// Secara teknis di blockchain, fee tidak kembali jika transaksi salah format.
			http.Error(w, "Format JSON Salah", 400)
			return
		}
		data.AppID = app.AppID

		// 5. Simpan ke Storage (SafePut dengan Rules Engine)
		err = sk.SafePut(app.AppID, data, app.Owner)
		if err != nil {
			// Opsional: Jika gagal tulis ke disk, Sultan bisa refund di sini
			// k.AddBalanceBVM(app.Owner, burnAmount, nil)
			http.Error(w, "Storage Error: "+err.Error(), 500)
			return
		}

		// 🚩 6. LOG EKONOMI KE TERMINAL SULTAN
		logger.Error("ECONOMY", fmt.Sprintf("🔥 BURNED: %d Atomic BVM dimusnahkan dari aktivitas App [%s]", burnAmount, app.AppID))

		// 7. Respon ke Klien
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":      "Data Tersimpan",
			"fee_burned":  burnAmount,
			"size_bytes":  len(bodyBytes),
			"burning_tag": "DEFLATIONARY_STORAGE_V1",
		})
	}
}

// HandleAppGet: Mengambil data dari gudang mandiri
func HandleAppGet(sk *keeper.StorageKeeper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		app := r.Context().Value("app_metadata").(types.AppContainer)
		key := r.URL.Query().Get("key")

		if key == "" {
			http.Error(w, "Key diperlukan", 400)
			return
		}

		// Sultan bisa modifikasi CheckRules untuk aksi "read" di sini jika perlu
		appDB, err := sk.GetAppStore(app.AppID)
		if err != nil {
			http.Error(w, "Gagal membuka database", 500)
			return
		}
		defer appDB.Close()

		var val string
		if err := appDB.Get(key, &val); err != nil {
			http.Error(w, "Data tidak ditemukan", 404)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"key": key, "value": val})
	}
}
