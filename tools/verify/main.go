package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go [path_to_source.go] [path_to_onchain.wasm]")
		return
	}

	sourceFile := os.Args[1]
	onChainWasm := os.Args[2]

	fmt.Println("🔍 Memulai Proses Verifikasi...")

	// 1. Kompilasi Ulang Source Code secara Independen
	tempWasm := "temp_verify.wasm"
	cmd := exec.Command("go", "build", "-o", tempWasm, sourceFile)
	cmd.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm")

	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Gagal Kompilasi: %v\n", err)
		return
	}
	defer os.Remove(tempWasm) // Bersihkan setelah selesai

	// 2. Hitung Hash dari Hasil Kompilasi Baru & Biner On-Chain
	hashLocal := calculateHash(tempWasm)
	hashOnChain := calculateHash(onChainWasm)

	// 3. Bandingkan!
	fmt.Printf("\n📍 Hash Lokal    : %s", hashLocal)
	fmt.Printf("\n📍 Hash On-Chain : %s\n", hashOnChain)

	if hashLocal == hashOnChain {
		fmt.Println("\n✅ VERIFIKASI SUKSES: Kode sumber COCOK dengan biner di blockchain!")
		fmt.Println("🛡️ Kontrak ini AMAN dan TRANSPARAN.")
	} else {
		fmt.Println("\n❌ VERIFIKASI GAGAL: Kode sumber TIDAK COCOK!")
		fmt.Println("⚠️ Peringatan: Ada potensi manipulasi kode.")
	}
}

func calculateHash(filePath string) string {
	f, _ := os.Open(filePath)
	defer f.Close()
	h := sha256.New()
	io.Copy(h, f)
	return hex.EncodeToString(h.Sum(nil))
}
