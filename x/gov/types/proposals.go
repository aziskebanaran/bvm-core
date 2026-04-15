package types

// Daftar Status Proposal
const (
    StatusPending  = "Pending"
    StatusApproved = "Approved"
    StatusRejected = "Rejected"
)

type SoftwareUpgradeProposal struct {
    Title            string
    Description      string
    FeatureName      string // Target Fitur: "WASM_ENGINE"
    ActivateAtHeight int64  // "Bom Waktu" Aktivasi
    Status           string
    Proposer         string // Alamat yang mengajukan
}
