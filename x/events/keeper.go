package events

import (
    "github.com/aziskebanaran/bvm-core/x/bvm/types"
    "fmt"
	"encoding/json"
	eventstypes "github.com/aziskebanaran/bvm-core/x/events/types"
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


// =========================================================================
// 📡 FUNGSI 1: EmitEvent (Gaya Lama - Murni Log Console)
// =========================================================================
// Tetap pertahankan 2 argumen ini agar modul P2P Jenderal Lolos Sensor 100%!
func EmitEvent(eventType string, data interface{}) {
	fmt.Printf("\n📢 [SYSTEM LOG] Type: %s\n", eventType)

	if m, ok := data.(map[string]interface{}); ok {
		for key, val := range m {
			fmt.Printf("   🔹 %s: %v\n", key, val)
		}
	} else {
		fmt.Printf("   🔹 Data: %+v\n", data)
	}
	fmt.Println("--------------------------------------------")
}

// =========================================================================
// 🪙 FUNGSI 2: EmitBatch (Gaya Baru Khusus Keuangan - Wajib Pahat LevelDB)
// =========================================================================
// Nama fungsi baru sesuai komando Jenderal, menerima 3 argumen untuk disk!
func EmitBatch(batch interface{}, eventType string, data interface{}) {
	fmt.Printf("\n📢 [STATE EVENT] Type: %s\n", eventType)

	if m, ok := data.(map[string]interface{}); ok {
		for key, val := range m {
			fmt.Printf("   🔹 %s: %v\n", key, val)
		}
	}
	fmt.Println("--------------------------------------------")

	if batch != nil {
		type BatchWriter interface {
			Put(key []byte, value []byte)
		}

		if b, ok := batch.(BatchWriter); ok {
			eventObj := eventstypes.NewEvent(eventType, data)
			keyDB := fmt.Sprintf("ev:%s:%s", eventType, eventObj.ID)

			importJSON, errMarshal := json.Marshal(eventObj)
			if errMarshal == nil {
				b.Put([]byte(keyDB), importJSON)
			}
		}
	}
}
