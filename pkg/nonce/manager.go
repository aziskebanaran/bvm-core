package nonce

import (
	"sync"
	"github.com/aziskebanaran/bvm-core/pkg/logger"  // 🚩 Gunakan Logger Pintar
	"github.com/aziskebanaran/bvm-core/pkg/storage"
)

type NonceManager struct {
	store storage.BVMStore
	mu    sync.RWMutex
	cache map[string]uint64
}

func NewNonceManager(store storage.BVMStore) *NonceManager {
	return &NonceManager{
		store: store,
		cache: make(map[string]uint64),
	}
}

func (m *NonceManager) GetNextNonce(address string) uint64 {
        m.mu.RLock()
        defer m.mu.RUnlock()

        // 1. Cek di RAM Cache
        if n, ok := m.cache[address]; ok {
                return n
        }

        // 2. Baca dari Store
        var n uint64
        err := m.store.Get("n:"+address, &n)
        if err != nil {

                return 0
        }

        return n
}

func (m *NonceManager) Increment(address string) error {
        m.mu.Lock()
        defer m.mu.Unlock()

        current := m.GetNextNonce(address)
        newNonce := current + 1

        m.cache[address] = newNonce
        return m.store.Put("n:"+address, newNonce)
}

// HealthCheckNonce: Memastikan konsistensi
func (m *NonceManager) HealthCheckNonce(address string) (bool, uint64, uint64) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cacheVal := m.cache[address]

	var dbVal uint64
	err := m.store.Get("n:"+address, &dbVal) // 🚩 Perbaikan Get
	if err != nil {
		dbVal = 0
	}

	if cacheVal == 0 && dbVal > 0 {
		return true, dbVal, dbVal
	}

	return cacheVal == dbVal, cacheVal, dbVal
}

// ManualOverride: Hak Veto Sultan
func (m *NonceManager) ManualOverride(address string, newNonce uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cache[address] = newNonce
	err := m.store.Put("n:"+address, newNonce) // 🚩 Perbaikan Put
	if err != nil {
		logger.Error("NONCE", "Gagal override: ", err)
		return err
	}

	logger.Success("NONCE", "👑 Sultan memaksakan Nonce ", address[:10], " menjadi ", newNonce)
	return nil
}

func (m *NonceManager) SetNonce(address string, newNonce uint64) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.cache[address] = newNonce
    return m.store.Put("n:"+address, newNonce)
}
