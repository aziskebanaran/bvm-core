package keeper

import (
	"bvm.core/pkg/storage"
	"bvm.core/x"
	"bvm.core/x/storage/types"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type StorageKeeper struct {
	mainStore storage.BVMStore
}

func NewStorageKeeper(mainStore storage.BVMStore) *StorageKeeper {
	return &StorageKeeper{
		mainStore: mainStore,
	}
}

// GetAllAppData memindai folder storage dan mengambil snapshot data untuk AI
func (k *StorageKeeper) GetAllAppData() (map[string][]byte, error) {
	appDataMap := make(map[string][]byte)
	basePath := "./data/apps_storage"

	// 1. Baca daftar folder aplikasi di dalam data/apps_storage
	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca folder storage: %v", err)
	}

	for _, f := range files {
		if f.IsDir() && strings.HasPrefix(f.Name(), "app_") {
			// Ekstrak AppID dari nama folder (misal: app_DompetKu_db -> DompetKu)
			appID := strings.TrimPrefix(f.Name(), "app_")
			appID = strings.TrimSuffix(appID, "_db")

			// 2. Buka database aplikasi tersebut
			appDB, err := k.GetAppStore(appID)
			if err != nil {
				continue // Skip jika gagal buka satu DB
			}

			// 3. Scan semua data di dalamnya menggunakan PrefixScan dari engine Sultan
			// Kita ambil semua data (prefix kosong "")
			results, err := appDB.PrefixScan("")
			appDB.Close() // Tutup segera setelah scan selesai

			if err == nil {
				// Gabungkan semua value menjadi satu byte array untuk "makanan" AI
				var combinedData []byte
				for _, val := range results {
					combinedData = append(combinedData, val...)
				}
				appDataMap[appID] = combinedData
			}
		}
	}

	return appDataMap, nil
}


func (k *StorageKeeper) GetAppStore(appID string) (storage.BVMStore, error) {
	path := fmt.Sprintf("./data/apps_storage/app_%s_db", appID)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, os.ModePerm)
	}
	// Gunakan cache kecil agar tidak membebani RAM saat scan massal
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
