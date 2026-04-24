package keeper

import (
    "github.com/aziskebanaran/bvm-core/pkg/logger"
    "github.com/aziskebanaran/bvm-core/x/bvm/types"
    "fmt"
)

func (k *Keeper) SyncState() uint64 {
    // 🚩 1. AMBIL KENYATAAN DARI DISK (Gunakan Key Baru)
    var diskHeight uint64
    err := k.Store.Get("m:height", &diskHeight)
    if err != nil {
        // Fallback ke key lama jika sedang migrasi, atau nol jika benar-benar baru
        _ = k.Store.Get("last_indexed_height", &diskHeight)
    }

    // 2. KUNCI TINGGI BLOK DI RAM
    k.Blockchain.Height = int64(diskHeight)

    // 🚩 3. AUDIT EKONOMI (Dinamis via Params)
    // Jangan gunakan angka 10 manual, gunakan fungsi Halving Sultan
    expectedSupply := k.Params.GetExpectedSupply(int64(diskHeight))

    if k.TotalSupplyBVM != expectedSupply {
        logger.Warning("SYNC", fmt.Sprintf("⚖️ Sinkronisasi Suplai: %d -> %d", k.TotalSupplyBVM, expectedSupply))
        k.TotalSupplyBVM = expectedSupply
        _ = k.Store.Put("m:supply", k.TotalSupplyBVM)
    }

    // 4. UPDATE HASH TERAKHIR (Key Konsisten)
    if diskHeight > 0 {
        var lastBlock types.Block
        blockKey := fmt.Sprintf("b:%d", diskHeight)
        err := k.Store.Get(blockKey, &lastBlock)

        if err == nil && lastBlock.Hash != "" {
            k.Blockchain.LatestHash = lastBlock.Hash
            // Pastikan Chain di RAM tidak kosong
            k.Blockchain.Chain = []types.Block{lastBlock}
        } else {
            logger.Error("SYNC", "❌ Data Blok Terakhir Hilang di Key "+blockKey)
        }
    } else {
        k.Blockchain.LatestHash = "0000000000000000000000000000000000000000000000000000000000000000"
    }

    logger.Success("SYNC", fmt.Sprintf("Kernel BVM Siap di Blok #%d", diskHeight))
    return diskHeight
}

func (k *Keeper) IndexBlock(block types.Block) {
    // 🚩 STRATEGI INDEXING INDUSTRI: 
    // Kita gunakan prefix "h:" (History) yang lebih pendek untuk hemat disk

    batch := k.Store.NewBatch()

    for _, tx := range block.Transactions {
        if tx.ID == "" {
            tx.ID = tx.GenerateID()
        }

        // PENGARSIPAN RIWAYAT (History Index)
        // Format: h:[address]:[txid]
        k.Store.PutToBatch(batch, "h:"+tx.From+":"+tx.ID, tx)
        k.Store.PutToBatch(batch, "h:"+tx.To+":"+tx.ID, tx)
    }

    // Catat bahwa blok ini sudah masuk index sejarah internal
    k.Store.PutToBatch(batch, "m:last_indexed_internal", uint64(block.Index))

    // Ledakkan sekaligus agar Atomic!
    _ = k.Store.WriteBatch(batch)
}


func (k *Keeper) AutoRecoverDatabase() {
    if len(k.Blockchain.PendingBlocks) == 0 {
        return
    }

    logger.Info("SYSTEM", fmt.Sprintf("🔧 Memulihkan %d blok & Menghitung ulang saldo...", len(k.Blockchain.PendingBlocks)))

    for _, block := range k.Blockchain.PendingBlocks {
        // 🚩 JANGAN cuma tulis data mentah. 
        // Panggil ExecuteBlock agar Saldo, Nonce, dan Supply ikut diperbarui di Disk.
        if err := k.ExecuteBlock(block); err != nil {
            logger.Error("SYSTEM", fmt.Sprintf("❌ Gagal eksekusi ulang blok #%d: %v", block.Index, err))
            return 
        }
    }

    // Jika semua sukses dieksekusi ke Disk, baru kosongkan RAM
    k.Blockchain.PendingBlocks = nil
    logger.Success("SYSTEM", "✅ Database sinkron! Semua transaksi telah dipahat ulang ke saldo.")
}
