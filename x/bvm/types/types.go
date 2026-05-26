package types


type Account struct {
    Address       string            `json:"address"`
    Username      string            `json:"username"`
    EthAddress    string            `json:"eth_address,omitempty"` // 🚩 Tambahkan untuk mapping 0x
    Balances      map[string]uint64 `json:"balances"`
    Nonce         uint64            `json:"nonce"`
    IsContract    bool              `json:"is_contract"`
    Status        string            `json:"status"`       // "active", "frozen"
    LastActivity  int64             `json:"last_activity"` // Untuk statistik wallet
}


type Block struct {
    // --- KELOMPOK 1: FIXED-SIZE (ANGKA) ---
    Version      int32         `json:"version"`
    Index        int64         `json:"index"`
    Timestamp    int64         `json:"timestamp"`
    Difficulty   int32         `json:"difficulty"`
    Nonce        int32         `json:"nonce"`
    Reward       uint64        `json:"reward"`
    TotalFee     uint64        `json:"total_fee"`

    // --- KELOMPOK 2: VARIABLE-SIZE (STRING) ---
    PrevHash     string        `json:"prev_hash"`
    MerkleRoot   string        `json:"merkle_root"`
    Miner        string        `json:"miner"`
    MinerName    string        `json:"miner_name"`
    StateRoot    string        `json:"state_root"`

    // --- DATA TAMBAHAN (TIDAK MASUK HASH) ---
    Hash         string        `json:"hash"`
    Transactions []Transaction `json:"transactions"`
    ParentHash   string        `json:"parent_hash"`
    Size         int           `json:"size"`
}

type NodeStatus struct {
    Status           string  `json:"status"`
    Height           int64   `json:"height"`
    LatestHash       string  `json:"latest_hash"`
    LastBlockTimestamp int64 `json:"last_block_timestamp"`

    Difficulty       int32   `json:"difficulty"`
    Reward           uint64  `json:"reward"`

    TotalSupply      uint64  `json:"total_supply"`
    TotalBurned      uint64  `json:"total_burned"`
    StateRoot        string  `json:"state_root"` // 👈 TAMBAHKAN INI (Wajib untuk WASM)

    TargetDifficulty int     `json:"target_difficulty"`
    PeerCount        int     `json:"peer_count"`
    ValidatorCount   int     `json:"validator_count"`
    Version          int32   `json:"version"`     // 👈 UBAH STRING KE INT (Biar sama dengan Block)

    InFlight         int64   `json:"in_flight"`
}

// WASMContract: Blueprint untuk aplikasi yang berjalan di atas BVM
type WASMContract struct {
    ContractAddr string `json:"contract_addr"`
    Owner        string `json:"owner"`
    CodeHash     string `json:"code_hash"` // Hash dari Bytecode WASM
    StateRoot    string `json:"state_root"` // Snapshot data contract
}


type UserProfile struct {
    Username  string `json:"username"`
    Address   string `json:"address"`
    CreatedAt int64  `json:"created_at"`
    Tier      string `json:"tier"` // Misal: "Basic", "Pro", "Merchant"
}
