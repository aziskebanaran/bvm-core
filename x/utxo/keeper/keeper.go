package keeper

import (
	 "encoding/json"
	"fmt"
	"strings"
	"github.com/aziskebanaran/bvm-core/pkg/storage"
	"github.com/aziskebanaran/bvm-core/x" // 🚩 Impor interface lintas modul
	"github.com/aziskebanaran/bvm-core/x/utxo/types"
)

type Keeper struct {
	store storage.BVMStore
	auth  x.AuthKeeper // 🏛️ Menteri Identitas
}

// NewKeeper: Inisialisasi unit UTXO
func NewKeeper(s storage.BVMStore, ak x.AuthKeeper) *Keeper {
	return &Keeper{
		store: s,
		auth:  ak,
	}
}
func (k *Keeper) SendAset(from, to, symbol string, amount, fee uint64, txID string, payload []byte, batch storage.Batch) error {
    // 1. BONGKAR INSTRUKSI PAYLOAD (Explicit Input)
    // Core tidak lagi mencari kepingan secara acak, tapi mengikuti daftar dari Gateway
    var data types.UTXOData
    if err := json.Unmarshal(payload, &data); err != nil {
        return fmt.Errorf("❌ Payload UTXO korup: %v", err)
    }

    needed := amount + fee
    var totalSelected uint64

    // =========================================================================
    // ⚔️ 2. VALIDASI & PEMBAKARAN (THE PURGE ENGINE EDITION)
    // =========================================================================
    if len(data.Inputs) == 0 {
        return fmt.Errorf("🚨 Payload Ilegal: Tidak ada kepingan input yang disertakan")
    }

    for _, inputID := range data.Inputs {
        var prevTxID string

        // 🚩 ARSITEKTUR ADAPTIF: Deteksi apakah kepingan ini adalah kepingan kembalian atau kepingan murni
        if strings.Contains(inputID, "-change-") {
            // Kasus Kepingan Kembalian: Format masuk biasanya "HASH-change-1-0" atau "HASH-change-1"
            // Kita wajib mengambil "HASH-change-1" secara utuh sebagai kunci DB Jenderal
            parts := strings.Split(inputID, "-change-")
            if len(parts) >= 2 {
                // parts[0] = HASH asli
                // parts[1] = sisa string di belakangnya (misal: "1" atau "1-0" atau "1-out-0")
                subParts := strings.Split(parts[1], "-")
                // subParts[0] dijamin adalah nomor urut kembalian ("1")
                prevTxID = fmt.Sprintf("%s-change-%s", parts[0], subParts[0])
            } else {
                prevTxID = inputID
            }
        } else {
            // Kasus Kepingan Murni: Format biasanya "HASH-0" atau "HASH"
            parts := strings.Split(inputID, "-")
            prevTxID = parts[0]
        }

        // 🛡️ PROTEKSI LAPIS BAJA: Cek eksistensi, kepemilikan, dan status ke database
        keping, found := k.GetUTXO(from, prevTxID)
        if !found {
            return fmt.Errorf("🚨 Kepingan %s tidak ditemukan di dompet %s", prevTxID, from[:10])
        }
        if keping.Status == "SPENT" {
            return fmt.Errorf("🚨 Kepingan %s sudah hangus (Double Spend Terdeteksi!)", prevTxID)
        }
        if keping.Symbol != symbol {
            return fmt.Errorf("🚨 Sabotase Simbol: Kepingan %s adalah %s, bukan %s", prevTxID, keping.Symbol, symbol)
        }

        totalSelected += keping.Amount

        // Eksekusi Pembakaran di dalam Batch menggunakan ID steril hasil penyaringan
        if err := k.BurnUTXO(from, prevTxID, batch); err != nil {
            return fmt.Errorf("❌ Gagal membakar kepingan %s: %v", prevTxID, err)
        }
    }


    // 3. CEK KESETIMBANGAN EKONOMI
    if totalSelected < needed {
        return fmt.Errorf("🚨 Dana tidak cukup! Kepingan: %d, Butuh (Amt+Fee): %d", totalSelected, needed)
    }

    // 4. JEMBATAN TARGET (Minting vs Account Liquidation)
    if strings.HasPrefix(to, "0x") {
        // 🚩 TARGET UTXO: Cetak kepingan baru dengan suffix unik agar tidak tabrakan ID
        // Gunakan suffix "-out-0" agar ID tetap unik dalam satu transaksi
        err := k.MintUTXO(to, to, "AUTO_USER", symbol, amount, txID+"-out-0", batch)
        if err != nil {
            return fmt.Errorf("❌ Gagal mencetak target UTXO: %v", err)
        }
        fmt.Printf("📥 [UTXO] Aset dipahat ke Brankas 0x: %s\n", to[:10])
    } else {
        // 🚩 TARGET ACCOUNT (bvmf): Biarkan ExecuteBlock yang menambah pendingChanges
        // Kita hanya perlu memastikan kepingan sudah terbakar di sini.
        fmt.Printf("🔓 [UTXO] Aset dicairkan ke Ruang Account (bvmf) untuk: %s\n", to)
    }

    // 5. MANAJEMEN KEMBALIAN (Change Management)
    change := totalSelected - needed
    if change > 0 {
        // Ambil metadata dari kepingan pertama untuk warisan Identitas (Address0x/Username)
        // Ini agar Sultan tidak kehilangan identitas profil saat kepingan pecah
        firstInputID := strings.Split(data.Inputs[0], "-")[0]
        heritage, _ := k.GetUTXO(from, firstInputID)
        
        // Cetak kepingan sisa kembali ke pengirim dengan suffix "-change-1"
        err := k.MintUTXO(
            from, 
            heritage.Address0x, 
            heritage.Username, 
            symbol, 
            change, 
            txID+"-change-1", 
            batch,
        )
        if err != nil {
            return fmt.Errorf("❌ Gagal mencetak kembalian: %v", err)
        }
        fmt.Printf("🔄 [UTXO] Kembalian %d %s dipulangkan ke %s\n", change, symbol, from[:10])
    }

    return nil
}


