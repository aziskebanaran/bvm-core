package types

const (
    // ModuleName mendefinisikan nama modul untuk logging dan prefix
    ModuleName = "gov"

    // UpgradeKeyPrefix adalah prefix database untuk menyimpan tinggi blok aktivasi fitur.
    // Hasil akhirnya di DB akan terlihat seperti "gov:upgrade:WASM_ENGINE"
    UpgradeKeyPrefix = "gov:upgrade:"

    // ProposalKeyPrefix digunakan untuk menyimpan detail data proposal itu sendiri
    ProposalKeyPrefix = "gov:prop:"
)

// GetUpgradeKey menghasilkan key lengkap untuk fitur tertentu
func GetUpgradeKey(feature string) string {
    return UpgradeKeyPrefix + feature
}
