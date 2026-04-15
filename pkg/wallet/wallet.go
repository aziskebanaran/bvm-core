package wallet

import (
	"bvm.core/pkg/client" // Kabel Sakti Sultan
	"bvm.core/pkg/logger" // Logger Berwarna
	"bvm.core/x/bvm/types"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"os"
    "encoding/json" // Hanya untuk simpan file lokal saja
)

// BVMWallet: Sekarang lebih ramping
type BVMWallet struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	Address    string `json:"address"`
	Nonce      uint64 `json:"nonce"`
}

// SaveWallet menyimpan wallet ke file
func SaveWallet(w *BVMWallet, filename string) error {
        data, err := json.MarshalIndent(w, "", "  ")
        if err != nil {
                return err
        }
        return os.WriteFile(filename, data, 0600)
}

// LoadWallet mengambil wallet dari file
func LoadWallet(filename string) (*BVMWallet, error) {
        data, err := os.ReadFile(filename)
        if err != nil {
                return nil, err
        }
        var w BVMWallet
        err = json.Unmarshal(data, &w)
        return &w, err
}

// CreateNewWallet menghasilkan dompet baru dengan format Hex
func CreateNewWallet() (*BVMWallet, error) {
    // 1. Ganti ke P256 (Mesin Baru Sultan)
    privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    if err != nil { return nil, err }

    // 2. Export ke Hex (Agar bisa disimpan di JSON)
    privBytes, _ := x509.MarshalECPrivateKey(privKey)
    pubBytes, _ := x509.MarshalPKIXPublicKey(&privKey.PublicKey)

    // 3. Buat Address Unik
    hash := sha256.Sum256(pubBytes)
    address := fmt.Sprintf("bvmf%s", hex.EncodeToString(hash[:10]))

    return &BVMWallet{
        PrivateKey: hex.EncodeToString(privBytes),
        PublicKey:  hex.EncodeToString(pubBytes),
        Address:    address,
        Nonce:      0,
    }, nil
}

// FetchNonce: Mengambil Nonce terbaru langsung dari Kernel (AuthKeeper)
func (w *BVMWallet) FetchNonce(c *client.BVMClient) uint64 {
	acc, err := c.GetAccount(w.Address)
	if err != nil {
		logger.Error("WALLET", fmt.Sprintf("Gagal ambil Nonce: %v", err))
		return w.Nonce
	}
	return acc.Nonce
}

func (w *BVMWallet) SignAndPack(c *client.BVMClient, to string, amount uint64, symbol string, memo string) (types.Transaction, error) {
    // 1. Ambil Info Jaringan & State
    info, err := c.GetNetworkInfo()
    if err != nil { return types.Transaction{}, err }

    state, err := c.GetSecureState(w.Address)
    if err != nil { return types.Transaction{}, err }

    // 2. Sinkronisasi Nonce (Disk + Mempool)
    mempoolStats, err := c.GetMempool()
    ramCount := uint64(0)
    if err == nil && mempoolStats != nil {
        for _, tx := range mempoolStats.Txs {
            if tx.From == w.Address { ramCount++ }
        }
    }
    finalNonce := state.Nonce + ramCount

    // 3. RAKIT TRANSAKSI (Gunakan NewTransaction agar Timestamp terisi otomatis)
    // Sesuai types/transaction.go: From, To, Amount, Fee, Symbol, Memo, Nonce, Params
    tx := types.NewTransaction(w.Address, to, amount, info.DynamicFee, symbol, memo, finalNonce, info.Params)

    // 🚩 PENTING: Masukkan PublicKey SEBELUM hitung Hash
    tx.PublicKey = w.PublicKey

    // 🚩 PENTING: Update ID agar sinkron dengan data yang sudah lengkap
    tx.ID = tx.GenerateID()

    // 4. HITUNG HASH BAKU (Memanggil CalculateHash yang baru Sultan cat tadi)
    hashBytes, err := tx.CalculateHash()
    if err != nil { return tx, err }

    // 5. PROSES TANDA TANGAN (SIGNING)
    privBytes, _ := hex.DecodeString(w.PrivateKey)
    rawPriv, err := x509.ParseECPrivateKey(privBytes)
    if err != nil { return tx, err }

    // Gunakan ASN1 sesuai standar P256 di Kernel
    sig, err := ecdsa.SignASN1(rand.Reader, rawPriv, hashBytes)
    if err != nil { return tx, err }

    tx.Signature = hex.EncodeToString(sig)

    // 6. Update Nonce Lokal
    w.Nonce = finalNonce + 1

    return tx, nil
}


func BroadcastTransaction(c *client.BVMClient, tx types.Transaction) error {
    // 🚩 PERBAIKAN: Tangkap txID dan err
    txID, err := c.BroadcastTX(tx) 

    if err != nil {
        logger.Error("NETWORK", fmt.Sprintf("Transaksi Gagal: %v", err))
        return err
    }

    // Sekarang Sultan bisa pamer TXID-nya di log!
    logger.Success("NETWORK", fmt.Sprintf("Berhasil! TXID: %s", txID))
    return nil
}


func LoadOrCreate(filename string) (*BVMWallet, error) {
    // 1. Cek apakah file dompet sudah ada
    if _, err := os.Stat(filename); err == nil {
        return LoadWallet(filename)
    }

    // 2. Jika tidak ada, buat baru
    newW, err := CreateNewWallet()
    if err != nil {
        return nil, err
    }

    // 3. Simpan agar permanen
    err = SaveWallet(newW, filename)
    if err != nil {
        return nil, err
    }

    fmt.Printf("🆕 [WALLET] Membuat dompet baru: %s\n", newW.Address)
    return newW, nil
}
