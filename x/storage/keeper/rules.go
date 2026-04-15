package keeper

import (
    "bvm.core/x/storage/types"
)

// CheckRules: Memvalidasi apakah aksi (read/write) diizinkan berdasarkan JSON Rules
func (k *StorageKeeper) CheckRules(app types.AppContainer, path string, action string, callerAddr string) bool {
    // 1. Ambil aturan untuk path spesifik atau gunakan default
    // Aturan tersimpan dalam app.Rules (map[string]interface{})

    ruleSet, ok := app.Rules[path].(map[string]interface{})
    if !ok {
        // Jika path tidak ada aturan spesifik, cek aturan root "."
        ruleSet, ok = app.Rules["."].(map[string]interface{})
        if !ok {
            return false // Tidak ada aturan = Akses Ditolak (Aman)
        }
    }

    // 2. Ambil otorisasi untuk aksi tersebut (misal: ".write" atau ".read")
    requiredAuth, ok := ruleSet["."+action].(string)
    if !ok {
        return false
    }

    // 3. Evaluator Sederhana ala BVM Guard
    switch requiredAuth {
    case "public":
        return true
    case "owner":
        // Hanya pemilik AppContainer yang boleh eksekusi
        return callerAddr == app.Owner
    case "auth != null":
        // Siapapun yang punya alamat dompet valid (bukan anonim)
        return callerAddr != ""
    default:

        return false
    }
}
