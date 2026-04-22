package api

import (
	"github.com/aziskebanaran/bvm-core/pkg/storage"
	"github.com/aziskebanaran/bvm-core/x"
	"github.com/aziskebanaran/bvm-core/x/bvm/types"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
	"fmt"
)

// HandleExplorer: Melihat detail blok berdasarkan Height
func HandleExplorer(k x.BVMKeeper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		heightStr := strings.TrimPrefix(r.URL.Path, "/api/explorer/")
		height, err := strconv.Atoi(heightStr)
		if err != nil {
			http.Error(w, "Format Height salah", http.StatusBadRequest)
			return
		}

		// 🚩 PERBAIKAN: Ambil chain melalui Jenderal
		chain := k.GetChain()

		if height >= 0 && height < len(chain) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(chain[height])
		} else {
			http.Error(w, "Blok tidak ditemukan di database Sultan", http.StatusNotFound)
		}
	}
}

// HandleAddressHistory: Melihat riwayat transaksi sebuah alamat
func HandleAddressHistory(store storage.BVMStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		addr := r.URL.Query().Get("address")
		if addr == "" {
			http.Error(w, "❌ Alamat BVM harus diisi", http.StatusBadRequest)
			return
		}

		// Store Sultan sudah canggih, bisa langsung ambil history
		history, err := store.GetAddressHistory(addr)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "Gagal mengambil riwayat dari Store",
			})
			return
		}

		if history == nil {
			history = []types.Transaction{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"address": addr,
			"history": history,
		})
	}
}

// HandleHolders: Menampilkan daftar semua pemilik saldo (Rich List)
func HandleHolders(k x.BVMKeeper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🚩 PERBAIKAN: Panggil menteri Bank melalui Jenderal
		holders := k.GetBank().GetAllBalances()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(holders)
	}
}


func HandleGetBlockByHeight(k x.BVMKeeper) http.HandlerFunc {
	    return func(w http.ResponseWriter, r *http.Request) {
	        // Mengambil angka setelah /api/block/
	        heightStr := strings.TrimPrefix(r.URL.Path, "/api/block/")
	        height, err := strconv.ParseUint(heightStr, 10, 64)
	        if err != nil {
	            http.Error(w, "Sultan, format height harus angka!", http.StatusBadRequest)
	            return
	        }

                // 🚩 PERBAIKAN: Jangan ambil seluruh chain ke RAM!
                // Minta Keeper untuk ambil SATU blok saja langsung dari Database (Disk)
                targetBlock, err := k.GetBlockByHeight(height) 

                if err != nil || targetBlock == nil {
                        w.Header().Set("Content-Type", "application/json")
                        w.WriteHeader(http.StatusNotFound)
                        json.NewEncoder(w).Encode(map[string]string{
                                "error": "Blok belum terbit atau tidak ditemukan di disk!",
                        })
                        return
                }

                w.Header().Set("Content-Type", "application/json")
                json.NewEncoder(w).Encode(targetBlock)
        }
}


func HandleRealTimeExplorer(k x.BVMKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Set header untuk streaming
        w.Header().Set("Content-Type", "text/event-stream")
        w.Header().Set("Cache-Control", "no-cache")
        w.Header().Set("Connection", "keep-alive")

        lastHeight := int64(-1)

        // Loop abadi untuk memantau jantung blok
        for {
            status := k.GetStatus()
            if status.Height > lastHeight {
                // Ada blok baru! Kirim ke penonton
                chain := k.GetChain()
                if len(chain) > 0 {
                    latest := chain[len(chain)-1]
                    data, _ := json.Marshal(latest)
                    fmt.Fprintf(w, "data: %s\n\n", data)
                    w.(http.Flusher).Flush() // Dorong data ke client
                    lastHeight = status.Height
                }
            }
            time.Sleep(2 * time.Second) // Cek setiap 2 detik
        }
    }
}
