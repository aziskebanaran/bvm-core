package keeper

import (
    "fmt" // 🚩 Tambahkan ini!
    "bvm.core/pkg/storage"
    "bvm.core/x/bvm/types"
)

const (
    BurnAddress = "bvmf000000000000000000000000000000000000burn"
)

type BankKeeper struct {
    Store  storage.BVMStore
    Batch  storage.Batch // 🚩 Akan diisi oleh Kernel saat ExecuteBlock
    Params *types.Params
}

func NewBankKeeper(store storage.BVMStore, params *types.Params) *BankKeeper {
    return &BankKeeper{
        Store:  store,
        Params: params,
    }
}

// --- 🛡️ PROTEKSI KEDAULATAN ---

func (bk *BankKeeper) isNative(symbol string) bool {
    return symbol == "BVM" || symbol == "" || symbol == bk.Params.NativeSymbol
}

// GetBalance: Mengambil saldo dengan aman
func (bk *BankKeeper) GetBalance(addr string, symbol string) uint64 {
    var balance uint64
    key := "t:" + symbol + ":" + addr
    if bk.isNative(symbol) {
        key = "a:" + addr
    }

    err := bk.Store.Get(key, &balance)
    if err != nil {
        return 0
    }
    return balance
}

func (bk *BankKeeper) SetBalance(addr string, amount uint64, symbol string) {
    key := "t:" + symbol + ":" + addr
    if bk.isNative(symbol) {
        key = "a:" + addr
    }

    if bk.Batch != nil {
        // 🚩 SOLUSI: Panggil via bk.Store, kirim bk.Batch sebagai argumen
        bk.Store.PutToBatch(bk.Batch, key, amount)
    } else {
        bk.Store.Put(key, amount)
    }
}


// AddBalance & SubBalance sekarang otomatis Atomic karena memanggil SetBalance di atas!
func (bk *BankKeeper) AddBalance(addr string, amount uint64, symbol string) error {
    current := bk.GetBalance(addr, symbol)
    bk.SetBalance(addr, current+amount, symbol)
    return nil
}

func (bk *BankKeeper) SubBalance(addr string, amount uint64, symbol string) error {
    current := bk.GetBalance(addr, symbol)
    if current < amount {
        return fmt.Errorf("🚨 Saldo %s tidak cukup! Punya: %d, Butuh: %d", symbol, current, amount)
    }
    bk.SetBalance(addr, current-amount, symbol)
    return nil
}

// --- 🗂️ METADATA AKUN ---

func (bk *BankKeeper) GetAccount(addr string) (types.Account, error) {
    var acc types.Account
    err := bk.Store.Get("acc:"+addr, &acc)
    if err != nil {
        return types.Account{Address: addr, Balances: make(map[string]uint64)}, nil
    }
    return acc, nil
}

func (bk *BankKeeper) SaveAccount(acc *types.Account) error {
    return bk.Store.Put("acc:"+acc.Address, acc)
}


func (bk *BankKeeper) GetOrCreateAccount(addr string) *types.Account {
    var dbAcc types.Account
    err := bk.Store.Get("acc:"+addr, &dbAcc)

    if err == nil {
        return &dbAcc // Kembalikan apa adanya dari Disk
    }

    return &types.Account{
        Address: addr,
        Status:  "active",
        Nonce:   0,
        // Balances dibiarkan kosong, BVM tidak disimpan di sini!
    }
}


// --- 📊 STATISTIK (Scan via Prefix) ---

func (bk *BankKeeper) GetTotalSupply(symbol string) uint64 {
    // Di sini Sultan bisa memanggil fungsi iterator Store untuk menghitung 
    // semua key yang diawali dengan "t:symbol:"
    return bk.ScanTotalSupplyFromDB(symbol)
}

func (bk *BankKeeper) GetTotalBurned(symbol string) uint64 {
    return bk.GetBalance(BurnAddress, symbol)
}

// --- Formalitas Interface ---
func (bk *BankKeeper) GetAccounts() map[string]*types.Account { return nil }
func (bk *BankKeeper) LoadInitialState()                     {}

