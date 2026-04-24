package keeper

import (
	"github.com/aziskebanaran/bvm-core/x/bvm/types"
)

func (k *Keeper) GetStatus() types.NodeStatus {
    var h int64
    var supply, burned uint64
    var lastHash string

    // 1. Tarik Data Utama Langsung dari Disk (Stateless)
    _ = k.Store.Get(k.keyMeta("height"), &h)
    _ = k.Store.Get(k.keyMeta("supply"), &supply)
    _ = k.Store.Get(k.keyMeta("burned"), &burned)
    _ = k.Store.Get(k.keyMeta("hash"), &lastHash)

    // 2. Ambil Timestamp dari Blok Terakhir di Disk
    var lastTs int64 = 0
    if h > 0 {
        var lastBlock types.Block
        // Ambil blok pada height saat ini dari prefix "b:"
        if err := k.Store.Get(k.keyBlock(h), &lastBlock); err == nil {
            lastTs = lastBlock.Timestamp
        }
    }

    // 3. Susun Status Node
	vCount := k.GetValidatorCount()

    return types.NodeStatus{
        Status:             "Online",
        Height:             h,
        LatestHash:         lastHash,
        LastBlockTimestamp: lastTs,

        // Kalkulasi dinamis berdasarkan Height di Disk
        Difficulty:         int32(k.GetNextDifficulty()),
        Reward:             k.GetSubsidiAtHeight(h+1, vCount),

        TotalSupply:        supply,
        TotalBurned:        burned,
        PeerCount:          k.P2P.CountActive(),
        Version:            1,
    }
}

func (k *Keeper) GetChain() []types.Block {
    // ✅ CARA SULTAN: Gunakan fungsi GetLatestBlocks yang efisien
    blocks, err := k.Store.GetLatestBlocks(10) // Ambil 10 blok terakhir saja
    if err != nil {
        return []types.Block{}
    }
    return blocks
}

func (k *Keeper) GetUserHistory(address string) []types.Transaction {
    // ✅ CARA SULTAN: Ambil langsung dari index history
    history, err := k.Store.GetAddressHistory(address)
    if err != nil {
        return []types.Transaction{}
    }
    return history
}

func (k *Keeper) GetBlockByHeight(height uint64) (*types.Block, error) {
    var block types.Block
    // Gunakan k.keyBlock(int64(height)) untuk mendapatkan prefix key "b:[height]"
    err := k.Store.Get(k.keyBlock(int64(height)), &block)
    if err != nil {
        return nil, err
    }

    return &block, nil
}
