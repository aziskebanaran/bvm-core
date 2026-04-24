package keeper

import (
    "time"
    "github.com/aziskebanaran/bvm-core/x"
    "github.com/aziskebanaran/bvm-core/x/bvm/types"
)

// GetParams: Memberikan akses ke Menteri Params (Interface)
func (k *Keeper) GetParams() x.ParamsKeeper {
    return k
}

// GetParamsData: Mengambil data struct Params (Atomic)
func (k *Keeper) GetParamsData() types.Params {
    if k.Blockchain == nil {
        return types.DefaultParams()
    }
    return k.Blockchain.Params
}

// GetDynamicFee: Mengikuti rumus dinamis di Params
func (k *Keeper) GetDynamicFee(mempoolSize int) uint64 {
    return k.GetParamsData().GetDynamicFee(mempoolSize)
}

func (k *Keeper) GetNextDifficultyForMiner(minerAddr string) int {
	params := k.GetParamsData()
	latest := k.GetLatestBlock()
	currentHeight := int64(k.GetLastHeight())

	// 1. Tentukan Base Difficulty (Default dari blok sebelumnya)
	baseDiff := latest.Difficulty
	if currentHeight == 0 {
		return params.MinDifficulty
	}

	actualDelay := time.Now().Unix() - latest.Timestamp 
	minDelay := params.GetMinBlockDelay()

	// 🚩 KOREKSI STRUKTUR IF-ELSE: } else if harus satu baris!
	if actualDelay < minDelay {
		// Jika terlalu ngebut, persulit secara instan!
		baseDiff = latest.Difficulty + 1
	} else if currentHeight%int64(params.AdjustmentBlock) == 0 {
		// --- 🚩 LOGIKA RETARGETING PERIODIK (Setiap 20 Blok) ---
		var prevRef types.Block
		err := k.Store.Get(k.keyBlock(currentHeight-int64(params.AdjustmentBlock)+1), &prevRef)

		if err == nil {
			actualWindowTime := latest.Timestamp - prevRef.Timestamp
			targetWindowTime := int64(params.TargetBlockTime) * int64(params.AdjustmentBlock)

			if actualWindowTime < targetWindowTime/2 {
				baseDiff = latest.Difficulty + 1
			} else if actualWindowTime > targetWindowTime*2 {
				if latest.Difficulty > int32(params.MinDifficulty) {
					baseDiff = latest.Difficulty - 1
				}
			}
		}
	}

	// 2. STAKING BOOST (Hadiah untuk Sultan yang Stake)
	if minerAddr != "" && k.Staking != nil {
		stakedAmount := k.Staking.GetValidatorStake(minerAddr)
		if stakedAmount >= params.MinStakeAmount {
			boost := int(stakedAmount / params.MinStakeAmount)
			baseDiff -= int32(boost)
		}
	}

	// 🚩 TAMBAHKAN: SKALA SERVER (FIREBASE ANALOGY)
	activeNodes := k.GetP2P().CountActive()
	serverMultiplier := activeNodes / 10 // Setiap 10 server, tambah 1 difficulty
	baseDiff += int32(serverMultiplier)

	// 3. Batas Bawah (Garis Merah)
	minAllowed := int32(params.MinDifficulty)
	if baseDiff < minAllowed {
	    baseDiff = minAllowed
	}

	return int(baseDiff)

}

// CalculateAvgBlockTime: Menggunakan sample dari AdjustmentBlock
func (k *Keeper) CalculateAvgBlockTime() int64 {
    params := k.GetParamsData()
    chain := k.Blockchain.Chain
    if len(chain) < 2 { return 0 }

    sampleSize := params.AdjustmentBlock
    if len(chain) < sampleSize { sampleSize = len(chain) }

    var totalTime int64
    count := 0
    for i := len(chain) - 1; i > len(chain)-sampleSize+1; i-- {
        diff := chain[i].Timestamp - chain[i-1].Timestamp
        totalTime += diff
        count++
    }

    if count == 0 { return 0 }
    return totalTime / int64(count)
}
