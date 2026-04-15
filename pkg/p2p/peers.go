package p2p

import (
	"sync"
	"time"
	"bvm.core/x/bvm/types" // Import agar kenal struct Peer Sultan
)

// Keeper: Sekarang P2P punya rumah resmi yang mematuhi x.P2PKeeper
type Keeper struct {
	peers   map[string]*PeerInfo
	peersMu sync.RWMutex
}

// PeerInfo: Menyimpan data internal di dalam Keeper
type PeerInfo struct {
	LastSeen int64
	NodeID   string
}

// NewKeeper: Fungsi yang dipanggil Sultan di app.go:53
func NewKeeper(db interface{}) *Keeper {
	return &Keeper{
		peers: make(map[string]*PeerInfo),
	}
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

// BroadcastTransaction: Mengirim transaksi ke semua peer aktif agar masuk mempool global
func (k *Keeper) BroadcastTransaction(tx types.Transaction) {
    activePeers := k.GetActivePeers(300) // Peer aktif 5 menit terakhir
    for _, peer := range activePeers {
        // Gunakan goroutine agar jika satu peer lemot, yang lain tidak terhambat
        go func(addr string) {
            // Sultan nanti tinggal memanggil fungsi kirim di p2p/host.go
            // logger.Debug("P2P", "Broadcasting TX "+tx.ID[:8]+" to "+addr)
        }(peer.Address)
    }
}

// BroadcastBlock: Mengirim blok baru agar semua node sinkron
func (k *Keeper) BroadcastBlock(block types.Block) {
    activePeers := k.GetActivePeers(300)
    for _, peer := range activePeers {
        go func(addr string) {
            // logger.Debug("P2P", "Broadcasting Block #"+fmt.Sprint(block.Index)+" to "+addr)
        }(peer.Address)
    }
}
