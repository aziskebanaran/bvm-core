package sdk

import (
    "unsafe"
)

// --- HOST FUNCTIONS (Deklarasi Murni - Tanpa Isi/Body) ---

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

// --- WRAPPER (Tetap seperti biasa) ---

func Transfer(from, to string, amount uint64, symbol string) bool {
    fPtr, fSize := stringToPtr(from)
    tPtr, tSize := stringToPtr(to)
    sPtr, sSize := stringToPtr(symbol)
    return host_transfer_token(fPtr, fSize, tPtr, tSize, amount, sPtr, sSize) == 1
}

func GetCaller() string {
    buf := make([]byte, 64)
    ptr, size := uint32(uintptr(unsafe.Pointer(&buf[0]))), uint32(len(buf))
    n := host_get_caller(ptr, size)
    return string(buf[:n])
}

func Mint(target string, amount uint64, symbol string) bool {
    aPtr, aSize := stringToPtr(target)
    sPtr, sSize := stringToPtr(symbol)
    return host_mint_token(aPtr, aSize, amount, sPtr, sSize) == 1
}

func Emit(tag, message string) {
    tPtr, tSize := stringToPtr(tag)
    mPtr, mSize := stringToPtr(message)
    host_emit_event(tPtr, tSize, mPtr, mSize)
}

func UpdateStake(address string, amount uint64, isAdding bool) bool {
    aPtr, aSize := stringToPtr(address)
    var op uint32 = 0
    if isAdding { op = 1 }
    return host_update_stake(aPtr, aSize, amount, op) == 1
}

// --- HELPERS ---

func stringToPtr(s string) (uint32, uint32) {
    if s == "" { return 0, 0 }
    return uint32(uintptr(unsafe.Pointer(unsafe.StringData(s)))), uint32(len(s))
}

func PtrToString(ptr uint32, size uint32) string {
    return unsafe.String((*byte)(unsafe.Pointer(uintptr(ptr))), size)
}
