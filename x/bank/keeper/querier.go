package keeper

import (
	"github.com/aziskebanaran/bvm-core/x/bvm/types"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// SearchAccount: Sekarang mengambil Nonce dari acc: tapi Saldo dari a:
func (bk *BankKeeper) SearchAccount(address string) (map[string]interface{}, bool) {
	// 1. Ambil Metadata (Status, Nonce, dll) dari prefix "acc:"
	acc := bk.GetOrCreateAccount(address)

	// 2. Ambil Saldo BVM asli dari prefix "a:" melalui GetBalance
	bvmBal := bk.GetBalance(address, "BVM")

	// Jika tidak ada saldo dan tidak ada metadata, baru kita anggap tidak ada
	if bvmBal == 0 && acc.Status == "" && acc.Nonce == 0 {
		return nil, false
	}

	// 3. Gabungkan! Saldo BVM dari kedaulatan Core ditampilkan sebagai bvm_balance
	return map[string]interface{}{
		"address":     acc.Address,
		"bvm_balance": bvmBal,      // <--- Ini yang muncul di CLI Dompet
		"balances":    acc.Balances, // Token L2 lainnya
		"nonce":       acc.Nonce,    // Tampilkan Nonce agar CLI tidak bingung
		"status":      acc.Status,
	}, true
}

func (bk *BankKeeper) GetAllBalances() map[string]map[string]uint64 {
    allHolders := make(map[string]map[string]uint64)

    // 🚩 Ambil DB langsung untuk iterasi
    db := bk.Store.GetDB()
    if db == nil {
        return allHolders
    }

    // Gunakan Iterator untuk menyisir semua saldo "a:"
    iter := db.NewIterator(util.BytesPrefix([]byte("a:")), nil)
    defer iter.Release()

    for iter.Next() {
        key := string(iter.Key())
        if len(key) <= 2 { continue }

        address := key[2:] // Ambil alamat setelah "a:"

        var bvmBal uint64
        // Gunakan Get langsung dari Store agar konsisten dengan cara simpan
        if err := bk.Store.Get(key, &bvmBal); err == nil {
            if _, exists := allHolders[address]; !exists {
                allHolders[address] = make(map[string]uint64)
            }
            allHolders[address]["BVM"] = bvmBal
        }

        // Ambil token L2 lainnya jika ada di metadata "acc:"
        var acc types.Account
        if err := bk.Store.Get("acc:"+address, &acc); err == nil {
            for sym, amt := range acc.Balances {
                allHolders[address][sym] = amt
            }
        }
    }

    return allHolders
}



func (bk *BankKeeper) GetPublicKeyFromLedger(bc *types.Blockchain, address string) string {
        searchLimit := 500
        startBlock := len(bc.Chain) - 1
        if startBlock < 0 { return "" }

        endBlock := startBlock - searchLimit
        if endBlock < 0 { endBlock = 0 }

        for i := startBlock; i >= endBlock; i-- {
                for _, tx := range bc.Chain[i].Transactions {
                        if tx.From == address && tx.PublicKey != "" {
                                return tx.PublicKey
                        }
                }
        }
        return ""
}


func (bk *BankKeeper) LoadAllAccountsFromDB() int {
	count := 0
	db := bk.Store.GetDB()

	if db == nil {
		return 0
	}

	iter := db.NewIterator(util.BytesPrefix([]byte("acc:")), nil)
	defer iter.Release()

	for iter.Next() {
		if iter.Error() != nil {
			break
		}
		count++
	}

	return count
}
