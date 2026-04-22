//go:build wasm
package sdk

import (
    "unsafe"
)

// --- 1. HOST FUNCTIONS (Deklarasi Murni untuk Kernel) ---

//go:wasmimport env transfer_token
func host_transfer_token(fP, fS, tP, tS uint32, am uint64, sP, sS uint32) uint32

//go:wasmimport env get_caller
func host_get_caller(ptr, size uint32) uint32

//go:wasmimport env mint_token
func host_mint_token(aP, aS uint32, am uint64, sP, sS uint32) uint32

//go:wasmimport env emit_event
func host_emit_event(tP, tS, mP, mS uint32)

//go:wasmimport env update_stake
func host_update_stake(aP, aS uint32, am uint64, op uint32) uint32

//go:wasmimport env get_validator_power
func host_get_validator_power(aP, aS uint32) uint64

//go:wasmimport env register_nexus
func host_register_nexus(idP, idS, ownP, ownS, tokP, tokS uint32, stake uint64) uint32

//go:wasmimport env lock_for_bridge
func host_lock_for_bridge(fP, fS, tP, tS uint32, am uint64) uint32

// --- 2. WRAPPERS (Fungsi yang dipanggil oleh Kontrak Sultan) ---

func Transfer(from, to string, amount uint64, symbol string) bool {
    fP, fS := stringToPtr(from)
    tP, tS := stringToPtr(to)
    sP, sS := stringToPtr(symbol)
    return host_transfer_token(fP, fS, tP, tS, amount, sP, sS) == 1
}

func GetCaller() string {
    buf := make([]byte, 64)
    ptr, size := uint32(uintptr(unsafe.Pointer(&buf[0]))), uint32(len(buf))
    n := host_get_caller(ptr, size)
    return string(buf[:n])
}

func Mint(target string, amount uint64, symbol string) bool {
    aP, aS := stringToPtr(target)
    sP, sS := stringToPtr(symbol)
    return host_mint_token(aP, aS, amount, sP, sS) == 1
}

func Emit(tag, message string) {
    tP, tS := stringToPtr(tag)
    mP, mS := stringToPtr(message)
    host_emit_event(tP, tS, mP, mS)
}

func UpdateStake(address string, amount uint64, isAdding bool) bool {
    aP, aS := stringToPtr(address)
    var op uint32 = 0
    if isAdding { op = 1 }
    return host_update_stake(aP, aS, amount, op) == 1
}

func RegisterNexus(id, owner, token string, stake uint64) bool {
    iP, iS := stringToPtr(id)
    oP, oS := stringToPtr(owner)
    tP, tS := stringToPtr(token)
    return host_register_nexus(iP, iS, oP, oS, tP, tS, stake) == 1
}

func LockForBridge(from, to string, amount uint64) bool {
    fP, fS := stringToPtr(from)
    tP, tS := stringToPtr(to)
    return host_lock_for_bridge(fP, fS, tP, tS, amount) == 1
}

// --- 1. HOST FUNCTIONS (Deklarasi) ---
//go:wasmimport env get_method
func host_get_method(ptr, size uint32) uint32

//go:wasmimport env get_arg_string
func host_get_arg_string(kP, kS, vP, vS uint32) uint32

//go:wasmimport env get_arg_uint64
func host_get_arg_uint64(kP, kS uint32) uint64

// --- 2. WRAPPERS (Tambahkan di bawah Wrapper lainnya) ---

func GetMethod() string {
    buf := make([]byte, 32)
    ptr, size := uint32(uintptr(unsafe.Pointer(&buf[0]))), uint32(len(buf))
    n := host_get_method(ptr, size)
    return string(buf[:n])
}

func GetArgString(key string) string {
    kP, kS := stringToPtr(key)
    buf := make([]byte, 128) // Buffer untuk nilai argumen
    vP, vS := uint32(uintptr(unsafe.Pointer(&buf[0]))), uint32(len(buf))
    n := host_get_arg_string(kP, kS, vP, vS)
    return string(buf[:n])
}

func GetArgUint64(key string) uint64 {
    kP, kS := stringToPtr(key)
    return host_get_arg_uint64(kP, kS)
}

// --- 3. HELPERS ---

func stringToPtr(s string) (uint32, uint32) {
    if s == "" { return 0, 0 }
    return uint32(uintptr(unsafe.Pointer(unsafe.StringData(s)))), uint32(len(s))
}

func PtrToString(ptr uint32, size uint32) string {
    return unsafe.String((*byte)(unsafe.Pointer(uintptr(ptr))), size)
}
