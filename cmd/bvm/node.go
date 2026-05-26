package main

import (
	"encoding/json" // 🚩 Tambahkan ini
	"os"            // 🚩 Tambahkan ini
	"fmt"

	"github.com/aziskebanaran/bvm-core/pkg/logger"
	"github.com/aziskebanaran/bvm-core/pkg/node"
	"github.com/aziskebanaran/bvm-core/pkg/storage"
	"github.com/aziskebanaran/bvm-core/x/app"
	"github.com/aziskebanaran/bvm-core/x/bvm/types"
	"github.com/aziskebanaran/bvm-core/x/miner" // 🚩 Impor package miner Sultan
	"github.com/spf13/cobra"
	"github.com/aziskebanaran/bvm-core/pkg/client"
	"time"
)

// getMinerAddress: Mengambil identitas miner secara dinamis dari file wallet
func getMinerAddress(homeDir string) (string, error) {
	// Konsolidasi: Selalu cari di dalam folder data (home)
	walletPath := fmt.Sprintf("%s/node_wallet.json", homeDir)

	if _, err := os.Stat(walletPath); os.IsNotExist(err) {
		return "", fmt.Errorf("file %s tidak ditemukan", walletPath)
	}

	data, err := os.ReadFile(walletPath)
	if err != nil {
		return "", fmt.Errorf("gagal membaca file wallet: %v", err)
	}

	var w struct {
		Address string `json:"address"`
	}
	if err := json.Unmarshal(data, &w); err != nil {
		return "", fmt.Errorf("format JSON wallet rusak: %v", err)
	}

	if w.Address == "" {
		return "", fmt.Errorf("alamat di dalam wallet kosong")
	}

	return w.Address, nil
}

func startNodeProvider(cmd *cobra.Command, args []string) {
    h, _ := cmd.Flags().GetString("home")
    useMiner, _ := cmd.Flags().GetBool("miner")
    isTestnet, _ := cmd.Flags().GetBool("testnet")
    
    // 🚩 1. PENETAPAN NEXUS URL DINAMIS (Satu-satunya sumber kebenaran)
    nexusURL := "http://localhost:9092" // Default Mainnet
    if isTestnet {
        nexusURL = "http://localhost:9094" // Port Testnet
    }
    // Override manual jika flag -n diset
    if cmd.Flags().Changed("nexus") {
        nexusURL, _ = cmd.Flags().GetString("nexus")
    }

    // 🚩 2. KONSOLIDASI LINGKUNGAN (Environment Injection)
    if isTestnet {
        os.Setenv("BVM_CHAIN_ID", "9999")
        os.Setenv("BVM_NETWORK_NAME", "BVM Atomic Testnet")
        logger.Info("SYSTEM", "🧪 SWITCHING TO TESTNET ENVIRONMENT (ChainID: 9999)")
    } else {
        os.Setenv("BVM_CHAIN_ID", "1989")
        os.Setenv("BVM_NETWORK_NAME", "BVM Mainnet")
        logger.Info("SYSTEM", "🌐 RUNNING ON MAINNET ENVIRONMENT (ChainID: 1989)")
    }

    logger.Info("SYSTEM", fmt.Sprintf("🏗️  Inisialisasi BVM Node di %s...", h))

    // 3. DATABASE: Gunakan path yang sudah terisolasi (via main.go)
        // 1. BUAT SATU INSTANS DATABASE UTAMA
        dbPath := fmt.Sprintf("%s/blockchain_db", h)
        os.MkdirAll(dbPath, 0755)
        store, err := storage.NewLevelDBStore(dbPath, 8)
        if err != nil {
                logger.Error("SYSTEM", "🚨 Gagal membuka database: ", err)
                panic(err)
        }

        // 🚩 2. SINKRONISASI SATU KALI JALAN (BOOTSTRAP)
        // Proses ini memblokir startup sampai sinkronisasi selesai, lalu BERHENTI.
        StartNodeWithSync(nexusURL, store)

        // 3. SETUP APP & KERNEL
        bc := types.NewBlockchain()
        bvmApp := app.NewApp(store, bc)
        // 🚩 TAMBAHKAN INI SEBAGAI BACKGROUND MONITOR
        go RunSyncEngine(nexusURL, store)
        bvmApp.Start()


        // 4. 👷 MOBILISASI MINER INTERNAL (ANTI-BENTROK DB / MURNI INTER-MEMORY)
        if useMiner {
                go func() {
                        // Beri jeda 5 detik agar pangkalan node P2P & HTTP siap sepenuhnya
                        time.Sleep(5 * time.Second)
                        logger.Success("MINER", "🏗️  Membangunkan Miner Internal Sultan...")

                        minerAddr, err := getMinerAddress(h)
                        if err != nil {
                                logger.Error("MINER", "🚨 KRITIKAL: Miner gagal aktif karena: ", err)
                                logger.Error("MINER", fmt.Sprintf("Silakan pastikan node_wallet.json tersedia di %s", h))
                                return
                        }

                        logger.Info("MINER", "👷 Alamat Miner Aktif: "+minerAddr)

                        // =========================================================================
                        // 🔒 KAITAN PENGUNCI GOROUTINE: Paksa thread Miner tunduk pada Jaringan aktif!
                        // =========================================================================
                        if isTestnet {
                                os.Setenv("BVM_CHAIN_ID", "9999")
                        } else {
                                os.Setenv("BVM_CHAIN_ID", "1989")
                        }
                        // =========================================================================

                        // Jalankan mesin engine penambang murni menggunakan Keeper lokal yang sama!
                        engine := miner.NewMinerEngine(bvmApp.BVMKeeper)
                        engine.Start(minerAddr)
                }()
        }


    // 7. START NODE (DENGAN PORT DINAMIS)
    apiPort := 8080
    p2pPort := 9090
    nodeName := "BVM-Mainnet-Node-01"
    
    if isTestnet {
        apiPort = 8081
        p2pPort = 9091
        nodeName = "BVM-Testnet-Node-01"
    }

    node.StartFullNode(bvmApp.BVMKeeper, bvmApp.Mempool, bvmApp.P2P, store, nodeName, apiPort, p2pPort)
}

func StartNodeWithSync(nexusAddr string, store storage.BVMStore) {
    clientNexus := client.NewBVMClient(nexusAddr)
    
    // Tanya Nexus tinggi blok terakhir
    nexusInfo, err := clientNexus.GetNetworkInfo()
    if err != nil {
        fmt.Printf("⚠️ Nexus Offline di %s: %v. Mode Offline...\n", nexusAddr, err)
        return
    }

    localHeight := getLocalHeight(store)

    // Jika tertinggal, blokir proses start sampai sinkron selesai (Bootstrap)
    if uint64(nexusInfo.Height) > localHeight {
        fmt.Printf("🔄 [BOOTSTRAP] Memulai sinkronisasi awal dari #%d ke #%d...\n", localHeight, nexusInfo.Height)
        err := clientNexus.FastSync(localHeight, uint64(nexusInfo.Height), store)
        if err != nil {
            fmt.Printf("❌ Bootstrap Gagal: %v\n", err)
        } else {
            fmt.Println("✅ [BOOTSTRAP] Sinkronisasi Awal Selesai!")
        }
    }
}


// 1. Fungsi untuk cek tinggi blok lokal di database Core
func getLocalHeight(store storage.BVMStore) uint64 {
	var h uint64
	store.Get("m:height", &h)
	return h
}

// 2. Fungsi untuk mengambil info dari Nexus Sultan
func fetchInfoFromNexus(nexusURL string) (*types.NetworkResponse, error) {
	c := client.NewBVMClient(nexusURL)
	return c.GetNetworkInfo()
}

// 3. Fungsi eksekutor FastSync
func performFastSync(nexusURL string, start uint64, target uint64, store storage.BVMStore) error {
	c := client.NewBVMClient(nexusURL)
	return c.FastSync(start, target, store)
}
