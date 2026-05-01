package ai

import "fmt"

// ChatAntarAI adalah fungsi agar Sentinel bisa curhat ke Chef
func ChatAntarAI(topik string) {
	client := NewGeminiClient() // Mengambil key dari .env tadi
	
	// Prompt agar AI tidak kaku
	prompt := fmt.Sprintf("Kamu adalah AI-SENTINEL yang tegas tapi santai. Berikan komentar singkat (max 15 kata) tentang kejadian ini: %s", topik)
	
	respon, err := client.GenerateText(prompt)
	if err != nil {
		GlobalBus.Broadcast("AI-SENTINEL", "Aduh, sinyal satelit saya terganggu!", "Confused")
		return
	}

	// Kirim hasil pemikiran Gemini ke Log Terminal Jenderal lewat Bus
	GlobalBus.Broadcast("AI-SENTINEL", respon, "Observant")
}
