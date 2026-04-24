package keeper

import (
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "time"
    "github.com/aziskebanaran/bvm-core/x/storage/types"
	"github.com/aziskebanaran/bvm-core/x"
)

// GenerateAPIKey: Membuat kunci rahasia baru untuk pengembang
func (k *StorageKeeper) RegisterApp(owner string, appID string, rules map[string]interface{}) (string, error) {
    // 1. Cek apakah AppID sudah dipakai
    var existing types.AppContainer
    if err := k.mainStore.Get("app:"+appID, &existing); err == nil {
        return "", fmt.Errorf("AppID %s sudah terdaftar", appID)
    }

    // 2. Buat Random Secret Key
    b := make([]byte, 24)
    rand.Read(b)
    secretKey := hex.EncodeToString(b)

    // 3. Simpan Metadata ke MainStore (Blockchain Core)
    newApp := types.AppContainer{
        AppID:     appID,
        Owner:     owner,
        APIKey:    secretKey, // Di dunia nyata, ini sebaiknya di-hash (Bcrypt/SHA256)
        Rules:     rules,
        CreatedAt: time.Now().Unix(),
    }

    // Simpan dengan prefix app:
    err := k.mainStore.Put("app:"+appID, newApp)
    return secretKey, err
}

func (k *StorageKeeper) ProcessAutoBilling(owner string, dataSize int, bvm x.BVMKeeper) (uint64, error) {
    // 1. Hitung biaya total dalam unit Atomic
    totalBurnAmount := k.CalculateStorageFee(dataSize, bvm)

    // 2. Validasi Saldo Pengembang
    if bvm.GetBalanceBVM(owner) < totalBurnAmount {
        return 0, fmt.Errorf("🚨 Saldo BVM tidak cukup untuk biaya pembakaran storage")
    }

    // 3. Eksekusi Pemotongan Saldo (Debit)
    // Nil dipersiapkan jika Sultan ingin menggunakan sistem batch nanti
    err := bvm.SubBalanceBVM(owner, totalBurnAmount, nil)
    if err != nil {
        return 0, err
    }

    // 4. KIRIM KE LUBANG HITAM (100% BURN)
    // Alamat "000..." secara teknis tidak memiliki kunci privat, koin terkunci selamanya.
    burnAddr := "bvmf000000000000000000000000000000000000burn"
    bvm.AddBalanceBVM(burnAddr, totalBurnAmount, nil)

    return totalBurnAmount, nil
}
