package main

import (
    "github.com/aziskebanaran/BVM.core/pkg/miner" 
    "github.com/aziskebanaran/BVM.core/pkg/wallet"
    "github.com/aziskebanaran/BVM.core/pkg/logger"
    "fmt"
	"time"
)

const CORE_URL = "http://127.0.0.1:8080"

func main() {
    fmt.Println("-------------------------------------------")
    fmt.Println("🚀 BVM EXTERNAL MINER PRO - GENERASI SULTAN")
    fmt.Println("-------------------------------------------")

    // 1. Load Identitas Sultan
    myWallet, err := wallet.LoadWallet("node_wallet.json")
    if err != nil {
        logger.Error("MINER", "❌ Gagal muat wallet node_wallet.json. Pastikan file ada!")
        return
    }

    logger.Info("MINER", fmt.Sprintf("👷 Alamat Penambang: %s", myWallet.Address))
    logger.Info("MINER", fmt.Sprintf("🔗 Menghubungkan ke Kernel: %s", CORE_URL))

    // 2. Jalankan Mesin
    // Tips: Tambahkan mekanisme recovery jika StartMining berhenti
    for {
        miner.StartMining(myWallet.Address, CORE_URL)

        // Jika sampai ke sini, berarti StartMining keluar (mungkin karena error network)
        logger.Error("MINER", "⚠️ Koneksi terputus. Mencoba rekoneksi dalam 10 detik...")
        time.Sleep(10 * time.Second)
    }
}
