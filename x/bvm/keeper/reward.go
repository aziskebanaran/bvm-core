package keeper

import (
	"time"
	"fmt"
	"github.com/aziskebanaran/bvm-core/pkg/logger"
	"github.com/aziskebanaran/bvm-core/pkg/storage"
	"github.com/aziskebanaran/bvm-core/x/events" // 🧾 Panggil modul kuitansi digital Jenderal
)

// 🎯 UPDATE SAKRAL: Tambahkan parameter minerAddress string di pintu depan fungsi
func (k *Keeper) DistributeBlockReward(height int64, minerAddress string, fees uint64, batch storage.Batch, pendingChanges map[string]int64) (uint64, uint64, error) {
        p := k.GetParamsData()
        activeValidators := k.GetValidatorCount()

        // 1. HITUNG SUBSIDI STANDAR
        subsidi := k.GetSubsidiAtHeight(height, activeValidators)

        // 2. HITUNG PEMBAGIAN FEE
        tip, burnFromFee := p.DistributeFee(fees)

        // =========================================================================
        // 🎯 LOGIKA MULTI-VAULT POOL (Mengambil daftar dari StorageKeeper)
        // =========================================================================
        blocksLeft := int64(p.HalvingInterval) - (height % int64(p.HalvingInterval))
        var totalBonusBVM uint64 = 0

        if blocksLeft > 0 {
                // Panggil daftar dari storage (Asumsi StorageKeeper sudah ter-inject di Keeper)
                // Jika belum ada fungsi GetRegisteredVaults, gunakan k.Storage.GetVaultList()
                vaults := k.GetRegisteredVaults() 

                for _, addr := range vaults {
                        balance := k.GetBalanceBVM(addr)
                        if balance > 0 {
                                // Potong proporsional berdasarkan jumlah vault yang aktif
                                bonus := balance / uint64(blocksLeft) / uint64(len(vaults))
                                if batch != nil && bonus > 0 {
                                        err := k.SubBalanceBVM(addr, bonus, batch)
                                        if err == nil {
                                                totalBonusBVM += bonus
                                        }
                                }
                        }
                }
        }

        subsidi += totalBonusBVM
        totalSubsidyBudget := subsidi
        // =========================================================================

        currentMiner := minerAddress
        k.EmitBlockRewardEvents(uint64(height), currentMiner, totalSubsidyBudget, tip, batch, pendingChanges)

        return totalSubsidyBudget + tip, burnFromFee, nil
}

// GetSubsidiAtHeight: Sekarang membagi subsidi murni HANYA dengan jumlah validator yang BENAR-BENAR AKTIF NYALA
func (k *Keeper) GetSubsidiAtHeight(height int64, activeValidatorCount int) uint64 {
	params := k.GetParamsData()

	if params.HalvingInterval <= 0 {
		return params.InitialReward
	}

	numHalvings := height / int64(params.HalvingInterval)
	if numHalvings >= 64 {
		return 0
	}

	// 1. Hitung total subsidi blok (100%)
	totalBlockSubsidi := params.InitialReward >> uint64(numHalvings)

	// 2. 🛡️ AMANKAN PENGALI: Jika yang nyala hanya 1 atau 2, bagi berdasarkan yang nyala itu saja!
	if activeValidatorCount <= 1 {
		return totalBlockSubsidi
	}

	return totalBlockSubsidi / uint64(activeValidatorCount)
}

// =========================================================================
// 🚀 PERAKITAN BARU: LOGIKA EKONOMI & KUITANSI RELAYER MEMPOOL (SULTAN MODULAR)
// =========================================================================

// DistributeRelayerReward: Memotong jatah Mempool, menyuntikkan balance, dan mencetak kuitansi resmi
func (k *Keeper) DistributeRelayerReward(relayer string, txID string, fee uint64, height uint64, pendingChanges map[string]int64, batch storage.Batch) uint64 {
	if relayer == "" {
		return 0
	}

	// Jatah upah Mempool Node: 10% dari total Gas Fee transaksi
	relayerCut := fee * 10 / 100
	if relayerCut == 0 {
		return 0
	}

	// Salurkan jatah ke map in-memory perubahan saldo
	pendingChanges[relayer] += int64(relayerCut)
	logger.Success("FEE-ROUTER", fmt.Sprintf("⛽ Upah Relayer 10%% (%d) dialirkan ke Mempool: %s", relayerCut, relayer[:10]))

	// 🧾 Pahat Kuitansi Sistem untuk Audit Transparansi Nexus
	// 🧾 GANTI MENJADI 'events.EmitBatch' JENDERAL!
	events.EmitBatch(batch, "DISTRIBUTE_RELAYER_FEE", map[string]interface{}{
		"height":       height,
		"tx_id":        txID,
		"relayer_node": relayer,
		"amount":       relayerCut,
		"symbol":       "BVM",
		"memo":         "Insentif penyaringan transaksi Mempool Standalone",
	})

	return relayerCut
}

// EmitBlockRewardEvents: Membagi rata anggaran subsidi murni ke semua node aktif, dan memberikan tip fee penuh ke miner utama
func (k *Keeper) EmitBlockRewardEvents(height uint64, miner string, totalSubsidyBudget uint64, tip uint64, batch storage.Batch, pendingChanges map[string]int64) {
        if k.Staking == nil { return }

        allValidators := k.Staking.GetValidators()

        // 1. INVENTARISASI NODE HIDUP (DENGAN SENSOR WAKTU KEDALUWARSA DISK)
        liveNodesMap := make(map[string]bool)
        if miner != "" { liveNodesMap[miner] = true }

        now := time.Now().Unix() // Detak jam runtime detik ini

        for _, v := range allValidators {
                var heartbeatTimestamp int64
                err := k.Store.Get("mempool_active:"+v.Address, &heartbeatTimestamp)

                // 🎯 SENSOR DISK ANTI HANTU: Lindungi dari validator eksternal yang sudah mati
                if v.IsActive && err == nil && heartbeatTimestamp > 0 && (now - heartbeatTimestamp < 180) {
                        liveNodesMap[v.Address] = true
                }
        }

        activeMiners := k.GetActiveMiners()
        for addr, lastPing := range activeMiners {
                if now - lastPing < 180 {
                        liveNodesMap[addr] = true
                } else {
                        k.DeleteActiveMiner(addr)
                }
        }

        liveValidators := make([]string, 0)
        for addr := range liveNodesMap {
                liveValidators = append(liveValidators, addr)
        }
        liveCount := len(liveValidators)
        if liveCount == 0 { liveCount = 1 }

        // =========================================================================
        // 🧮 DISTRIBUSI BAGI RATA ADIL SULTAN (KODE ASLI JENDERAL AMAN)
        // =========================================================================
        subsidyPerNode := totalSubsidyBudget / uint64(liveCount)
        remainder := totalSubsidyBudget % uint64(liveCount)

        for _, valAddr := range liveValidators {
                if valAddr == miner {
                        actualMinerReward := subsidyPerNode + tip + remainder
                        pendingChanges[valAddr] += int64(actualMinerReward)

                        events.EmitBatch(batch, "BLOCK_REWARD_MINER", map[string]interface{}{
                                "height":      height,
                                "beneficiary": valAddr,
                                "amount":      actualMinerReward,
                                "symbol":      "BVM",
                                "live_nodes":  liveCount,
                                "type":        "MINER_SUBSIDY_AND_GAS_TIP",
                        })
                } else {
                        pendingChanges[valAddr] += int64(subsidyPerNode)

                        events.EmitBatch(batch, "BLOCK_REWARD_VALIDATOR_SHARE", map[string]interface{}{
                                "height":      height,
                                "beneficiary": valAddr,
                                "amount":      subsidyPerNode,
                                "symbol":      "BVM",
                                "live_nodes":  liveCount,
                                "type":        "PASSIVE_VALIDATOR_SUBSIDY",
                        })
                }
        }
}
