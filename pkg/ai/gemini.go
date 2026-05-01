package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type GeminiClient struct {
	APIKey string
	Model  string
}


func NewGeminiClient() *GeminiClient {
	return &GeminiClient{
		APIKey: os.Getenv("GEMINI_API_KEY"),
		// 🚩 GUNAKAN NAMA INI (Sesuai curl Jenderal yang sukses)
		Model:  "gemini-flash-latest", 
	}
}


func (c *GeminiClient) GenerateText(prompt string) (string, error) {
	// 🚩 UBAH: v1 -> v1beta
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", c.Model, c.APIKey)

	// Struktur request sesuai standar Google
	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": prompt},
				},
			},
		},
	}

	jsonData, _ := json.Marshal(payload)
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Gemini Menolak (Status %d): %s", resp.StatusCode, string(body))
	}

	// Parsing hasil
	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if len(result.Candidates) > 0 && len(result.Candidates[0].Content.Parts) > 0 {
		return result.Candidates[0].Content.Parts[0].Text, nil
	}

	return "AI sedang melamun...", nil
}
