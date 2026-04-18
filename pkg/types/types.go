package types

// Account: Versi SDK untuk informasi pengguna
type Account struct {
	Address    string            `json:"address"`
	Username   string            `json:"username"`
	Balances   map[string]uint64 `json:"balances"`
	Nonce      uint64            `json:"nonce"`
	IsContract bool              `json:"is_contract"`
	Status     string            `json:"status"`
}

// Block: Diperlukan oleh Miner & Explorer di sisi SDK
type Block struct {
	Version      int32         `json:"version"`
	Index        int64         `json:"index"`
	Timestamp    int64         `json:"timestamp"`
	Difficulty   int32         `json:"difficulty"`
	Nonce        int32         `json:"nonce"`
	Reward       uint64        `json:"reward"`
	TotalFee     uint64        `json:"total_fee"`
	PrevHash     string        `json:"prev_hash"`
	MerkleRoot   string        `json:"merkle_root"`
	Miner        string        `json:"miner"`
	MinerName    string        `json:"miner_name"`
	StateRoot    string        `json:"state_root"`
	Hash         string        `json:"hash"`
	Transactions []Transaction `json:"transactions"`
	ParentHash   string        `json:"parent_hash"`
	Size         int           `json:"size"`
}

// NodeStatus: Informasi dashboard jaringan
type NodeStatus struct {
	Status             string `json:"status"`
	Height             int64  `json:"height"`
	LatestHash         string `json:"latest_hash"`
	LastBlockTimestamp int64  `json:"last_block_timestamp"`
	Difficulty         int32  `json:"difficulty"`
	Reward             uint64 `json:"reward"`
	TotalSupply        uint64 `json:"total_supply"`
	TotalBurned        uint64 `json:"total_burned"`
	StateRoot          string `json:"state_root"`
	TargetDifficulty   int    `json:"target_difficulty"`
	PeerCount          int    `json:"peer_count"`
	ValidatorCount     int    `json:"validator_count"`
	Version            int32  `json:"version"`
	InFlight           int64  `json:"in_flight"`
}

// WalletState: Digunakan oleh Wallet/Nexus untuk sinkronisasi cepat
type WalletState struct {
	Address        string `json:"address"`
	BalanceAtomic  uint64 `json:"balance_atomic"`
	BalanceDisplay string `json:"balance_display"`
	Nonce          uint64 `json:"nonce"`
	Symbol         string `json:"symbol"`
	Status         string `json:"status,omitempty"`
}

// Params: Konstitusi Ekonomi & Teknis BVM (Versi SDK)
type Params struct {
	NetworkName       string  `json:"network_name"`
	NativeSymbol      string  `json:"native_symbol"`
	TargetBlockTime   int     `json:"target_block_time"`
	AdjustmentBlock   int     `json:"adjustment_block"`
	MinDifficulty     int     `json:"min_difficulty"`
	InitialReward     uint64  `json:"initial_reward"`
	HalvingInterval   int     `json:"halving_interval"`
	MaxSupply         uint64  `json:"max_supply"`
	CurrentBaseFee    uint64  `json:"current_base_fee"`
	BurnAddress       string  `json:"burn_address"`
	MaxValidators     int     `json:"max_validators"`
	MinStakeAmount    uint64  `json:"min_stake_amount"`
	UnbondingPeriod   int     `json:"unbonding_period"`
	AutoStakePercent  float64 `json:"auto_stake_percent"`
	L2_BatchThreshold int     `json:"l2_batch_threshold"`
}

// NetworkResponse: Struktur data gabungan untuk Miner & Wallet Sultan (Versi SDK)
type NetworkResponse struct {
	Params      Params  `json:"params"`
	Height      int64   `json:"height"`
	LatestHash  string  `json:"latest_hash"`
	Difficulty  int     `json:"difficulty"`
	Reward      uint64  `json:"reward"`
	DynamicFee  uint64  `json:"dynamic_fee"`
	MempoolSize int     `json:"mempool_size"`
	NetworkName string  `json:"network_name"`
}
