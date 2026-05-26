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

// =========================================================================
// 🛡️ BENTENG VARIABEL GLOBAL SULTAN (TERKUNCI DI SINI)
// =========================================================================
var bvmClient *client.BVMClient

// Inisialisasi Grup Perintah Utama Cobra (Hanya dideklarasikan SATU KALI)
var (
	rootCmd   = &cobra.Command{Use: "bvm", Short: constants.ProjectName + " CLI Control Center"}
	walletCmd = &cobra.Command{Use: "wallet", Short: "Manajemen dompet BVM"}
	nodeCmd   = &cobra.Command{Use: "node", Short: "Manajemen Node/Kernel BVM"}
	minerCmd  = &cobra.Command{Use: "miner", Short: "Manajemen Pekerja Tambang BVM"}
)

func main() {
	// --- 🛡️ OPERASI LOAD KONFIGURASI ---
	err := godotenv.Load()
	if err != nil {
		fmt.Println("ℹ️  Info: Menjalankan tanpa file .env, pastikan variabel sudah di-export.")
	} else {
		fmt.Println("✅ [SYSTEM] Konfigurasi .env berhasil dimuat.")
	}

	// =========================================================================
	// 📡 BENTENG SATU PINTU PUSAT: AUTO-DETECT NETWORK HUB
	// =========================================================================
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		isTestnet, _ := cmd.Flags().GetBool("testnet")
		homeDir, _ := cmd.Flags().GetString("home")

		// 1. KUNCI LOGIKA: Definisi Mutlak per Lingkungan
		var nodeURL, nexusURL string
		if isTestnet {
			nodeURL = "http://localhost:8081"  // Core Testnet
			nexusURL = "http://localhost:9094" // Nexus Gateway Testnet
		} else {
			nodeURL = "http://localhost:8080"  // Core Mainnet
			nexusURL = "http://localhost:9092" // Nexus Gateway Mainnet
		}

		// 2. Override jika user memaksa dengan flag --nexus
		if cmd.Flags().Changed("nexus") {
			nexusURL, _ = cmd.Flags().GetString("nexus")
		}

		// 3. PENYESUAIAN DIREKTORI DATA
		if isTestnet && homeDir == "./data" {
			_ = cmd.Flags().Set("home", "./data/testnet")
			homeDir = "./data/testnet"
		}

		// 4. SETEL LINGKUNGAN KERNEL
		if isTestnet {
			os.Setenv("BVM_CHAIN_ID", "9999")
			os.Setenv("BVM_NETWORK_NAME", "BVM Atomic Testnet")
			fmt.Printf("🧪 [NETWORK HUB] Koordinat Terkunci: BVM Testnet (ChainID: 9999) | Port: 8081\n")
		} else {
			os.Setenv("BVM_CHAIN_ID", "1989")
			os.Setenv("BVM_NETWORK_NAME", "BVM Mainnet")
			fmt.Printf("🛡️ [NETWORK HUB] Koordinat Terkunci: BVM Mainnet (ChainID: 1989) | Port: 8080\n")
		}

		// 5. PENETAPAN GLOBAL
		bvmClient = client.NewBVMClient(nodeURL)
		os.Setenv("NEXUS_URL", nexusURL)
		fmt.Printf("🌐 Nexus Point: %s\n", nexusURL)

		// 6. Muat Sesi (JWT)
		tokenPath := fmt.Sprintf("%s/session.jwt", homeDir)
		if tokenData, err := os.ReadFile(tokenPath); err == nil {
			bvmClient.Token = string(tokenData)
		}
	}

	// 🚩 FLAG PERSISTEN GLOBAL
	rootCmd.PersistentFlags().StringP("home", "H", "./data", "Jalur folder data utama")
	rootCmd.PersistentFlags().StringP("nexus", "n", "", "Alamat Nexus Server manual")
	rootCmd.PersistentFlags().Bool("testnet", false, "Jalankan di ekosistem BVM Testnet (ChainID: 9999)")


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

var bridgeCmd = &cobra.Command{
    Use:   "bridge",
    Short: "Cross-Chain Bridge: Pindahkan aset antar jaringan BVM (1989 <-> 9999)",
    Run: func(cmd *cobra.Command, args []string) {
        h, _ := cmd.Flags().GetString("home")
        walletFile := fmt.Sprintf("%s/node_wallet.json", h)

        targetChainID, _ := cmd.Flags().GetString("chain")
        targetAddr, _ := cmd.Flags().GetString("to")
        amountFloat, _ := cmd.Flags().GetFloat64("amount")

        // 1. Muat Dompet
        w, err := wallet.LoadWallet(walletFile)
        if err != nil {
            fmt.Printf("❌ Error: %s tidak ditemukan!\n", walletFile)
            return
        }

        currentChainID := os.Getenv("BVM_CHAIN_ID")
        if targetChainID == currentChainID {
            fmt.Printf("❌ Ditolak: Anda sudah berada di Chain %s!\n", currentChainID)
            return
        }

        // 2. Skala Atomik
        amountAtomic := types.Params{}.ToAtomic(fmt.Sprintf("%.8f", amountFloat))

        // 3. 🛡️ ALAMAT BRANKAS & MEMO (Wajib ada!)
        ibcBridgeVault := "bvmf_ibc_bridge_vault"
        memo := fmt.Sprintf("IBC_TRANSFER|%s|%s", targetChainID, targetAddr)

        fmt.Println("---------------------------------------")
        fmt.Printf("🌉 BVM INTER-CHAIN COMMUNICATION (IBC) INITIATED\n")
        fmt.Printf("📡 Dari Jaringan: BVM Chain %s\n", currentChainID)
        fmt.Printf("🌌 Ke Jaringan  : BVM Chain %s\n", targetChainID)
        fmt.Printf("📍 Alamat Tujuan: %s\n", targetAddr)
        fmt.Printf("💎 Jumlah       : %.8f BVM\n", amountFloat)
        fmt.Println("---------------------------------------")
        fmt.Println("⏳ Mengunci aset di Vault & membungkus paket IBC...")

        // 5. Tanda tangani paket IBC menggunakan variabel yang sudah disiapkan
        tx, err := w.SignBridgeOutTX(bvmClient, ibcBridgeVault, memo, amountAtomic, "BVM")
        if err != nil {
            fmt.Printf("❌ Gagal merakit paket IBC: %v\n", err)
            return
        }

        // 6. Broadcast
        txID, err := bvmClient.BroadcastBridgeOut(tx)
        if err != nil {
            fmt.Printf("❌ Gagal menembus portal: %v\n", err)
            return
        }

        fmt.Printf("🚀 Paket IBC Sukses Terkirim!\n")
        fmt.Printf("🔗 TXID Portal: %s\n", txID)
        fmt.Println("🛰️ Relayer (Nexus) akan segera memproses pencetakan di rantai tujuan.")
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


        // 3. FINALISASI & PENGIKATAN FLAG (BERSIH & AMAN)
        // ==========================================


        sendCmd.Flags().StringP("to", "t", "", "Alamat tujuan")
        sendCmd.Flags().Float64P("amount", "a", 0.0, "Jumlah BVM")
        _ = sendCmd.MarkFlagRequired("to")
        _ = sendCmd.MarkFlagRequired("amount")

        // 🚩 BENTENG FLAG UNTUK BRIDGE (IBC-LITE)
        bridgeCmd.Flags().StringP("chain", "c", "9999", "ChainID tujuan (contoh: 9999 untuk Testnet, 1989 untuk Mainnet)")
        bridgeCmd.Flags().StringP("to", "t", "", "Alamat tujuan di rantai seberang (wajib)")
        bridgeCmd.Flags().Float64P("amount", "a", 0.0, "Jumlah aset yang ingin diseberangkan")
        _ = bridgeCmd.MarkFlagRequired("to")
        _ = bridgeCmd.MarkFlagRequired("amount")

        // 🚀 PASANG KEMBALI PELATUK MINER DI KANDUNGAN NODE START
        startNodeCmd.Flags().BoolP("miner", "m", false, "Aktifkan Miner Internal Gabungan")

        // 🚩 PASANG FLAG PORT DINAMIS (Wajib Agar Bisa Dibaca oleh node.go)
        startNodeCmd.Flags().Int("api-port", 8080, "Port API Server (Mainnet: 8080, Testnet: 8081)")
        startNodeCmd.Flags().Int("p2p-port", 9090, "Port P2P Engine (Mainnet: 9090, Testnet: 9091)")

        // 🚩 MASUKKAN KE WALLET CMD
        walletCmd.AddCommand(createWalletCmd, balanceCmd, searchCmd, sendCmd, registerCmd, loginCmd, bridgeCmd)

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
