package keeper

import ()



func (k *Keeper) DistributeBlockReward(height int64, fees uint64) (uint64, uint64, error) {
    p := k.GetParamsData()

    // Ambil jumlah validator aktif dari state (atau dari list validator Sultan)
    // Asumsi k.GetAllValidators(ctx) atau sejenisnya tersedia
    activeValidators := k.GetValidatorCount() 

    // 1. HITUNG SUBSIDI (Kirim jumlah validator ke sini)
    subsidi := k.GetSubsidiAtHeight(height, activeValidators)

    // 2. HITUNG PEMBAGIAN FEE
    tip, burnFromFee := p.DistributeFee(fees)

    // 3. TOTAL HADIAH
    minerTotal := subsidi + tip

    return minerTotal, burnFromFee, nil
}

// GetSubsidiAtHeight: Sekarang membagi subsidi murni dengan jumlah validator
func (k *Keeper) GetSubsidiAtHeight(height int64, validatorCount int) uint64 {
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

    // 2. 🚩 PEMBAGIAN: Bagi total subsidi dengan jumlah validator aktif
    if validatorCount <= 1 {
        return totalBlockSubsidi
    }

    return totalBlockSubsidi / uint64(validatorCount)
}
