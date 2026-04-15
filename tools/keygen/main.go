package tools

import (
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
    "crypto/sha256"
    "crypto/x509"
    "encoding/hex"
    "fmt"
    "log"
)

func main() {
    privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    if err != nil {
        log.Fatal(err)
    }

    privBytes, _ := x509.MarshalECPrivateKey(privateKey)
    privHex := hex.EncodeToString(privBytes)

    pubBytes, _ := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
    pubHex := hex.EncodeToString(pubBytes)

    pubHash := sha256.Sum256(pubBytes)
    address := fmt.Sprintf("bvmf%x", pubHash[:12])

    fmt.Println("--------------------------------------------------")
    fmt.Println("🔐 BVM KEY GENERATOR (SULTAN EDITION)")
    fmt.Println("--------------------------------------------------")
    fmt.Printf("📍 ADDRESS     : %s\n", address)
    fmt.Printf("🔓 PUBLIC KEY  : %s\n", pubHex)
    fmt.Printf("🔑 PRIVATE KEY : %s\n", privHex)
    fmt.Println("--------------------------------------------------")
    fmt.Println("⚠️  SIMPAN PRIVATE KEY ANDA! JANGAN SAMPAI HILANG!")
    fmt.Println("--------------------------------------------------")
}
