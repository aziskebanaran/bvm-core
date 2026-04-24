package keeper

import (
	"strings"
	"crypto/sha256"

	"github.com/aziskebanaran/bvm-core/pkg/logger"
	"github.com/aziskebanaran/bvm-core/pkg/storage"
	"github.com/aziskebanaran/bvm-core/x"
	"github.com/aziskebanaran/bvm-core/x/bvm/types"
	staketypes "github.com/aziskebanaran/bvm-core/x/staking/types"
	"fmt"
)

type Keeper struct {
	Store      storage.BVMStore
	Blockchain *types.Blockchain
	Params     *types.Params
        TotalSupplyBVM uint64
        TotalBurnedBVM uint64

	// Rantai Komando (Menteri-Menteri Sultan)
	NonceMgr x.NonceKeeper
	Bank     x.BankKeeper
	Auth     x.AuthKeeper
	Mempool  x.MempoolKeeper
	Staking  x.StakingKeeper
	Wasm     x.WasmKeeper
	P2P      x.P2PKeeper
        Storage  x.StorageModuleKeeper
	Factory x.FactoryKeeper
}

func NewKeeper(
	store storage.BVMStore,
	bc *types.Blockchain,
	params *types.Params,
	bk x.BankKeeper,
	ak x.AuthKeeper,
	nk x.NonceKeeper,
	mk x.MempoolKeeper,
	sk x.StakingKeeper,
	wk x.WasmKeeper,
	pk x.P2PKeeper,
        stk x.StorageModuleKeeper,
	fk x.FactoryKeeper,
) *Keeper {
	k := &Keeper{
		Store:      store,
		Blockchain: bc,
		Params:     params,
		Bank:       bk,
		Auth:       ak,
		NonceMgr:   nk,
		Mempool:    mk,
		Staking:    sk,
		Wasm:       wk,
		P2P:        pk,
                Storage:    stk,
		Factory: fk,
	}

	k.InitialSync()
	return k
}

func (k *Keeper) InitialSync() {
    var diskHeight uint64
    if err := k.Store.Get(k.keyMeta("height"), &diskHeight); err != nil {
        diskHeight = 0
    }

    logger.Info("SYSTEM", fmt.Sprintf("🔍 Sinkronisasi Siklus (Tinggi: %d)...", diskHeight))

    // Kita periksa blok terakhir yang ada di Disk
    if diskHeight > 0 {
        lastBlock, err := k.Store.GetBlockByHeight(int(diskHeight))
        if err == nil {
            // 🚩 DISINI RENCANA SULTAN BERAKSI:
            // Jalankan validasi adaptif Sultan.
            if err := k.ValidateBlockTransactions(lastBlock); err != nil {
                logger.Error("SYSTEM", "🚨 Checkpoint Gagal: "+err.Error())
                // Jika benar-benar rusak parah, baru kita mundur 1 langkah
                k.Blockchain.Height = int64(diskHeight - 1)
            } else {
                // Jika valid (atau dimaafkan karena checkpoint), lanjut!
                k.Blockchain.Height = int64(diskHeight)
                k.Blockchain.LatestHash = lastBlock.Hash
                k.Blockchain.Chain = []types.Block{lastBlock}
                logger.Success("SYSTEM", "✅ Siklus Aman. Kernel Aktif.")
            }
        }
    } else {
        k.Blockchain.Height = 0
        k.Blockchain.LatestHash = strings.Repeat("0", 64)
    }

    k.TotalSupplyBVM = k.Params.GetExpectedSupply(k.Blockchain.Height)
}

func (k *Keeper) ProcessTransaction(tx types.Transaction) error {
    // 1. Validasi Signature
    if !k.Auth.VerifyTransaction(tx) {
        return fmt.Errorf("❌ INVALID SIGNATURE: Hash transaksi tidak cocok dengan tanda tangan")
    }

    // 2. LOGIKA FIREWALL NONCE
    actualInDisk := k.GetNextNonce(tx.From)
    lastInRAM := k.Mempool.GetHighestNonce(tx.From)

    // Tentukan ambang batas lompatan (Jump)
    expectedNextNonce := actualInDisk
    if lastInRAM >= actualInDisk {
        expectedNextNonce = lastInRAM + 1
    }

    // A. Filter Basi (Sudah Masuk Blok)
    if tx.Nonce < actualInDisk {
        return fmt.Errorf("❌ NONCE BASI: DB butuh %d, Anda kirim %d", actualInDisk, tx.Nonce)
    }

    // B. Filter Lompatan (Mencegah Gap)
    // Sultan hanya boleh kirim Nonce yang urut, 
    // ATAU mengirim ulang Nonce yang sedang mengantre (untuk Update/Replace).
    if tx.Nonce > expectedNextNonce {
        return fmt.Errorf("❌ NONCE JUMP: Harusnya %d, Anda kirim %d", expectedNextNonce, tx.Nonce)
    }

    // 🚩 CATATAN: Filter "DUPLICATE" dihapus dari sini! 
    // Kita serahkan ke Mempool.Add(tx) untuk melakukan replaceIfDuplicate.
    // 3. CEK SALDO (Filter Ekonomi)
    // Ingat Jenderal: Pengirim harus punya saldo untuk (Amount + Fee)
    totalRequired := tx.Fee
    if tx.Symbol == "BVM" {
        totalRequired += tx.Amount
    }

    // Gunakan GetBalanceBVM agar lebih spesifik ke koin utama
    if k.GetBalanceBVM(tx.From) < totalRequired {
        return fmt.Errorf("🚨 Saldo BVM tidak cukup untuk membayar biaya & transfer!")
    }


    // 4. KIRIM KE MEMPOOL
    // Di sini Mempool akan menjalankan m.replaceIfDuplicate(tx) jika Nonce sama.
    if err := k.Mempool.Add(tx); err != nil {
        return err // Mempool sekarang sudah pintar menghandle duplicate
    }

    return nil
}


func (k *Keeper) GetSecureBalance(address string) (types.WalletState, bool) {
    balanceAtomic := k.GetBalanceBVM(address)

    nonce := k.GetNextNonce(address)

    if balanceAtomic == 0 && nonce == 0 {
        return types.WalletState{}, false
    }

    return types.WalletState{
        Address:        address,
        BalanceAtomic:  balanceAtomic,
        BalanceDisplay: k.Params.FormatDisplay(balanceAtomic),
        Nonce:          nonce,
        Symbol:         k.Params.NativeSymbol,
        Status:         "active",
    }, true
}

// CalculateBatchHash: Mengompres 10 blok menjadi 1 Hash Acuan tunggal
func (k *Keeper) CalculateBatchHash(startHeight, endHeight int) string {
    var combinedHashes string

    for i := startHeight; i <= endHeight; i++ {
        block, err := k.Store.GetBlockByHeight(i)
        if err == nil {
            combinedHashes += block.Hash
        }
    }

    // Jika kosong (misal database korup), berikan nilai default
    if combinedHashes == "" {
        return strings.Repeat("0", 64)
    }

    // Hashing gabungan 10 hash menjadi 1 hash baru
    hash := sha256.Sum256([]byte(combinedHashes))
    return fmt.Sprintf("%x", hash)
}


func (k *Keeper) FinalizeBlock(block types.Block) {
    // Karena ExecuteBlock sudah menyimpan blok ke disk, 
    // di sini kita fokus pada logika kompresi siklus.

    if block.Index > 0 && block.Index % 10 == 0 {
        start := block.Index - 9
        if start < 1 { start = 1 }

        // 🚩 HITUNG ANCHOR (KOMPRESI)
        anchorHash := k.CalculateBatchHash(int(start), int(block.Index))

        // Simpan ke Metadata agar blok selanjutnya (kelipatan 10 + 1) bisa mengambilnya
        _ = k.Store.Put(k.keyMeta("cycle_anchor"), anchorHash)

        logger.Success("COMPRESSOR", fmt.Sprintf("📦 Siklus %d-%d Berhasil Dikunci! Anchor: %s", 
            start, block.Index, anchorHash[:16]))
    }
}


// --- 1. HELPER PREFIX (Rahasia Brankas Sultan) ---

func (k *Keeper) keyAcc(addr string) string  { return "a:" + addr }
func (k *Keeper) keyBlock(idx int64) string { return fmt.Sprintf("b:%d", idx) }
func (k *Keeper) keyMeta(attr string) string { return "m:" + attr }

// --- 2. LOGIKA SALDO (PINTAR & ATOMIC) ---

// SetBalanceBVM adalah satu-satunya pintu untuk mengubah saldo BVM di disk/batch
func (k *Keeper) SetBalanceBVM(addr string, amount uint64, batch storage.Batch) error {
    key := k.keyAcc(addr)

    // Jika sedang dalam proses blok (batch tidak nil), masukkan ke antrean
    if batch != nil {
        return k.Store.PutToBatch(batch, key, amount)
    }

    // Jika di luar proses blok (misal: pemberian saldo awal/genesis), langsung pahat
    return k.Store.Put(key, amount)
}

// --- 2. LOGIKA SALDO (Sudah Menggunakan Helper) ---
func (k *Keeper) GetBalanceBVM(addr string) uint64 {
    var balance uint64
    err := k.Store.Get(k.keyAcc(addr), &balance)
    if err != nil {
        return 0
    }
    return balance
}

// AddBalanceBVM: Sekarang mendukung Single & Batch secara otomatis
func (k *Keeper) AddBalanceBVM(addr string, amount uint64, batch storage.Batch) {
    oldBal := k.GetBalanceBVM(addr)
    newBal := oldBal + amount

    if batch != nil {
        // Jika ada batch, masukkan ke antrean blok
        k.Store.PutToBatch(batch, k.keyAcc(addr), newBal)
    } else {
        // Jika tidak ada batch (manual/simulasi), langsung pahat
        k.Store.Put(k.keyAcc(addr), newBal)
    }
}

// SubBalanceBVM: Versi aman dengan pengecekan saldo
func (k *Keeper) SubBalanceBVM(addr string, amount uint64, batch storage.Batch) error {
    oldBal := k.GetBalanceBVM(addr)
    if oldBal < amount {
        return fmt.Errorf("🚨 Saldo BVM %s tidak cukup! Kurang: %d", addr[:8], amount-oldBal)
    }

    newBal := oldBal - amount
    if batch != nil {
        k.Store.PutToBatch(batch, k.keyAcc(addr), newBal)
    } else {
        k.Store.Put(k.keyAcc(addr), newBal)
    }
    return nil
}


// --- DELEGASI INTERFACE (MEMENUHI BVMKeeper) ---

func (k *Keeper) GetBank() x.BankKeeper       { return k.Bank }
func (k *Keeper) GetAuth() x.AuthKeeper       { return k.Auth }
func (k *Keeper) GetMempool() x.MempoolKeeper { return k.Mempool }
func (k *Keeper) GetStaking() x.StakingKeeper { return k.Staking }
func (k *Keeper) GetWasm() x.WasmKeeper       { return k.Wasm }
func (k *Keeper) GetNonce() x.NonceKeeper     { return k.NonceMgr }
func (k *Keeper) GetP2P() x.P2PKeeper         { return k.P2P }
func (k *Keeper) GetMining() x.MiningKeeper   { return k }
func (k *Keeper) GetStore() storage.BVMStore {
    return k.Store
}


// --- 2. LOGIKA BLOCKCHAIN ---

func (k *Keeper) GetLatestBlock() types.Block {
	if k.Blockchain == nil || len(k.Blockchain.Chain) == 0 {
		return types.Block{}
	}
	return k.Blockchain.Chain[len(k.Blockchain.Chain)-1]
}

func (k *Keeper) GetLastBlockHash() string {
	latest := k.GetLatestBlock()
	if latest.Hash == "" {
		return "0000000000000000000000000000000000000000000000000000000000000000"
	}
	return latest.Hash
}

func (k *Keeper) GetLastHeight() int {
	if k.Blockchain == nil {
		return 0
	}
	return int(k.Blockchain.Height)
}

// --- 3. LOGIKA EKONOMI & FEE ---

func (k *Keeper) CalculateDynamicFee() uint64 {
	return k.Blockchain.Params.GetDynamicFee(k.Mempool.Count())
}

func (k *Keeper) SaveAccount(addr string, balance uint64) error {
    return k.Store.Put(k.keyAcc(addr), balance)
}

func (k *Keeper) SearchAccount(query string) (interface{}, bool) {
    state, found := k.GetSecureBalance(query)
    return state, found
}

func (k *Keeper) GetPendingTransactions() []types.Transaction {
    return k.Mempool.GetPendingTransactions()
}

func (k *Keeper) ValidateWithAI() (int, string) {
    return 200, "AI Engine: All systems nominal"
}

func (k *Keeper) FromAtomic(amount uint64) string {
    return k.Params.FormatDisplay(amount)
}

func (k *Keeper) ToAtomic(amount float64) uint64 {
    amountStr := fmt.Sprintf("%.8f", amount)
    return k.Params.ToAtomic(amountStr)
}


func (k *Keeper) GetNextNonce(address string) uint64 {
    return k.NonceMgr.GetNextNonce(address)
}


func (k *Keeper) GetValidatorObjects() ([]staketypes.Validator, error) {
    // 1. Ambil data asli dari Menteri Staking
    rawValidators := k.Staking.GetValidators()
    params := k.GetParamsData()

    // 🚩 PERBAIKAN: Inisialisasi dengan slice kosong, BUKAN nil
    // Ini rahasia agar 'jq' menampilkan [] dan bukan 'null'
    result := []staketypes.Validator{}

    for _, v := range rawValidators {
        dynamicPower := int64(0)
        if params.MinStakeAmount > 0 {
            dynamicPower = int64(v.StakedAmount / params.MinStakeAmount)
        }

        if dynamicPower == 0 && v.Status == "Active" {
            dynamicPower = 1
        }

        result = append(result, staketypes.Validator{
            Address:      v.Address,
            PubKey:       v.PubKey,
            StakedAmount: v.StakedAmount,
            SelfStake:    v.SelfStake,
            Power:        dynamicPower,
            Commission:   v.Commission,
            IsActive:     v.IsActive,
            Status:       v.Status,
        })
    }

    return result, nil
}

func (k *Keeper) GetValidatorCount() int {
    // 🚩 PERBAIKAN: Nama fungsi harus presisi sesuai interface
    validators := k.Staking.GetValidators() 
    count := len(validators)

    // Safety guard: Jika jaringan baru berjalan atau daftar kosong
    if count < 1 {
        return 1
    }
    return count
}

// 🚩 PERBAIKAN: Gunakan (k *Keeper) bukan BaseKeeper
func (k *Keeper) GetCloudStorage() x.StorageModuleKeeper {
    return k.Storage
}

// --- DELEGASI FACTORY ---

func (k *Keeper) GetFactory() x.FactoryKeeper {
    return k.Factory
}
