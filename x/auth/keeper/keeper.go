package keeper

import (
	"bvm.core/pkg/logger"   // 🚩 Pakai Logger Pintar
	"bvm.core/pkg/nonce"
	"bvm.core/pkg/storage" // 🚩 Ganti DB ke Store
	"bvm.core/x/bvm/types"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/hex"
	"fmt"
)

type AuthKeeper struct {
	Store    storage.BVMStore   // 🚩 Gunakan asisten gudang
	NonceMgr nonce.NonceKeeper  // 🚩 Gunakan Interface agar fleksibel
}

func NewAuthKeeper(store storage.BVMStore, nm nonce.NonceKeeper) *AuthKeeper {
	return &AuthKeeper{
		Store:    store,
		NonceMgr: nm,
	}
}

// --- 1. MANAJEMEN PUBLIC KEY (The Identity Vault) ---

func (a *AuthKeeper) GetPublicKey(address string) (string, error) {
	var pubKey string
	// 🚩 Gunakan Store: Otomatis handle data
	err := a.Store.Get("pubkey:"+address, &pubKey)
	if err != nil {
		return "", fmt.Errorf("pubkey tidak ditemukan")
	}
	return pubKey, nil
}

func (a *AuthKeeper) SetPublicKey(address string, pubKey string) error {
	return a.Store.Put("pubkey:"+address, pubKey)
}

// VerifySignature: Gerbang Utama
func (a *AuthKeeper) VerifySignature(tx types.Transaction) bool {
	// 1. Bypass untuk Reward System
	if tx.From == "SYSTEM_REWARD" || tx.From == "" {
		return true
	}

	// 2. Ambil Public Key
	pubKeyHex := tx.PublicKey
	if pubKeyHex == "" {
		stored, err := a.GetPublicKey(tx.From)
		if err != nil {
			logger.Error("AUTH", "Public Key tidak ditemukan untuk: ", tx.From[:10])
			return false
		}
		pubKeyHex = stored
	}

	// 3. Decode Signature
	sigBytes, err := hex.DecodeString(tx.Signature)
	if err != nil {
		logger.Error("AUTH", "Signature korup dari: ", tx.From[:10])
		return false
	}

	// 4. Deteksi Protokol & Eksekusi
	var isValid bool
	if len(sigBytes) == 64 {
		isValid = a.verifyEd25519Legacy(tx, pubKeyHex)
	} else {
		isValid = a.verifyECDSAModern(tx, pubKeyHex)
	}

	if !isValid {
		logger.Warning("AUTH", "Tanda tangan PALSU dari: ", tx.From[:10])
	}
	return isValid
}

// --- LOGIKA VERIFIKASI (Legacy vs Modern) ---

func (a *AuthKeeper) verifyEd25519Legacy(tx types.Transaction, pubKeyHex string) bool {
	pubKeyBytes, _ := hex.DecodeString(pubKeyHex)
	sigBytes, _ := hex.DecodeString(tx.Signature)
	data := fmt.Sprintf("%s-%s-%f", tx.From, tx.To, tx.Amount)
	return ed25519.Verify(pubKeyBytes, []byte(data), sigBytes)
}

func (a *AuthKeeper) verifyECDSAModern(tx types.Transaction, pubKeyHex string) bool {
	hashBytes, err := tx.CalculateHash()
	if err != nil {
		return false
	}
	pubBytes, _ := hex.DecodeString(pubKeyHex)
	pubKeyInterface, err := x509.ParsePKIXPublicKey(pubBytes)
	if err != nil {
		return false
	}
	ecdsaPubKey, ok := pubKeyInterface.(*ecdsa.PublicKey)
	if !ok {
		return false
	}
	sigBytes, _ := hex.DecodeString(tx.Signature)
	return ecdsa.VerifyASN1(ecdsaPubKey, hashBytes, sigBytes)
}

// VerifyTransaction: Pintu masuk untuk x/app
func (a *AuthKeeper) VerifyTransaction(tx types.Transaction) bool {
	// 1. Cek Tanda Tangan
	if !a.VerifySignature(tx) {
		return false
	}

	// 2. Cek Nonce (PENTING: Validasi urutan di sini!)
	expectedNonce := a.NonceMgr.GetNextNonce(tx.From)
	if tx.Nonce != expectedNonce {
		logger.Warning("AUTH", "Nonce Mismatch! Akun ", tx.From[:10], " Butuh: ", expectedNonce, " Dapat: ", tx.Nonce)
		return false
	}

	return true
}

func (a *AuthKeeper) GetAccount(addr string) (types.Account, error) {
	return types.Account{
		Address: addr,
		Nonce:   a.NonceMgr.GetNextNonce(addr),
	}, nil
}

// SetAccount: Menyimpan data akun ke Store (Wajib untuk memenuhi interface)
func (k *AuthKeeper) SetAccount(acc types.Account) error {
	return k.Store.Put("acc:"+acc.Address, acc)
}

// 🚩 Tambahkan ini agar Keeper tidak perlu panggil NonceMgr langsung
func (a *AuthKeeper) GetNextNonce(addr string) uint64 {
    return a.NonceMgr.GetNextNonce(addr)
}

// 🚩 Update ini agar sinkron ke Database saat dipanggil ExecuteBlock
func (a *AuthKeeper) IncrementNonce(addr string) error {
    // 1. Naikkan di RAM (NonceMgr)
    a.NonceMgr.Increment(addr)

    // 2. PAHAT KE DISK (Kunci Utama Sultan)
    newNonce := a.NonceMgr.GetNextNonce(addr)
    return a.Store.Put("nonce:"+addr, newNonce)
}

func (a *AuthKeeper) CheckNonceIntegrity(addr string) (bool, uint64, uint64) {
    return a.NonceMgr.HealthCheckNonce(addr)
}
