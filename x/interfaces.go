package x

import (
    "bvm.core/x/bvm/types"
    "bvm.core/pkg/storage"
    banktypes    "bvm.core/x/bank/types" 
    stakingtypes "bvm.core/x/staking/types" 
    mempooltypes "bvm.core/x/mempool/types"
    wasmtypes    "bvm.core/x/wasm/types"
    authtypes    "bvm.core/x/auth/types" 
    storagetypes "bvm.core/x/storage/types"
)

// ParamsKeeper: Kontrak Konstitusi Ekonomi BVM
type ParamsKeeper interface {
    GetParamsData() types.Params
    GetDynamicFee(mempoolSize int) uint64
    FromAtomic(uint64) string
    ToAtomic(float64) uint64

}

type BankKeeper    = banktypes.BankKeeper

type AuthKeeper    = authtypes.AuthKeeper

type StakingKeeper = stakingtypes.StakingKeeper 

type MempoolKeeper = mempooltypes.MempoolKeeper

type WasmKeeper    = wasmtypes.WasmKeeper 

type NonceKeeper interface {
    GetNextNonce(address string) uint64
    Increment(address string) error
    HealthCheckNonce(address string) (bool, uint64, uint64)
    ManualOverride(address string, newNonce uint64) error
    SetNonce(address string, newNonce uint64) error
}

type BVMStore interface {
    SaveBlock(block types.Block) error
    GetBlockByHeight(height int) (types.Block, error)
    LoadFullChain() ([]types.Block, error)
    GetAddressHistory(addr string) ([]types.Transaction, error)
    Put(key string, value interface{}) error
    Get(key string, target interface{}) error
    LoadLedger(bc *types.Blockchain) error
    Close() error
}

// P2PKeeper: Menteri Hubungan Internasional (Antar Node)
type P2PKeeper interface {
    GetPeers() []types.Peer
    GetActivePeers(timeoutSeconds int64) []types.Peer // Filter otomatis node aktif
    AddPeer(ip string, nodeID string) error
    CountActive() int
    RemovePeer(ip string) error // Untuk memutus koneksi node nakal
    BroadcastTransaction(tx types.Transaction) // Siarkan transaksi baru
    BroadcastBlock(block types.Block)          // Siarkan blok yang baru dipahat
}


// 🚩 Definisikan Interface untuk Storage
type StorageModuleKeeper interface {
    GetAppStore(appID string) (storage.BVMStore, error)
    GetAppMetadata(appID string) (storagetypes.AppContainer, error)
    RegisterApp(owner string, appID string, rules map[string]interface{}) (string, error)
    SafePut(appID string, data storagetypes.UserData, callerAddr string) error
    CheckRules(app storagetypes.AppContainer, path string, action string, callerAddr string) bool
}


// BVMKeeper: Jenderal Lapangan (Kernel)
type BVMKeeper interface {
    // --- 1. CORE OPERATIONS (Pintu Masuk & Keluar) ---
    GetLatestBlock() types.Block
    GetLastBlockHash() string
    GetLastHeight() int
    ProcessTransaction(tx types.Transaction) error
    GetSecureBalance(address string) (types.WalletState, bool)
    SearchAccount(query string) (interface{}, bool)
    GetNextNonce(address string) uint64

    GetStore() storage.BVMStore
    // Tambahkan ini agar API bisa ambil data upgrades
    GetAllScheduledUpgrades() interface{}

    GetValidatorObjects() ([]stakingtypes.Validator, error)
    IsFeatureActive(feature string, height int64) bool

	//Consensus
    GetParams() ParamsKeeper
    GetParamsData() types.Params
    GetDynamicFee(mempoolSize int) uint64
    GetNextDifficultyForMiner(minerAddr string) int
    CalculateAvgBlockTime() int64
    GetNextDifficulty() int

	//Mining
    PrepareNewWork(minerAddr string, minerName string) (types.Block, error)
    SolveBlockLogic(block *types.Block, quit chan bool) bool
    SubmitMinedBlock(block types.Block) error
    GetDifficulty() int
    SetDifficulty(newDiff int)

	//Validation
    ProcessBlock(newBlock types.Block) error
    ValidateConsensus(newBlock types.Block) error
    ValidateBlockTransactions(block types.Block) error
    VerifyBlock(newBlock types.Block) bool

        //Execution
    CreateNextBlock(minerAddr string) types.Block
    ExecuteBlock(block types.Block) error
    CommitBlock(block types.Block) error

	//Reward
    DistributeBlockReward(height int64, fees uint64) (uint64, uint64, error)
    GetSubsidiAtHeight(height int64) uint64

	//Minting
    GetCurrentReward(height int64) uint64
    GetInflationStats() (string, int64)


    // --- 4. AI & SECURITY (Sentinel) ---
    ValidateWithAI() (int, string)
    AutoRecoverDatabase()
    SyncState() uint64

    FromAtomic(amount uint64) string
    ToAtomic(amount float64) uint64
    // --- 5. RANTAI KOMANDO (Akses Menteri) ---
    SetBalanceBVM(addr string, amount uint64, batch storage.Batch) error
    AddBalanceBVM(addr string, amount uint64, batch storage.Batch)
    SubBalanceBVM(addr string, amount uint64, batch storage.Batch) error
    GetBalanceBVM(addr string) uint64

    GetBank() BankKeeper
    GetAuth() AuthKeeper
    GetNonce() NonceKeeper
    GetStatus() types.NodeStatus
    GetChain() []types.Block
    GetMempool() MempoolKeeper
    GetStaking() StakingKeeper
    GetP2P() P2PKeeper
    GetWasm() WasmKeeper
    GetPendingTransactions() []types.Transaction

    GetCloudStorage() StorageModuleKeeper
}

// MiningKeeper: Kini murni sebagai 'Kontrak Perhitungan'
type MiningKeeper interface {
    GetNextDifficulty() int
    GetNextDifficultyForMiner(minerAddr string) int
    ValidateConsensus(block types.Block) error
}

type MinerEngine interface {
    Start(minerAddr string)
    // Sultan bisa tambah fungsi Stop jika ingin mematikan tambang via API
    Stop() 
}
