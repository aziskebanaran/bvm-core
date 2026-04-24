//go:build !wasm
package sdk

import (
	"fmt"
	"os"
	"github.com/aziskebanaran/bvm-core/pkg/client"
	"github.com/aziskebanaran/bvm-core/pkg/wallet"
)

const CoreURL = "http://localhost:8080"

func RegisterNexus(id, owner, token string, stake uint64) bool {
	fmt.Printf("🌐 [SDK-STD] Memulai Registrasi Nexus: %s...\n", id)

	c := client.NewBVMClient(CoreURL)
	// Memuat wallet nexus_operator.json yang ada di folder Nexus
	w, err := wallet.LoadWallet("./nexus_operator.json")
	if err != nil {
		fmt.Printf("❌ [SDK-ERR] Wallet nexus_operator.json tidak ditemukan!\n")
		return false
	}

	// Rakit transaksi Registrasi (Kirim stake ke SYSTEM_RESERVE)
	tx, err := w.SignAndPack(c, "SYSTEM_RESERVE", stake, "BVM", "Nexus Registration: "+id)
	if err != nil {
		fmt.Printf("❌ [SDK-ERR] Gagal merakit transaksi: %v\n", err)
		return false
	}

	// Broadcast ke Kernel
	txID, err := c.BroadcastTX(tx)
	if err != nil {
		fmt.Printf("❌ [SDK-ERR] Core Menolak: %v\n", err)
		return false
	}

	fmt.Printf("✅ [SDK-STD] Nexus Terdaftar On-Chain! TXID: %s\n", txID)
	return true
}

func LockForBridgeWithMemo(from, to string, amount uint64, memo string) bool {
    c := client.NewBVMClient(CoreURL)
    w, _ := wallet.LoadWallet("./nexus_operator.json")

    tx, err := w.SignAndPack(c, to, amount, "BVM", memo)
    if err != nil { 
        fmt.Printf("❌ [SDK-ERR] Gagal Sign: %v\n", err)
        return false 
    }

    txID, err := c.BroadcastTX(tx)
    if err != nil {
        fmt.Printf("❌ [SDK-ERR] Node Menolak: %v\n", err)
        return false
    }

    fmt.Printf("✅ [SDK] Anchor Sah! TXID: %s | Memo: %s\n", txID, memo)
    return true
}

// Fungsi lama tetap ada agar Nexus versi lama tidak crash
func LockForBridge(from, to string, amount uint64) bool {
    return LockForBridgeWithMemo(from, to, amount, "L2_ANCHOR_SIMPLE")
}
func EnsureCloudAccess(appID, owner string) (string, error) {
    // 1. Cek apakah sudah ada key tersimpan
    keyPath := "./.cloud_api.key"
    if dat, err := os.ReadFile(keyPath); err == nil {
        return string(dat), nil
    }

    // 2. Jika tidak ada, panggil kurir pendaftaran di Client
    fmt.Printf("🛡️ [SDK] Mendaftarkan '%s' ke BVM Cloud Storage...\n", appID)
    c := client.NewBVMClient(CoreURL)
    
    // Panggil endpoint /api/storage/register
    apiKey, err := c.RegisterApp(appID, owner)
    if err != nil {
        return "", fmt.Errorf("Gagal registrasi Cloud: %v", err)
    }

    // 3. Simpan key secara lokal untuk penggunaan masa depan
    os.WriteFile(keyPath, []byte(apiKey), 0600)
    fmt.Println("✅ [SDK] API Key Cloud berhasil didapatkan dan disimpan!")
    
    return apiKey, nil
}

// Tambahkan di bvm_std.go
func GetState(key string, value interface{}) error { return nil }
func PutState(key string, val uint64) { }
func GetBlockHeight() uint64 { return 6000 } // Default ke era modern


// Fungsi dummy lainnya tetap biarkan agar tidak error saat compile
func Transfer(from, to string, amount uint64, symbol string) bool { return true }
func GetCaller() string { return "std_caller" }
func Mint(target string, amount uint64, symbol string) bool { return true }
func Emit(tag, message string) { fmt.Printf("📝 [EVENT] %s: %s\n", tag, message) }
func UpdateStake(address string, amount uint64, isAdding bool) bool { return true }
func PtrToString(ptr uint32, size uint32) string { return "std_string" }
// Tambahkan di deretan fungsi dummy di bagian bawah bvm_std.go
func GetMethod() string { return "transfer" }
func GetArgString(key string) string { return "" }
func GetArgUint64(key string) uint64 { return 0 }
