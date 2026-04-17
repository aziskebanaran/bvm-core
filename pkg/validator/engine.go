package validator

import (
	"github.com/aziskebanaran/BVM.core/pkg/logger"
    "github.com/aziskebanaran/BVM.core/x/bvm/keeper" // 👈 Import keeper langsung
    "github.com/aziskebanaran/BVM.core/x/bvm/types"
)

// Engine sekarang menggunakan pointer langsung ke Keeper (Si Jenderal)
type Engine struct {
    Keeper *keeper.Keeper
}

func NewEngine(k *keeper.Keeper) *Engine {
    return &Engine{Keeper: k}
}

// ValidateBlock: Memeriksa integritas Blok DAN Transaksi
func (e *Engine) ValidateBlock(newBlock types.Block, bc *types.Blockchain) bool {
    // 1. Ambil Blok Terakhir dari Chain
    if len(bc.Chain) == 0 { return false }
    prevBlock := bc.Chain[len(bc.Chain)-1]

    // 2. Validasi Struktur Dasar (Index & PrevHash)
    if newBlock.Index != prevBlock.Index+1 || newBlock.PrevHash != prevBlock.Hash {
        return false
    }

    // 3. Validasi Proof of Work (HasValidTarget)
    // Gunakan fungsi yang sudah ada di types.Block agar konsisten
    if !newBlock.HasValidTarget(newBlock.Hash) {
        return false
    }

    // 🚩 UPDATE VALIDASI OTORITAS MINER
    validators, _ := e.Keeper.GetValidatorObjects() 
    isAuthorized := false
    
    for _, v := range validators {
        // Cek apakah alamat cocok DAN statusnya harus "Active"
        if v.Address == newBlock.Miner && v.Status == "Active" {
            isAuthorized = true
            break
        }
    }

    if !isAuthorized {
        logger.Error("VALIDATOR", "❌ Blok ditolak: Miner tidak terdaftar atau sedang di-Jail!")
        return false
    }

    // 4. VALIDASI TRANSAKSI
    for _, tx := range newBlock.Transactions {
        // A. Verifikasi Tanda Tangan (Sudah Bagus)
        if !e.Keeper.Auth.VerifyTransaction(tx) {
            return false
        }

        // 🚩 B. PERBAIKAN: Tanya saldo ke Bank, bukan baca langsung dari struct Account
        // Kita harus mengecek saldo koin yang sesuai dengan 'tx.Symbol'
        currentBalance := e.Keeper.Bank.GetBalance(tx.From, tx.Symbol)
        
        // Pastikan saldo cukup untuk (Jumlah Kirim + Biaya)
        if currentBalance < (tx.Amount + tx.Fee) {
            return false
        }

        // C. Cek Nonce (Ambil data akun terbaru dari Auth)
        acc, _ := e.Keeper.Auth.GetAccount(tx.From)
        if tx.Nonce < acc.Nonce {
            return false
        }
    }

    return true
}
