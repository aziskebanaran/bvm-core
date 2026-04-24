package types

import (
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "time"
    "github.com/cbergoon/merkletree"
)

type Transaction struct {
    ID        string  `json:"id"`
    From      string  `json:"from"`
    To        string  `json:"to"`
    Amount    uint64  `json:"amount"`
    Fee       uint64  `json:"fee"`
    Symbol    string  `json:"symbol"` 

    Memo      string  `json:"memo"`   

    Nonce     uint64  `json:"nonce"`
    Timestamp int64   `json:"timestamp"`

    Type      string  `json:"type"`
    Layer     int     `json:"layer"`
    Payload   []byte  `json:"payload"`
    ZKP_Proof string  `json:"zkp_proof"`

    PublicKey string  `json:"public_key"`
    Signature string  `json:"signature"`
}

func (t Transaction) CalculateHash() ([]byte, error) {
    // 🚩 URUTAN SULTAN: 100% Sesuai urutan variabel di struct (Tanpa ID & Signature)
    // Menggunakan pemisah titik dua (:) sesuai keinginan awal Sultan
    data := fmt.Sprintf("%s:%s:%d:%d:%s:%s:%d:%d:%s:%d:%s:%s:%s",
        t.From,                         // 1.  From
        t.To,                           // 2.  To
        t.Amount,                       // 3.  Amount
        t.Fee,                          // 4.  Fee
        t.Symbol,                       // 5.  Symbol
        t.Memo,                         // 6.  Memo
        t.Nonce,                        // 7.  Nonce
        t.Timestamp,                    // 8.  Timestamp
        t.Type,                         // 9.  Type
        t.Layer,                        // 10. Layer
        hex.EncodeToString(t.Payload),  // 11. Payload
        t.ZKP_Proof,                    // 12. ZKP_Proof
        t.PublicKey,                    // 13. PublicKey (INI DIA!)
    )

    h := sha256.Sum256([]byte(data))
    return h[:], nil
}

// GetID: Fungsi pembantu untuk mendapatkan string Hex (TXID)
func (t *Transaction) GenerateID() string {
    hash, _ := t.CalculateHash()
    return hex.EncodeToString(hash)
}

// Equals: Wajib ada untuk Merkle Tree Sultan
func (t Transaction) Equals(other merkletree.Content) (bool, error) {
    ot, ok := other.(Transaction)
    if !ok { return false, nil }

    // Jika ID sudah ada, bandingkan ID saja (lebih cepat)
    if t.ID != "" && ot.ID != "" {
        return t.ID == ot.ID, nil
    }

    h1, _ := t.CalculateHash()
    h2, _ := ot.CalculateHash()
    return hex.EncodeToString(h1) == hex.EncodeToString(h2), nil
}

// 🚩 PERBAIKAN: Ganti float64 menjadi uint64 pada parameter 'amount' dan 'fee'
func NewTransaction(from, to string, amount, fee uint64, symbol, memo string, nonce uint64, p Params) Transaction {

    tx := Transaction{
        From:      from,
        To:        to,
        Amount:    amount, // ✅ SEKARANG COCOK: uint64 diisi oleh uint64
        Fee:       fee,    // ✅ SEKARANG COCOK: uint64 diisi oleh uint64
        Symbol:    p.GetNative(symbol),
        Memo:      memo,
        Nonce:     nonce,
        Timestamp: time.Now().Unix(),
        Layer:     1,
        Type:      "transfer",
        Payload:   []byte{},
        ZKP_Proof: "",
    }

    tx.ID = tx.GenerateID()
    return tx
}

// --- Fungsi 2: Untuk Eksekusi Smart Contract (WASM) ---
func NewContractTransaction(from, to string, fee uint64, payload []byte, nonce uint64) Transaction {
    tx := Transaction{
        From:      from,
        To:        to,              // Alamat Kontrak (bvmwasm...)
        Amount:    0,               // Kontrak biasanya tidak mengirim saldo utama
        Fee:       fee,             // Biaya eksekusi
        Symbol:    "BVM",           // Biaya selalu dibayar pakai native token
        Nonce:     nonce,
        Timestamp: time.Now().Unix(),
        Type:      "contract_call", // Penanda untuk WASM Keeper
        Payload:   payload,         // Data perintah untuk kontrak
        Layer:     1,
    }
    tx.ID = tx.GenerateID()
    return tx
}

// --- Fungsi 3: Untuk Pendaftaran User Baru (Identity) ---
func NewRegisterTransaction(from, username string, fee uint64, nonce uint64) Transaction {
    // Kita bungkus username ke dalam Payload agar ID transaksi tetap unik
    payload := []byte(fmt.Sprintf(`{"username":"%s"}`, username))

    tx := Transaction{
        From:      from,
        To:        "SYSTEM_AUTH",    // Alamat virtual untuk registrasi
        Amount:    0,                // Daftar tidak butuh kirim koin ke siapa-siapa
        Fee:       fee,              // Biaya administrasi pendaftaran
        Symbol:    "BVM",
        Nonce:     nonce,
        Timestamp: time.Now().Unix(),
        Type:      "user_register",  // Kode kunci untuk AuthKeeper
        Payload:   payload,
        Layer:     1,
        Memo:      fmt.Sprintf("Register User: %s", username),
    }
    tx.ID = tx.GenerateID()
    return tx
}
