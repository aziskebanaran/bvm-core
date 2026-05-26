package types

const (
    ModuleName = "utxo"
    // Prefix utama untuk menyimpan kepingan koin di LevelDB
    UTXOPrefix = "u:" 
)

// KeyUTXO: Menghasilkan kunci unik untuk setiap kepingan koin
func KeyUTXO(id string) string {
    return UTXOPrefix + id
}
