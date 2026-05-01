package main

import (
	"fmt"
	"os/exec"
)

func main() {
	fmt.Println("🤖 BVM AI-Agent: Memulai Operasi Pabrik Konten...")

	// 1. SIMULASI: Mendapatkan Metadata dari AI (Misal: Gemini menghasilkan gambar)
	// Di dunia nyata, ini akan memanggil fungsi di pkg/ai/gemini.go
	contentHash := "QmXoypizjW3WknFiJnKLwHCnL72vedxjQkDDP1mXWo6uco" // Contoh CID IPFS
	mediaType := "image"

	fmt.Printf("🎨 Konten AI Baru Terdeteksi! \nType: %s \nHash: %s\n", mediaType, contentHash)

	// 2. EKSEKUSI: Mengirim ke Blockchain via BVM-Wallet
	// Kita memanggil kontrak 'media_notary' yang sudah Jenderal build tadi
	fmt.Println("🔗 Mengirim data ke Notaris Blockchain...")
	
	// Format perintah: bvm-wallet contract call [NAMA_KONTRAK] [METHOD] --args
	cmd := exec.Command("./bvm-wallet", "contract", "call", "media_notary", "mint_asset", 
		"--arg", fmt.Sprintf("ipfs_hash=%s", contentHash),
		"--arg", fmt.Sprintf("media_type=%s", mediaType))

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ Gagal mencatat: %v\nLog: %s\n", err, string(out))
		return
	}

	fmt.Println("✅ Operasi Sukses! Aset AI telah terdaftar secara permanen.")
	fmt.Printf("📄 Transaksi Log: %s\n", string(out))
}
