package main

import "github.com/aziskebanaran/bvm-core/pkg/sdk" // 🚩 Gunakan Impor Resmi

//go:wasmexport handle
func handle(fromPtr, fromSize, toPtr, toSize uint32, amount uint64, symPtr, symSize uint32) {
    // 🚩 Gunakan prefix sdk.
    caller := sdk.GetCaller()

    targetAddr := sdk.PtrToString(fromPtr, fromSize)
    tokenSymbol := sdk.PtrToString(symPtr, symSize)

    if caller == "" {
        sdk.Emit("ERROR", "Caller tidak teridentifikasi!")
        return
    }

    // Eksekusi Mint melalui SDK
    if sdk.Mint(targetAddr, amount, tokenSymbol) {
        sdk.Emit("SUCCESS", "Minting koin " + tokenSymbol + " berhasil oleh: " + caller)
    } else {
        sdk.Emit("SECURITY", "Gagal! Anda bukan Owner dari aset ini.")
    }
}

func main() {}
