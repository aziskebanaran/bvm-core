package mempool

import (
	"github.com/aziskebanaran/bvm-core/x/bvm/types"
	"context"
)

type Mempool struct {
    Engine *MempoolEngine
}

// 🚩 Ubah fungsi NewMempool Sultan menjadi:
func NewMempool(params *types.Params, k MempoolKeeper) *Mempool {
    return &Mempool{
        // Oper keeper (k) ke Engine
        Engine: NewMempoolEngine(params, k),
    }
}

// Add: Wrapper untuk menambah transaksi
func (m *Mempool) Add(tx types.Transaction) error {
    return m.Engine.Add(tx)
}

// GetPendingTransactions: Jembatan untuk API agar bisa melihat seluruh antrean
func (m *Mempool) GetPendingTransactions() []types.Transaction {
    return m.Engine.GetPendingTransactions()
}

// GetTransactions: Untuk fase persiapan (Mining)
func (m *Mempool) GetTransactions(limit int) []types.Transaction {
    return m.Engine.GetTransactions(limit)
}

// PullTransactions: Untuk fase finalisasi (Execution)
func (m *Mempool) PullTransactions(limit int) []types.Transaction {
    return m.Engine.PullTransactions(limit)
}

// RemoveUsedTransactions: Hapus yang sudah di-mine
func (m *Mempool) RemoveUsedTransactions(txsInBlock []types.Transaction) {
    m.Engine.RemoveUsedTransactions(txsInBlock)
}

// Count: Cek jumlah antrean
func (m *Mempool) Count() int {
    return m.Engine.Queue.Len()
}

// Clear: Sekarang memanggil Flush dari Engine agar satu komando
func (m *Mempool) Clear() {
    // Tuan tidak perlu m.mu.Lock di sini karena Engine sudah punya lock sendiri
    m.Engine.Flush() 
}

// Flush: Tetap ada sebagai alias jika diperlukan
func (m *Mempool) Flush() {
    m.Engine.Flush()
}

// StartHeartbeat: Memanggil detak jantung dari Engine
func (m *Mempool) StartHeartbeat(ctx context.Context) {
    m.Engine.StartHeartbeat(ctx)
}

func (m *Mempool) GetNotifyChan() chan bool {
    return m.Engine.NotifyChan
}

// Tambahkan di x/mempool/mempool.go
func (m *Mempool) GetHighestNonce(address string) uint64 {
    return m.Engine.GetHighestNonce(address)
}
