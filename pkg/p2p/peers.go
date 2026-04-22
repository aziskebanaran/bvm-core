package p2p

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "sync"
    "time"

    "github.com/aziskebanaran/bvm-core/x/bvm/types"
)


// Keeper: Sekarang P2P punya rumah resmi yang mematuhi x.P2PKeeper
type Keeper struct {
	peers   map[string]*PeerInfo
	peersMu sync.RWMutex
}

type PeerInfo struct {
    LastSeen  int64
    NodeID    string
    P2PPort   int    // Tambahkan ini agar sinkron
    APIPort   int    // Tambahkan ini agar sinkron
    FailCount int    // Tambahkan ini untuk Self-Healing
}


// NewKeeper: Fungsi yang dipanggil Sultan di app.go:53
func NewKeeper(db interface{}) *Keeper {
        return &Keeper{
                peers: make(map[string]*PeerInfo),
        }
}

// Tambahkan fungsi ini di bawah struct Keeper Sultan
func (k *Keeper) savePeersToDisk() {
    const peerFile = "data_nexus/peers.json"
    peers := k.GetPeers()
    data, _ := json.MarshalIndent(peers, "", "    ")
    os.WriteFile(peerFile, data, 0644)
}

// --- Implementasi Interface x.P2PKeeper (Sesuai Konstitusi interfaces.go) ---

// AddPeer: Mencatat atau memperbarui teman di jaringan
func (k *Keeper) AddPeer(ip string, nodeID string) error {
	k.peersMu.Lock()
	defer k.peersMu.Unlock()

	k.peers[ip] = &PeerInfo{
		LastSeen: time.Now().Unix(),
		NodeID:   nodeID,
	}
	return nil
}

// GetPeers: Mengambil semua daftar teman dalam format types.Peer
func (k *Keeper) GetPeers() []types.Peer {
	k.peersMu.RLock()
	defer k.peersMu.RUnlock()

	var list []types.Peer
	for ip, info := range k.peers {
		list = append(list, types.Peer{
			Address:  ip,
			NodeID:   info.NodeID,
			LastSeen: info.LastSeen,
		})
	}
	return list
}

// GetActivePeers: Filter node yang aktif dalam kurun waktu tertentu
func (k *Keeper) GetActivePeers(timeout int64) []types.Peer {
	k.peersMu.RLock()
	defer k.peersMu.RUnlock()

	now := time.Now().Unix()
	var active []types.Peer
	for ip, info := range k.peers {
		if now-info.LastSeen < timeout {
			active = append(active, types.Peer{
				Address:  ip,
				NodeID:   info.NodeID,
				LastSeen: info.LastSeen,
			})
		}
	}
	return active
}

// CountActive: Menghitung berapa banyak node yang online (5 menit terakhir)
func (k *Keeper) CountActive() int {
	return len(k.GetActivePeers(300))
}

// RemovePeer: Mengusir node dari daftar (misal karena curang atau mati)
func (k *Keeper) RemovePeer(ip string) error {
	k.peersMu.Lock()
	defer k.peersMu.Unlock()

	delete(k.peers, ip)
	return nil
}

func (k *Keeper) BroadcastTransaction(tx types.Transaction) {
    activePeers := k.GetActivePeers(300)
    for _, peer := range activePeers {
        go func(p types.Peer) {
            // Kita asumsikan API port adalah port P2P - 1010 (atau port statis)
            // Atau Sultan bisa menyimpan API port di dalam struct PeerInfo
            targetURL := fmt.Sprintf("http://%s/api/mempool", p.Address)

            payload, _ := json.Marshal(tx)
            resp, err := http.Post(targetURL, "application/json", bytes.NewBuffer(payload))
            if err != nil {
                // Jika gagal, mungkin node tersebut offline
                return
            }
            defer resp.Body.Close()
        }(peer)
    }
}

func (k *Keeper) BroadcastBlock(block types.Block) {
    activePeers := k.GetActivePeers(300)
    for _, peer := range activePeers {
        go func(addr string) {
            // Kita coba kirim ke port API standar tetangga
            err := SendToPeer(addr, "/api/mine", block) 
            if err != nil {
                // fmt.Printf("⚠️ Gagal kirim blok ke %s: %v\n", addr, err)
            }
        }(peer.Address)
    }
}

func (k *Keeper) LoadPeersFromFile(filePath string) {
    data, err := os.ReadFile(filePath)
    if err != nil {
        return // File belum ada, tidak masalah
    }

    var savedPeers []map[string]string
    json.Unmarshal(data, &savedPeers)

    for _, p := range savedPeers {
        // Daftarkan kembali ke memori RAM Keeper
        k.AddPeer(p["address"], p["name"])
        fmt.Printf("📡 [P2P] Menghubungkan ulang ke teman: %s\n", p["address"])
    }
}

// MarkFailure: Menambah hitungan gagal. Jika sudah 3x, hapus peer.
func (k *Keeper) MarkFailure(ip string) {
    k.peersMu.Lock()
    defer k.peersMu.Unlock()

    if p, ok := k.peers[ip]; ok {
        p.FailCount++
        if p.FailCount >= 3 {
            delete(k.peers, ip)
            // Simpan perubahan ke peers.json agar permanen
            k.savePeersToDisk() 
        }
    }
}

// ResetSuccess: Jika berhasil konek lagi, nolkan hitungan gagalnya
func (k *Keeper) ResetSuccess(ip string) {
    k.peersMu.Lock()
    defer k.peersMu.Unlock()
    if p, ok := k.peers[ip]; ok {
        p.FailCount = 0
        p.LastSeen = time.Now().Unix()
    }
}
