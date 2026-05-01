package ai

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

type GrokClient struct {
	APIKey string
}

func NewGrokClient(apiKey string) *GrokClient {
	return &GrokClient{APIKey: apiKey}
}

func (c *GrokClient) Chat(prompt string) (string, error) {
	url := "https://api.x.ai/v1/chat/completions"

	payload := map[string]interface{}{
		"model": "grok-beta", // Atau versi terbaru di 2026
		"messages": []interface{}{
			map[string]interface{}{"role": "user", "content": prompt},
		},
	}

	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	json.NewDecoder(resp.Body).Decode(&result)
	if len(result.Choices) > 0 {
		return result.Choices[0].Message.Content, nil
	}

	return "Grok sedang sibuk, Jenderal!", nil
}
