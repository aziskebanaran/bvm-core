package keeper

import (
	"github.com/aziskebanaran/bvm-core/x/bvm/types"
	"github.com/aziskebanaran/bvm-core/pkg/logger"
	"fmt"

)

func (k *Keeper) ProcessBlock(newBlock types.Block) error {
    // 1. Cek Fisik & Urutan
    if !k.VerifyBlock(newBlock) {
        return fmt.Errorf("❌ Struktur atau Mata Rantai tidak valid")
    }

    // 2. Cek Waktu
    if err := k.ValidateBlockTime(newBlock); err != nil {
        return err
    }

    // 3. Cek Ekonomi (Reward & Diff)
    if err := k.ValidateConsensus(newBlock); err != nil {
        return err
    }

    // 4. Validasi Transaksi
    if err := k.ValidateBlockTransactions(newBlock); err != nil {
        return err
    }

    // 5. Eksekusi Logika Internal (Hitung jatah final)
    if err := k.ExecuteBlock(newBlock); err != nil {
        return err
    }

    // 🚩 TAMBAHKAN INI: Update Statistik Global sebelum Patenkan ke Store
    k.UpdateGlobalStats(newBlock)

    // 6. Patenkan ke Store
    return k.CommitBlock(newBlock)
}


// ValidateConsensus: Menjamin keadilan Hybrid PoW + PoS
func (k *Keeper) ValidateConsensus(newBlock types.Block) error {
	// --- A. Cek Kesulitan (PoW + Boost) ---
	expectedDiff := int32(k.GetNextDifficultyForMiner(newBlock.Miner))
	if newBlock.Difficulty < expectedDiff {
		return fmt.Errorf("❌ Miner %s curang! Diff %d < Minimal %d",
			newBlock.Miner[:8], newBlock.Difficulty, expectedDiff)
	}

	// --- B. Cek Reward (Ekonomi BVM) ---
	vCount := k.GetValidatorCount()

        expectedSubsidi := k.GetSubsidiAtHeight(int64(newBlock.Index), vCount)
	if newBlock.Reward != expectedSubsidi {
		return fmt.Errorf("❌ Subsidi tidak valid! Blok minta %d, aturan bilang %d",
			newBlock.Reward, expectedSubsidi)
	}

	return nil
}

func (k *Keeper) ValidateBlockTransactions(block types.Block) error {
    isNewCycle := (block.Index % 10 == 1)

    if isNewCycle && block.Index > 1 {
        // Hanya verifikasi apakah anchor-nya ada, tidak perlu menulis lagi.
        var anchor string
        k.Store.Get(k.keyMeta("cycle_anchor"), &anchor)

        if anchor != "" {
            logger.Info("VALIDATOR", fmt.Sprintf("🛡️ Checkpoint #%d Sah via Anchor.", block.Index))
        }
    }

    // Cek integritas transaksi dasar
    for _, tx := range block.Transactions {
        if tx.ID == "" || tx.From == "" {
            return fmt.Errorf("❌ Struktur transaksi rusak")
        }
    }
    return nil
}

func (k *Keeper) VerifyBlock(newBlock types.Block) bool {
    // 1. Cek Fisik: Hash blok harus sesuai dengan isinya
    if newBlock.Hash != newBlock.CalculateBlockHash() {
        logger.Error("VALIDATOR", "❌ Hash blok tidak valid (Calculated != Provided)")
        return false
    }

    var lastHeight int64
    var lastHash string
    _ = k.Store.Get(k.keyMeta("height"), &lastHeight)
    _ = k.Store.Get(k.keyMeta("hash"), &lastHash)

    // Jika ini blok pertama (Genesis), langsung sahkan
    if lastHeight == 0 { return true }

    // 2. CEK APAKAH INI AWAL SIKLUS (Blok 11, 21, 31...)
    if newBlock.Index > 1 && (newBlock.Index-1)%10 == 0 {
        var anchor string
        _ = k.Store.Get(k.keyMeta("cycle_anchor"), &anchor)

        // Validasi: Blok #21 harus menyambung ke Anchor, bukan ke blok sembarang
        if newBlock.PrevHash != anchor {
            logger.Error("VALIDATOR", fmt.Sprintf("❌ Blok #%d tidak menyambung ke Anchor!", newBlock.Index))
            return false
        }
        return true // Sah sebagai awal siklus baru
    }

    // 3. VALIDASI RANTAI NORMAL (Blok 2-9, 12-19, dst)
    // Blok harus menyambung ke hash blok tepat di belakangnya
    if newBlock.PrevHash != lastHash {
        logger.Error("VALIDATOR", fmt.Sprintf("❌ Rantai Putus di #%d! Prev: %s... ≠ Last: %s...", 
            newBlock.Index, newBlock.PrevHash[:8], lastHash[:8]))
        return false
    }

    return true
}


func (k *Keeper) ValidateBlockTime(block types.Block) error {
    var lastHeight int64
    _ = k.Store.Get(k.keyMeta("height"), &lastHeight)

    if lastHeight == 0 { return nil } // Genesis bebas

    var lastBlock types.Block
    if err := k.Store.Get(k.keyBlock(lastHeight), &lastBlock); err != nil {
        return nil 
    }

    minDelay := k.Params.GetMinBlockDelay()
    actualDelay := block.Timestamp - lastBlock.Timestamp

    if block.Timestamp <= lastBlock.Timestamp {
        return fmt.Errorf("🚨 Pelanggaran Waktu: Timestamp mundur!")
    }

    if actualDelay < minDelay {
        return fmt.Errorf("🚨 Blok terlalu cepat! Butuh jeda %d detik", minDelay)
    }
    return nil
}

func (k *Keeper) UpdateGlobalStats(block types.Block) {
    var supply, burned uint64
    // Gunakan k.keyMeta agar konsisten dengan querier.go
    _ = k.Store.Get(k.keyMeta("supply"), &supply)
    _ = k.Store.Get(k.keyMeta("burned"), &burned)

    // 1. Tambah Supply (Subsidi Blok)
    supply += block.Reward

    // 2. Hitung Pembakaran (Accumulated Fees)
    for _, tx := range block.Transactions {
        burned += tx.Fee
    }

    // 3. Simpan permanen
    _ = k.Store.Put(k.keyMeta("supply"), supply)
    _ = k.Store.Put(k.keyMeta("burned"), burned)

    logger.Success("STATS", fmt.Sprintf("📊 Height #%d | 🔥 Burned: %d | 💰 Supply: %d", block.Index, burned, supply))
}
