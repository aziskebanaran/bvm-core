package keeper

import (
	"strings"

	"github.com/syndtr/goleveldb/leveldb/util"
	"github.com/vmihailenco/msgpack/v5" // 🚩 WAJIB: Gunakan codec yang sama dengan Sultan Engine!
	"github.com/aziskebanaran/bvm-core/x/utxo/types"
)

// GetAllUTXOHolders: Mengambil daftar semua saldo kepingan UNSPENT secara dinamis berbasis Params Jaringan & Msgpack
func (k *Keeper) GetAllUTXOHolders(nativeSymbol string) map[string]uint64 {
	utxoHoldersMap := make(map[string]uint64)

	// Panggil GetDB() murni dari k.store (Pastikan interface BVMStore sudah mencatat GetDB)
	db := k.store.GetDB() 
	if db == nil {
		return utxoHoldersMap
	}

	// Sisir LevelDB murni menggunakan prefix "u:" dari keys UTXO Jenderal
	prefixKey := []byte(types.UTXOPrefix)
	iter := db.NewIterator(util.BytesPrefix(prefixKey), nil)
	defer iter.Release()

	for iter.Next() {
		var keping types.UTXO
		// 🚀 SINKRONISASI BINARI SULTAN: Bongkar biner murni msgpack dari disk!
		if err := msgpack.Unmarshal(iter.Value(), &keping); err == nil {
			// Validasi status UNSPENT dan kecocokan Simbol Natif
			if keping.Status != "SPENT" && (keping.Symbol == nativeSymbol || keping.Symbol == "") {
				ownerAddr := keping.Address
				if ownerAddr == "" {
					ownerAddr = keping.Address0x
				}

				if strings.HasPrefix(ownerAddr, "0x") {
					utxoHoldersMap[ownerAddr] += keping.Amount
				}
			}
		}
	}

	return utxoHoldersMap
}
