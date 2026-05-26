package keeper

import (
	"github.com/aziskebanaran/bvm-core/x/utxo/types"
	"github.com/vmihailenco/msgpack/v5"
)

func (k *Keeper) GetUTXO(addr, utxoID string) (types.UTXO, bool) {
	var keping types.UTXO
	err := k.store.Get(types.KeyUTXO(addr+":"+utxoID), &keping)
	return keping, err == nil
}

func (k *Keeper) ListUnspent(addr, symbol string) []types.UTXO {
	var list []types.UTXO
	// Scan berdasarkan prefix pemilik (u:bvmf...)
	prefix := types.KeyUTXO(addr)
	results, _ := k.store.PrefixScan(prefix)

	for _, val := range results {
		var keping types.UTXO
		if err := msgpack.Unmarshal(val, &keping); err == nil {
			if keping.Status == "UNSPENT" && (symbol == "" || keping.Symbol == symbol) {
				list = append(list, keping)
			}
		}
	}
	return list
}

func (k *Keeper) GetTotalBalance(addr, symbol string) uint64 {
    var total uint64
    // Scan berdasarkan prefix pemilik (u:bvmf...)
    prefix := types.KeyUTXO(addr)
    results, err := k.store.PrefixScan(prefix)
    if err != nil {
        return 0
    }

    for _, val := range results {
        var keping types.UTXO
        // Gunakan msgpack sesuai standar performa Jenderal
        if err := msgpack.Unmarshal(val, &keping); err == nil {
            if keping.Status == "UNSPENT" && (symbol == "" || keping.Symbol == symbol) {
                total += keping.Amount
            }
        }
    }
    return total
}
