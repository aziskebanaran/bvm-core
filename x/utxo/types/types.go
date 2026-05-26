package types

import "fmt"

// UTXO: Jantung Ekonomi Nexus (BVM + 0x + Social)
type UTXO struct {
    TxID      string `json:"tx_id" msgpack:"tx_id"`
    Index     int    `json:"index" msgpack:"index"`
    Address   string `json:"address" msgpack:"address"`
    Address0x string `json:"address_0x" msgpack:"address_0x"`
    Username  string `json:"username,omitempty" msgpack:"username,omitempty"`
    Symbol    string `json:"symbol" msgpack:"symbol"`
    Amount    uint64 `json:"amount" msgpack:"amount"`
    AssetID   string `json:"asset_id,omitempty" msgpack:"asset_id,omitempty"`
    Metadata  string `json:"metadata,omitempty" msgpack:"metadata,omitempty"`
    Type      string `json:"type" msgpack:"type"`
    Status    string `json:"status" msgpack:"status"` 
    Timestamp int64  `json:"timestamp" msgpack:"timestamp"`
}

// GetID: Memberikan koordinat unik kepingan di Database
func (u UTXO) GetID() string {
    return fmt.Sprintf("%s-%d", u.TxID, u.Index)
}

type UTXOData struct {
    Inputs  []string `json:"inputs"`  // Daftar ID kepingan (TxID-Index)
    Outputs []uint64 `json:"outputs"` // Nilai kepingan baru yang akan dicetak
}

// Transaction: Membungkus mutasi kepingan (Hancur -> Lahir)
type Transaction struct {
    ID        string   `json:"id" msgpack:"id"`
    Inputs    []Input  `json:"inputs" msgpack:"inputs"`
    Outputs   []Output `json:"outputs" msgpack:"outputs"`
    Fee       uint64   `json:"fee" msgpack:"fee"`
    Timestamp int64    `json:"timestamp" msgpack:"timestamp"`
}

type Input struct {
    PrevTxID  string `json:"prev_txid" msgpack:"prev_txid"`
    Index     int    `json:"index" msgpack:"index"`
    Signature string `json:"signature" msgpack:"signature"`
}

type Output struct {
    To     string `json:"to" msgpack:"to"` // Bisa berupa bvmf, 0x, atau @username
    Amount uint64 `json:"amount" msgpack:"amount"`
}
