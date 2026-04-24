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
        // --- 🚩 JALUR 1: POST (MENGUBAH ATURAN / KONSTITUSI) ---
        if r.Method == http.MethodPost {
            HandleUpdateConstitution(k)(w, r)
            return
        }

        // --- 🚩 JALUR 2: GET (MELIHAT STATUS & PARAMETER) ---
        // 1. Ambil Data dari Keeper Sultan
        paramsManager := k.GetParams()
        paramsData := paramsManager.GetParamsData()
        status := k.GetStatus()
        pendingTxs := k.GetPendingTransactions()
        mempoolSize := len(pendingTxs)

        // 2. Susun Respon JSON (Informasi Publik)
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

        w.Header().Set("Content-Type", "application/json")
        if err := json.NewEncoder(w).Encode(response); err != nil {
            w.WriteHeader(http.StatusInternalServerError)
        }
    }
}

func HandleUpdateConstitution(k x.BVMKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Keamanan tetap terjaga: Hanya localhost
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

        currentHeight := uint64(k.GetLastHeight())
        var err error

        // 🚩 STRATEGI TRANSISI (The Bridge Logic)
        if currentHeight < 6000 {
            // ERA KLASIK: Masih suntik langsung ke Database (Legacy Mode)
            key := "gov:upgrade:" + req.Feature
            err = k.GetStore().Put(key, req.Height)
            logger.Info("GOV", fmt.Sprintf("⚠️ Legacy Update: %s dijadwalkan via DB", req.Feature))
        } else {
            // ERA MODERN: Gunakan Smart Contract (The Shift)
            // k.CallGovernance adalah jembatan yang kita buat di keeper/bridge.go
            _, err = k.CallGovernance("ScheduleNewUpgrade", req.Feature, req.Height)
            logger.Success("GOV", fmt.Sprintf("⚖️ Modern Update: %s dijadwalkan via Contract", req.Feature))
        }

        if err != nil {
            http.Error(w, "Gagal update: "+err.Error(), 500)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "status":  "SUCCESS",
            "mode":    "MODERN_WASM",
            "feature": req.Feature,
            "at_height": req.Height,
        })
    }
}
