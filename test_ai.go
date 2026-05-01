package main

import (
    "fmt"
    "github.com/aziskebanaran/bvm-core/pkg/ai"
    "github.com/joho/godotenv"
)

func main() {
    godotenv.Load()
    client := ai.NewGeminiClient()
    res, err := client.GenerateText("Katakan halo dalam gaya komandan robot!")
    if err != nil {
        fmt.Println("Gagal:", err)
    } else {
        fmt.Println("Hasil Gemini:", res)
    }
}
