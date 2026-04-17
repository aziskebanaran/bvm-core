package app

import (
	"sync"
	"fmt"
	"context"
	"time"

	"bvm.core/pkg/logger"
	"bvm.core/pkg/nonce"
	"bvm.core/pkg/p2p"
	"bvm.core/pkg/storage"
	"bvm.core/x" // 🚩 Penting untuk interface
	authkeeper "bvm.core/x/auth/keeper"
	bankkeeper "bvm.core/x/bank/keeper"
	bvmkeeper "bvm.core/x/bvm/keeper"
	"bvm.core/x/bvm/types"
	"bvm.core/x/mempool"
	"bvm.core/x/miner"
	stakingkeeper "bvm.core/x/staking/keeper"
	wasmkeeper "bvm.core/x/wasm/keeper"
        storagekeeper "bvm.core/x/storage/keeper"
)

type BaseApp struct {
	Mu sync.RWMutex

	BVMKeeper  x.BVMKeeper      // 🚩 Gunakan interface x.BVMKeeper
	BankKeeper x.BankKeeper     // 🚩 Gunakan interface x.BankKeeper
	AuthKeeper x.AuthKeeper     // 🚩 Gunakan interface x.AuthKeeper
	WasmKeeper x.WasmKeeper
	Staking    x.StakingKeeper

	Mempool    x.MempoolKeeper
	Miner      x.MinerEngine
	Store      storage.BVMStore
	Blockchain *types.Blockchain
	P2P        x.P2PKeeper      // 🚩 Gunakan interface x.P2PKeeper
        StorageKeeper x.StorageModuleKeeper
}

func NewApp(store storage.BVMStore, bc *types.Blockchain) *BaseApp {
        logger.Info("SYSTEM", "Mempersiapkan Kabinet BVM...")

	params := types.DefaultParams()

        nm := nonce.NewNonceManager(store)

        authK := authkeeper.NewAuthKeeper(store, nm)

        mp := mempool.NewMempool(&params, authK)

	bankK := bankkeeper.NewBankKeeper(store, &params)

        // Pastikan NewKeeper ini sudah Sultan ubah di file masing-masing (Poin 3 di atas)
        wasmK := wasmkeeper.NewKeeper(store)

	stakingK := stakingkeeper.NewKeeper(store, bankK, wasmK)

        p2pK := p2p.NewKeeper(store)

        storageK := storagekeeper.NewStorageKeeper(store)

        // 🚩 SEKARANG HANYA 9 ARGUMEN (Tanpa store.GetDB())
        bvmK := bvmkeeper.NewKeeper(
                store,         // 1. BVMStore
                bc,            // 2. Blockchain
		&params,
                bankK,         // 3. Bank
                authK,         // 4. Auth
                nm,            // 5. Nonce
                mp,            // 6. Mempool
                stakingK,      // 7. Staking
                wasmK,         // 8. Wasm
                p2pK,          // 9. P2P
		storageK,      // 11 🚩
        )

        minerE := miner.NewMinerEngine(bvmK)

        return &BaseApp{
                BVMKeeper:  bvmK,
                BankKeeper: bankK,
                AuthKeeper: authK,
                Mempool:    mp,
                WasmKeeper: wasmK,
                Staking:    stakingK,
                Miner:      minerE,
                Store:      store,
                Blockchain: bc,
                P2P:        p2pK,
		StorageKeeper: storageK,
        }
}

func (app *BaseApp) Start() {
    logger.Info("SYSTEM", "🚀 Memulai BVM Engine (Atomic Mode)...")

    // --- 🚩 LANGKAH 1: LOCK & RECOVERY (KEDAULATAN DATABASE) ---
    app.Mu.Lock() // Kunci seluruh kabinet

    // Pulihkan data yang tersisa di RAM (PendingBlocks) ke Disk
    app.BVMKeeper.AutoRecoverDatabase()

    // Sinkronkan tinggi blok dan suplai sesuai kenyataan di Disk
    app.BVMKeeper.SyncState()

    // Load saldo-saldo dari Disk ke RAM Bank
    app.BankKeeper.LoadAllAccountsFromDB()

    app.Mu.Unlock()

    // 2. DETAK JANTUNG OTOMATIS
    ctx := context.Background()
    app.Mempool.StartHeartbeat(ctx)

    logger.Success("SYSTEM", fmt.Sprintf("Kernel aktif di Blok #%d", app.Blockchain.Height))

    // --- 🧠 AI LEARNING ENGINE (DI LUAR LOCK UTAMA) ---
    go func() {
        // Beri jeda 10 detik setelah start agar sistem stabil dulu
        time.Sleep(10 * time.Second)

        for {
            app.ExportSnapshotToAI()

            // AI istirahat sejenak agar tidak membebani CPU Termux
            time.Sleep(1 * time.Minute)
        }
    }()
}

func (app *BaseApp) ProcessTransaction(tx types.Transaction) error {
    app.Mu.Lock()
    // 1. Validasi dan Masukkan ke Mempool via Keeper
    err := app.BVMKeeper.ProcessTransaction(tx)
    app.Mu.Unlock() // Buka lock segera setelah validasi selesai

    if err != nil {
        return err
    }

    // --- TAMBAHAN SULTAN: PROPAGASI P2P ---
    go func() {
        if app.P2P != nil {
            app.P2P.BroadcastTransaction(tx)
        }
    }()

    logger.Success("MEMPOOL", fmt.Sprintf("📦 TXID %s masuk antrean", tx.ID[:8]))
    return nil
}


func (app *BaseApp) Stop() {
	logger.Info("SYSTEM", "🛑 Mematikan BVM Engine...")

}

// ExportSnapshotToAI sekarang memiliki fitur "Self-Cleaning"
func (app *BaseApp) ExportSnapshotToAI() {
    logger.Info("AI-CHEF", "👨‍🍳 Sedang menyiapkan makanan untuk AI...")

    // 1. Ambil data snapshot
    allData, err := app.StorageKeeper.GetAllAppData()
    if err != nil {
        logger.Error("AI-CHEF", "❌ Gagal mengambil data snapshot")
        return
    }

    // 2. Berikan ke AI Sentinel
    status, pesan := types.AI_DeepLearning(allData)

    // 🚩 TAMBAHAN: Logika Pruning (Pembersihan RAM/Cache)
    // Setelah AI belajar, kita pastikan map data dibersihkan dari memori
    for k := range allData {
        delete(allData, k)
    }
    allData = nil

    if status != 0 {
        logger.Error("AI-SENTINEL", pesan)
    } else {
        logger.Success("AI-SENTINEL", "🧠 Pelajaran diterima: " + pesan)
    }

    // 🚩 AUTO-PRUNING LOG: Jika blok kelipatan 1000, bersihkan cache database
    if app.Blockchain.Height % 1000 == 0 {
        logger.Info("SYSTEM", "🧹 Membersihkan cache storage untuk menjaga stabilitas...")
        // Di sini kita bisa panggil fungsi pembersih database jika diperlukan
    }
}
