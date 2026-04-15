package main

import (
	"bvm.core/pkg/wallet"
	"bvm.core/pkg/client"
	"bvm.core/x/bvm/types"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

const CORE_URL = "http://127.0.0.1:8080"

func main() {
	if len(os.Args) < 2 {
		fmt.Println("💡 Gunakan perintah: create, check, atau send")
		return
	}

	bvm := client.NewBVMClient(CORE_URL)
	command := os.Args[1]

	// Helper untuk mencari dompet di berbagai lokasi
	getWalletPath := func() string {
		paths := []string{"../node_wallet.json", "node_wallet.json", "new_wallet.json"}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
		return "../node_wallet.json"
	}

	switch command {
	case "create":
		newWallet, err := wallet.CreateNewWallet()
		if err != nil {
			fmt.Println("❌ Gagal membuat wallet:", err)
			return
		}
		walletPath := "../node_wallet.json"
		if err := wallet.SaveWallet(newWallet, walletPath); err != nil {
			wallet.SaveWallet(newWallet, "node_wallet.json")
		}
		fmt.Printf("✨ Wallet baru tercipta!\n👤 Address: %s\n", newWallet.Address)

case "check":
    walletFile := getWalletPath()
    myWallet, err := wallet.LoadWallet(walletFile)
    if err != nil {
        fmt.Println("❌ Dompet tidak ditemukan.")
        return
    }

    // 🚩 PERBAIKAN: Gunakan GetSecureState agar data 100% sama dengan Node
    state, err := bvm.GetSecureState(myWallet.Address)
    if err != nil {
        fmt.Println("❌ Kernel Offline. Jalankan 'bvm node start' dulu.")
        return
    }

    fmt.Println("---------------------------------------")
    fmt.Printf("👤 ALAMAT : %s\n", state.Address)
    fmt.Printf("💰 SALDO  : %s %s\n", state.BalanceDisplay, state.Symbol) // Pakai string display!
    fmt.Printf("🔢 NONCE  : %d\n", state.Nonce)
    fmt.Printf("📂 FILE   : %s\n", walletFile)
    fmt.Println("---------------------------------------")
    getHistory(myWallet.Address)


        case "send":
                if len(os.Args) < 3 {
                        fmt.Println("💡 Gunakan: send [ALAMAT] [JUMLAH]")
                        return
                }

                var toAddr string
                var amountStr string

                if len(os.Args) >= 4 && (len(os.Args[2]) >= 3 && os.Args[2][:3] == "to=") {
                        for _, arg := range os.Args {
                                if len(arg) > 3 && arg[:3] == "to=" { toAddr = arg[3:] }
                                if len(arg) > 7 && arg[:7] == "amount=" { amountStr = arg[7:] }
                        }
                } else {
                        toAddr = os.Args[2]
                        if len(os.Args) > 3 { amountStr = os.Args[3] }
                }


                // 1. Parsing Float dari Input
                amountFloat, _ := strconv.ParseFloat(amountStr, 64)
                if toAddr == "" || amountFloat <= 0 {
                        fmt.Println("❌ Format salah! Contoh: send bvmf123 50")
                        return
                }

		   // 🚩 PERBAIKAN: Konversi ke Atomic menggunakan angka statis atau Params
	        amountAtomic := uint64(amountFloat * 1e8) // 100.000.000

	        walletFile := getWalletPath()
	        senderWallet, err := wallet.LoadWallet(walletFile)

		    // 🚩 SINKRONISASI NONCE: Jangan pakai nonce dari file lokal, tanya Node!
	        state, _ := bvm.GetSecureState(senderWallet.Address)
	        senderWallet.Nonce = state.Nonce 

	        fmt.Printf("🛰️  Menandatangani Transaksi (Nonce: %d)...\n", senderWallet.Nonce)
	        signedTx, err := senderWallet.SignAndPack(bvm, toAddr, amountAtomic, "BVM", "Transfer via BVM-Wallet")


                // 3. KIRIM KE JARINGAN
                fmt.Printf("🚀 Mengirim %.8f BVM ke %s...\n", amountFloat, toAddr)

                txID, err := bvm.BroadcastTX(signedTx)
                if err != nil {
                    fmt.Printf("❌ Jaringan Menolak: %v\n", err)
                    return
                }

                wallet.SaveWallet(senderWallet, walletFile)

                fmt.Println("--------------------------------------------------")
                fmt.Printf("✅ BERHASIL! Transaksi Sah Telah Terkirim.\n")
                fmt.Printf("🎫 TX ID     : %s\n", txID)
                fmt.Printf("🔢 New Nonce : %d\n", senderWallet.Nonce)
                fmt.Println("--------------------------------------------------")


	}


}


func getHistory(address string) {
    resp, err := http.Get(CORE_URL + "/api/history?address=" + address)
    if err != nil || resp.StatusCode != 200 {
        fmt.Println("⚠️ Gagal mengambil riwayat dari server.")
        return
    }
    defer resp.Body.Close()

    var history []struct {
        Height int               `json:"height"`
        Time   int64             `json:"time"`
        Tx     types.Transaction `json:"tx"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&history); err != nil {
        return
    }

    // PENGAMAN 1: Cek jika riwayat kosong
    if len(history) == 0 {
        fmt.Println("\n📭 Belum ada riwayat transaksi untuk alamat ini.")
        return
    }

    fmt.Println("\n📜 RIWAYAT TRANSAKSI TERAKHIR:")

    // PENGAMAN 2: Batasi tampilan maksimal 10 transaksi saja
    limit := 10
    if len(history) < limit {
        limit = len(history)
    }

for i := 0; i < limit; i++ {
    entry := history[i]

    // Konversi atomic ke float hanya untuk tampilan riwayat
    displayAmount := float64(entry.Tx.Amount) / 100000000.0

    toAddr := entry.Tx.To
        // PENGAMAN 3: Cek panjang alamat sebelum di-slice [:10] agar tidak PANIC
        displayAddr := toAddr
        if len(toAddr) > 10 {
            displayAddr = toAddr[:10] + "..."
        }

    fmt.Printf("[%d] %-13s -> %12.8f BVM (%s)\n",
        entry.Height,
        displayAddr,
        displayAmount, // 🚩 Sekarang sudah benar tampilannya
        time.Unix(entry.Time, 0).Format("15:04"))

    }
}
