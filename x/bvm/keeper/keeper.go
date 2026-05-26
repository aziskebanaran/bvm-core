package keeper

import (
	"strings"
	"crypto/sha256"
	"sync"
	"time"

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

        ActiveMinersMu sync.RWMutex
        ActiveMiners   map[string]int64

	NonceMgr x.NonceKeeper
	Bank     x.BankKeeper
	Auth     x.AuthKeeper
	Mempool  x.MempoolKeeper
	Staking  x.StakingKeeper
	Wasm     x.WasmKeeper
	P2P      x.P2PKeeper
        Storage  x.StorageModuleKeeper
	Factory x.FactoryKeeper
	UTXO    x.UTXOKeeper
	Bridge   x.BridgeKeeper
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
	uk x.UTXOKeeper,
	bridgeK x.BridgeKeeper,
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
		UTXO:       uk,
		Bridge:     bridgeK,

		ActiveMiners: make(map[string]int64),
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
    // 🚩 PINTU VIP JEMBATAN (BRIDGE)
    isBridgeTx := tx.Type == "bridge_in" || tx.Type == "submit_proof"

    // Jika ini transaksi bridge, lewati verifikasi signature & nonce untuk debugging
    if isBridgeTx {
        return k.Mempool.Add(tx)
    }

    // 1. Validasi Signature
    if !k.Auth.VerifyTransaction(tx) {
        return fmt.Errorf("❌ INVALID SIGNATURE: Hash transaksi tidak cocok dengan tanda tangan")
    }                                                                                                         // 🚩 PINTU VIP UTXO (0x): Langsung lewat tanpa Firewall Nonce
  // 🚩 PENYEMPURNAAN JALUR VIP UTXO
    isUTXO := strings.HasPrefix(tx.From, "0x") || tx.Type == "utxo_move"

    if isUTXO {
        // A. Cek Ketersediaan Kepingan (PENTING!)
        // Jangan biarkan kepingan double-spend masuk ke Mempool
        if !k.UTXO.IsInputAvailable(tx) {
            return fmt.Errorf("🚨 Kepingan aset sudah terpakai atau tidak valid!")
        }

        // B. Cek Minimal Fee (Pencegahan Spam)
        // Walau UTXO, Fee tetap harus divalidasi agar validator mau kerja
        if tx.Fee < k.Params.MinGasPrice {
            return fmt.Errorf("🚨 Fee terlalu rendah untuk Jalur VIP")
        }

        return k.Mempool.Add(tx)
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

func (k *Keeper) ProcessUTXOTransaction(tx types.Transaction) error {
    logger.Info("DEBUG", "🔍 Memulai validasi VIP untuk TX: "+tx.ID[:8])

    // 1. Validasi Signature
    if !k.Auth.VerifyTransactionUTXO(tx) {
        logger.Error("DEBUG", "❌ Gagal di Signature")
        return fmt.Errorf("❌ INVALID SIGNATURE UTXO")
    }

    // 2. Cek Tipe
    if tx.Type != "utxo_move" && tx.Type != "transfer" {
        logger.Error("DEBUG", "❌ Gagal di Tipe: "+tx.Type)
        return fmt.Errorf("❌ ILLEGAL TYPE")
    }

    // 3. Cek Kepingan
    if !k.UTXO.IsInputAvailable(tx) {
        logger.Error("DEBUG", "❌ Gagal di Input: Kepingan tidak tersedia di database")
        return fmt.Errorf("🚨 Kepingan aset tidak valid!")
    }

    // 🚩 KALIBRASI PERTAHANAN SULTAN: Cegah Eksploitasi Spam Jalur VIP UTXO!
    // Pastikan Miner tetap mendapatkan upah yang adil sesuai Konstitusi L1 Params
    if k.Params != nil && tx.Fee < k.Params.MinGasPrice {
        logger.Error("DEBUG", fmt.Sprintf("❌ Gagal di Validasi Ekonomi: Fee %d kurang dari MinGasPrice %d", tx.Fee, k.Params.MinGasPrice))
        return fmt.Errorf("🚨 Fee terlalu rendah untuk Jalur VIP UTXO!")
    }

    // 4. Masuk Antrean
    logger.Info("MEMPOOL", "💎 Mencoba memasukkan ke antrean...")
    if err := k.Mempool.Add(tx); err != nil {
        logger.Error("DEBUG", "❌ Gagal di Mempool.Add: "+err.Error())
        return fmt.Errorf("🚨 Gagal masuk Mempool: %v", err)
    }

    logger.Success("MEMPOOL", "📥 Berhasil! Cek './bvm mempool' sekarang.")
    return nil
}

func (k *Keeper) GetSecureBalance(address string) (types.WalletState, bool) {
    balanceAtomic := k.GetBalanceBVM(address)
    nonce := k.GetNextNonce(address)

    if balanceAtomic == 0 && nonce == 0 {
        return types.WalletState{}, false
    }

    // 🚩 PERBAIKAN: Mengambil simbol dari Params, bukan hardcode "BVM"
    nativeSymbol := k.Params.NativeSymbol

    return types.WalletState{
        Address:        address,
        BalanceAtomic:  balanceAtomic,
        BalanceDisplay: k.Params.FormatDisplay(balanceAtomic),
        Nonce:          nonce,
        Symbol:         nativeSymbol, // ✅ Dinamis sesuai Params
        Status:         "active",
        Type:           "ACCOUNT_BASED",
    }, true
}

func (k *Keeper) GetSecureBalanceUTXO(address string) (types.WalletState, bool) {
    // 1. Ambil saldo (Hanya tangkap 1 variabel sesuai fungsi aslinya)
    nativeSymbol := k.Params.NativeSymbol
    balanceAtomic := k.UTXO.GetTotalBalance(address, nativeSymbol)

    // 2. Filter: Jika saldo kosong, berarti alamat tidak punya kepingan di 0x
    if balanceAtomic == 0 {
        return types.WalletState{}, false
    }

    // 3. Rakit WalletState dengan kedaulatan UTXO
    return types.WalletState{
        Address:        address,
        BalanceAtomic:  balanceAtomic,
        BalanceDisplay: k.Params.FormatDisplay(balanceAtomic),
        Symbol:         nativeSymbol,
        Nonce:          0,             // 🚩 Wilayah merdeka Nonce
        Status:         "active",
        Type:           "UTXO_BASED",
        UTXOCount:      0,             // Sementara kita nol kan sampai Jenderal update k.UTXO
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

func (k *Keeper) GetUTXO() x.UTXOKeeper {
    return k.UTXO
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
    // 🎯 PENERJEMAH SAKRAL SULTAN (BERBASIS RAM REAL-TIME)
    // Kita panggil fungsi internal k.GetActiveMiners() yang sudah Jenderal buat dengan aman di bawah
    activeMiners := k.GetActiveMiners()
    now := time.Now().Unix()

    realActiveCount := 0
    for _, lastPing := range activeMiners {
        // Hanya hitung miner/validator yang detak jantungnya segar di bawah 180 detik (3 menit)
        if now - lastPing < 180 {
            realActiveCount++
        }
    }

    // Safety guard: Jika miner kedua mati atau hanya ada internal miner tunggal
    // Kembali ke angka 1 agar params mengembalikan 10 BVM utuh
    if realActiveCount <= 1 {
        return 1
    }

    return realActiveCount
}


// 🚩 PERBAIKAN: Gunakan (k *Keeper) bukan BaseKeeper
func (k *Keeper) GetCloudStorage() x.StorageModuleKeeper {
    return k.Storage
}

// --- DELEGASI FACTORY ---

func (k *Keeper) GetFactory() x.FactoryKeeper {
    return k.Factory
}

func (k *Keeper) extractUniqueWallets(txs []types.Transaction) int {
    wallets := make(map[string]bool)
    for _, tx := range txs {
        wallets[tx.From] = true
        // Jika Jenderal ingin menghitung penerima juga:
        // wallets[tx.To] = true 
    }
    return len(wallets)
}

func (k *Keeper) GetRegisteredVaults() []string {
    // Pastikan Storage adalah objek yang memiliki metode GetVaultList()
    // Karena field Anda di struct adalah 'Storage', maka panggilannya:
    return k.Storage.GetVaultList()
}

func (k *Keeper) InitSystemVaults() {
    vaultAddr := "bvmf_market_system_vault"
    
    // 🚩 PERBAIKAN: Gunakan GetBalanceBVM untuk cek keberadaan
    // Di BVM, jika saldo tidak ada di DB, dia akan mengembalikan 0.
    // Kita cek apakah akun ini sudah pernah "disentuh" atau belum.
    
    currentBal := k.GetBalanceBVM(vaultAddr)
    
    // Jika saldo 0, kita inisialisasi agar alamat ini terdaftar di State
    // Kita bisa melakukan SaveAccount untuk memastikannya tertulis di LevelDB
    if currentBal == 0 {
        err := k.SaveAccount(vaultAddr, 0) // Inisialisasi saldo 0
        if err != nil {
            logger.Error("SYSTEM", "❌ Gagal mengaktifkan Brankas Market: "+err.Error())
        } else {
            fmt.Printf("🏦 [SYSTEM] Brankas Market berhasil diaktifkan: %s\n", vaultAddr)
        }
    }
}


// 👑 SUNTIKKAN METODE BARU: Agar api/mempool.go bisa memasukkan data lewat Interface x.BVMKeeper!
func (k *Keeper) UpdateActiveMiner(address string, timestamp int64) {
        k.ActiveMinersMu.Lock()
        k.ActiveMiners[address] = timestamp
        k.ActiveMinersMu.Unlock()
}

// 👑 SUNTIKKAN METODE BARU: Agar reward.go bisa menyedot data RAM dengan aman
func (k *Keeper) GetActiveMiners() map[string]int64 {
        k.ActiveMinersMu.RLock()
        defer k.ActiveMinersMu.RUnlock()
        
        // Kita kloning map-nya demi keamanan data bersama
        clone := make(map[string]int64)
        for addr, ts := range k.ActiveMiners {
                clone[addr] = ts
        }
        return clone
}

// 👑 SUNTIKKAN METODE BARU: Untuk membersihkan miner yang sudah expired dari RAM
func (k *Keeper) DeleteActiveMiner(address string) {
        k.ActiveMinersMu.Lock()
        delete(k.ActiveMiners, address)
        k.ActiveMinersMu.Unlock()
}
