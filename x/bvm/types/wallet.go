package types

import (
	"fmt"
	"strings"
)

type WalletState struct {
        Address        string            `json:"address"`
        Username       string            `json:"username,omitempty"`
        EthAddress     string            `json:"eth_address,omitempty"`
        MappedAddress  string            `json:"mapped_address,omitempty"` // Alias 0x
        ActivityCount  int               `json:"activity_count,omitempty"` // Total transaksi history
        LastSeen       int64             `json:"last_seen,omitempty"`      // Timestamp terakhir aktif
        LastMemo       string            `json:"last_memo,omitempty"`      // Memo terakhir
        BalanceAtomic  uint64            `json:"balance_atomic"`
        BalanceDisplay string            `json:"balance_display"`
        Symbol         string            `json:"symbol"`
        Nonce          uint64            `json:"nonce"`
        Status         string            `json:"status"`
        UTXOCount      int               `json:"utxo_count,omitempty"`
        Type           string            `json:"type"`
        Assets         map[string]uint64 `json:"assets,omitempty"`
}

// --- FUNGSI GENERATOR (FACTORY) ---
// NewBVMWalletState: Menggunakan symbol dari Params
func NewBVMWalletState(acc Account, displayBal string, symbol string) WalletState {
        return WalletState{
                Address:        acc.Address,
                Username:       acc.Username,
                EthAddress:     acc.EthAddress,
                BalanceAtomic:  acc.Balances[symbol], // ✅ Dinamis mencari di map balances
                BalanceDisplay: displayBal,
                Symbol:         symbol,               // ✅ Dinamis
                Nonce:          acc.Nonce,
                Status:         acc.Status,
                Type:           "ACCOUNT_BASED",
                Assets:         acc.Balances,
        }
}

// NewUTXOWalletState: Menggunakan symbol dari Params
func NewUTXOWalletState(acc Account, utxoBal uint64, displayBal string, count int, symbol string) WalletState {
        addr := acc.Address
        if acc.EthAddress != "" {
                addr = acc.EthAddress
        }

        return WalletState{
                Address:        addr,
                Username:       acc.Username,
                EthAddress:     acc.EthAddress,
                BalanceAtomic:  utxoBal,
                BalanceDisplay: displayBal,
                Symbol:         symbol, // ✅ Dinamis
                UTXOCount:      count,
                Status:         acc.Status,
                Type:           "UTXO_BASED",
        }
}

// --- FUNGSI HELPER (UTILITY) ---

// IsNexusReady: Mengecek apakah wallet sudah terhubung ke ekosistem 0x
func (ws WalletState) IsNexusReady() bool {
	return ws.EthAddress != "" || strings.HasPrefix(ws.Address, "0x")
}

// Summary: Memberikan ringkasan singkat untuk log atau chat AI
func (ws WalletState) Summary() string {
	return fmt.Sprintf("[%s] %s (%s): %s %s", 
		ws.Type, ws.Username, ws.Address[:10], ws.BalanceDisplay, ws.Symbol)
}
