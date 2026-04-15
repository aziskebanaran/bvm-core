package logger

import (
	"fmt"
	"time"
)

// Warna Terminal (ANSI)
const (
	Reset  = "\033[0m"
	Green  = "\033[32m"
	Blue   = "\033[34m"
	Yellow = "\033[33m"
	Red    = "\033[31m"
	Cyan   = "\033[36m"
)

// helper: Menggabungkan semua input menjadi satu string rapi
func format(args ...interface{}) string {
	return fmt.Sprint(args...)
}

// Info: Pesan umum (Biru)
func Info(module string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	msg := format(args...)
	fmt.Printf("%s[%s]%s [%s%s%s] %s\n", Blue, timestamp, Reset, Cyan, module, Reset, msg)
}

// Success: Berhasil (Hijau)
func Success(module string, args ...interface{}) {
	msg := format(args...)
	fmt.Printf("✅ [%s%s%s] %s%s%s\n", Green, module, Reset, Green, msg, Reset)
}

// Error: Gagal (Merah)
func Error(module string, args ...interface{}) {
	msg := format(args...)
	fmt.Printf("❌ [%s%s%s] %s%s%s\n", Red, module, Reset, Red, msg, Reset)
}

// Warning: Peringatan (Kuning)
func Warning(module string, args ...interface{}) {
	msg := format(args...)
	fmt.Printf("⚠️ [%s%s%s] %s%s%s\n", Yellow, module, Reset, Yellow, msg, Reset)
}

// MiningProgress: Titik-titik progres
func MiningProgress() {
	fmt.Printf("%s.%s", Yellow, Reset)
}
