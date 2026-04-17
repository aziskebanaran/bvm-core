package keeper

import (
    "github.com/aziskebanaran/BVM.core/pkg/storage" // 🚩 Import Store
    "github.com/aziskebanaran/BVM.core/x"
    staketypes "github.com/aziskebanaran/BVM.core/x/staking/types"
    "github.com/aziskebanaran/BVM.core/x/staking"
    "fmt"
    "sync"
)

type Keeper struct {
    Store         storage.BVMStore
    Bank          x.BankKeeper
    StakingEngine *staking.StakingEngine
    Mu            sync.RWMutex
}

func NewKeeper(store storage.BVMStore, bk x.BankKeeper, wk x.WasmKeeper) *Keeper { // 👈 Tambah wk
    engine := staking.NewStakingEngine(store, wk) // 👈 Oper wk ke sini
    engine.LoadAll()

    return &Keeper{
        Store:         store,
        Bank:          bk,
        StakingEngine: engine,
    }
}

// Stake: Logika utama Sultan
func (k *Keeper) Stake(addr string, amount uint64) error {
        k.Mu.Lock()
        defer k.Mu.Unlock()

        // 1. Cek Saldo di Bank
        // 🚩 PERBAIKAN: Ambil saldo dulu, baru bandingkan. 
        // Jangan menggabungkan assignment dan perbandingan dalam satu baris if tanpa pemisah ';'.
        balance := k.Bank.GetBalance(addr, "BVM") 
        if balance < amount {
                return fmt.Errorf("❌ saldo tidak mencukupi untuk staking (Saldo: %.4f, Minta: %.4f)", balance, amount)
        }

        // 2. Kunci di Engine
        err := k.StakingEngine.LockForStake(addr, amount)
        if err != nil { return err }

        // 3. POTONG SALDO di Bank (Pindah dari dompet ke vault staking)
        k.Bank.SubBalance(addr, amount, "BVM")

        // 4. Simpan ke Database agar permanen
        k.StakingEngine.SaveToDB(addr)

        fmt.Printf("🔒 [STAKING] %s berhasil stake %.4f BVM\n", addr, amount)
        return nil
}

func (k *Keeper) Unstake(addr string, amount uint64) error {
    k.Mu.Lock()
    defer k.Mu.Unlock()

    // 🚩 1. VALIDASI DI ENGINE: Apakah user punya saldo stake yang cukup?
    // Kita harus memanggil StakingEngine untuk mengurangi jatah stake mereka dulu.
    // Jika saldo stake mereka < amount, fungsi ini harus gagal.
    err := k.StakingEngine.UnlockFromStake(addr, amount)
    if err != nil {
        return fmt.Errorf("❌ gagal unstake: %v", err)
    }

    // 🚩 2. KEMBALIKAN SALDO KE BANK: Setelah ditarik dari engine, masukkan ke dompet.
    k.Bank.AddBalance(addr, amount, "BVM")

    // 🚩 3. PERSISTENSI: Simpan perubahan status stake ke database
    k.StakingEngine.SaveToDB(addr)

    fmt.Printf("🔓 [UNSTAKE] %s berhasil menarik kembali %.4f BVM ke dompet\n", addr, amount)
    return nil
}


func (k *Keeper) Delegate(delegator, validator string, amount uint64) error {
    // Logika delegasi suara/koin Sultan
    fmt.Printf("🗳️ [DELEGATE] %s -> %s: %.4f\n", delegator, validator, amount)
    return nil
}


// AutoDelegate: Untuk hadiah blok otomatis
func (k *Keeper) AutoDelegate(addr string, amount uint64) error {
	err := k.StakingEngine.AutoDelegate(addr, amount)
	if err == nil {
		k.StakingEngine.SaveToDB(addr)
	}
	return err
}

// Method tambahan agar API tidak error
func (k *Keeper) GetValidators() []staketypes.Validator {
	raw := k.StakingEngine.QueryTopValidators(100)
	var result []staketypes.Validator
	for _, v := range raw {
		result = append(result, *v)
	}
	return result
}

func (k *Keeper) GetTopValidators(n int) []staketypes.Validator {
    // Panggil engine Sultan yang sudah canggih
    raw := k.StakingEngine.QueryTopValidators(n)
    var result []staketypes.Validator
    for _, v := range raw {
        if v != nil {
            result = append(result, *v)
        }
    }
    return result
}


// GetValidatorStake: Memenuhi janji di x/interfaces.go
func (k *Keeper) GetValidatorStake(addr string) uint64 {
        return k.StakingEngine.GetValidatorStake(addr)
}

func (k *Keeper) ProcessIncentive(addr string, amount uint64) error {
    // Gunakan StakingEngine Sultan untuk menambah saldo stake secara otomatis
    err := k.StakingEngine.ProcessIncentive(addr, amount)
    if err != nil {
        return err
    }

    // Jangan lupa simpan ke database agar tidak hilang saat mati lampu
    k.StakingEngine.SaveToDB(addr)
    return nil
}

// ModifyValidatorPower: Jembatan yang dipanggil oleh WASM (dpos.go)
func (k *Keeper) ModifyValidatorPower(address string, amount uint64, isAdding bool) error {
    k.Mu.Lock()
    defer k.Mu.Unlock()

    var err error
    if isAdding {
        err = k.StakingEngine.LockForStake(address, amount)
    } else {
        err = k.StakingEngine.UnlockFromStake(address, amount)
    }

    if err == nil {
        // 🚩 Krusial: Update Power berdasarkan StakedAmount terbaru
        // Kita asumsikan 1 BVM = 1 Power
        if v, ok := k.StakingEngine.Validators[address]; ok {
            v.Power = int64(v.StakedAmount) 
            if v.StakedAmount > 0 {
                v.IsActive = true
                v.Status = "ACTIVE"
            }
        }
        // Simpan ke Database permanen
        k.StakingEngine.SaveToDB(address)
    }

    return err
}


func (k *Keeper) GetValidatorObjects() ([]staketypes.Validator, error) {
    // Memanggil mesin engine Sultan
    raw := k.StakingEngine.QueryTopValidators(100)
    
    // Inisialisasi slice kosong (agar tidak null di JSON)
    var result []staketypes.Validator = []staketypes.Validator{}
    
    for _, v := range raw {
        if v != nil {
            result = append(result, *v)
        }
    }
    return result, nil
}


func (k *Keeper) QueryTopValidators(n int) []*staketypes.Validator {
    return k.StakingEngine.QueryTopValidators(n)
}

// GetValidatorPower: Digunakan WASM untuk melihat kekuatan voting
func (k *Keeper) GetValidatorPower(address string) uint64 {
    return k.StakingEngine.GetValidatorStake(address)
}
