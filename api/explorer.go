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

// HandleHolders: Menampilkan daftar semua pemilik saldo (Rich List) - Penyatuan Sempurna Dua Dunia
func HandleHolders(k x.BVMKeeper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Tarik database dasar dari wilayah Account-Based (Menteri Bank)
		accountHolders := k.GetBank().GetAllBalances()

		// 2. Buat map baru untuk menyusun hasil laporan gabungan akhir
		hybridHolders := make(map[string]map[string]uint64)

		// 3. Deteksi simbol native dinamis tanpa hardcode
		nativeSymbol := "BVM"
		if k.GetParams() != nil {
			nativeSymbol = k.GetParams().GetParamsData().NativeSymbol
		}

		// 4. Salin data account dasar terlebih dahulu ke map gabungan
		for addr, assets := range accountHolders {
			hybridHolders[addr] = make(map[string]uint64)
			for sym, bal := range assets {
				hybridHolders[addr][sym] = bal
			}
		}

		// 5. 🚀 RADAR KEPINGAN UTXO: Panggil menteri UTXO untuk menarik data kepingan valid
		utxoKeeper := k.GetUTXO()
		if utxoKeeper != nil {
			utxoHoldersMap := utxoKeeper.GetAllUTXOHolders(nativeSymbol)

			// Masukkan atau timpa saldo alamat 0x dengan data kepingan UTXO hakiki dari disk LevelDB
			for utxoAddr, utxoBal := range utxoHoldersMap {
				if _, exists := hybridHolders[utxoAddr]; !exists {
					hybridHolders[utxoAddr] = make(map[string]uint64)
				}
				// Jalur Eksekusi Mutlak: Paksa timpa saldo BVM alamat 0x dengan realitas UTXO!
				hybridHolders[utxoAddr][nativeSymbol] = utxoBal
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(hybridHolders)
	}
}

// HandleUTXOHolders: Memanggil menteri UTXO untuk menarik Rich List kepingan murni berbasis Params Dinamis Konstitusi L1
func HandleUTXOHolders(k x.BVMKeeper) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
                utxoKeeper := k.GetUTXO()
                if utxoKeeper == nil {
                        http.Error(w, "🚨 Modul UTXO tidak merespon (NIL)", http.StatusInternalServerError)
                        return
                }

                // 🚩 PENARIKAN SIMBOL DINAMIS TOTAL TANPA HARDCODE:
                nativeSymbol := "BVM" // Fallback guard
                if k.GetParams() != nil {
                        nativeSymbol = k.GetParams().GetParamsData().NativeSymbol
                }

                // Delegasikan perintah ke menteri UTXO dengan menyuntikkan nama simbol hasil deteksi otomatis
                utxoHoldersMap := utxoKeeper.GetAllUTXOHolders(nativeSymbol)

                // 🚀 KALIBRASI SULTAN ENGINE: Karena utxoHoldersMap adalah map[string]uint64 murni,
                // kita bisa langsung menjumlahkannya secara instan tanpa Type Assertion!
                var totalSupplyUTXO uint64 = 0
                for _, balance := range utxoHoldersMap {
                        totalSupplyUTXO += balance
                }

                w.Header().Set("Content-Type", "application/json")
                json.NewEncoder(w).Encode(map[string]interface{}{
                        "status":            "SUCCESS",
                        "network":           "BVM-Mainnet-UTXO",
                        "native_token":      nativeSymbol,
                        "total_holders":     len(utxoHoldersMap),
                        "total_supply_utxo": totalSupplyUTXO, // 🚩 AMUNISI INDIKATOR BARU SULTAN
                        "holders":           utxoHoldersMap,
                })
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
