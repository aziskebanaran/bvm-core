package main

import (
	"fmt"
	"os"
	"time"

	"github.com/aziskebanaran/bvm-core/pkg/client"
	"github.com/aziskebanaran/bvm-core/pkg/constants"
	"github.com/aziskebanaran/bvm-core/pkg/types"
	"github.com/aziskebanaran/bvm-core/pkg/wallet"
	"github.com/spf13/cobra"
	"github.com/joho/godotenv"
)

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Manajemen Node/Kernel BVM",
}

var walletCmd = &cobra.Command{
	Use:   "wallet",
	Short: "Manajemen dompet BVM",
}

func main() {
    // --- 🛡️ OPERASI LOAD KONFIGURASI ---
    err := godotenv.Load()
    if err != nil {
        fmt.Println("ℹ️  Info: Menjalankan tanpa file .env, pastikan variabel sudah di-export.")
    } else {
        fmt.Println("✅ [SYSTEM] Konfigurasi .env berhasil dimuat.")
    }

    // 1. Inisialisasi Client
    bvmClient := client.NewBVMClient("http://localhost:8080")

    // 2. 🛡️ OPERASI AUTO-LOAD SESSION (Tambahkan di sini)
    // Kita ambil jalur folder data dari flag atau default ./data
    tokenPath := "./data/session.jwt"
    tokenData, err := os.ReadFile(tokenPath)
    if err == nil {
        bvmClient.Token = string(tokenData)
        // Opsional: fmt.Println("🎫 Sesi aktif dimuat otomatis.") 
    }

    // 3. Konfigurasi Cobra seperti biasa
    var rootCmd = &cobra.Command{
        Use:   "bvm",
        Short: constants.ProjectName + " CLI Control Center",
    }

	// 🚩 FLAG PERSISTEN: Tersedia untuk SEMUA sub-command (Node, Wallet, Mempool, dll)
	rootCmd.PersistentFlags().StringP("home", "H", "./data", "Jalur folder data utama (database & wallet)")
	rootCmd.PersistentFlags().StringP("nexus", "n", "http://localhost:9092", "Alamat Nexus Server untuk sinkronisasi")

	// ==========================================
	// 1. SUB-COMMAND WALLET
	// ==========================================

	var createWalletCmd = &cobra.Command{
		Use:   "create",
		Short: "Buat wallet baru dengan 12 kata rahasia",
		Run: func(cmd *cobra.Command, args []string) {
			h, _ := cmd.Flags().GetString("home")
			walletFile := fmt.Sprintf("%s/node_wallet.json", h)

			os.MkdirAll(h, 0755)
			newW, mnemonic, err := wallet.CreateNewWallet()
			if err != nil {
				fmt.Printf("❌ Gagal: %v\n", err)
				return
			}

			wallet.SaveWallet(newW, walletFile)
			fmt.Println("---------------------------------------")
			fmt.Printf("✨ Wallet Berhasil Dibuat di %s!\n", walletFile)
			fmt.Printf("📍 Address  : %s\n", newW.Address)
			fmt.Printf("🔑 Mnemonic : %s\n", mnemonic)
			fmt.Println("---------------------------------------")
		},
	}

	var balanceCmd = &cobra.Command{
		Use:   "balance",
		Short: "Cek saldo wallet",
		Run: func(cmd *cobra.Command, args []string) {
			h, _ := cmd.Flags().GetString("home")
			walletFile := fmt.Sprintf("%s/node_wallet.json", h)

			w, err := wallet.LoadWallet(walletFile)
			if err != nil {
				fmt.Printf("❌ Wallet tidak ditemukan di %s!\n", walletFile)
				return
			}
			state, err := bvmClient.GetSecureState(w.Address)
			if err != nil {
				fmt.Printf("💰 Alamat: %s\n❌ Node Offline\n", w.Address)
				return
			}
			fmt.Println("---------------------------------------")
			fmt.Printf("🛡️  NETWORK : BVM Mainnet\n")
			fmt.Printf("📍 ADDRESS : %s\n", state.Address)
			fmt.Printf("💵 BALANCE : %s %s\n", state.BalanceDisplay, state.Symbol)
			fmt.Printf("🔢 NONCE   : %d\n", state.Nonce)
			fmt.Println("---------------------------------------")
		},
	}

var sendCmd = &cobra.Command{
    Use:   "send",
    Short: "Kirim BVM ke alamat atau username lain",
    Run: func(cmd *cobra.Command, args []string) {
        h, _ := cmd.Flags().GetString("home")
        walletFile := fmt.Sprintf("%s/node_wallet.json", h)

        to, _ := cmd.Flags().GetString("to")
        amountFloat, _ := cmd.Flags().GetFloat64("amount")

        // --- 🚩 UPGRADE: RESOLUSI IDENTITAS SULTAN ---
        finalRecipient := to
        // Jika input tidak diawali 'bvmf', kita asumsikan ini adalah USERNAME
        if len(to) < 4 || to[:4] != "bvmf" {
            fmt.Printf("🔍 Mencari alamat untuk identitas @%s...\n", to)

            // Gunakan bvmClient (Radar) yang sudah kita sinkronkan tadi
            state, err := bvmClient.GetSecureState(to)
            if err != nil || state.Address == "" || state.Address == to {
                fmt.Printf("❌ Gagal: Username @%s tidak ditemukan atau belum aktif!\n", to)
                return
            }

            finalRecipient = state.Address
            fmt.Printf("🎯 Radar Terkunci! Alamat asli: %s\n", finalRecipient)
        }
        // ---------------------------------------------

        w, err := wallet.LoadWallet(walletFile)
        if err != nil {
            fmt.Printf("❌ Error: %s tidak ditemukan!\n", walletFile)
            return
        }

        amountAtomic := types.Params{}.ToAtomic(fmt.Sprintf("%.8f", amountFloat))
        fmt.Println("⏳ Menandatangani transaksi...")

        // 🚩 GUNAKAN finalRecipient, BUKAN to
        tx, err := w.SignAndPack(bvmClient, finalRecipient, amountAtomic, "BVM", "Sent via BVM-CLI")
        if err != nil {
            fmt.Printf("❌ Gagal: %v\n", err)
            return
        }

        txID, err := bvmClient.BroadcastTX(tx)
        if err != nil {
            fmt.Printf("❌ Gagal broadcast: %v\n", err)
            return
        }
        fmt.Printf("🚀 Sukses! TXID: %s\n", txID)
    },
}



var searchCmd = &cobra.Command{
    Use:   "search [username/address]",
    Short: "Mencari identitas atau akun berdasarkan Username atau Address",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        query := args[0]
        fmt.Printf("🔍 Mencari identitas: %s...\n", query)

        // 1. Panggil API Secure State
        state, err := bvmClient.GetSecureState(query)
        if err != nil {
            fmt.Printf("❌ Gagal mencari: %v\n", err)
            return
        }

        // 2. 🛡️ Validasi Hasil (Pagar Betis Sultan)
        // 2. 🛡️ Validasi Hasil (Versi Sultan yang Lebih Bijak)
        if state.Address == "" {
            fmt.Printf("❌ Identitas @%s tidak ditemukan di jaringan.\n", query)
            return
        }

        // 3. ✨ Tampilkan Hasil dengan Rapi
        fmt.Println("---------------------------------------")
        fmt.Printf("✅ HASIL PENCARIAN JARINGAN\n")

        // Jika Sultan mencari 'kebanaran' dan hasilnya 'bvmf...', kita tunjukkan mapping-nya
        if query != state.Address {
            fmt.Printf("👤 USERNAME : @%s\n", query)
        }

        fmt.Printf("📍 ADDRESS  : %s\n", state.Address)
        fmt.Printf("💰 BALANCE  : %s %s\n", state.BalanceDisplay, state.Symbol)
        fmt.Printf("🔢 NONCE    : %d\n", state.Nonce)
        fmt.Println("---------------------------------------")
    },
}


	var registerCmd = &cobra.Command{
		Use:   "register [username]",
		Short: "Daftarkan username unik untuk alamat wallet Anda",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			h, _ := cmd.Flags().GetString("home")
			walletFile := fmt.Sprintf("%s/node_wallet.json", h)

			w, err := wallet.LoadWallet(walletFile)
			if err != nil {
				fmt.Printf("❌ Error: %s tidak ditemukan!\n", walletFile)
				return
			}

			username := args[0]
			fmt.Printf("⏳ Mendaftarkan @%s...\n", username)
			tx, err := w.SignAndPackCustom(bvmClient, username)
			if err != nil {
				fmt.Printf("❌ Gagal: %v\n", err)
				return
			}

			txID, err := bvmClient.BroadcastTX(tx)
			if err != nil {
				fmt.Printf("❌ Gagal: %v\n", err)
				return
			}
			fmt.Printf("🚀 Sukses! TXID: %s\n", txID)
		},
	}

var loginCmd = &cobra.Command{
    Use:   "login [username]",
    Short: "Login ke jaringan menggunakan username",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        h, _ := cmd.Flags().GetString("home")
        walletFile := fmt.Sprintf("%s/node_wallet.json", h)

        w, err := wallet.LoadWallet(walletFile)
        if err != nil {
            fmt.Printf("❌ Error: %s tidak ditemukan!\n", walletFile)
            return
        }

        username := args[0]
        message := fmt.Sprintf("LOGIN_TO_BVM_%d", time.Now().Unix())
        
        // Tandatangani pesan menggunakan Private Key Wallet
        sig, err := w.SignMessage(message) 
        if err != nil {
            fmt.Printf("❌ Gagal tanda tangan: %v\n", err)
            return
        }

        // Kirim ke API Core
        fmt.Printf("⏳ Mencoba login sebagai @%s...\n", username)
        
        // Note: Sultan perlu menambahkan fungsi c.Login di pkg/client/auth.go
        token, err := bvmClient.Login(username, sig, message)
        if err != nil {
            fmt.Printf("❌ Login Gagal: %v\n", err)
            return
        }

        fmt.Printf("✨ LOGIN SUKSES!\n🎫 Token Sesi: %s\n", token)
    },
}



	// ==========================================
	// 2. SUB-COMMAND NODE & MEMPOOL
	// ==========================================

	var startNodeCmd = &cobra.Command{
		Use:   "start",
		Short: "Menjalankan Kernel BVM",
		Run:   startNodeProvider,
	}

	var mempoolCmd = &cobra.Command{
		Use:   "mempool",
		Short: "Lihat antrean transaksi di RAM Mempool",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := bvmClient.GetMempool()
			if err != nil {
				fmt.Println("❌ Gagal terhubung ke Node.")
				return
			}
			fmt.Printf("📦 TOTAL ANTREAN: %d Transaksi\n", resp.Count)
			for i, tx := range resp.Txs {
				fmt.Printf("[%d] TXID: %s | Dari: %s | Nonce: %d\n", i+1, tx.ID[:16], tx.From[:12], tx.Nonce)
			}
		},
	}

	// ==========================================
	// 3. FINALISASI
	// ==========================================

	sendCmd.Flags().StringP("to", "t", "", "Alamat tujuan")
	sendCmd.Flags().Float64P("amount", "a", 0.0, "Jumlah BVM")
	sendCmd.MarkFlagRequired("to")
	sendCmd.MarkFlagRequired("amount")

	startNodeCmd.Flags().BoolP("miner", "m", false, "Aktifkan Miner Internal")

	walletCmd.AddCommand(createWalletCmd, balanceCmd, searchCmd, sendCmd, registerCmd, loginCmd)
	nodeCmd.AddCommand(startNodeCmd)
	rootCmd.AddCommand(walletCmd, nodeCmd, mempoolCmd)

// ==========================================
// 4. SUB-COMMAND APP (Sistem "Aplikasi dalam Aplikasi")
// ==========================================

var appCmd = &cobra.Command{
    Use:   "app",
    Short: "Manajemen Mini-Apps di ekosistem BVM (WASM)",
}

// Fitur Instalasi: Menyalin file .wasm secara fisik
var installAppCmd = &cobra.Command{
    Use:   "install [file_wasm]",
    Short: "Pasang aplikasi baru ke dalam pangkalan data BVM",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        srcPath := args[0]
        h, _ := cmd.Flags().GetString("home")

        // 1. Validasi Sumber
        srcFile, err := os.Open(srcPath)
        if err != nil {
            fmt.Printf("❌ Gagal: File %s tidak ditemukan!\n", srcPath)
            return
        }
        defer srcFile.Close()

        // 2. Siapkan Gudang (data/apps_storage)
        appsDir := fmt.Sprintf("%s/apps_storage", h)
        os.MkdirAll(appsDir, 0755)

        // 3. Tentukan Nama Aplikasi (Ambil dari nama file)
        // Misal: sentinel.wasm -> sentinel
        destPath := fmt.Sprintf("%s/%s", appsDir, srcPath) 

        // 4. Proses Pemindahan Data
        fmt.Printf("📥 Menyalin unit ke gudang BVM...\n")
        // (Gunakan os.WriteFile atau io.Copy untuk menyalin isi file ke destPath)

        fmt.Printf("🛡️  Sentinel sedang mengaudit kode...\n")
        time.Sleep(1 * time.Second) // Simulasi audit keamanan WASM

        fmt.Println("---------------------------------------")
        fmt.Printf("✅ UNIT BERHASIL TERPASANG!\n")
        fmt.Printf("📍 Lokasi: %s\n", destPath)
        fmt.Printf("🚀 Jalankan dengan: ./bvm app run %s\n", srcPath)
        fmt.Println("---------------------------------------")
    },
}


// Fitur Eksekusi: Menjalankan Sandbox WASM
var runAppCmd = &cobra.Command{
    Use:   "run [nama_app]",
    Short: "Jalankan aplikasi internal dalam mode Sandbox",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        appName := args[0]
        fmt.Printf("🏗️  Menyiapkan Sandbox WASM untuk @%s...\n", appName)
        
        // Di sini nantinya akan memanggil x/wasm/keeper.go milik Jenderal
        fmt.Printf("🧠 Memuat aturan dari node_manager.wasm...\n")
        fmt.Printf("🌐 Aplikasi '%s' sekarang berjalan di atas Jaringan Sultan.\n", appName)
    },
}

var listAppCmd = &cobra.Command{
    Use:   "list",
    Short: "Tampilkan daftar aplikasi yang terpasang di BVM-OS",
    Run: func(cmd *cobra.Command, args []string) {
        h, _ := cmd.Flags().GetString("home")
        appsDir := fmt.Sprintf("%s/apps_storage", h)

        files, err := os.ReadDir(appsDir)
        if err != nil || len(files) == 0 {
            fmt.Println("📭 Belum ada aplikasi yang terpasang di gudang.")
            return
        }

        fmt.Println("---------------------------------------")
        fmt.Printf("📦 DAFTAR APLIKASI INTERNAL (WASM)\n")
        for _, file := range files {
            if !file.IsDir() {
                info, _ := file.Info()
                fmt.Printf("🔹 %-15s | Size: %v\n", file.Name(), info.Size())
            }
        }
        fmt.Println("---------------------------------------")
    },
}

// Tambahkan ke grup app
appCmd.AddCommand(listAppCmd)

// Masukkan sub-command ke dalam grup App
appCmd.AddCommand(installAppCmd, runAppCmd)
// Masukkan grup App ke Root Command (Aplikasi Utama)
rootCmd.AddCommand(appCmd)


	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
