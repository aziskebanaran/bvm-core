package miner

import (
	 "bvm.core/x/bvm/types"
    "bvm.core/pkg/client"
    "bvm.core/pkg/logger"
    "fmt"
    "time"
	"strings"
)

func StartMining(address string, kernelURL string) {
    bvmClient := client.NewBVMClient(kernelURL)
    const minerName = "BVM-EXTERNAL-PRO"
    
    // 🚩 PERBAIKAN 1: Inisialisasi lastMinedHeight dari status jaringan terbaru
    status, _ := bvmClient.GetNodeStatus(address)
    lastMinedHeight := int64(status.Height)

    for {
        work, err := bvmClient.GetWork(address, minerName)
        if err != nil {
            logger.Error("MINER", "Gagal mengambil kerjaan. Coba lagi dalam 5s...")
            time.Sleep(5 * time.Second)
            continue
        }

        block, ok := work.(types.Block)
        // 🚩 PERBAIKAN 2: Proteksi jika Kernel mengirim blok kosong atau blok lama
	if !ok || int64(block.Index) <= lastMinedHeight {

            time.Sleep(1 * time.Second)
            continue
        }

        logger.Info("MINER", fmt.Sprintf("🚀 Menambang Blok #%d | Diff: %d", block.Index, block.Difficulty))

        found := false
        target := strings.Repeat("0", int(block.Difficulty))

        // PoW Loop
        for i := 0; ; i++ {
            // 🚩 PERBAIKAN 3: Cek status lebih jarang agar tidak membebani network
            if i % 5000000 == 0 { 
                currentStatus, _ := bvmClient.GetNodeStatus(address)
                if int64(currentStatus.Height) >= int64(block.Index) {
                    logger.Info("MINER", fmt.Sprintf("⏩ Blok #%d sudah ditemukan. Skip!", block.Index))
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
                logger.Success("MINER", fmt.Sprintf("✅ Blok #%d TERSEGEL!", block.Index))
                lastMinedHeight = int64(block.Index)
            }
        }
    }
}
