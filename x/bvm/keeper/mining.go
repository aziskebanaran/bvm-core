package keeper

import (
    "github.com/aziskebanaran/bvm-core/x/bvm/types"
    "github.com/aziskebanaran/bvm-core/pkg/logger"
    "fmt"
    "strings"
    "time"
)

// x/bvm/keeper/mining.go

func (k *Keeper) PrepareNewWork(minerAddr string, minerName string) (types.Block, error) {
    // 1. CREATE NEXT BLOCK
    // Ini adalah 'Pabrik' yang mengambil TX (Get), Status, Diff, dan Reward.
    block := k.CreateNextBlock(minerAddr)
    
    // Set identitas penambang
    block.MinerName = minerName

    // 2. RETURN UNTUK DISERAHKAN KE MESIN (SolveBlockLogic)
    // Kita tidak memanggil ProcessBlock di sini karena bloknya BELUM ditambang (Hash belum tembus).
    // Submit dan ProcessBlock nanti dipanggil di fungsi SubmitMinedBlock.
    
    return block, nil
}


// SolveBlockLogic: Mesin Proof of Work (PoW) yang dipusatkan di Keeper
// quit: channel untuk menghentikan mining jika ada blok lain yang lebih dulu sah
func (k *Keeper) SolveBlockLogic(block *types.Block, quit chan bool) bool {
    target := strings.Repeat("0", int(block.Difficulty))

    for {
        select {
        case <-quit:
            // Berhenti jika Jenderal memberi sinyal (misal: blok baru sudah ditemukan orang lain)
            return false
        default:
            // Hitung Hash
            hash := block.CalculateBlockHash()

            // Cek apakah target difficulty tembus
            if strings.HasPrefix(hash, target) {
                block.Hash = hash
                return true
            }

            // Naikkan Nonce
            block.Nonce++

            // Cegah Overheat: Beri nafas ke CPU setiap 1 juta percobaan
            if block.Nonce%1000000 == 0 {
                time.Sleep(1 * time.Millisecond)
            }
        }
    }
}

func (k *Keeper) SubmitMinedBlock(block types.Block) error {
    // 1. CEK JEDA WAKTU (Sesuai Logika Sultan)
    lastHeight := k.GetLastHeight()
    lastBlock, _ := k.Store.GetBlockByHeight(lastHeight)
    elapsed := time.Now().Unix() - lastBlock.Timestamp

    if elapsed < 6 {
        wait := 6 - elapsed
        logger.Info("MINER", fmt.Sprintf("⏳ Menunggu jeda %d detik...", wait))
        time.Sleep(time.Duration(wait) * time.Second)
        block.Timestamp = time.Now().Unix()
        block.Hash = block.CalculateBlockHash()
    }

    // 2. PROSES KE JENDERAL
    err := k.ProcessBlock(block)

    // 🚩 SOLUSI SULTAN: BERITAHU MEMPOOL APAPUN HASILNYA
    // Dengan memanggil Pull, kita memaksa Mempool untuk "Sinkron" dengan Disk.
    // Jika blok sukses, Pull akan membuang TX yang sudah sah.
    // Jika blok gagal (karena basi), Pull juga akan membersihkan sampahnya.
    k.Mempool.PullTransactions(len(block.Transactions))

    if err != nil {
        return fmt.Errorf("❌ Blok ditolak Jenderal: %v", err)
    }

    return nil
}


// GetDifficulty: Akses cepat ke target kesulitan saat ini
func (k *Keeper) GetDifficulty() int {
    return k.GetNextDifficulty()
}

func (k *Keeper) SetDifficulty(newDiff int) {
    k.Blockchain.Params.MinDifficulty = newDiff
}


func (k *Keeper) GetNextDifficulty() int {
    return k.GetNextDifficultyForMiner("")
}

