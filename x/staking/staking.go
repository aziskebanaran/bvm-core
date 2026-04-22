package staking

import (
    "github.com/aziskebanaran/bvm-core/pkg/storage" // 🚩 Gunakan Store Sultan
    "github.com/aziskebanaran/bvm-core/x"
    "github.com/aziskebanaran/bvm-core/x/staking/types"
    "fmt"
    "sync"
    "sort"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type StakingEngine struct {
    Validators  map[string]*types.Validator
    Delegations map[string]uint64
    Mu          sync.RWMutex
    Store       storage.BVMStore // 🚩 Ganti DB jadi Store
    Wasm        x.WasmKeeper
}

func NewStakingEngine(store storage.BVMStore, wk x.WasmKeeper) *StakingEngine {

    return &StakingEngine{
        Validators:  make(map[string]*types.Validator),
	Delegations: make(map[string]uint64),
        Store:       store,
        Wasm:        wk,
    }
}

// SaveToDB: Sekarang otomatis pakai MsgPack/JSON via Store
func (s *StakingEngine) SaveToDB(addr string) {
    v, ok := s.Validators[addr]
    if ok {
        s.Store.Put("val:"+addr, v)
    }
}

func (s *StakingEngine) LoadAll() {
    // 1. Ambil DB mentah melalui iterator untuk mendapatkan KEY dan VALUE
    db := s.Store.GetDB() 
    if db == nil {
        return
    }

    // Gunakan prefix scan yang presisi
    iter := db.NewIterator(util.BytesPrefix([]byte("val:")), nil)
    defer iter.Release()

    s.Mu.Lock()
    defer s.Mu.Unlock()

    count := 0
    for iter.Next() {
        key := string(iter.Key())

        // 🚩 GUNAKAN s.Store.Get agar proses Unmarshal-nya 
        // sama persis dengan s.Store.Put saat SaveToDB
        var v types.Validator
        err := s.Store.Get(key, &v) 
        if err == nil {
            s.Validators[v.Address] = &v
            count++
        }
    }

    if count > 0 {
        fmt.Printf("✅ [STAKING] Memori Pulih! %d Validator siap bertugas.\n", count)
    } else {
        fmt.Println("❌ [STAKING] Disk terbaca, tapi tidak ada data validator yang valid.")
    }
}


func (s *StakingEngine) LockForStake(addr string, amount uint64) error {
    s.Mu.Lock()
    defer s.Mu.Unlock()

    // 1. Inisialisasi jika validator baru pertama kali muncul
    if _, exists := s.Validators[addr]; !exists {
        s.Validators[addr] = &types.Validator{
            Address:      addr,
            Status:       "ACTIVE",
            IsActive:     true,
            Commission:   0.1,
            StakedAmount: 0,
            Power:        0,
        }
    }

    // 2. Tambahkan saldo koin fisik (StakedAmount)
    s.Validators[addr].StakedAmount += amount

    // 🚩 3. Sinkronisasi Power (Bobot Suara untuk Konsensus)
    // Di BVM Sultan, kita gunakan rasio 1:1 antara BVM dan Power.
    // Kita konversi uint64 ke int64 untuk field Power.
    s.Validators[addr].Power = int64(s.Validators[addr].StakedAmount)

    // 4. Pastikan status tetap aktif karena ada saldo
    s.Validators[addr].IsActive = true
    s.Validators[addr].Status = "ACTIVE"

    return nil
}

// UnlockFromStake: Mengurangi jatah stake validator (Pintu keluar dana)
func (s *StakingEngine) UnlockFromStake(addr string, amount uint64) error {
    s.Mu.Lock()
    defer s.Mu.Unlock()

    v, exists := s.Validators[addr]
    if !exists {
        return fmt.Errorf("validator %s tidak terdaftar di engine", addr)
    }

    // 🚩 VALIDASI KRUSIAL: Jangan biarkan rakyat menarik lebih dari yang mereka punya
    if v.StakedAmount < amount {
        return fmt.Errorf("saldo stake tidak cukup (Ada: %.4f, Minta Tarik: %.4f)", v.StakedAmount, amount)
    }

    // Kurangi tumpukan stake di engine
    v.StakedAmount -= amount

    // Jika stake jadi nol, Sultan bisa memilih tetap ACTIVE atau mengubah statusnya
    if v.StakedAmount <= 0 {
        v.Status = "INACTIVE" 
    }

    return nil
}


func (s *StakingEngine) QueryTopValidators(n int) []*types.Validator {
    s.Mu.RLock()
    defer s.Mu.RUnlock()

     //fmt.Printf("🔍 [DEBUG] Jumlah validator di RAM: %d\n", len(s.Validators))

    var list []*types.Validator
    for _, v := range s.Validators {
        list = append(list, v)
    }
    sort.Slice(list, func(i, j int) bool {
        return list[i].StakedAmount > list[j].StakedAmount
    })
    if len(list) > n { return list[:n] }
    return list
}

func (s *StakingEngine) GetValidatorStake(addr string) uint64 {
    s.Mu.RLock()
    defer s.Mu.RUnlock()

    nativeStake := uint64(0)
    if v, ok := s.Validators[addr]; ok {
        nativeStake = v.StakedAmount
    }

    // ✅ Sekarang s.Wasm sudah dikenal!
    // Kita asumsikan ada fungsi QueryDPosPower di WasmKeeper Sultan
    wasmStake := uint64(0)
    if s.Wasm != nil {
        // Panggil state dari Smart Contract DPoS
        wasmStake = s.Wasm.GetContractBalance(addr, "DPoS_CONTRACT")
    }

    return nativeStake + wasmStake
}

func (s *StakingEngine) ProcessIncentive(addr string, amount uint64) error {
    s.Mu.Lock()
    defer s.Mu.Unlock()
    if _, exists := s.Validators[addr]; !exists {
        return fmt.Errorf("validator tidak ditemukan")
    }
    s.Validators[addr].StakedAmount += amount // Hadiah langsung di-stake ulang
    return nil
}


func (s *StakingEngine) AutoDelegate(addr string, amount uint64) error {
    if s == nil { return nil }

    s.Mu.Lock()
    // 1. Pastikan Validator Sultan terdaftar di RAM
    if _, exists := s.Validators[addr]; !exists {
        s.Validators[addr] = &types.Validator{
            Address:      addr,
            Status:       "ACTIVE",
            IsActive:     true,
            StakedAmount: 0,
            Power:        0,
        }
    }

    // 2. Update saldo stake & power
    s.Validators[addr].StakedAmount += amount
    s.Validators[addr].Power = int64(s.Validators[addr].StakedAmount)
    s.Mu.Unlock()

    // 🚩 JURUS PAMUNGKAS: Paksa tulis ke LevelDB detik ini juga!
    s.SaveToDB(addr) 

    fmt.Printf("💾 [STAKING] Data validator %s berhasil dipahat ke Disk!\n", addr[:12])
    return nil
}
