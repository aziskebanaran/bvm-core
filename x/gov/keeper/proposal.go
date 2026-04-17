package keeper

import (
    "fmt"
    "github.com/aziskebanaran/BVM.core/pkg/logger"
    "github.com/aziskebanaran/BVM.core/x/gov/types"
	"github.com/aziskebanaran/BVM.core/x"
	"github.com/aziskebanaran/BVM.core/pkg/storage"
)

type GovKeeper struct {
    Store      storage.BVMStore
    BVMKeeper  x.BVMKeeper // Agar Gov bisa tanya ke BVM soal Height
}

func NewGovKeeper(store storage.BVMStore, bvm x.BVMKeeper) GovKeeper {
    return GovKeeper{
        Store:     store,
        BVMKeeper: bvm,
    }
}



// SubmitProposal menerima pengajuan upgrade baru dari User/Sultan
func (gk *GovKeeper) SubmitProposal(proposer string, title string, feature string, targetHeight int64) error {
    // 1. Validasi Dasar: Target blok harus di masa depan
    currentHeight := gk.BVMKeeper.GetLastHeight()
    if targetHeight <= int64(currentHeight) {
        return fmt.Errorf("target blok (%d) harus lebih tinggi dari blok saat ini (%d)", targetHeight, currentHeight)
    }

    // 2. Rakit Struktur Proposal
    newProposal := types.SoftwareUpgradeProposal{
        Title:            title,
        Description:      fmt.Sprintf("Aktivasi fitur %s", feature),
        FeatureName:      feature,
        ActivateAtHeight: targetHeight,
        Status:           types.StatusPending,
        Proposer:         proposer,
    }

    // 3. Simpan ke database dengan prefix proposal
    // Key: gov:prop:[Judul_Proposal]
    key := types.ProposalKeyPrefix + title
    err := gk.Store.Put(key, newProposal)
    if err != nil {
        return fmt.Errorf("gagal menyimpan proposal: %v", err)
    }

    logger.Info("GOV", fmt.Sprintf("📩 Proposal Baru: '%s' untuk Fitur [%s] di Blok #%d", 
        title, feature, targetHeight))
    
    return nil
}


// ExecutePassedUpgrade dijalankan ketika proposal sudah melewati masa voting dan disetujui.
func (gk *GovKeeper) ExecutePassedUpgrade(proposal types.SoftwareUpgradeProposal, currentHeight int64) error {
    // 1. Validasi: Jangan sampai mengaktifkan fitur di blok yang sudah terlanjur lewat
    if proposal.ActivateAtHeight <= currentHeight {
        return fmt.Errorf("tinggi blok aktivasi (%d) harus lebih besar dari blok saat ini (%d)", 
            proposal.ActivateAtHeight, currentHeight)
    }

    // 2. Tentukan Key menggunakan standarisasi dari types
    key := types.GetUpgradeKey(proposal.FeatureName)
    
    // 3. Simpan tinggi blok aktivasi ke Store
    // Nilai ini yang akan dibaca oleh k.IsFeatureActive di BVM Keeper
    err := gk.Store.Put(key, proposal.ActivateAtHeight)
    if err != nil {
        return fmt.Errorf("gagal menyimpan status upgrade ke database: %v", err)
    }

    // 4. Update status proposal menjadi Approved/Sah
    proposal.Status = types.StatusApproved
    gk.Store.Put(types.ProposalKeyPrefix+proposal.Title, proposal)

    logger.Success("GOV", fmt.Sprintf("⚖️ PROPOSAL DISAHKAN: Fitur [%s] akan aktif otomatis pada Blok #%d", 
        proposal.FeatureName, proposal.ActivateAtHeight))
    
    return nil
}

