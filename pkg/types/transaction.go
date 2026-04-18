package types

import (
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "time"
)

// Transaction versi SDK (Tanpa Merkle Tree)
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

// CalculateHash: WAJIB SAMA PERSIS dengan versi x/ agar tanda tangan valid!
func (t Transaction) CalculateHash() ([]byte, error) {
    data := fmt.Sprintf("%s:%s:%d:%d:%s:%s:%d:%d:%s:%d:%s:%s:%s",
        t.From,
        t.To,
        t.Amount,
        t.Fee,
        t.Symbol,
        t.Memo,
        t.Nonce,
        t.Timestamp,
        t.Type,
        t.Layer,
        hex.EncodeToString(t.Payload),
        t.ZKP_Proof,
        t.PublicKey,
    )
    h := sha256.Sum256([]byte(data))
    return h[:], nil
}

func (t *Transaction) GenerateID() string {
    hash, _ := t.CalculateHash()
    return hex.EncodeToString(hash)
}

// NewTransaction versi Ringan untuk SDK
func NewTransaction(from, to string, amount, fee uint64, symbol, memo string, nonce uint64) Transaction {
    tx := Transaction{
        From:      from,
        To:        to,
        Amount:    amount,
        Fee:       fee,
        Symbol:    symbol, // Di SDK kita isi string langsung
        Memo:      memo,
        Nonce:     nonce,
        Timestamp: time.Now().Unix(),
        Layer:     1,
        Type:      "transfer",
        Payload:   []byte{},
    }
    tx.ID = tx.GenerateID()
    return tx
}

// Tambahkan ini ke pkg/types/transaction.go jika belum ada
func NewRegisterTransaction(from, username string, fee uint64, nonce uint64) Transaction {
    payload := []byte(fmt.Sprintf(`{"username":"%s"}`, username))
    tx := Transaction{
        From:      from,
        To:        "SYSTEM_AUTH",
        Amount:    0,
        Fee:       fee,
        Symbol:    "BVM",
        Nonce:     nonce,
        Timestamp: time.Now().Unix(),
        Type:      "user_register",
        Payload:   payload,
        Layer:     1,
        Memo:      fmt.Sprintf("Register User: %s", username),
    }
    tx.ID = tx.GenerateID()
    return tx
}
