package ai

import (
    "fmt"
    "math/rand"
    "strings"
)

type Orchestrator struct {
    // Gemini Client tetap ada di struktur tapi kita abaikan sementara
    Gemini *GeminiClient
}

func NewOrchestrator() *Orchestrator {
    return &Orchestrator{
        Gemini: NewGeminiClient(),
    }
}

func (o *Orchestrator) ReportNewBlock(height int64, txCount int, activeWallets int) {
    // 1. Database komentar internal (Mode Offline)
    sentinelQuotes := []string{
        "Blok sah. Radar menunjukkan area bersih dari intrusi.",
        "Satu blok lagi terpahat. Pertahanan perimeter diperkuat.",
        "Siklus selesai. Data terenkripsi dengan sempurna, Jenderal.",
        "Laporan intelijen: Blok stabil. Tidak ada anomali terdeteksi.",
        "Mesin berderu pelan. Blok ini milik kita sekarang.",
    }

    // 2. Pilih komentar secara acak agar lebih hidup
    comment := sentinelQuotes[rand.Intn(len(sentinelQuotes))]

    // 3. Tambahkan info teknis ke komentar
    status := fmt.Sprintf("%s [#%d | TX: %d]", comment, height, txCount)

    // 4. Siarkan ke Log Jenderal
    GlobalBus.Broadcast("AI-SENTINEL", strings.TrimSpace(status), "Vigilant")

    // 5. Logika Tambahan AI-CHEF
    if txCount > 5 {
        GlobalBus.Broadcast("AI-CHEF", "Dapur rekaman transaksi mendidih, Jenderal!", "Busy")
    }
}
