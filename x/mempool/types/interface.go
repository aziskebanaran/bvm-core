package types

import (
	"github.com/aziskebanaran/bvm-core/x/bvm/types"
	"context"
)

// MempoolKeeper: Kontrak untuk antrean transaksi
type MempoolKeeper interface {
    Add(tx types.Transaction) error
    GetTransactions(limit int) []types.Transaction
    PullTransactions(limit int) []types.Transaction
    RemoveUsedTransactions(txsInBlock []types.Transaction)
    GetPendingTransactions() []types.Transaction
    GetNotifyChan() chan bool
    StartHeartbeat(ctx context.Context)
    GetHighestNonce(address string) uint64
    Flush()
    Clear()
    Count() int
}
