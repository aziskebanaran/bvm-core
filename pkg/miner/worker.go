package miner

import (
	"github.com/aziskebanaran/bvm-core/pkg/client"
	"github.com/aziskebanaran/bvm-core/pkg/logger"
	"github.com/aziskebanaran/bvm-core/x/bvm/types"
	"fmt"
	"strings"
	"time"
)

func StartMining(address string, kernelURL string) {
	bvmClient := client.NewBVMClient(kernelURL)
	const minerName = "BVM-EXTERNAL-PRO"

       // --- 🛡️ PROTEKSI SULTAN ---
        status, err := bvmClient.GetNodeStatus(address)
        if err != nil || status == nil {
        logger.Error("MINER", "⚠️ Gagal koneksi awal ke Nexus/Core. Pastikan Nexus AKTIF di port 9092.")
           return // Keluar dari fungsi agar loop di main.go yang menangani retry
        }

	// Ambil status awal untuk sinkronisasi Height
	lastMinedHeight := int64(status.Height)

	for {
		// --- 🚩 SENSOR DETAK JANTUNG (HEARTBEAT) ---
		// Periksa Mempool sebelum meminta pekerjaan (GetWork)
		txs, err := bvmClient.GetMempoolTxs()
		if err != nil {
			logger.Error("MINER", "Gagal cek Mempool. Kernel mungkin offline. Re-sync 5s...")
			time.Sleep(5 * time.Second)
			continue
		}

		// Jika tidak ada transaksi, jangan "ngegas". 
		// Istirahat 5 detik untuk menghemat CPU dan membersihkan log Kernel.
		if len(txs) == 0 {
			time.Sleep(5 * time.Second)
			continue
		}

		// --- 🚩 AMBIL PEKERJAAN ---
		work, err := bvmClient.GetWork(address, minerName)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		block, ok := work.(types.Block)
		if !ok || int64(block.Index) <= lastMinedHeight {
			// Jika blok sudah ketinggalan atau sudah ada yang garap, tunggu sejenak.
			time.Sleep(2 * time.Second)
			continue
		}

		logger.Info("MINER", fmt.Sprintf("🔥 Transaksi Terdeteksi! Menambang Blok #%d | Diff: %d", block.Index, block.Difficulty))

		found := false
		target := strings.Repeat("0", int(block.Difficulty))

		// PoW Loop
		for i := 0; ; i++ {
			// Cek interupsi setiap 2.000.000 hash agar tidak terlalu sering menembak network
			if i%2000000 == 0 {
				currentStatus, _ := bvmClient.GetNodeStatus(address)
				if int64(currentStatus.Height) >= int64(block.Index) {
					logger.Info("MINER", fmt.Sprintf("⏩ Blok #%d sudah disegel miner lain. Abort!", block.Index))
					break
				}
			}

			block.Nonce = int32(i)
			hash := block.CalculateBlockHash()

			if strings.HasPrefix(hash, target) {
				block.Hash = hash
				found = true
				break
			}
		}

		if found {
			err := bvmClient.SubmitBlock(block)
			if err == nil {
				logger.Success("MINER", fmt.Sprintf("🧱 BERHASIL! Blok #%d telah dipahat ke Blockchain!", block.Index))
				lastMinedHeight = int64(block.Index)
			}
		}
	}
}
