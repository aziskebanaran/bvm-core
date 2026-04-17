package api

import (
        "github.com/aziskebanaran/bvm-core/x"
        "encoding/json"
        "net/http"
	"fmt"
	"strings"
	"github.com/aziskebanaran/bvm-core/pkg/logger"
)

func HandleParams(k x.BVMKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 1. Ambil Data Parameter Jaringan
        paramsManager := k.GetParams()
        paramsData := paramsManager.GetParamsData()

        // 2. Ambil Status Terkini & Ukuran Mempool
        status := k.GetStatus()
        pendingTxs := k.GetPendingTransactions()
        mempoolSize := len(pendingTxs)

        // 3. Susun Respon JSON
        response := map[string]interface{}{
            "params":       paramsData,
            "height":       status.Height,
            "latest_hash":  status.LatestHash,
            "difficulty":   status.Difficulty,
            "mempool_size": mempoolSize,
            "dynamic_fee":  paramsManager.GetDynamicFee(mempoolSize),
            "network_name": "BVM-Mainnet-Beta",
            "status":       "Operational",
        }

        // 4. Kirim Respon
        w.Header().Set("Content-Type", "application/json")
        if err := json.NewEncoder(w).Encode(response); err != nil {
            w.WriteHeader(http.StatusInternalServerError)
        }
    }
}

func HandleUpdateConstitution(k x.BVMKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Keamanan: Hanya terima perintah dari localhost (Sultan sendiri)
        if !strings.HasPrefix(r.RemoteAddr, "127.0.0.1") && !strings.HasPrefix(r.RemoteAddr, "[::1]") {
            http.Error(w, "Akses Ilegal! Hanya Sultan yang boleh ubah Konstitusi.", 403)
            return
        }

        var req struct {
            Feature string `json:"feature"`
            Height  uint64 `json:"height"`
        }
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Format JSON salah", 400)
            return
        }

        // 🚩 SUNTIKAN PANAS (Hot-Inject)
        // Kita gunakan k.GetStore() yang sudah terbuka dan aman dari lock
        key := "gov:upgrade:" + req.Feature
        err := k.GetStore().Put(key, req.Height)

        if err != nil {
            http.Error(w, "Gagal update: "+err.Error(), 500)
            return
        }

        logger.Success("GOV", fmt.Sprintf("🛡️ Aturan Baru: %s aktif di blok #%d", req.Feature, req.Height))

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{"status": "SUCCESS", "message": "Konstitusi diperbarui tanpa restart!"})
    }
}
