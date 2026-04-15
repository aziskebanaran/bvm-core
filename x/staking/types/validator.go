package types

// Validator Lengkap (Gabungan Versi Core & Staking)
type Validator struct {
    Address      string  `json:"address"`
    PubKey       string  `json:"pub_key"`
    StakedAmount uint64  `json:"staked_amount"`
    SelfStake    uint64  `json:"self_stake"`
    Power        int64   `json:"power"`      // Untuk logika DPoS
    Commission   float64 `json:"commission"`
    IsActive     bool    `json:"is_active"`
    Status       string  `json:"status"`     // "Active", "Jailed", "Candidate"
}

type ValidatorSet struct {
    Validators []Validator `json:"validators"`
    UpdatedAt  int64       `json:"updated_at"`
}
