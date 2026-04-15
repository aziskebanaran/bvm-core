package constants

import (
    "os"
    "fmt" 
)



const (
    ProjectName    = "BVM Network"
    ProjectVersion = "v0.1.5-alpha" // Update versi agar Sultan tahu ini mesin baru

    GenesisHash    = "BVM_GENESIS_001"

    // Identitas Aset
    CoinDenom     = "bvm"
    CoinSymbol    = "BVM"
    CoinDecimals  = 6

    GenesisAddress = "bvm1genesis_key"

    MinValidatorStake = 1000.0 

)

// --- FUNGSI DINAMIS PORT (PENTING!) ---
// GetRPCPort: Mengembalikan nomor port saja (tanpa titik dua di depan)
// Agar lebih mudah digabung: "http://localhost:" + GetRPCPort()
func GetRPCPort() string {
    port := os.Getenv("PORT_RPC")
    if port == "" {
        return "8080" // Cukup angkanya saja, Sultan
    }
    // Jika user terlanjur input ":8080", kita bersihkan titik duanya
    if port[0] == ':' {
        return port[1:]
    }
    return port
}

// GetP2PPort: Ambil dari ENV "PORT_P2P", jika kosong pakai 9000
func GetP2PPort() int {
    portStr := os.Getenv("PORT_P2P")
    if portStr == "" {
        return 9000 // Jalur Antar Node
    }
    
    // Konversi string ke int sederhana
    var p2pPort int
    _, err := fmt.Sscanf(portStr, "%d", &p2pPort)
    if err != nil {
        return 9000
    }
    return p2pPort
}
