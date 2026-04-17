package types

import "github.com/aziskebanaran/BVM.core/x/bvm/types"

type BankKeeper interface {
    // --- 1. Keeper (Core State Management) ---
    GetAccount(addr string) (types.Account, error)
    SaveAccount(acc *types.Account) error
    AddBalance(addr string, amount uint64, symbol string) error

    SubBalance(addr string, amount uint64, symbol string) error
    GetBalance(addr string, symbol string) uint64
    GetTokenMetadata(symbol string) (TokenMetadata, bool)

    // --- 2. In (Validation & Entry) ---
    HoldForMempool(tx types.Transaction) error
    ReleaseFromMempool(tx types.Transaction)
    ValidateSend(tx types.Transaction) error
    GetOrCreateAccount(addr string) *types.Account
    IsFrozen(symbol string) bool // 🚩 Baru: Untuk proteksi di ValidateSend

    // --- 3. Out (Execution & Distribution) ---
    Transfer(tx types.Transaction) error
    FinalizeTransaction(tx types.Transaction)
    Mint(addr string, amount uint64, symbol string)
    Burn(addr string, amount uint64, symbol string) error
    CreateToken(owner string, symbol string, totalSupply uint64) error
    HandleMsgCreateToken(msg MsgCreateToken) error

    // --- 4. Audit & Admin (Authority) ---
    CheckSupplyIntegrity(symbol string)
    CalculateBalanceManual(bc *types.Blockchain, addr string, symbol string) uint64
    FreezeToken(caller string, symbol string) error   // 🚩 Baru: Otoritas Creator
    UnfreezeToken(caller string, symbol string) error // 🚩 Baru: Otoritas Creator
    ScanTotalSupplyFromDB(symbol string) uint64

    // --- 5. Querier (Information Retrieval) ---
    SearchAccount(address string) (map[string]interface{}, bool)
    GetAllBalances() map[string]map[string]uint64
    LoadAllAccountsFromDB() int
    GetPublicKeyFromLedger(bc *types.Blockchain, address string) string
}
