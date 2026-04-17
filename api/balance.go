package api

import (
    "github.com/aziskebanaran/bvm-core/x"
    "encoding/json"
    "net/http"
)

func HandleBalance(k x.BVMKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        addr := r.URL.Query().Get("address")
        if addr == "" { addr = r.URL.Query().Get("addr") }

        // 🚩 GUNAKAN JALUR VVIP GetSecureBalance (Stateless)
        state, exists := k.GetSecureBalance(addr)

        w.Header().Set("Content-Type", "application/json")
        if !exists {
            // Jika akun baru, tetap kirim data kosong yang valid
            json.NewEncoder(w).Encode(map[string]interface{}{
                "address": addr,
                "balance_atomic": 0,
                "balance_display": "0.00000000",
                "nonce": 0,
                "symbol": "BVM",
            })
            return
        }

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

        // 2. 🚩 SCAN TOKEN (Logika Taktis Sultan)
        // Karena BVMStore Sultan berbasis Key-Value, kita butuh cara
        // untuk tahu token apa saja yang dimiliki user.
        // Untuk sekarang, kita ambil "Daftar Saldo" dari Bank.

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
            "assets":  allBalances, // Sultan bisa kembangkan loop untuk token t: di sini
        })
    }
}
