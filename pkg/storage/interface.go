package storage

import (
    "bvm.core/x/bvm/types"
    "github.com/syndtr/goleveldb/leveldb"
)

// 🚩 1. Definisikan tipe Batch agar bisa dipanggil di luar
type Batch interface{}

type BVMStore interface {
    GetDB() *leveldb.DB
    SaveBlock(block types.Block) error
    Put(key string, value interface{}) error
    Get(key string, target interface{}) error
    GetAddressHistory(addr string) ([]types.Transaction, error)
    GetBlockByHeight(height int) (types.Block, error)
    LoadFullChain() ([]types.Block, error)

    // 🚩 2. Gunakan tipe Batch yang sudah didefinisikan tadi
    GetLatestBlocks(limit int) ([]types.Block, error)
    PrefixScan(prefix string) ([][]byte, error)
    NewBatch() Batch
    PutToBatch(batch Batch, key string, value interface{}) error
    WriteBatch(batch Batch) error

    Close() error
}
