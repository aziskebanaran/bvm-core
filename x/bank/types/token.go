package types

// TokenMetadata menyimpan sertifikat kepemilikan token
type TokenMetadata struct {
    Symbol      string `json:"symbol"`
    Owner       string `json:"owner"`
    TotalSupply uint64 `json:"total_supply"`
    MintFee     uint64 `json:"mint_fee"` // Biaya BVM untuk setiap kali Mint
    IsFrozen    bool   `json:"is_frozen"`
}
