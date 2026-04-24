package keeper

import (
)


func (k *Keeper) GetCurrentReward(height int64) uint64 {
        // 🚩 PERUBAHAN: Tambahkan jumlah validator agar reward yang tampil real-time
        vCount := k.GetValidatorCount()
        return k.GetSubsidiAtHeight(height, vCount)
}

func (k *Keeper) GetInflationStats() (string, int64) {
        // 🚩 Ambil height sebagai int64
        height := int64(k.GetLastHeight()) 
        params := k.GetParamsData()

        // 🚩 Sekarang height + 1 sudah valid sebagai int64
        nextReward := k.GetCurrentReward(height + 1)

        if params.HalvingInterval <= 0 {
                return params.FormatDisplay(nextReward), 0
        }

        // 🚩 Perhitungan sisa blok menuju Halving
        halvingIn := int64(params.HalvingInterval) - (height % int64(params.HalvingInterval))

        return params.FormatDisplay(nextReward), halvingIn
}
