package types

// UserProfile: Struktur KTP Digital User BVM
type UserProfile struct {
    Username  string `json:"username"`
    Address   string `json:"address"`
    CreatedAt int64  `json:"created_at"`
    Tier      string `json:"tier"`
}
