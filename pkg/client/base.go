package client

import (
	 "encoding/json"
	"net/http"
	"time"
	"fmt"
	"github.com/aziskebanaran/BVM.core/x/bvm/types"
)

type BVMClient struct {
	BaseURL string
	HTTP    *http.Client
}

func NewBVMClient(url string) *BVMClient {
	return &BVMClient{
		BaseURL: url,
		HTTP:    &http.Client{Timeout: 120 * time.Second},
	}
}

// GetNetworkInfo: Mengambil data gabungan (Params + Realtime Status)
func (c *BVMClient) GetNetworkInfo() (*types.NetworkResponse, error) {
    resp, err := c.HTTP.Get(c.BaseURL + "/api/params")
    if err != nil { 
        return nil, fmt.Errorf("🌐 Jaringan Offline: %v", err) 
    }
    defer resp.Body.Close()

    var info types.NetworkResponse
    
    // Bongkar JSON langsung ke Struct agar aman dan cepat
    if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
        return nil, fmt.Errorf("❌ Format data tidak dikenal: %v", err)
    }
    
    return &info, nil
}

// MempoolResponse: Struktur data untuk menangkap antrean dari API
type MempoolResponse struct {
    Count   int                 `json:"count"`
    Txs     []types.Transaction `json:"txs"`
    Status  string              `json:"status"`
    Message string              `json:"message"`
}

// GetMempool: Fungsi kurir untuk mengambil antrean transaksi di RAM
func (c *BVMClient) GetMempool() (*MempoolResponse, error) {
    // Tembak ke endpoint /api/mempool sesuai Router Sultan
    resp, err := c.HTTP.Get(c.BaseURL + "/api/mempool")
    if err != nil {
        return nil, fmt.Errorf("🌐 Node Offline: %v", err)
    }
    defer resp.Body.Close()

    var result MempoolResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("❌ Gagal bongkar data Mempool: %v", err)
    }

    return &result, nil
}
