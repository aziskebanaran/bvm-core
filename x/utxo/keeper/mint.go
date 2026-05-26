package keeper

import (
	"fmt"
	"time"
	"github.com/aziskebanaran/bvm-core/pkg/storage"
	"github.com/aziskebanaran/bvm-core/x/utxo/types"
)

func (k *Keeper) MintUTXO(addr, addr0x, username, symbol string, amount uint64, utxoID string, batch storage.Batch) error {
	keping := types.UTXO{
		TxID:      utxoID,
		Index:     0,
		Address:   addr,
		Address0x: addr0x,
		Username:  username,
		Symbol:    symbol,
		Amount:    amount,
		Status:    "UNSPENT",
		Timestamp: time.Now().Unix(),
		Type:      "COIN",
	}

	dbKey := types.KeyUTXO(addr + ":" + utxoID)
	if batch != nil {
		return k.store.PutToBatch(batch, dbKey, keping)
	}
	return k.store.Put(dbKey, keping)
}

func (k *Keeper) BurnUTXO(addr, utxoID string, batch storage.Batch) error {
	keping, found := k.GetUTXO(addr, utxoID)
	if !found {
		return fmt.Errorf("❌ Kepingan %s tidak ditemukan", utxoID)
	}

	keping.Status = "SPENT"
	dbKey := types.KeyUTXO(addr + ":" + utxoID)

	if batch != nil {
		return k.store.PutToBatch(batch, dbKey, keping)
	}
	return k.store.Put(dbKey, keping)
}
