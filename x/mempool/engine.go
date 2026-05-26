package mempool

import (
    "github.com/aziskebanaran/bvm-core/x/bvm/types"
    "github.com/aziskebanaran/bvm-core/pkg/logger"
    "container/heap"
    "sync"
	"fmt"
	"time"
	"context"
)

// PriorityQueue: Implementasi Heap agar TX Fee tertinggi selalu di atas
type PriorityQueue []*types.Transaction

func (pq PriorityQueue) Len() int { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool { return pq[i].Fee > pq[j].Fee } // Urutkan Fee Terbesar
func (pq PriorityQueue) Swap(i, j int)      { pq[i], pq[j] = pq[j], pq[i] }

func (pq *PriorityQueue) Push(x interface{}) {
    item := x.(*types.Transaction)
    *pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
    old := *pq
    n := len(old)
    item := old[n-1]
    *pq = old[0 : n-1]
    return item
}

type MempoolKeeper interface {
    GetNextNonce(address string) uint64
}

type MempoolEngine struct {
    Queue *PriorityQueue
    Mu    sync.RWMutex
    PendingNonces map[string]uint64
    Params        *types.Params
    Keeper        MempoolKeeper // 🚩 Tambahkan ini agar bisa cek DB
    NotifyChan    chan bool
}

func NewMempoolEngine(params *types.Params, k MempoolKeeper) *MempoolEngine {
    pq := &PriorityQueue{}
    heap.Init(pq)
    return &MempoolEngine{
        Queue: pq,
        PendingNonces: make(map[string]uint64),
        Params:        params,
        Keeper:        k, // 🚩 Masukkan keeper di sini
        NotifyChan:    make(chan bool, 1),
    }
}

func (m *MempoolEngine) Add(tx types.Transaction) error {
    m.Mu.Lock()
    defer m.Mu.Unlock()

    // 🚩 PINTU VIP: UTXO, Bridge In, dan Proof (Merdeka Nonce)
    // Semua transaksi tipe ini diizinkan masuk tanpa interogasi Nonce
    isVIP := (len(tx.From) > 2 && tx.From[:2] == "0x") || 
             tx.Type == "utxo_move" || 
             tx.Type == "bridge_in" || 
             tx.Type == "submit_proof"

    if isVIP {
        m.replaceIfDuplicate(tx)
        heap.Push(m.Queue, &tx)
        return nil // 🚀 LANGSUNG LOLOS MASUK ANTRIAN BLOK!
    }

    // =========================================================================
    // 🏦 JALUR AKUN BIASA (Account Based)
    // =========================================================================
    actualNonceOnDisk := m.Keeper.GetNextNonce(tx.From)

    if lastNonce, ok := m.PendingNonces[tx.From]; ok {
        if lastNonce < actualNonceOnDisk {
            delete(m.PendingNonces, tx.From)
        }
    }

    if tx.Nonce < actualNonceOnDisk {
        return fmt.Errorf("❌ NONCE EXPIRED")
    }

    m.PendingNonces[tx.From] = tx.Nonce
    m.replaceIfDuplicate(tx)
    heap.Push(m.Queue, &tx)
    return nil
}


func (m *MempoolEngine) GetPendingTransactions() []types.Transaction {
    m.Mu.RLock()
    defer m.Mu.RUnlock()
    var txs []types.Transaction
    for _, tx := range *m.Queue {
        txs = append(txs, *tx)
    }
    return txs
}

func (m *MempoolEngine) GetTransactions(limit int) []types.Transaction {
    m.Mu.RLock()
    defer m.Mu.RUnlock()

    var txs []types.Transaction
    count := 0
    for _, tx := range *m.Queue {
        if count >= limit {
            break
        }

        // 🚩 PERLUASAN BYPASS VIP: Masukkan Bridge dan Proof ke area aman
        isVIP := (len(tx.From) > 2 && tx.From[:2] == "0x") || 
                 tx.Type == "utxo_move" || 
                 tx.Type == "bridge_in" || 
                 tx.Type == "submit_proof"

        // 🚀 Jika VIP, langsung masukkan ke antrean pahat tanpa syarat Nonce
        if isVIP || tx.Nonce >= m.Keeper.GetNextNonce(tx.From) {
            txs = append(txs, *tx)
            count++
        }
    }
    return txs
}


func (m *MempoolEngine) PullTransactions(limit int) []types.Transaction {
    m.Mu.Lock()
    defer m.Mu.Unlock()

    var finalTxs []types.Transaction

    for m.Queue.Len() > 0 && len(finalTxs) < limit {
        tx := heap.Pop(m.Queue).(*types.Transaction)

        // 🚩 BYPASS UTXO & BRIDGE UNTUK PULL ENGINE
        isVIP := (len(tx.From) > 2 && tx.From[:2] == "0x") ||
                 tx.Type == "utxo_move" ||
                 tx.Type == "bridge_in" ||
                 tx.Type == "submit_proof"

        if isVIP {
            finalTxs = append(finalTxs, *tx)
            // Bridge/UTXO tidak memerlukan penghapusan dari PendingNonces
            continue
        }

        // 🏦 JALUR ACCOUNT BIASA (KODE ASLI JENDERAL - AMAN)
        actualNonce := m.Keeper.GetNextNonce(tx.From)

        if tx.Nonce < actualNonce {
            logger.Warning("MEMPOOL", fmt.Sprintf("🗑️  Membuang TX Basi %s (Nonce %d, Harusnya %d)", tx.ID[:8], tx.Nonce, actualNonce))

            delete(m.PendingNonces, tx.From)
            continue 
        }

        if tx.Nonce > actualNonce {
            heap.Push(m.Queue, tx)
            break 
        }

        finalTxs = append(finalTxs, *tx)
        delete(m.PendingNonces, tx.From)
    }

    return finalTxs
}

func (m *MempoolEngine) RemoveUsedTransactions(txsInBlock []types.Transaction) {
    m.Mu.Lock()
    defer m.Mu.Unlock()

    // 1. Buat peta ID transaksi yang sudah masuk blok agar cepat dicari
    usedIDs := make(map[string]bool)
    for _, tx := range txsInBlock {
        usedIDs[tx.ID] = true
        delete(m.PendingNonces, tx.From) // Hapus cache Nonce
    }

    // 2. Filter ulang antrean RAM (Gudang Belakang)
    newQueue := PriorityQueue{}
    for _, tx := range *m.Queue {
        // Hanya masukkan kembali transaksi yang TIDAK ada di dalam blok
        if !usedIDs[tx.ID] {
            newQueue = append(newQueue, tx)
        }
    }

    // 3. Pasang kembali antrean yang sudah bersih dan rapikan Heap
    *m.Queue = newQueue
    heap.Init(m.Queue)

    logger.Info("MEMPOOL", fmt.Sprintf("🧹 %d TX Dibuang | Sisa Antrean: %d", len(usedIDs), m.Queue.Len()))
}


func (m *MempoolEngine) Flush() {
    m.Mu.Lock()
    defer m.Mu.Unlock()
    *m.Queue = (*m.Queue)[:0]

	m.PendingNonces = make(map[string]uint64)
}

func (m *MempoolEngine) StartHeartbeat(ctx context.Context) {
    // 1. Ambil napas sesuai Konstitusi Sultan (Params)
    blockTime := time.Duration(m.Params.TargetBlockTime) * time.Second
    ticker := time.NewTicker(blockTime)

    logger.Info("MEMPOOL", fmt.Sprintf("⏱️ Jantung diaktifkan! Ritme: %v", blockTime))

    go func() {
        defer ticker.Stop()
        for {
            select {
            case <-ticker.C:
                // 🚩 Sinyal rutin 60 detik (atau sesuai params)
                select {
                case m.NotifyChan <- true:
                    // Log ini akan muncul di terminal Sultan setiap 1 menit
                    logger.Info("MEMPOOL", "💓 Detak Jantung: Mengirim instruksi pahat blok.")
                default:
                    // Jika miner masih sibuk (chan penuh), kita diam saja agar tidak macet
                }
            case <-ctx.Done():
                logger.Warning("MEMPOOL", "🛑 Detak Jantung dihentikan (Context Done).")
                return
            }
        }
    }()
}

// GetHighestNonce: Mencari nonce tertinggi milik alamat tertentu di antrean
func (m *MempoolEngine) GetHighestNonce(address string) uint64 {
    m.Mu.RLock()
    defer m.Mu.RUnlock()

    // Ambil dari map cache yang sudah kita buat sebelumnya
    if lastNonce, ok := m.PendingNonces[address]; ok {
        return lastNonce
    }
    return 0
}

// replaceIfDuplicate: Mencari dan menghapus transaksi lama dengan Nonce yang sama
func (m *MempoolEngine) replaceIfDuplicate(newTx types.Transaction) {
    // Kita tidak perlu lock lagi karena fungsi ini dipanggil di dalam m.Mu.Lock() milik Add
    for i, tx := range *m.Queue {
        if tx.From == newTx.From && tx.Nonce == newTx.Nonce {
            // Ditemukan hantu Nonce lama! Kita buang dari Queue.
            heap.Remove(m.Queue, i)
            logger.Warning("MEMPOOL", fmt.Sprintf("♻️  Mengganti transaksi lama (Nonce %d) dengan yang baru.", newTx.Nonce))
            break 
        }
    }
}
