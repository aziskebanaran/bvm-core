package sdk

// BVM-20: Standar Token Sultan (Terinspirasi dari ERC-20)
type IBVM20 interface {
    // Fungsi Wajib (Read-Only)
    TotalSupply() uint64
    BalanceOf(owner string) uint64
    Allowance(owner, spender string) uint64

    // Fungsi Wajib (State Changing)
    Transfer(to string, amount uint64) bool
    Approve(spender string, amount uint64) bool
    TransferFrom(from, to string, amount uint64) bool

    // Fungsi Opsional (Untuk Pengembang)
    Mint(to string, amount uint64) bool
    Burn(from string, amount uint64) bool
}

// Struktur Dasar Data Token agar seragam di DB
type TokenMetadata struct {
    Name     string `json:"name"`
    Symbol   string `json:"symbol"`
    Decimals uint8  `json:"decimals"`
    Owner    string `json:"owner"`
}
