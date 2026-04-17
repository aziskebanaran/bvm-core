package keeper

import (
        "fmt"
        "github.com/aziskebanaran/bvm-core/x/bvm/types"
)

func (k *BankKeeper) ValidateSend(tx types.Transaction) error {
    // 1. Cek apakah token sedang di-freeze (Sultan punya field IsFrozen di interface)
    if k.IsFrozen(tx.Symbol) {
        return fmt.Errorf("token %s sedang dibekukan oleh otoritas", tx.Symbol)
    }

    // 2. Ambil akun pengirim
    acc, err := k.GetAccount(tx.From)
    if err != nil {
        return fmt.Errorf("akun pengirim tidak ditemukan di sistem bank")
    }

    // 3. Cek Saldo Token
    currentBal := acc.Balances[tx.Symbol]
    if currentBal < tx.Amount {
        return fmt.Errorf("saldo %s tidak mencukupi (punya: %d, butuh: %d)", 
            tx.Symbol, currentBal, tx.Amount)
    }

    return nil
}

// HoldForMempool: SEKARANG MEMOTONG SALDO (Untuk Token Non-BVM)
func (bk *BankKeeper) HoldForMempool(tx types.Transaction) error {
    // Ingat Sultan: BVM Utama sudah diurus oleh Keeper, 
    // Bank hanya mengurus token L2 (GOLD, USD, dll)
    if bk.isNative(tx.Symbol) {
        return nil 
    }

    currentBal := bk.GetBalance(tx.From, tx.Symbol)
    totalNeeded := tx.Amount + tx.Fee

    if currentBal < totalNeeded {
        return fmt.Errorf("insufficient %s balance", tx.Symbol)
    }

    // POTONG SALDO SEKARANG (Locking mechanism)
    bk.SetBalance(tx.From, currentBal - totalNeeded, tx.Symbol)
    return nil
}

// ReleaseFromMempool: MENGEMBALIKAN SALDO jika TX gagal masuk antrean
func (bk *BankKeeper) ReleaseFromMempool(tx types.Transaction) {
    if bk.isNative(tx.Symbol) {
        return
    }

    currentBal := bk.GetBalance(tx.From, tx.Symbol)
    bk.SetBalance(tx.From, currentBal + tx.Amount + tx.Fee, tx.Symbol)
}
