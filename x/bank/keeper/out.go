package keeper

import (
	"github.com/aziskebanaran/BVM.core/pkg/logger"
	"github.com/aziskebanaran/BVM.core/x/bvm/types"
	banktypes "github.com/aziskebanaran/BVM.core/x/bank/types"
	"fmt"
)

func (bk *BankKeeper) Transfer(tx types.Transaction) error {
    if tx.Symbol == "BVM" || tx.Symbol == "" {
        return fmt.Errorf("transfer BVM harus melalui Protocol Keeper, bukan Bank")
    }

    totalDeduct := tx.Amount + tx.Fee

    // 🚩 SubBalance & AddBalance internal harus menggunakan bk.safePut 
    // agar masuk ke Batch jika sedang dalam mode eksekusi blok.
    if err := bk.SubBalance(tx.From, totalDeduct, tx.Symbol); err != nil {
        return err
    }

    bk.AddBalance(tx.To, tx.Amount, tx.Symbol)

    return nil
}

// FinalizeTransaction di Bank sekarang hanya fokus pada validasi akhir
func (bk *BankKeeper) FinalizeTransaction(tx types.Transaction) {
    // Bank tidak lagi menghitung tip/burn di sini.
    // Tugas ini dipindahkan ke Jenderal Pusat (BVM Keeper).
    logger.Info("BANK", fmt.Sprintf("🏁 Transaksi %s selesai diproses", tx.ID[:8]))
}


func (bk *BankKeeper) Mint(addr string, amount uint64, symbol string) {
    if symbol == "BVM" || symbol == "" { return }

    current := bk.GetBalance(addr, symbol)
    key := "t:" + symbol + ":" + addr

    // ✅ Sekarang menggunakan safePut
    bk.safePut(key, current+amount)

    logger.Info("BANK", fmt.Sprintf("✨ [ATOMIC] Minted %d %s ke %s", amount, symbol, addr[:8]))
}

func (bk *BankKeeper) Burn(addr string, amount uint64, symbol string) error {
    if symbol == "BVM" || symbol == "" { return nil }

    current := bk.GetBalance(addr, symbol)
    if current < amount {
        return fmt.Errorf("saldo %s tidak mencukupi untuk burn", symbol)
    }

    key := "t:" + symbol + ":" + addr
    // ✅ Sekarang menggunakan safePut
    bk.safePut(key, current-amount)
    return nil
}

// CreateToken sekarang menangani pembakaran BVM dan pencetakan Supply awal
func (bk *BankKeeper) CreateToken(owner string, symbol string, totalSupply uint64) error {
    // 1. Cek Duplikasi: Jangan biarkan ada simbol ganda
    _, exists := bk.GetTokenMetadata(symbol)
    if exists {
        return fmt.Errorf("🚨 Token %s sudah ada di jagat BVM", symbol)
    }

    // 2. Kalkulasi Burn Fee (Rumus Sultan 1:1.000.000)
    // Kita paksa minimal 1 BVM agar tidak ada token gratisan
    burnRequired := (totalSupply + 999999) / 1000000 

    // 3. Eksekusi Pembakaran BVM
    // Pastikan SubBalance bisa memotong saldo BVM milik owner
    err := bk.SubBalance(owner, burnRequired, "BVM")
    if err != nil {
        return fmt.Errorf("❌ Gagal membakar BVM: %v", err)
    }

    // 4. Siapkan Metadata
    md := banktypes.TokenMetadata{
        Symbol:      symbol,
        Owner:       owner,
        TotalSupply: totalSupply, // Langsung isi dengan supply perdana
        MintFee:     burnRequired, // Catat berapa BVM yang sudah dikorbankan
        IsFrozen:    false,
    }

    // 5. Simpan Metadata ke Store
    key := "md:" + symbol
    if err := bk.Store.Put(key, md); err != nil {
        return err
    }

    // 6. Cetak Saldo Token Baru ke Owner
    bk.AddBalance(owner, totalSupply, symbol)

    logger.Success("ECONOMY", fmt.Sprintf("🔥 BURN & MINT: %d BVM dibakar untuk mencetak %d %s", 
        burnRequired, totalSupply, symbol))

    return nil
}


func (bk *BankKeeper) HandleMsgCreateToken(msg banktypes.MsgCreateToken) error {
    // Panggil fungsi CreateToken yang sudah kita buat tadi
    // Rumus 1:1.000.000 otomatis dijalankan di dalamnya
    return bk.CreateToken(msg.Owner, msg.Symbol, msg.TotalSupply)
}


func (bk *BankKeeper) GetTokenMetadata(symbol string) (banktypes.TokenMetadata, bool) {
    var md banktypes.TokenMetadata
    key := "md:" + symbol
    err := bk.Store.Get(key, &md)
    if err != nil {
        return md, false
    }
    return md, true
}


func (bk *BankKeeper) safePut(key string, value interface{}) {
    if bk.Batch != nil {
        // 🚩 KOREKSI: Gunakan 'value', bukan 'md'
        bk.Store.PutToBatch(bk.Batch, key, value)
    } else {
        bk.Store.Put(key, value) // Langsung pahat (untuk testing/simulasi)
    }
}
