//go:build !wasm
package sdk

import "fmt"

func Transfer(from, to string, amount uint64, symbol string) bool {
    fmt.Printf("📡 [SDK-STD] Transfer: %d %s dari %s ke %s\n", amount, symbol, from, to)
    return true
}

func GetCaller() string { return "std_caller" }

func Mint(target string, amount uint64, symbol string) bool {
    fmt.Printf("🪙 [SDK-STD] Mint: %d %s ke %s\n", amount, symbol, target)
    return true
}

func Emit(tag, message string) {
    fmt.Printf("📝 [EVENT] %s: %s\n", tag, message)
}

func UpdateStake(address string, amount uint64, isAdding bool) bool {
    fmt.Printf("🥩 [SDK-STD] Update Stake: %s (%v)\n", address, isAdding)
    return true
}

func RegisterNexus(id, owner, token string, stake uint64) bool {
    fmt.Printf("🌐 [SDK-STD] Register Nexus: %s\n", id)
    return true
}

func LockForBridge(from, to string, amount uint64) bool {
    fmt.Printf("🔗 [SDK-STD] Lock Bridge: %d\n", amount)
    return true
}

func PtrToString(ptr uint32, size uint32) string { return "std_string" }
