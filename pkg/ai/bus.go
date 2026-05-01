package ai

import (
	"fmt"
	"sync"
)

// AIMessage adalah struktur pesan yang dikirim antar AI
type AIMessage struct {
	Sender  string
	Content string
	Mood    string
}

// AIBus adalah jalur distribusi pesan (Radio AI)
type AIBus struct {
	history []AIMessage
	mu      sync.Mutex
}

// GlobalBus adalah instance tunggal agar semua file di pkg/ai bisa mengaksesnya
var GlobalBus = &AIBus{}

// Broadcast untuk menyiarkan pesan ke terminal Jenderal
func (b *AIBus) Broadcast(sender, content, mood string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	msg := AIMessage{
		Sender:  sender,
		Content: content,
		Mood:    mood,
	}
	b.history = append(b.history, msg)

	// Tampilkan log yang bergaya dan tidak kaku
	fmt.Printf("[%s] %s: \"%s\" [%s]\n", 
		sender, 
		getEmoji(sender), 
		content, 
		mood,
	)
}

func getEmoji(name string) string {
	switch name {
	case "AI-SENTINEL":
		return "🛡️"
	case "AI-CHEF":
		return "👨‍🍳"
	default:
		return "🤖"
	}
}
