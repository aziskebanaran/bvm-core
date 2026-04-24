package keeper

import (
    "github.com/aziskebanaran/bvm-core/pkg/logger"
    "fmt"
	"os"
)

func (k *Keeper) IsFeatureActive(feature string, height int64) bool {
    // ERA KLASIK (Gunakan LevelDB Native)
    if height < 6000 {
        var activationHeight int64
        key := "gov:upgrade:" + feature
        if err := k.Store.Get(key, &activationHeight); err != nil {
            return false
        }
        return height >= activationHeight
    }

    // ERA MODERN (Gunakan Jembatan Kontrak)
    active, err := k.CallGovernance("IsFeatureEnabled", feature)
    if err != nil {
        // Fallback: Jika kontrak gagal, demi keamanan kita anggap fitur belum aktif
        return false
    }

    return active.(bool)
}


// GetAllScheduledUpgrades mengambil semua jadwal upgrade untuk API
func (k *Keeper) GetAllScheduledUpgrades() interface{} {
    features := []string{"WASM_ENGINE", "STAKING_V2", "BURN_FEE_V2"}
    
    type UpgradeStatus struct {
        Feature          string `json:"feature"`
        ActivationHeight int64  `json:"activation_height"`
    }

    var results []UpgradeStatus
    for _, f := range features {
        var actHeight int64
        key := "gov:upgrade:" + f
        if err := k.Store.Get(key, &actHeight); err == nil {
            results = append(results, UpgradeStatus{
                Feature:          f,
                ActivationHeight: actHeight,
            })
        }
    }
    return results
}

func (k *Keeper) InitializeGovernance() error {
    // 1. Cek apakah Kontrak sudah ada di folder build
    contractPath := "build/node_manager.wasm"
    bytecode, err := os.ReadFile(contractPath)
    if err != nil {
        return fmt.Errorf("❌ Gagal memuat Konstitusi: %v", err)
    }

    // 2. Daftarkan ke Mesin WASM Core
    // Kita berikan ID khusus "system_gov_manager" agar mudah dipanggil oleh Bridge
    addr, err := k.Wasm.DeployContract("system_gov_manager", bytecode)
    if err != nil {
        return fmt.Errorf("❌ Gagal menanamkan Konstitusi ke State: %v", err)
    }

    logger.Success("GOV", fmt.Sprintf("🛡️ Konstitusi Aktif! Alamat: %s", addr))
    return nil
}
