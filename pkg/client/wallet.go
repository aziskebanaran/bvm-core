package client

import (
	"bytes"
	 "encoding/json"
	"github.com/aziskebanaran/bvm-core/x/bvm/types"
	"fmt"
	"net/http"
)

// GetBalance: Mengambil saldo koin spesifik (Multi-Token)
func (c *BVMClient) GetBalance(address string, symbol string) (uint64, error) {
    url := fmt.Sprintf("%s/api/balance?address=%s&symbol=%s", c.BaseURL, address, symbol)
    
    resp, err := c.HTTP.Get(url)
    if err != nil {
        return 0, fmt.Errorf("🛰️ Kernel Offline: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        if resp.StatusCode == http.StatusNotFound { return 0, nil }
        return 0, fmt.Errorf("🚨 Kernel Error (Status: %d)", resp.StatusCode)
    }

    var result struct {
        Balance uint64 `json:"balance"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return 0, fmt.Errorf("❌ Gagal baca format saldo: %v", err)
    }

    return result.Balance, nil
}

// GetSecureState: Mengambil data saldo & nonce dari jalur VVIP Core
func (c *BVMClient) GetSecureState(address string) (*types.WalletState, error) {
    url := fmt.Sprintf("%s/api/balance?address=%s", c.BaseURL, address)
    resp, err := c.HTTP.Get(url)
    if err != nil { return nil, err }
    defer resp.Body.Close()

    // Gunakan types.WalletState yang kita buat di types.go tadi!
    var state types.WalletState
    if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
        return nil, fmt.Errorf("gagal parsing data core: %v", err)
    }

    return &state, nil
}


// GetAccount: Mengambil data akun lengkap (Termasuk Nonce & Map Balances)
func (c *BVMClient) GetAccount(address string) (*types.Account, error) {
    // 1. Tembak endpoint akun
    resp, err := c.HTTP.Get(fmt.Sprintf("%s/api/account?address=%s", c.BaseURL, address))
    if err != nil { return nil, err }
    defer resp.Body.Close()

    var acc types.Account
    if err := json.NewDecoder(resp.Body).Decode(&acc); err != nil {
        return nil, fmt.Errorf("format data dari Node tidak valid: %v", err)
    }

    // 🚩 2. VERIFIKASI INTEGRITAS (Pagar Betis Sultan)
    if acc.Address == "" {
        acc.Address = address
    }
    
    // Pastikan map balances terinisialisasi agar CLI tidak crash saat membaca acc.Balances["BVM"]
    if acc.Balances == nil {
        acc.Balances = make(map[string]uint64)
    }

    return &acc, nil
}

// GetNextNonce: Mengambil angka nonce terbaru yang diharapkan oleh Jenderal
func (c *BVMClient) GetNextNonce(address string) (uint64, error) {
    url := fmt.Sprintf("%s/api/nonce?address=%s", c.BaseURL, address)
    resp, err := c.HTTP.Get(url)
    if err != nil { return 0, err }
    defer resp.Body.Close()

    var result struct {
        Nonce uint64 `json:"nonce"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return 0, fmt.Errorf("gagal parsing nonce: %v", err)
    }

    return result.Nonce, nil
}

func (c *BVMClient) BroadcastTX(tx types.Transaction) (string, error) {
    payload, _ := json.Marshal(tx)
    resp, err := c.HTTP.Post(c.BaseURL+"/api/send", "application/json", bytes.NewBuffer(payload))
    if err != nil { return "", err }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        var errorResp struct { Message string `json:"message"` }
        json.NewDecoder(resp.Body).Decode(&errorResp)
        return "", fmt.Errorf("Node Menolak: %s", errorResp.Message)
    }

    var result struct {
        TxID string `json:"tx_id"`
    }
    json.NewDecoder(resp.Body).Decode(&result)
    return result.TxID, nil
}


func (c *BVMClient) GetHistory(address string) ([]types.Transaction, error) {
	resp, err := c.HTTP.Get(fmt.Sprintf("%s/api/history?address=%s", c.BaseURL, address))
	if err != nil { return nil, err }
	defer resp.Body.Close()

	var history []types.Transaction
	err = json.NewDecoder(resp.Body).Decode(&history)
	return history, err
}

