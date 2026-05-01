package state

import (
    "github.com/aziskebanaran/bvm-lib/game"
    "github.com/aziskebanaran/bvm-lib/storage"
    "fmt"
)

type InventoryManager struct {
    db *storage.KVStore
}

// SaveInventory: Simpan tas ke LevelDB
func (m *InventoryManager) SaveInventory(inv game.Inventory) error {
    key := fmt.Sprintf("inv:%s", inv.OwnerAddress)
    // Langsung simpan struct! KVStore Jenderal yang sudah pakai Msgpack akan urus sisanya.
    return m.db.Put(key, inv)
}

// LoadInventory: Ambil tas dari LevelDB
func (m *InventoryManager) LoadInventory(playerAddr string) (*game.Inventory, error) {
    key := fmt.Sprintf("inv:%s", playerAddr)
    var inv game.Inventory
    
    // KVStore akan Unmarshal otomatis dari byte ke struct
    err := m.db.Get(key, &inv)
    if err != nil {
        return nil, err
    }
    return &inv, nil
}
