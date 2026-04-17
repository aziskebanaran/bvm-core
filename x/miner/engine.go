package miner

import (
    "github.com/aziskebanaran/bvm-core/x"
    "github.com/aziskebanaran/bvm-core/pkg/logger"
    "fmt"
	"time"
)

type MinerEngine struct {
    Keeper x.BVMKeeper // Mengacu ke Jenderal (BVMKeeper)
}

func NewMinerEngine(k x.BVMKeeper) *MinerEngine {
    return &MinerEngine{Keeper: k}
}

func (m *MinerEngine) Start(minerAddr string) {
    go func() {
        logger.Success("MINER", "👷 Mesin Miner Internal Aktif (Mode Sinyal Jantung)")

        // 1. Ambil radio kontrol dari Mempool
        notify := m.Keeper.GetMempool().GetNotifyChan()

        for {
            // 🚩 BERHENTI: Menunggu instruksi detak jantung 60 detik (atau sinyal TX baru)
            <-notify 

            // Beri waktu 500ms agar database benar-benar stabil setelah blok sebelumnya selesai dikomit
            time.Sleep(500 * time.Millisecond)

            // 2. Minta Paket Kerja ke Keeper (Jenderal)
            block, err := m.Keeper.PrepareNewWork(minerAddr, "BVM-INTERNAL")
            if err != nil {
                // Jangan menyerah, coba lagi di siklus berikutnya
                continue
            }

            // 3. Validasi Ketinggian (Anti-Basi)
            // Jika index yang ditambang tidak lebih tinggi dari yang ada di disk, abaikan.
            if int64(block.Index) <= int64(m.Keeper.GetStatus().Height) {
                continue
            }

            logger.Info("MINER", fmt.Sprintf("⚒️ Sinyal Diterima! Mulai Memahat Blok #%d [Diff: %d]", 
                block.Index, block.Difficulty))

            // 4. Proses Hashing (PoW)
            quit := make(chan bool)
            if success := m.Keeper.SolveBlockLogic(&block, quit); success {
                // 5. Setor Hasil ke Kernel
                if err := m.Keeper.SubmitMinedBlock(block); err != nil {
                    // Jika ditolak (misal karena alasan waktu atau konsensus), lapor ke Sultan
                    logger.Warning("MINER", fmt.Sprintf("💤 Sinyal ditolak Jenderal: %v", err))
                } else {
                    // Jika sukses, blok sudah sah dan mempool otomatis bersih di CommitBlock
                    logger.Success("MINER", fmt.Sprintf("💎 Blok #%d Sukses Terpahat!", block.Index))
                }
            }
        }
    }()
}


// Stop: Mematikan mesin tambang Sultan
func (m *MinerEngine) Stop() {
	logger.Info("MINER", "Mesin tambang dihentikan oleh Sultan.")
	// Jika Sultan pakai channel untuk mining, tambahkan logic stop di sini
}
