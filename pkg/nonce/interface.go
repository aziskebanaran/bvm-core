package nonce

// NonceKeeper adalah kontrak yang harus dipenuhi oleh NonceManager.
// Ini adalah "Janji" modul nonce kepada modul lain (seperti Auth/BVM).
type NonceKeeper interface {
    // GetNextNonce: Digunakan API/Wallet untuk tahu nomor antrean berikutnya
    GetNextNonce(address string) uint64

    // Increment: Digunakan Kernel saat transaksi berhasil masuk blok
    Increment(address string) error

    // HealthCheckNonce: Alat audit Sultan untuk cek RAM vs Database
    HealthCheckNonce(address string) (bool, uint64, uint64)

    // ManualOverride: Hak Veto Sultan untuk memperbaiki nonce macet
    ManualOverride(address string, newNonce uint64) error

    SetNonce(address string, newNonce uint64) error
}
