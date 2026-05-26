package types

import (
    "github.com/aziskebanaran/bvm-core/pkg/storage"
	bvmtypes "github.com/aziskebanaran/bvm-core/x/bvm/types"
)

// UTXOKeeper: Jantung Ekonomi Kepingan Aset
type UTXOKeeper interface {
    // Jalur Utama Transaksi (Tanpa Nonce)
    SendAset(from, to, symbol string, amount, fee uint64, txID string, payload []byte, batch storage.Batch) error
    // Logistik & Audit Kepingan
    MintUTXO(addr, addr0x, username, symbol string, amount uint64, utxoID string, batch storage.Batch) error
    BurnUTXO(addr, utxoID string, batch storage.Batch) error
    // Intelijen & Querier
    GetUTXO(addr, utxoID string) (UTXO, bool)
    ListUnspent(addr, symbol string) []UTXO
    GetTotalBalance(addr, symbol string) uint64

    IsInputAvailable(tx bvmtypes.Transaction) bool
    GetAllUTXOHolders(nativeSymbol string) map[string]uint64
}
