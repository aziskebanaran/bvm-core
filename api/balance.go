package api

import (
    "github.com/aziskebanaran/bvm-core/x"
	"github.com/aziskebanaran/bvm-core/x/bvm/types"
    "encoding/json"
    "net/http"
	"strings"
)

func HandleBalance(k x.BVMKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        addr := r.URL.Query().Get("address")
        if addr == "" { addr = r.URL.Query().Get("addr") }

        var state types.WalletState
        var exists bool

        // 🚩 LOGIKA NAVIGASI SULTAN: Deteksi Wilayah Otomatis
        if strings.HasPrefix(addr, "0x") {
            // Jika alamat 0x, panggil fungsi khusus UTXO yang Jenderal buat di Keeper
            state, exists = k.GetSecureBalanceUTXO(addr)
        } else {
            // Jika bvmf, panggil fungsi Account biasa
            state, exists = k.GetSecureBalance(addr)
        }

        w.Header().Set("Content-Type", "application/json")
        
        // Jika benar-benar tidak ditemukan di kedua wilayah
        if !exists {
            json.NewEncoder(w).Encode(map[string]interface{}{
                "address": addr,
                "balance_atomic": 0,
                "balance_display": "0.00000000",
                "symbol": "BVM",
                "type": "NONE",
            })
            return
        }

        // Kirim WalletState yang lengkap (Isinya sudah ada Type, Balance, dll)
        json.NewEncoder(w).Encode(state)
    }
}


func HandleGetAccount(k x.BVMKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        address := r.URL.Query().Get("address")
        if address == "" { address = r.URL.Query().Get("addr") }

        // 1. Ambil Saldo Utama (BVM) - Langsung dari Keeper
        bvmBal := k.GetBalanceBVM(address)
        nextNonce := k.GetNonce().GetNextNonce(address)

        // 2. 🚩 SCAN TOKEN
        allBalances := make(map[string]interface{})
        allBalances["BVM"] = map[string]interface{}{
            "atomic":  bvmBal,
            "display": k.FromAtomic(bvmBal),
        }

        // 3. Gabungkan Metadata Akun (Jika ada status/nama di 'acc:')
        accMeta, _ := k.GetBank().GetAccount(address)

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "address": address,
            "nonce":   nextNonce,
            "status":  accMeta.Status,
            "assets":  allBalances,
        })
    }
}

func HandleUTXOBalance(k x.BVMKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 1. Ambil Parameter dari URL
        addr := r.URL.Query().Get("address")
        symbol := r.URL.Query().Get("symbol")

        if addr == "" {
            http.Error(w, "❌ Alamat (address) wajib diisi", http.StatusBadRequest)
            return
        }

        // 2. Panggil Menteri UTXO melalui Kernel (BVMKeeper)
        utxoKeeper := k.GetUTXO()
        balance := utxoKeeper.GetTotalBalance(addr, symbol)

        // 3. Kirim Laporan ke Klien
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "address": addr,
            "symbol":  symbol,
            "balance": balance,
            "type":    "UTXO_BASED",
            "status":  "SUCCESS",
        })
    }
}

func HandleUTXOList(k x.BVMKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        addr := r.URL.Query().Get("address")
        symbol := r.URL.Query().Get("symbol") // Opsional

        if addr == "" {
            http.Error(w, "❌ Alamat wajib diisi", http.StatusBadRequest)
            return
        }

        // Panggil Menteri UTXO melalui Kernel
        utxoKeeper := k.GetUTXO()
        list := utxoKeeper.ListUnspent(addr, symbol)

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "address": addr,
            "count":   len(list),
            "utxos":   list,
            "status":  "SUCCESS",
        })
    }
}
