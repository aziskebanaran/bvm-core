package api

import (
	"github.com/aziskebanaran/bvm-core/x"
	"github.com/aziskebanaran/bvm-core/x/bvm/types"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleMine(k x.BVMKeeper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Decode secepat kilat
		var incomingBlock types.Block
		if err := json.NewDecoder(r.Body).Decode(&incomingBlock); err != nil {
			http.Error(w, "Format JSON salah", http.StatusBadRequest)
			return
		}

		// 2. AMBIL STATUS TERKINI DARI JENDERAL
		// Kita butuh Height dan InFlight untuk validasi sebelum proses berat
		status := k.GetStatus()
		incomingIdx := int64(incomingBlock.Index)

		// 🛡️ PROTEKSI: Jangan proses blok yang sudah usang atau sedang diproses
		if incomingIdx <= status.Height || (status.InFlight > 0 && incomingIdx == status.InFlight) {
			w.WriteHeader(http.StatusConflict)
			fmt.Fprintf(w, "⚠️ Blok #%d sudah usang atau sedang diproses node lain", incomingIdx)
			return
		}

		// 3. VALIDASI HASH & DIFFICULTY (Tanpa Lock)
		// Verifikasi ini menggunakan CPU, biarkan berjalan di luar lock
		if !k.VerifyBlock(incomingBlock) {
			http.Error(w, "❌ Blok Tidak Sah (Hash/Difficulty Salah)", http.StatusNotAcceptable)
			return
		}

		// 4. EKSEKUSI PENULISAN (Proses Berat ke Disk/DB)
		// Fungsi ProcessBlock di dalam Keeper Sultan seharusnya sudah menangani 
		// pengaturan InFlight dan Height secara internal.
		if err := k.ProcessBlock(incomingBlock); err != nil {
			http.Error(w, fmt.Sprintf("Gagal memproses blok: %v", err), http.StatusInternalServerError)
			return
		}

		// 5. RESPON SUKSES
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "✅ Sukses Mematenkan Blok #%d di BVM Mainnet", incomingIdx)
	}
}

func HandleGetWork(k x.BVMKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")

        // 1. Identifikasi siapa yang minta kerjaan
        minerAddr := r.URL.Query().Get("address")
        if minerAddr == "" {
            http.Error(w, "Alamat miner wajib diisi (?address=...)", http.StatusBadRequest)
            return
        }

        // 2. MINTA KERNEL MERAKIT BLOK BARU (Tinggi + 1)
        // Fungsi CreateNextBlock Sultan sudah pintar mengambil transaksi dari mempool
        newBlock := k.CreateNextBlock(minerAddr)

        // 3. KIRIM KE MINER
        json.NewEncoder(w).Encode(newBlock)
        // Log opsional agar Sultan tahu ada miner yang kerja
        // fmt.Printf("📡 Memberikan tugas Blok #%d kepada miner %s\n", newBlock.Index, minerAddr[:12])
    }
}
