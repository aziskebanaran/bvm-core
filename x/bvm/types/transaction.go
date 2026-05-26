package types

import (
    "crypto/sha256"
    "encoding/hex"
    "os"
    "strconv"
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
    ChainID   uint64  `json:"chain_id"`
    Relayer   string  `json:"relayer"`
}

func (t Transaction) CalculateHash() ([]byte, error) {

  if t.ChainID == 0 {
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

	data := fmt.Sprintf("%s:%s:%d:%d:%s:%s:%d:%d:%s:%d:%s:%s:%s:%d",
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
	t.ChainID,                      // 14.
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


// 🎯 FUNGSI 1: NewTransaction
func NewTransaction(from, to string, amount, fee uint64, symbol, memo string, nonce uint64, chainID uint64, p Params) Transaction {
    // 🚀 INTERCEPTOR LIVE CORE: Paksa ikuti RAM lingkungan
    if envChainID := os.Getenv("BVM_CHAIN_ID"); envChainID != "" {
        if val, err := strconv.ParseUint(envChainID, 10, 64); err == nil {
            chainID = val
        }
    }

    tx := Transaction{
        From:      from,
        To:        to,
        Amount:    amount,
        Fee:       fee,
        Symbol:    p.GetNative(symbol),
        Memo:      memo,
        Nonce:     nonce,
        Timestamp: time.Now().Unix(),
        Layer:     1,
        Type:      "transfer",
        Payload:   []byte{},
        ZKP_Proof: "",
        ChainID:   chainID,
    }
    tx.ID = tx.GenerateID()
    return tx
}


// 🎯 FUNGSI 2: NewContractTransaction
func NewContractTransaction(from, to string, fee uint64, payload []byte, nonce uint64, chainID uint64) Transaction {
    if envChainID := os.Getenv("BVM_CHAIN_ID"); envChainID != "" {
        if val, err := strconv.ParseUint(envChainID, 10, 64); err == nil {
            chainID = val
        }
    }

    tx := Transaction{
        From:      from,
        To:        to,
        Amount:    0,
        Fee:       fee,
        Symbol:    "BVM",
        Nonce:     nonce,
        Timestamp: time.Now().Unix(),
        Type:      "contract_call",
        Payload:   payload,
        Layer:     1,
        ChainID:   chainID,
    }
    tx.ID = tx.GenerateID()
    return tx
}


// 🎯 FUNGSI 3: NewRegisterTransaction
func NewRegisterTransaction(from, username string, fee uint64, nonce uint64, chainID uint64) Transaction {
    if envChainID := os.Getenv("BVM_CHAIN_ID"); envChainID != "" {
        if val, err := strconv.ParseUint(envChainID, 10, 64); err == nil {
            chainID = val
        }
    }

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
        ChainID:   chainID,
    }
    tx.ID = tx.GenerateID()
    return tx
}


// 🎯 FUNGSI 4: NewUTXOMoveTransaction
func NewUTXOMoveTransaction(from, to string, amount, fee uint64, symbol string, inputs []string, memo string, nonce uint64, timestamp int64, chainID uint64) Transaction {
    if envChainID := os.Getenv("BVM_CHAIN_ID"); envChainID != "" {
        if val, err := strconv.ParseUint(envChainID, 10, 64); err == nil {
            chainID = val
        }
    }

    inputStr := ""
    for i, in := range inputs {
        inputStr += fmt.Sprintf("\"%s\"", in)
        if i < len(inputs)-1 {
            inputStr += ","
        }
    }
    payloadStr := fmt.Sprintf("{\"inputs\":[%s]}", inputStr)
    payload := []byte(payloadStr)

    tx := Transaction{
        From:      from,
        To:        to,
        Amount:    amount,
        Fee:       fee,
        Symbol:    symbol,
        Memo:      memo,
        Nonce:     nonce,
        Timestamp: timestamp,
        Type:      "utxo_move",
        Layer:     2,
        Payload:   payload,
        ChainID:   chainID,
    }
    tx.ID = tx.GenerateID()
    return tx
}

// 🎯 FUNGSI 5: NewRelayerTransaction
func NewRelayerTransaction(mempoolAddr string, actionType string, payload []byte, nonce uint64, chainID uint64) Transaction {
    if envChainID := os.Getenv("BVM_CHAIN_ID"); envChainID != "" {
        if val, err := strconv.ParseUint(envChainID, 10, 64); err == nil {
            chainID = val
        }
    }

    tx := Transaction{
        From:      mempoolAddr,
        To:        "SYSTEM_REWARD_HUB",
        Amount:    0,
        Fee:       0,
        Symbol:    "BVM",
        Nonce:     nonce,
        Timestamp: time.Now().Unix(),
        Type:      "mempool_report",
        Layer:     1,
        Payload:   payload,
        ChainID:   chainID,
        Relayer:   mempoolAddr,
    }
    tx.ID = tx.GenerateID()
    return tx
}

// 🎯 FUNGSI 6: NewBridgeOutTransaction (User berangkat)
func NewBridgeOutTransaction(from, to string, amount, fee uint64, symbol string, targetChainID uint64, nonce uint64, currentChainID uint64) Transaction {
    if envChainID := os.Getenv("BVM_CHAIN_ID"); envChainID != "" {
        if val, err := strconv.ParseUint(envChainID, 10, 64); err == nil {
            currentChainID = val
        }
    }

    // Payload menyimpan informasi tiket tujuan
    payloadStr := fmt.Sprintf(`{"target_chain_id":%d}`, targetChainID)
    
    tx := Transaction{
        From:      from,
        To:        to,     // Alamat tujuan di chain seberang
        Amount:    amount,
        Fee:       fee,
        Symbol:    symbol,
        Nonce:     nonce,
        Timestamp: time.Now().Unix(),
        Type:      "bridge_out", // 🚩 Langsung tulis string-nya di sini
        Layer:     1,
        Payload:   []byte(payloadStr),
        ChainID:   currentChainID,
    }
    tx.ID = tx.GenerateID()
    return tx
}

// 🎯 FUNGSI 7: NewBridgeInTransaction (Paket tiba dari Nexus)
func NewBridgeInTransaction(nexusAddr, to string, amount uint64, symbol string, refTxID string, nonce uint64, currentChainID uint64) Transaction {
    if envChainID := os.Getenv("BVM_CHAIN_ID"); envChainID != "" {
        if val, err := strconv.ParseUint(envChainID, 10, 64); err == nil {
            currentChainID = val
        }
    }

    // Payload menyimpan bukti ID Transaksi dari chain asal untuk cegah double-claim
    payloadStr := fmt.Sprintf(`{"ref_tx_id":"%s"}`, refTxID)

    tx := Transaction{
        From:      nexusAddr, // Yang menandatangani adalah Nexus Relayer!
        To:        to,        // Penerima akhir
        Amount:    amount,
        Fee:       0,         // Fee digratiskan karena sudah dibayar di rantai asal
        Symbol:    symbol,
        Nonce:     nonce,
        Timestamp: time.Now().Unix(),
        Type:      "bridge_in", // 🚩 Langsung tulis string-nya di sini
        Layer:     1,
        Payload:   []byte(payloadStr),
        ChainID:   currentChainID,
        Relayer:   nexusAddr,
    }
    tx.ID = tx.GenerateID()
    return tx
}

// 🎯 FUNGSI 8: NewSubmitProofTransaction (Bukti dari Nexus ke BVM)
func NewSubmitProofTransaction(relayerAddr string, refTxID string, sourceChainID uint64, nonce uint64, chainID uint64) Transaction {
    if envChainID := os.Getenv("BVM_CHAIN_ID"); envChainID != "" {
        if val, err := strconv.ParseUint(envChainID, 10, 64); err == nil {
            chainID = val
        }
    }

    // Payload menyimpan bukti bahwa transaksi di rantai asal sudah dikunci
    payloadStr := fmt.Sprintf(`{"ref_tx_id":"%s", "source_chain_id":%d}`, refTxID, sourceChainID)

    tx := Transaction{
        From:      relayerAddr,
        To:        "SYSTEM_BRIDGE", // Alamat sistem untuk memproses bukti
        Amount:    0,
        Fee:       0,
        Symbol:    "BVM",
        Nonce:     nonce,
        Timestamp: time.Now().Unix(),
        Type:      "submit_proof", // 🚩 Langsung menggunakan string sesuai pola fungsi lainnya
        Layer:     1,
        Payload:   []byte(payloadStr),
        ChainID:   chainID,
        Relayer:   relayerAddr,
    }
    tx.ID = tx.GenerateID()
    return tx
}
