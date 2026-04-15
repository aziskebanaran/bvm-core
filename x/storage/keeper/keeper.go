package keeper

import (
    "bvm.core/pkg/storage"
    "bvm.core/x/storage/types"
	"bvm.core/x"
    "fmt"
    "os"
)


type StorageKeeper struct {
    // Keeper utama untuk mencatat daftar aplikasi yang aktif
    mainStore storage.BVMStore 
}

// Constructor untuk StorageKeeper
func NewStorageKeeper(mainStore storage.BVMStore) *StorageKeeper {
    return &StorageKeeper{
        mainStore: mainStore,
    }
}

func (k *StorageKeeper) GetAppStore(appID string) (storage.BVMStore, error) {
    // 1. Buat folder jika belum ada
    path := fmt.Sprintf("./data/apps_storage/app_%s_db", appID)
    if _, err := os.Stat(path); os.IsNotExist(err) {
        os.MkdirAll(path, os.ModePerm)
    }

    // 2. Panggil engine.go dengan RAM kecil (2MB saja cukup untuk database sampingan)
    return storage.NewLevelDBStore(path, 2)
}

func (k *StorageKeeper) PutUserData(appID string, data types.UserData) error {
    // 1. Buka Gudang Mandiri si User
    appDB, err := k.GetAppStore(appID)
    if err != nil {
        return err
    }
    // Pastikan database ditutup setelah selesai agar tidak bocor RAM
    // Tapi karena kita pakai sistem pool, lebih baik dikelola Manager. 
    // Untuk sekarang, kita tutup manual:
    defer appDB.Close()

    // 2. Simpan Data ke dalam Gudang tersebut
    return appDB.Put(data.Key, data.Value)
}

func (k *StorageKeeper) SafePut(appID string, data types.UserData, callerAddr string) error {
    // 1. Ambil Metadata Aplikasi dari MainStore
    var app types.AppContainer
    err := k.mainStore.Get("app:"+appID, &app)
    if err != nil {
        return fmt.Errorf("aplikasi %s tidak terdaftar", appID)
    }

    // 2. JALANKAN CHECKRULES (The Guard)
    if !k.CheckRules(app, data.Key, "write", callerAddr) {
        return fmt.Errorf("🚫 BVM-GUARD: Pelanggaran aturan tulis pada path '%s'", data.Key)
    }

    // 3. Jika Lolos, Buka Gudang Mandiri dan Simpan
    appDB, err := k.GetAppStore(appID)
    if err != nil {
        return err
    }
    defer appDB.Close()

    return appDB.Put(data.Key, data.Value)
}

func (k *StorageKeeper) GetAppMetadata(appID string) (types.AppContainer, error) {
    var app types.AppContainer
    // Kita ambil dari database utama dengan prefix "app:"
    err := k.mainStore.Get("app:"+appID, &app)
    if err != nil {
        return types.AppContainer{}, fmt.Errorf("aplikasi %s tidak ditemukan", appID)
    }
    return app, nil
}

// CalculateStorageFee: Menghitung biaya total yang akan dimusnahkan (Burn)
func (k *StorageKeeper) CalculateStorageFee(dataSize int, bvm x.BVMKeeper) uint64 {
    dataSizeKB := float64(dataSize) / 1024.0

    // Rumus: Base + (Size * Rate)
    rawTotal := types.BaseFeeBVM + (dataSizeKB * types.RatePerKBBVM)

    // Konversi langsung ke unit terkecil (Atomic) menggunakan mesin Kernel
    return bvm.ToAtomic(rawTotal)
}
