package wallet

import (
	"github.com/aziskebanaran/bvm-core/pkg/client" // Kabel Sakti Sultan
	"github.com/aziskebanaran/bvm-core/pkg/logger" // Logger Berwarna
	"github.com/aziskebanaran/bvm-core/pkg/types"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
 	"encoding/json" // Hanya untuk simpan file lokal saja
	"github.com/tyler-smith/go-bip39"
)

type BVMWallet struct {
    Mnemonic   string `json:"mnemonic,omitempty"` // 🚩 TAMBAHKAN INI
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
    if err != nil { return nil, err }

    var w BVMWallet
    if err := json.Unmarshal(data, &w); err != nil {
        return nil, err
    }

    // 🚩 TAMBAHAN: Deteksi otomatis tipe wallet
    if w.Mnemonic == "" {
        fmt.Printf("⚠️  [WALLET] Dompet Legacy terdeteksi (Tanpa Mnemonic). Segera migrasi!\n")
    }

    return &w, nil
}


func CreateNewWallet() (*BVMWallet, string, error) {
    // 1. Buat 12 kata baru
    mnemonic, err := GenerateMnemonic()
    if err != nil { return nil, "", err }

    // 2. Turunkan wallet (Nonce otomatis 0 di dalam sini)
    wallet, err := CreateFromMnemonic(mnemonic, 0)
    if err != nil { return nil, "", err }

    return wallet, mnemonic, nil
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
    tx := types.NewTransaction(w.Address, to, amount, info.DynamicFee, symbol, memo, finalNonce)

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
    // 1. Jika file sudah ada, pakai yang lama (Legacy Wallet)
    if _, err := os.Stat(filename); err == nil {
        return LoadWallet(filename)
    }

    // 2. Jika tidak ada, buat baru (Mnemonic Wallet)
    // 🚩 PERHATIKAN: Sekarang kita tangkap mnemonic-nya
    newW, mnemonic, err := CreateNewWallet() 
    if err != nil {
        return nil, err
    }

    // 3. Simpan agar permanen
    err = SaveWallet(newW, filename)
    if err != nil {
        return nil, err
    }

    fmt.Printf("🆕 [WALLET] Dompet Baru Terpahat: %s\n", newW.Address)
    fmt.Printf("⚠️  MNEMONIC: %s\n", mnemonic) // Tampilkan sekali saat pembuatan
    return newW, nil
}


func CreateFromMnemonic(mnemonic string, index int) (*BVMWallet, error) {
    if !bip39.IsMnemonicValid(mnemonic) {
        return nil, fmt.Errorf("mnemonic tidak valid")
    }

    // 1. Generate Seed
    seed := bip39.NewSeed(mnemonic, "")

    // 2. Buat Hash dari Seed + Index (Agar deterministik)
    combined := append(seed, []byte(fmt.Sprintf("%d", index))...)
    hashPriv := sha256.Sum256(combined)

    // 3. Bangkitkan Kunci P256 secara Manual & Aman
    // Kita buat struct PrivateKey kosong lalu isi nilai D-nya dengan hash kita
    privKey := new(ecdsa.PrivateKey)
    privKey.Curve = elliptic.P256()

    // Gunakan math/big untuk mengubah hash menjadi angka besar
    privKey.D = new(big.Int).SetBytes(hashPriv[:])

    // Hitung koordinat X dan Y (Public Key) secara matematis berdasarkan D
    privKey.PublicKey.X, privKey.PublicKey.Y = privKey.Curve.ScalarBaseMult(hashPriv[:])

    // 4. Proses Marshal (Sekarang DIJAMIN tidak nil)
    privBytes, err := x509.MarshalECPrivateKey(privKey)
    if err != nil { return nil, err }

    pubBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
    if err != nil { return nil, err }

    hash := sha256.Sum256(pubBytes)
    address := fmt.Sprintf("bvmf%s", hex.EncodeToString(hash[:10]))

    return &BVMWallet{
        Mnemonic:   mnemonic,
        PrivateKey: hex.EncodeToString(privBytes),
        PublicKey:  hex.EncodeToString(pubBytes),
        Address:    address,
        Nonce:      0,
    }, nil
}


// GenerateMnemonic: Menciptakan 12 kata rahasia baru untuk User
func GenerateMnemonic() (string, error) {
    // 1. Buat entropy (kerandoman) 128-bit untuk 12 kata
    entropy, err := bip39.NewEntropy(128)
    if err != nil {
        return "", err
    }

    // 2. Ubah biner menjadi kata-kata manusia
    return bip39.NewMnemonic(entropy)
}

// SignAndPackCustom: Menggunakan "pabrik" NewRegisterTransaction agar standar
func (w *BVMWallet) SignAndPackCustom(c *client.BVMClient, username string) (types.Transaction, error) {
    // 1. Ambil Info Jaringan & State Nonce
    info, err := c.GetNetworkInfo()
    if err != nil { return types.Transaction{}, err }
    
    state, err := c.GetSecureState(w.Address)
    if err != nil { return types.Transaction{}, err }

    // 2. RAKIT MENGGUNAKAN STANDAR SULTAN (Fungsi 3 di types/transaction.go)
    // Fungsi ini otomatis mengisi: To="SYSTEM_AUTH", Type="user_register", Payload, dll.
    tx := types.NewRegisterTransaction(w.Address, username, info.DynamicFee, state.Nonce)

    // 3. Tambahkan Kunci Publik (Wajib untuk Verifikasi Signature di Kernel)
    tx.PublicKey = w.PublicKey
    
    // Update ID karena ada tambahan PublicKey dalam perhitungan Hash
    tx.ID = tx.GenerateID()

    // 4. PROSES TANDA TANGAN (ASN1 P256)
    hashBytes, _ := tx.CalculateHash()
    privBytes, _ := hex.DecodeString(w.PrivateKey)
    rawPriv, _ := x509.ParseECPrivateKey(privBytes)

    sig, err := ecdsa.SignASN1(rand.Reader, rawPriv, hashBytes)
    if err != nil { return tx, err }

    tx.Signature = hex.EncodeToString(sig)

    return tx, nil
}
