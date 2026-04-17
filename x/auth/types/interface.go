package types

import "github.com/aziskebanaran/BVM.core/x/bvm/types"

// AuthKeeper: Intelijen & Identitas (Menteri Keamanan)
type AuthKeeper interface {
    // --- 1. Verifikasi ---
    VerifyTransaction(tx types.Transaction) bool
    VerifySignature(tx types.Transaction) bool

    // --- 2. Identitas (Public Key) ---
    GetPublicKey(address string) (string, error)
    SetPublicKey(address string, pubKey string) error

    // --- 3. Manajemen Nonce & Akun ---
    GetAccount(addr string) (types.Account, error)
    SetAccount(acc types.Account) error
    IncrementNonce(addr string) error
    GetNextNonce(addr string) uint64
    CheckNonceIntegrity(addr string) (bool, uint64, uint64)

    RegisterUser(username string, address string) error
    GetProfile(identifier string) (types.UserProfile, error)
    GenerateUserToken(username string, address string) (string, error)
    VerifyManualSignature(address string, message string, signature string) bool
}
