package mempool

import (
	"github.com/aziskebanaran/BVM.core/x/bvm/types"
    "sync"
    "fmt"
)

// MempoolStats: Raport kesehatan antrean Sultan
type MempoolStats struct {
    TotalTransactions int     `json:"total_tx"`
    TotalVolume       uint64 `json:"total_volume"` // Total BVM yang sedang dikirim
    TotalFees         uint64 `json:"total_fees"`   // Potensi keuntungan Miner
    AverageFee        uint64 `json:"average_fee"`
    Mu                sync.RWMutex
}

// GetStats: Menghitung statistik secara mendalam dari Heap
func (m *MempoolEngine) GetStats() MempoolStats {

    m.Mu.RLock() // 🚩 Pasang gembok baca
    defer m.Mu.RUnlock()

    stats := MempoolStats{
        TotalTransactions: m.Queue.Len(),
    }

    if stats.TotalTransactions == 0 {
        return stats
    }

    var sumVolume uint64
    var sumFees uint64

    // Iterasi seluruh isi Heap Sultan
    for _, tx := range *m.Queue {
        sumVolume += tx.Amount
        sumFees += tx.Fee
    }

    stats.TotalVolume = sumVolume
    stats.TotalFees = sumFees
    stats.AverageFee = sumFees / uint64(stats.TotalTransactions)

    return stats
}

func (m *MempoolEngine) CheckPriority(tx types.Transaction) {
    // 🚩 Sekarang panggil langsung dari Params yang sudah ada di dalam Engine
    if m.Params != nil && m.Params.IsHighPriority(tx.Fee) {
        fmt.Printf("\n🚀 [PRIORITY ALERT] Transaksi Paus Terdeteksi!\n")
        fmt.Printf("   💎 Pengirim : %s\n", tx.From)
        fmt.Printf("   💰 Jumlah   : %s\n", m.Params.FormatDisplay(tx.Amount)) // Pakai FormatDisplay biar cantik
        fmt.Printf("   💸 Fee      : %s (Prioritas Tinggi)\n", m.Params.FormatDisplay(tx.Fee))
        fmt.Printf("--------------------------------------------\n")
    }
}
