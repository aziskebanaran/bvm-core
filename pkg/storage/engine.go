package storage

import (
        "github.com/aziskebanaran/bvm-core/x/bvm/types"
        "fmt"
        "sort"
        "github.com/syndtr/goleveldb/leveldb"
        "github.com/syndtr/goleveldb/leveldb/util"
        "github.com/vmihailenco/msgpack/v5"
        "github.com/syndtr/goleveldb/leveldb/opt"
        "github.com/syndtr/goleveldb/leveldb/filter" // 🚩 TAMBAHKAN INI, JENDERAL!
)


type LevelDBStore struct {
	db *leveldb.DB
}

// NewLevelDBStore sekarang menerima parameter cacheMB agar fleksibel
func NewLevelDBStore(path string, cacheMB int) (*LevelDBStore, error) {
    options := &opt.Options{
        Compression: opt.SnappyCompression,

        // 🚩 DINAMIS: Inti Blockchain bisa 8MB, App User bisa 2MB
        BlockCacheCapacity: cacheMB * opt.MiB,

        Filter:                 filter.NewBloomFilter(10),
        OpenFilesCacheCapacity: 50,
    }

    db, err := leveldb.OpenFile(path, options)
    if err != nil {
        return nil, fmt.Errorf("❌ Gagal akses LevelDB di %s: %v", path, err)
    }
    return &LevelDBStore{db: db}, nil
}


func (s *LevelDBStore) GetDB() *leveldb.DB { return s.db }

// 🚩 STANDAR BARU: Simpan Blok SEKALIGUS Indeks Transaksinya
func (s *LevelDBStore) SaveBlock(block types.Block) error {
    // 1. Simpan Bloknya dulu seperti biasa
    err := s.Put(fmt.Sprintf("b:%d", block.Index), block)
    if err != nil {
        return err
    }

    // 2. Langsung Indeks semua transaksi yang ada di dalam blok tersebut
    if len(block.Transactions) > 0 {
        return s.IndexTransactions(block.Transactions)
    }

    return nil
}

// 🚩 USULAN SULTAN: Indexing Transaksi agar pencarian secepat kilat (O(1))
func (s *LevelDBStore) IndexTransactions(txs []types.Transaction) error {
    batch := s.NewBatch()
    for _, tx := range txs {
        // ✅ SEKARANG BENAR: Menggunakan tx.ID sesuai struct Transaction
        key := fmt.Sprintf("tx:%s", tx.ID) 

        err := s.PutToBatch(batch, key, tx)
        if err != nil {
            return err
        }
    }
    return s.WriteBatch(batch)
}

func (s *LevelDBStore) Put(key string, value interface{}) error {
	data, err := msgpack.Marshal(value)
	if err != nil {
		return err
	}
	return s.db.Put([]byte(key), data, nil)
}

func (s *LevelDBStore) Get(key string, target interface{}) error {
	data, err := s.db.Get([]byte(key), nil)
	if err != nil {
		return err
	}
	return msgpack.Unmarshal(data, target)
}

// 🚩 STANDAR BARU: History menggunakan prefix h:[addr]
func (s *LevelDBStore) GetAddressHistory(addr string) ([]types.Transaction, error) {
	var history []types.Transaction
	prefix := []byte(fmt.Sprintf("h:%s:", addr)) // Kita samakan dengan backfill.go
	iter := s.db.NewIterator(util.BytesPrefix(prefix), nil)
	defer iter.Release()
	for iter.Next() {
		var tx types.Transaction
		if err := msgpack.Unmarshal(iter.Value(), &tx); err == nil {
			history = append(history, tx)
		}
	}
	return history, nil
}

func (s *LevelDBStore) GetBlockByHeight(height int) (types.Block, error) {
	var block types.Block
	key := fmt.Sprintf("b:%d", height)
	err := s.Get(key, &block)
	if err != nil {
		return types.Block{}, fmt.Errorf("blok #%d tidak ditemukan", height)
	}
	return block, nil
}

// Digunakan untuk API Explorer Sultan
func (s *LevelDBStore) LoadFullChain() ([]types.Block, error) {
	var chain []types.Block
	iter := s.db.NewIterator(util.BytesPrefix([]byte("b:")), nil)
	defer iter.Release()
	for iter.Next() {
		var block types.Block
		if err := msgpack.Unmarshal(iter.Value(), &block); err == nil {
			chain = append(chain, block)
		}
	}
	sort.Slice(chain, func(i, j int) bool {
		return chain[i].Index < chain[j].Index
	})
	return chain, nil
}

// --- BATCH OPERATIONS (Jantung Performa) ---
func (s *LevelDBStore) NewBatch() Batch {
        return &leveldb.Batch{}
}

// Gunakan tipe Batch untuk argumen pertama
func (s *LevelDBStore) PutToBatch(batch Batch, key string, value interface{}) error {
        b, ok := batch.(*leveldb.Batch) // Casting ini tetap diperlukan dan sudah benar
        if !ok {
                return fmt.Errorf("batch tidak valid")
        }
        data, err := msgpack.Marshal(value)
        if err != nil {
                return err
        }
        b.Put([]byte(key), data)
        return nil
}

// Gunakan tipe Batch untuk argumen pertama
func (s *LevelDBStore) WriteBatch(batch Batch) error {
        b, ok := batch.(*leveldb.Batch)
        if !ok {
                return fmt.Errorf("tipe batch tidak valid")
        }

   return s.db.Write(b, &opt.WriteOptions{Sync: true})
}


// PrefixScan: Mencari semua data yang kuncinya diawali dengan prefix tertentu
func (s *LevelDBStore) PrefixScan(prefix string) ([][]byte, error) {
	var results [][]byte
	iter := s.db.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	defer iter.Release()

	for iter.Next() {
		// Kita ambil nilainya saja
		value := make([]byte, len(iter.Value()))
		copy(value, iter.Value())
		results = append(results, value)
	}

	if err := iter.Error(); err != nil {
		return nil, err
	}
	return results, nil
}

// GetLatestBlocks: Mengambil N blok terbaru untuk dashboard real-time
func (s *LevelDBStore) GetLatestBlocks(limit int) ([]types.Block, error) {
    var blocks []types.Block
    // Gunakan iterator yang mulai dari kunci terakhir (tertinggi)
    iter := s.db.NewIterator(util.BytesPrefix([]byte("b:")), nil)
    defer iter.Release()

    count := 0
    // Gerak mundur dari blok terbaru
    if iter.Last() {
        for count < limit {
            var block types.Block
            if err := msgpack.Unmarshal(iter.Value(), &block); err == nil {
                blocks = append(blocks, block)
            }
            if !iter.Prev() { break }
            count++
        }
    }
    return blocks, nil
}



func (s *LevelDBStore) Close() error { return s.db.Close() }
