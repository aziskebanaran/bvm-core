package keeper

import (
    "fmt"
)

// CallGovernance: Memanggil kontrak manajemen di folder luar secara internal
func (k *Keeper) CallGovernance(method string, args ...interface{}) (interface{}, error) {
    // 1. Definisikan Alamat Kontrak Governance Utama
    const GovContractAddr = "system_gov_manager"

    // 2. Lakukan Query ke Mesin WASM
    // Jenderal menggunakan k.Wasm (WasmKeeper) yang sudah ada di interface
    result, err := k.Wasm.QueryContract(GovContractAddr, method, args...)
    if err != nil {
        return nil, fmt.Errorf("⚖️ GOV_BRIDGE: Gagal memanggil metode [%s]: %v", method, err)
    }

    return result, nil
}
