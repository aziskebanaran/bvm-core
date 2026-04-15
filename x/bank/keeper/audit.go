package keeper

import (
	"bvm.core/pkg/logger"
	"bvm.core/x/bvm/types"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func (bk *BankKeeper) CheckSupplyIntegrity(symbol string) {
	if bk.isNative(symbol) { return }

	actualSum := bk.ScanTotalSupplyFromDB(symbol)
	
	var recordedSupply uint64
	bk.Store.Get("supply:"+symbol, &recordedSupply)

	if actualSum != recordedSupply {
		logger.Warning("BANK_AUDIT", fmt.Sprintf("⚠️ Supply %s TIDAK MATCH! Fisik: %d | Record: %d", symbol, actualSum, recordedSupply))
	} else {
		logger.Success("BANK_AUDIT", fmt.Sprintf("✅ Integritas %s Sinkron", symbol))
	}
}

func (bk *BankKeeper) ScanTotalSupplyFromDB(symbol string) uint64 {
	var total uint64
	db := bk.Store.GetDB()
	prefix := "t:" + symbol + ":"
	iter := db.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	defer iter.Release()

	for iter.Next() {
		var bal uint64
		if err := bk.Store.Get(string(iter.Key()), &bal); err == nil {
			total += bal
		}
	}
	return total
}

func (bk *BankKeeper) FreezeToken(caller string, symbol string) error {
	if bk.isNative(symbol) {
		return fmt.Errorf("❌ Koin utama tidak bisa dibekukan")
	}

	var creator string
	bk.Store.Get("owner:"+symbol, &creator)

	if creator == "" || creator != caller {
		return fmt.Errorf("🚫 Hanya pemilik token yang punya otoritas ini")
	}

	bk.Store.Put("frozen:"+symbol, true)
	logger.Warning("BANK", fmt.Sprintf("❄️ %s telah DIBEKUKAN oleh %s", symbol, caller[:8]))
	return nil
}

func (bk *BankKeeper) UnfreezeToken(caller string, symbol string) error {
    var owner string
    bk.Store.Get("owner:"+symbol, &owner)

    if owner == "" || owner != caller {
        return fmt.Errorf("🚫 Anda bukan creator token ini")
    }

    bk.Store.Put("frozen:"+symbol, false)
    logger.Success("BANK", fmt.Sprintf("🔥 Token %s telah diaktifkan kembali", symbol))
    return nil
}

func (bk *BankKeeper) IsFrozen(symbol string) bool {

    if bk.isNative(symbol) {
        return false
    }

    var frozen bool
    bk.Store.Get("frozen:"+symbol, &frozen)
    return frozen
}

func (bk *BankKeeper) CalculateBalanceManual(bc *types.Blockchain, addr string, symbol string) uint64 {
	if bk.isNative(symbol) { return 0 }

	var balance uint64 = 0

	start := 0
	if len(bc.Chain) > 1000 {
		start = len(bc.Chain) - 1000
	}

	for i := start; i < len(bc.Chain); i++ {
		for _, tx := range bc.Chain[i].Transactions {
			if tx.Symbol != symbol { continue }
			if tx.To == addr { balance += tx.Amount }
			if tx.From == addr {
				totalDeduct := tx.Amount + tx.Fee
				if balance >= totalDeduct {
					balance -= totalDeduct
				} else {
					balance = 0 // Terjadi jika ada data korup di masa lalu
				}
			}
		}
	}
	return balance
}

