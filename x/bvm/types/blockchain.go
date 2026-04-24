package types

import (
    "encoding/json"
    "os"
    "sync"
)

type Blockchain struct {
    Chain      []Block
    Mempool    []Transaction
    Params     Params
    Difficulty int
    Mu         sync.RWMutex
    Height     int64
    InFlight   int64
    LatestHash string

    // 🚩 PENGUATAN SULTAN: Kantong Darurat (Buffer RAM)
    // Ini tempat menyimpan blok jika database sedang macet/error
    PendingBlocks []Block 
}

func NewBlockchain() *Blockchain {
    params := DefaultParams()
    bc := &Blockchain{
        Chain:         make([]Block, 0), // Gunakan make untuk alokasi memori yang bersih
        Mempool:       []Transaction{},
        Params:        params,
        Difficulty:    params.MinDifficulty,
        Height:        0,
        InFlight:      0,
        LatestHash:    "0000000000000000000000000000000000000000000000000000000000000000",
        PendingBlocks: []Block{}, // Inisialisasi kantong kosong
    }

    // 1. Coba muat dari genesis.json (External Genesis)
    file, err := os.ReadFile("genesis.json")
    if err == nil {
        var genesisData struct {
            Params       Params `json:"params"`
            GenesisBlock Block  `json:"genesis_block"`
        }
        if err := json.Unmarshal(file, &genesisData); err == nil {
            bc.Params = genesisData.Params
            bc.Difficulty = genesisData.Params.MinDifficulty
            if len(bc.Chain) == 0 {
                // Hitung ulang hash genesis untuk memastikan integritas
                genesisData.GenesisBlock.Hash = genesisData.GenesisBlock.CalculateBlockHash()
                bc.Chain = append(bc.Chain, genesisData.GenesisBlock)
                bc.Height = int64(genesisData.GenesisBlock.Index)
                bc.LatestHash = genesisData.GenesisBlock.Hash
            }
        }
    }

    // 2. Jika masih kosong, buat Genesis Block default (Hardcoded)
    if len(bc.Chain) == 0 {
        genesis := Block{
            Version:      1,
            Index:        0,
            Timestamp:    1742854200,
            Transactions: []Transaction{},
            PrevHash:     "0000000000000000000000000000000000000000000000000000000000000000",
            Difficulty:   int32(bc.Params.MinDifficulty),
            Reward:       bc.Params.InitialReward,
            Miner:        "BVM_GENESIS",
            TotalFee:     0,
            MerkleRoot:   "0000000000000000000000000000000000000000000000000000000000000000",
            StateRoot:    "0000000000000000000000000000000000000000000000000000000000000000",
        }
        // 🚩 PENTING: Hitung hash asli genesis blok ke-0
        genesis.Hash = genesis.CalculateBlockHash()

        bc.Chain = append(bc.Chain, genesis)
        bc.Height = 0
        bc.LatestHash = genesis.Hash
    }

    return bc
}

type StateReader interface {
    Get(key string, v interface{}) error
}

// LoadState: Sinkronisasi awal dengan database
func (bc *Blockchain) LoadState(store StateReader) {
    var height uint64
    // 🚩 KOREKSI: Gunakan key m:height (Key Meta) agar sama dengan keeper/execution.go
    if err := store.Get("m:height", &height); err == nil {
        bc.Height = int64(height)
    } else {
        bc.Height = 0 
    }

    // Sinkronisasi LatestHash juga dari database (m:hash)
    var lastHash string
    if err := store.Get("m:hash", &lastHash); err == nil {
        bc.LatestHash = lastHash
    }
}
