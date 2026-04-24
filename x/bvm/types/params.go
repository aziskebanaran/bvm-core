package types

import (
	"fmt"
	"strconv"
        "strings"
)

const (
	// BVM_UNIT: 1 BVM = 10^8 (100.000.000) unit terkecil.
	// Ini menghilangkan penggunaan float64 dalam perhitungan saldo.
	BVM_UNIT uint64 = 100000000
)

// Params: Konstitusi Ekonomi & Teknis BVM
type Params struct {
	NetworkName     string `json:"network_name"`
	NativeSymbol    string `json:"native_symbol"`

	TargetBlockTime int `json:"target_block_time"`
	AdjustmentBlock int `json:"adjustment_block"`
	MinDifficulty   int `json:"min_difficulty"`

	// 💰 Kebijakan Moneter (Sekarang menggunakan uint64)
	InitialReward   uint64 `json:"initial_reward"`   // Satuan: Atomic Unit
	HalvingInterval int    `json:"halving_interval"` // Satuan: Jumlah Blok
	MaxSupply       uint64 `json:"max_supply"`       // Satuan: Atomic Unit

	CurrentBaseFee uint64 `json:"current_base_fee"`
	BurnAddress    string `json:"burn_address"`

	MaxValidators    int     `json:"max_validators"`
	MinStakeAmount   uint64  `json:"min_stake_amount"`
	UnbondingPeriod  int     `json:"unbonding_period"`
	AutoStakePercent float64 `json:"auto_stake_percent"`


	L2_BatchThreshold int     `json:"l2_batch_threshold"`
}

// DefaultParams: Inisialisasi dengan Rumus Matematika Murni
func DefaultParams() Params {
	// 🚩 VARIABEL INDUK
	blockTime := 60               // 60 detik
	baseRewardCoins := uint64(10) // 10 BVM per blok

	// 🚩 KONVERSI KE ATOMIC UNIT
	initialReward := baseRewardCoins * BVM_UNIT

	// 🚩 RUMUS OTOMATis
	// Halving setiap 1 tahun: (365 hari * 24 jam * 3600 detik) / blockTime
	oneYearInBlocks := (365 * 24 * 3600) / blockTime

	// 🚩 RUMUS MAX SUPPLY (Deret Geometri)
	// S_max = InitialReward * HalvingInterval * 2
	// Hasilnya akan selalu bulat dan presisi karena menggunakan uint64
	calculatedMaxSupply := initialReward * uint64(oneYearInBlocks) * 2

        baseDiff := 4
	    if blockTime < 30 {
	        baseDiff = 6 // Jika blok sangat cepat, persulit dari awal
	    }

	return Params{
		NetworkName:      "BVM Atomic Mainnet",
		NativeSymbol:     "BVM",
		TargetBlockTime:  blockTime,
		AdjustmentBlock:  20,
	        MinDifficulty:    baseDiff,

		// Moneter Otomatis
		InitialReward:    initialReward,
		HalvingInterval:  oneYearInBlocks,
		MaxSupply:        calculatedMaxSupply,

		// Fee: 0.001% dari Reward Awal (Misal 10 BVM -> 10.000 Unit)
		CurrentBaseFee:   10000,
		BurnAddress:      "bvmf000000000000000000000000000000000000burn",

		MaxValidators:    21,
		MinStakeAmount:   1000 * BVM_UNIT,
		UnbondingPeriod:  (14 * 24 * 3600) / blockTime,
		AutoStakePercent: 0.20,


		 L2_BatchThreshold: 100,
    }
}

// NetworkResponse: Struktur data gabungan untuk Miner & Wallet Sultan
type NetworkResponse struct {
    Params             Params  `json:"params"`
    Height             int64   `json:"height"`
    LatestHash         string  `json:"latest_hash"`
    Difficulty         int     `json:"difficulty"`

    // 🚩 PERUBAHAN: Sekarang menggunakan uint64
    // Wallet/Miner akan menerima angka seperti 1000000000 (10 BVM)
    Reward             uint64  `json:"reward"`
    DynamicFee         uint64  `json:"dynamic_fee"`

    MempoolSize        int     `json:"mempool_size"`
    NetworkName        string  `json:"network_name"`
}

// GetNative: Memberikan simbol utama jaringan jika input kosong
func (p Params) GetNative(symbol string) string {
    if symbol == "" {
        return p.NativeSymbol
    }
    return symbol
}

// --- 🛠️ FUNGSI EKONOMI (Helper untuk Audit & Miner) ---

func (p Params) DistributeFee(fee uint64) (tip uint64, burn uint64) {
    tip = fee / 2
    burn = fee - tip
    return
}

func (p Params) GetDynamicFee(mempoolSize int) uint64 {
    if mempoolSize > 50 {
        return (p.CurrentBaseFee * uint64(mempoolSize)) / 50
    }
    return p.CurrentBaseFee
}

func (p Params) GetExpectedSupply(height int64) uint64 {
    if height < 1 { return 0 }

    numHalvings := height / int64(p.HalvingInterval)
    interval := uint64(p.HalvingInterval)
    var totalAccumulated uint64 = 0

    // Batasi loop maksimal 64 (karena uint64 akan habis setelah 64 kali halving)
    limit := numHalvings
    if limit > 64 { limit = 64 }

    for i := int64(0); i < limit; i++ {
        // Reward setiap fase: Initial >> i
        rewardInPhase := p.InitialReward >> uint64(i)
        totalAccumulated += rewardInPhase * interval
    }

    // Hitung sisa blok di fase halving saat ini
    if numHalvings < 64 {
        currentReward := p.InitialReward >> uint64(numHalvings)
        blocksInCurrentInterval := uint64(height % int64(p.HalvingInterval))
        totalAccumulated += blocksInCurrentInterval * currentReward
    }

    return totalAccumulated
}

// GetBlockReward: Sekarang membagi reward berdasarkan jumlah pekerja aktif
func (p Params) GetBlockReward(height int64, activeValidators int) uint64 {
    if height < 0 { return 0 }

    // 1. Hitung jatah total blok pada fase ini (Halving logic)
    numHalvings := height / int64(p.HalvingInterval)
    if numHalvings >= 64 { return 0 }

    totalRewardForBlock := p.InitialReward >> uint64(numHalvings)

    // 2. 🚩 PEMBAGIAN ADIL: Bagi total jatah dengan jumlah validator
    // Jika tidak ada validator (safety check), gunakan 1 agar tidak crash (division by zero)
    if activeValidators <= 1 {
        return totalRewardForBlock
    }

    return totalRewardForBlock / uint64(activeValidators)
}


// --- 🧮 UNIT CONVERTER (The Bridge) ---

func (p Params) ToAtomic(amountStr string) uint64 {
    amountStr = strings.TrimSpace(amountStr)
    if amountStr == "" || amountStr == "." { return 0 }

    amountStr = strings.Replace(amountStr, ",", ".", -1)

    // Pastikan format ".5" menjadi "0.5"
    if strings.HasPrefix(amountStr, ".") {
        amountStr = "0" + amountStr
    }

    parts := strings.Split(amountStr, ".")
    var result uint64

    // 1. Bagian Depan
    if len(parts[0]) > 0 {
        whole, _ := strconv.ParseUint(parts[0], 10, 64)
        result = whole * BVM_UNIT
    }

    // 2. Bagian Belakang (Desimal)
    if len(parts) > 1 {
        fracStr := parts[1]
        if len(fracStr) > 8 {
            fracStr = fracStr[:8]
        }
        // Gunakan padding agar tepat 8 digit
        for len(fracStr) < 8 {
            fracStr += "0"
        }
        frac, _ := strconv.ParseUint(fracStr, 10, 64)
        result += frac
    }
    return result
}


// FormatDisplay: Memberikan string cantik "10.50000000"
// Sangat penting untuk Log dan Explorer agar Sultan mudah membaca saldo.
func (p Params) FormatDisplay(amount uint64) string {
    // Pastikan menggunakan BVM_UNIT dari konstanta
    return fmt.Sprintf("%d.%08d", amount/BVM_UNIT, amount%BVM_UNIT)
}

// GetMinBlockDelay: Menentukan jeda minimal antar blok (misal 1/10 dari target)
func (p Params) GetMinBlockDelay() int64 {
    // Jika Target 60 detik, maka minimal ada jeda 6 detik antar blok
    // Ini adalah "Satpam" otomatis agar database tidak bengkak
    return int64(p.TargetBlockTime / 10)
}

// GetAdjustmentFactor: Menghitung seberapa jauh penyimpangan waktu blok
func (p Params) GetAdjustmentFactor(actualTime, targetTime int64) int64 {
    if actualTime < targetTime {
        return 2 // Percepat kenaikan Difficulty
    }
    return 1
}

// GetUnit: Memberikan nilai 1 koin dalam unit terkecil (10^8)
func (p Params) GetUnit() uint64 {
    return BVM_UNIT
}

// IsHighPriority: Fungsi pusat untuk menentukan apakah sebuah fee masuk kategori "Paus"
func (p Params) IsHighPriority(fee uint64) bool {
    // Definisi prioritas: Misal jika fee >= 0.5 BVM
    // Kita hitung dinamis: (BVM_UNIT * 5) / 10
    threshold := (BVM_UNIT * 5) / 10 
    return fee >= threshold
}
