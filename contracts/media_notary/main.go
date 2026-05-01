package main

import (
	"github.com/aziskebanaran/bvm-core/pkg/sdk"
	"fmt"
)

// Karena kita di lingkungan WASM murni, Go butuh fungsi main 
// yang akan dieksekusi oleh Virtual Machine.
func main() {}

//go:export invoke
func invoke() uint32 {
	method := sdk.GetMethod()

	switch method {
	case "mint_asset":
		return MintAsset()
	default:
		return 0 // Error: Metode tidak ditemukan
	}
}

func MintAsset() uint32 {
	// 1. Ambil data dari argument yang dikirim saat transaksi
	ipfsHash := sdk.GetArgString("ipfs_hash")
	mediaType := sdk.GetArgString("media_type")
	
	if ipfsHash == "" {
		sdk.Emit("ERROR", "Hash IPFS tidak boleh kosong")
		return 0
	}

	// 2. Ambil siapa yang memanggil (pengirim transaksi)
	caller := sdk.GetCaller()

	// 3. Catat di State BVM
	// Catatan: Karena PutState di bvm_wasm.go Jenderal saat ini 
	// hanya menerima uint64, kita akan simpan tanda sukses dulu.
	// (Untuk menyimpan string panjang, Jenderal perlu upgrade PutState di kernel nanti)
	
	key := fmt.Sprintf("media:%s", ipfsHash)
	sdk.PutState(key, 1) // Tandai sebagai '1' (Terverifikasi)

	sdk.Emit("SUCCESS", fmt.Sprintf("Aset %s dicatat oleh %s", mediaType, caller))
	
	return 1 // Sukses
}
