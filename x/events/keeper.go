package events

import (
    "github.com/aziskebanaran/bvm-core/x/bvm/types"
    "fmt"
)

// --- 1. DEFINISI KEEPER (PONDASI UTAMA) ---
type Keeper struct {
    Mempool    MempoolKeeper  // Untuk k.Mempool.Count()
    Blockchain BlockchainInfo // Untuk data blok & difficulty
}

// --- 2. INTERFACE (Kabel Penghubung ke Modul Lain) ---
type MempoolKeeper interface {
    Count() int
    Flush()
}

type BlockchainInfo interface {
    GetDifficulty() int
    SetDifficulty(int)
    CalculateAvgBlockTime() float64
    GetBalance(string) float64
}

// --- 3. FUNGSI-FUNGSI SULTAN (Sekarang Sudah Legal) ---

func (k *Keeper) RunHealthCheck() {
    stats := types.NetworkHealth{
        PendingTxCount:   k.Mempool.Count(),
        AverageBlockTime: k.Blockchain.CalculateAvgBlockTime(), // Tambahkan .Blockchain
    }

    diffAdj, message := types.AI_Sentinel(stats)

    if diffAdj != 0 {
        fmt.Printf("🤖 [AI SENTINEL] %s\n", message)
        // Gunakan k.Blockchain karena Set/Get ada di sana
        k.Blockchain.SetDifficulty(k.Blockchain.GetDifficulty() + diffAdj)
    }
}

func (k *Keeper) ValidateWithAI() (int, string) {
    stats := types.NetworkHealth{
        AverageBlockTime: k.Blockchain.CalculateAvgBlockTime(),
        PendingTxCount:   k.Mempool.Count(),
    }
    return types.AI_Sentinel(stats)
}

func (k *Keeper) ExecuteMiningCycle(minerAddress string) {
    adjustment, message := k.ValidateWithAI()

    if adjustment == -99 {
        fmt.Printf("\n🚨 [AI SENTINEL] %s\n", message)
        k.Mempool.Flush()
        return 
    }

    if adjustment != 0 {
        fmt.Printf("\n⚠️ [AI SENTINEL] %s\n", message)
        newDiff := k.Blockchain.GetDifficulty() + adjustment
        k.Blockchain.SetDifficulty(newDiff)
        fmt.Printf("✅ [AI SENTINEL] Difficulty baru: %d\n", newDiff)
    }
}

func (k *Keeper) GetRequiredDifficulty(minerAddr string) int {
    params := types.DefaultParams()
    balance := k.Blockchain.GetBalance(minerAddr) // Tambahkan .Blockchain

    if balance >= 100 {
        return 1
    }
    return params.MinDifficulty
}

// EmitEvent: Pengeras suara universal untuk seluruh modul BVM
// Fungsi ini harus diawali huruf KAPITAL agar bisa diakses dari luar folder events
func EmitEvent(eventType string, data interface{}) {
    // 1. Cetak ke Console (Agar Sultan bisa memantau secara real-time)
    fmt.Printf("\n📢 [EVENT LOG] Type: %s\n", eventType)
    
    // 2. Jika data berupa Map, tampilkan isinya secara rapi
    if m, ok := data.(map[string]interface{}); ok {
        for key, val := range m {
            fmt.Printf("   🔹 %s: %v\n", key, val)
        }
    } else {
        fmt.Printf("   🔹 Data: %+v\n", data)
    }
    fmt.Println("--------------------------------------------")

    // 💡 TIPS SULTAN: 
    // Di sini Sultan bisa menambahkan logika untuk menyimpan event ke Database
    // atau mengirim notifikasi Push ke aplikasi DompetKu Sultan.
}
