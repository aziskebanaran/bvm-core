package types

type MsgCreateToken struct {
    Owner       string `json:"owner"`        // Alamat Sultan/User
    Symbol      string `json:"symbol"`       // Contoh: "GOLD", "SAHAM"
    TotalSupply uint64 `json:"total_supply"` // Jumlah yang ingin dicetak
}
