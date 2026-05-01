package types

import (
    "github.com/aziskebanaran/bvm-core/pkg/storage" // 🚩 Tambahkan ini!
    "github.com/aziskebanaran/bvm-core/x/bvm/types"
)

// WasmKeeper: Antarmuka utama untuk mesin Smart Contract
type WasmKeeper interface {
    DeployContract(owner string, bytecode []byte) (string, error)
    ExecuteContract(contractAddr string, caller string, payload []byte) error
    ExecuteContractWithBatch(batch storage.Batch, contractAddr string, caller string, payload []byte) error
    RegisterBVMFunctions()
    VerifyZKP(proof string, publicInputs string) bool
    GetContractBalance(addr string, contractID string) uint64
    QueryContract(contractAddr string, method string, args ...interface{}) (interface{}, error)
    ValidateBlock(block types.Block) error
    HandleAINFTMint(batch storage.Batch, caller string, taskID string, metaData []byte) error
}

// --- JEMBATAN KOMUNIKASI (DITAMBAHKAN) ---

// BankKeeper: Apa yang bisa diminta WASM kepada Modul Bank
type BankKeeper interface {
    // WASM perlu fungsi ini untuk mengecek saldo koin
    GetBalance(addr string, symbol string) uint64
}

// Agar WASM bisa mengambil akses ke modul-modul lain
type BVMSystem interface {
    GetBank() BankKeeper
    // Kedepannya bisa ditambah: GetAuth(), GetStaking(), dll.
}
