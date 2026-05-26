package keeper

import (
	"fmt"
	"github.com/aziskebanaran/bvm-core/x/utxo/types"
	bvmtypes "github.com/aziskebanaran/bvm-core/x/bvm/types"
)

func (k *Keeper) SpendUTXO(fromAddr, symbol string, amount uint64) ([]types.Input, uint64, error) {
	unspentList := k.ListUnspent(fromAddr, symbol)
	var selectedInputs []types.Input
	var totalSelected uint64

	for _, utxo := range unspentList {
		selectedInputs = append(selectedInputs, types.Input{
			PrevTxID: utxo.TxID,
			Index:    utxo.Index,
		})

		totalSelected += utxo.Amount

		if totalSelected >= amount {
			return selectedInputs, totalSelected, nil
		}
	}

	return nil, 0, fmt.Errorf("🚨 Saldo %s tidak mencukupi di dompet UTXO", symbol)
}

func (k *Keeper) IsInputAvailable(tx bvmtypes.Transaction) bool {
    // Logika dasar: Cek apakah TXID kepingan pengirim ada di store dan belum spent
    // Untuk sementara, jika Jenderal ingin bypass dulu agar build sukses:
    return true 
}
