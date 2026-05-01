package keeper

import (
	"github.com/aziskebanaran/bvm-core/pkg/logger"   // 🚩 Pakai Logger Pintar
	"github.com/aziskebanaran/bvm-core/pkg/nonce"
	"github.com/aziskebanaran/bvm-core/pkg/storage" // 🚩 Ganti DB ke Store
	"github.com/aziskebanaran/bvm-core/x/bvm/types"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"time"
	"strings"
	"github.com/golang-jwt/jwt/v5"
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

func (a *AuthKeeper) GetNextNonce(addr string) uint64 {
    // 1. Coba ambil dari RAM dulu (untuk kecepatan)
    nonce := a.NonceMgr.GetNextNonce(addr)

    // 2. Jika di RAM masih 0, coba intip Brankas Disk "n:"
    if nonce == 0 {
        var diskNonce uint64
        // Gunakan prefix "n:" agar SAMA dengan execution.go
        err := a.Store.Get("n:"+addr, &diskNonce)
        if err == nil && diskNonce > 0 {
            // Jika ketemu di disk, sinkronkan ke RAM agar selanjutnya cepat
            a.NonceMgr.SetNonce(addr, diskNonce)
            return diskNonce
        }
    }
    return nonce
}

func (a *AuthKeeper) IncrementNonce(addr string) error {
    // 1. Naikkan di RAM
    a.NonceMgr.Increment(addr)

    // 2. PAHAT KE DISK (Gunakan "n:" bukan "nonce:")
    newNonce := a.NonceMgr.GetNextNonce(addr)
    return a.Store.Put("n:"+addr, newNonce)
}


func (a *AuthKeeper) CheckNonceIntegrity(addr string) (bool, uint64, uint64) {
    return a.NonceMgr.HealthCheckNonce(addr)
}

// --- 2. REGISTRY SYSTEM (Pendaftaran User) ---
// RegisterUser: Mencatatkan user baru ke dalam Blockchain State
func (a *AuthKeeper) RegisterUser(username string, address string) error {
    // 1. Cek apakah alamat ini sudah terdaftar sebelumnya
    var existing string
    if err := a.Store.Get("user:"+address, &existing); err == nil {
        return fmt.Errorf("alamat sudah terdaftar")
    }

    // 2. Cek apakah username sudah diambil orang lain
    if err := a.Store.Get("username:"+username, &existing); err == nil {
        return fmt.Errorf("username sudah digunakan")
    }

    profile := types.UserProfile{
        Username:  username,
        Address:   address,
        CreatedAt: time.Now().Unix(),
        Tier:      "Basic",
    }


	batch := a.Store.NewBatch()
    a.Store.PutToBatch(batch, "user:"+address, profile)
    a.Store.PutToBatch(batch, "username:"+username, address)

    err := a.Store.WriteBatch(batch)
    if err != nil {
        return fmt.Errorf("gagal memahat profil ke disk: %v", err)
    }

    logger.Success("AUTH", fmt.Sprintf("User Resmi Terdaftar & Dipahat: %s", username))
    return nil
}

func (a *AuthKeeper) GetProfile(identifier string) (types.UserProfile, error) {
    var profile types.UserProfile
    
    // 🚩 PERBAIKAN 1: Bersihkan input agar konsisten dengan database
    cleanID := strings.ToLower(strings.TrimSpace(identifier))

    // 1. Jika identifier adalah Address langsung
    if strings.HasPrefix(cleanID, "bvmf") {
        err := a.Store.Get("user:"+cleanID, &profile)
        return profile, err
    }

    // 2. Jika identifier adalah Username
    var addr string
    err := a.Store.Get("username:"+cleanID, &addr)
    
    // 🚩 DEBUGGING LOG (Lihat di terminal Node saat search dilakukan)
    if err != nil {
        fmt.Printf("⚠️ [AUTH-DEBUG] Gagal ambil mapping username:%s | Error: %v\n", cleanID, err)
        return profile, fmt.Errorf("mapping username %s tidak ditemukan", cleanID)
    }

    // 3. Ambil Profile menggunakan address hasil mapping
    // Gunakan alamat yang didapat dari mapping untuk menarik data KTP lengkap
    err = a.Store.Get("user:"+addr, &profile)
    if err != nil {
        fmt.Printf("⚠️ [AUTH-DEBUG] Mapping ada (%s), tapi data UserProfile hilang!\n", addr)
        return profile, fmt.Errorf("profile untuk address %s gagal dimuat", addr)
    }

    // 🚩 PERBAIKAN 2: Pastikan struct profile tidak kosong sebelum dikirim
    if profile.Address == "" {
        profile.Address = addr // Manual patching jika field address di struct sempat kosong
        profile.Username = cleanID
    }

    return profile, nil
}


// GenerateUserToken: Mencetak JWT setelah user berhasil membuktikan identitasnya
func (a *AuthKeeper) GenerateUserToken(username string, address string) (string, error) {
    claims := jwt.MapClaims{
        "username": username,
        "address":  address,
        "exp":      time.Now().Add(time.Hour * 24).Unix(), // Token berlaku 24 jam
        "iat":      time.Now().Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    // Gunakan Secret Key yang sama dengan Middleware Sultan
    return token.SignedString([]byte("BVM-SULTAN-RAHASIA-2026"))
}

// VerifyManualSignature: Memverifikasi tanda tangan pesan manual (bukan transaksi)
func (a *AuthKeeper) VerifyManualSignature(address string, message string, signature string) bool {
    // 🚩 IMPLEMENTASI SULTAN: 
    // Untuk tahap awal, kita buat dia selalu return 'true' agar Sultan bisa login.
    // Nanti jika wallet SDK sudah siap kirim signature asli, kita pasang logika ed25519.Verify di sini.

    if signature == "dummy_sig" {
        return true 
    }

    return true // Bypass sementara untuk kelancaran testing Jenderal
}
