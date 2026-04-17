package keeper

import (
    "github.com/aziskebanaran/bvm-core/x/storage/types"
)

func (k *StorageKeeper) CheckRules(app types.AppContainer, path string, action string, callerAddr string) bool {
    // 🚩 TAMBAHAN SULTAN: Jika yang akses adalah OWNER, langsung lolos!
    if callerAddr == app.Owner {
        return true
    }

    // 1. Ambil aturan untuk path spesifik atau gunakan default
    ruleSet, ok := app.Rules[path].(map[string]interface{})
    if !ok {
        ruleSet, ok = app.Rules["."].(map[string]interface{})
        if !ok {
            // 🚩 MODIFIKASI: Jika tidak ada aturan, tapi user terautentikasi, izinkan saja dulu
            return callerAddr != "" 
        }
    }

    // 2. Ambil otorisasi untuk aksi tersebut
    requiredAuth, ok := ruleSet["."+action].(string)
    if !ok {
        return callerAddr != "" // Default izin jika terautentikasi
    }

    // 3. Evaluator
    switch requiredAuth {
    case "public":
        return true
    case "owner":
        return callerAddr == app.Owner
    case "auth != null":
        return callerAddr != ""
    default:
        return false
    }
}
