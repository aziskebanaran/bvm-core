package keeper

import (
    "encoding/hex"
    "github.com/aziskebanaran/bvm-core/pkg/storage"
    "github.com/aziskebanaran/bvm-core/x/bridge/types"
    "github.com/ethereum/go-ethereum/crypto" // 🚩 Gunakan library crypto standar
)

type Keeper struct {
    Store storage.BVMStore
}

func NewKeeper(store storage.BVMStore) *Keeper {
    return &Keeper{Store: store}
}

// IsAuthorizedRelayer memeriksa apakah relayer terdaftar di KVStore
func (k *Keeper) IsAuthorizedRelayer(address string) bool {
    var relayers []string
    // 🚩 Menggunakan konstanta dari types agar tidak hardcode
    err := k.Store.Get(types.RelayerPrefix+"list", &relayers)
    if err != nil {
        return false
    }

    for _, r := range relayers {
        if r == address {
            return true
        }
    }
    return false
}

// SetAuthorizedRelayers untuk mengupdate daftar via Governance
func (k *Keeper) SetAuthorizedRelayers(relayers []string) error {
    // 🚩 Menggunakan konstanta yang sama
    return k.Store.Put(types.RelayerPrefix+"list", relayers)
}

// 1. Verifikasi Signature Relayer
func (k *Keeper) VerifyRelayerSignature(payload []byte, signatureHex string) bool {
    var pubKeyHex string
    err := k.Store.Get(types.RelayerPrefix+"pubkey", &pubKeyHex)
    if err != nil { return false }

    pubBytes, _ := hex.DecodeString(pubKeyHex)
    sigBytes, _ := hex.DecodeString(signatureHex)

    // Verifikasi langsung tanpa menyimpan pubKey ke variabel jika tidak dipakai
    return crypto.VerifySignature(pubBytes, crypto.Keccak256(payload), sigBytes[:64])
}

// 2. Verifikasi Lock (State-based)
func (k *Keeper) VerifySourceChainLock(refTxID string) bool {
    var isLocked bool
    // Key harus konsisten dengan yang disuntikkan oleh transaksi "submit_proof"
    err := k.Store.Get(types.RelayerPrefix+"lock_proof:"+refTxID, &isLocked)
    return err == nil && isLocked == true
}

// 3. Tambahan: Fungsi untuk register PubKey Relayer (Gunakan saat inisialisasi)
func (k *Keeper) SetRelayerPubKey(pubKeyHex string) error {
    return k.Store.Put(types.RelayerPrefix+"pubkey", pubKeyHex)
}

// Fungsi ini dipanggil saat Relayer mengirim transaksi "submit_proof"
func (k *Keeper) RecordLockProof(refTxID string) error {
    return k.Store.Put(types.RelayerPrefix+"lock_proof:"+refTxID, true)
}
