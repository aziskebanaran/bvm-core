package main

import (
    "github.com/aziskebanaran/BVM.core/pkg/sdk" // Pastikan jalurnya presisi seperti ini
)

// Biarkan fungsi main tetap kosong karena ini persyaratan WASM
func main() {}

//go:wasmexport delegate
func delegate(minerPtr, minerSize uint32, amount uint64) {
    // Gunakan prefix sdk. untuk semua fungsi dari bvm.go
    caller := sdk.GetCaller()
    minerAddr := sdk.PtrToString(minerPtr, minerSize)

    if amount <= 0 {
        sdk.Emit("DPos_ERR", "Delegasi gagal: Jumlah harus lebih dari 0!")
        return
    }

    // Eksekusi update stake
    success := sdk.UpdateStake(minerAddr, amount, true)

    if success {
        sdk.Emit("DPos_OK", caller+" mendelagasikan koin ke "+minerAddr)
    } else {
        sdk.Emit("DPos_FAIL", "Gagal memperbarui database staking!")
    }
}
